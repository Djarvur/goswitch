# CI-раннер e2e — self-hosted, живая GNOME-сессия (план 02-07, D-19)

Владелец осознанно выбрал CI-контур уже в Фазе 2 (D-19): зелёная матрица
на раннере — часть phase gate. Раннер — это ЭТА машина: e2e-стенду нужна
настоящая GNOME Wayland-сессия с фокусом окон, эфемерная ubuntu-latest
дать её не может. Workflow `.github/workflows/e2e-matrix.yml` запускается
вручную (`workflow_dispatch`) и — после мержа фазы в `main` — по ночному
`schedule` (план 04-09, раздел ниже), прогоняет `mise run e2e-matrix` —
ту же задачу, что и локальный запуск (D-09), — и выгружает `e2e-report.txt`
артефактом.

Инструкция воспроизводима без root: всё живёт в `$HOME` (каталог раннера,
user unit, окружение сессии).

## Требования к машине

Из STACK (research Фазы 0, проверено на цели живьём):

| Требование                                    | Проверка                                                       |
|-----------------------------------------------|----------------------------------------------------------------|
| Живая GNOME Wayland-сессия (не headless)      | `echo $WAYLAND_DISPLAY` → `wayland-0`                          |
| Пользователь в группе `input`                 | `id -nG | grep -w input`                                       |
| `/dev/uinput` доступен на запись              | `test -w /dev/uinput && echo ok`                               |
| ydotool                                       | `ydotool --version` (0.1.8 — открывает `/dev/uinput` напрямую) |
| python3-gi + AT-SPI                           | `/usr/bin/python3 -c "import gi; gi.require_version('Atspi','2.0')"` |
| gir1.2-atspi-2.0, at-spi2-core                | `dpkg -l gir1.2-atspi-2.0 at-spi2-core`                        |
| Поверхности матрицы                           | `zenity --version`; `google-chrome --version`; `gnome-text-editor --version` |
| gedit — поверхность матрицы v3 (план 04-05)   | `gedit --version` (GTK3-поколение IM/AT-SPI; в дефолтный GNOME НЕ входит — `sudo apt install gedit`, репозиторий universe) |
| gh CLI, аутентифицирован с правом admin на репозитории | `gh api repos/Djarvur/goswitch/actions/runners` → HTTP 200    |
| curl, tar                                     | `command -v curl tar`                                          |

## Установка раннера

Тарболл берём под `linux-x64` из релизов actions/runner (версию не
фиксируем в этой доке — актуальную смотрите на странице релизов; факт
установленной версии попадает в SUMMARY плана и видна в
`~/actions-runner/run.sh --version`):

```console
$ mkdir -p ~/actions-runner && cd ~/actions-runner
$ # URL последнего linux-x64-тарболла — со страницы
$ # https://github.com/actions/runner/releases/latest
$ curl -o actions-runner-linux-x64.tar.gz -L <URL-тарболла>
$ tar xzf actions-runner-linux-x64.tar.gz
```

Регистрационный токен одноразовый, живёт минуты. Минтится через gh CLI
(нужен admin на репозитории) или из UI: GitHub → Djarvur/goswitch →
Settings → Actions → Runners → New self-hosted runner. Токен НЕ
логируется и НЕ коммитится. Endpoint минта — POST (живая находка 02-07:
`gh api` без `--method POST` делает GET и получает HTTP 404):

```console
$ cd ~/actions-runner
$ TOKEN=$(gh api --method POST repos/Djarvur/goswitch/actions/runners/registration-token --jq .token)
$ ./config.sh --url https://github.com/Djarvur/goswitch --token "$TOKEN" \
    --labels gnome --unattended
$ unset TOKEN
```

Ярлык `gnome` — обязательный: workflow ищет раннера по
`runs-on: [self-hosted, gnome]`. Дефолтные ярлыки (self-hosted, Linux,
X64) config.sh ставит сам.

## systemd user unit (графическая сессия)

Раннер обязан жить ВНУТРИ сессии владельца — иначе у стенда нет ни
Wayland-композитора, ни session-bus, ни IBus. Прецедент — юнит демона
`dist/systemd/user/goswitchd.service` (тот же паттерн «после
graphical-session.target»). Файл
`~/.config/systemd/user/goswitch-ci-runner.service`:

```ini
[Unit]
Description=goswitch CI runner (e2e-matrix, live GNOME session)
PartOf=graphical-session.target
After=graphical-session.target

[Service]
ExecStart=%h/actions-runner/run.sh
# Некорректная смерть раннера (kill -9) — systemd поднимет его снова;
# незавершённый job GitHub отдаст другому/новому подключению.
Restart=on-failure

[Install]
WantedBy=graphical-session.target
```

### Окружение сессии

Юниту нужны переменные графической сессии:
`WAYLAND_DISPLAY`, `XDG_RUNTIME_DIR`, `DBUS_SESSION_BUS_ADDRESS`
(плюс `DISPLAY` для X11-фолбэк-путей). GDM при входе импортирует их в
user-менеджер — раннер наследует их автоматически. Проверка и ручной
импорт, если чего-то не хватает:

```console
$ systemctl --user show-environment | grep -E 'WAYLAND_DISPLAY|XDG_RUNTIME_DIR|DBUS_SESSION_BUS_ADDRESS'
$ # при отсутствии (например, вход без GDM):
$ systemctl --user import-environment WAYLAND_DISPLAY XDG_RUNTIME_DIR DBUS_SESSION_BUS_ADDRESS DISPLAY
```

Альтернатива импорту — `EnvironmentFile=` в юните (файл вне git, в
`$HOME`); выбранный на этой машине способ (импорт через GDM) зафиксирован
в SUMMARY плана.

### Включение и проверка

```console
$ systemctl --user daemon-reload
$ systemctl --user enable --now goswitch-ci-runner
$ systemctl --user is-active goswitch-ci-runner     # → active
$ gh api repos/Djarvur/goswitch/actions/runners --jq '.runners[] | {name, status, labels: [.labels[].name]}'
```

Ожидание: раннер со статусом `online` и ярлыками `self-hosted, Linux,
X64, gnome`.

## Bootstrap-регистрация воркфлоуса

GitHub перечисляет `workflow_dispatch`-воркфлоусы по файлу на ветке ПО
УМОЛЧАНИЮ (default branch, здесь — `main`). Пока `e2e-matrix.yml` нет на
`main`, `gh workflow run e2e-matrix.yml` отвечает «could not find any
workflows named e2e-matrix» — даже если файл давно живёт на фаза-ветке.

Первый запуск поэтому идёт заявленной последовательностью:

1. Проверить регистрацию: `gh api repos/Djarvur/goswitch/actions/workflows --jq '.workflows[].path'`.
2. Попытка диспатча с фаза-ветки до регистрации фиксирует фактическое
   поведение (ожидаемо — workflow-not-found).
3. Bootstrap: свежая ветка от default branch, в диффе — ТОЛЬКО файл
   `.github/workflows/e2e-matrix.yml` (минимальный chore-PR, без кода и
   шагов — T-02-07-04):

```console
$ git fetch origin
$ git checkout -b chore/e2e-matrix-workflow origin/main
$ # добавить единственный файл .github/workflows/e2e-matrix.yml
$ git add .github/workflows/e2e-matrix.yml && git commit -m "ci: register e2e-matrix workflow (dispatch bootstrap)"
$ git push -u origin chore/e2e-matrix-workflow
$ gh pr create --fill --head chore/e2e-matrix-workflow
$ gh pr merge --squash --delete-branch
```

4. Диспатч с нужной веткой: `gh workflow run e2e-matrix.yml --ref <ветка>`.
   Определение воркфлоуса берётся от default branch, checkout — от `ref`
   (код матрицы живёт на фаза-ветке до гейта фазы).
5. Ждать: `gh run watch` / `gh run list --workflow e2e-matrix.yml`;
   артефакт: `gh run download <run-id> -n e2e-report`.

(Уточнение 03-07, живое доказательство: исполнится ОПРЕДЕЛЕНИЕ воркфлоуса
от диспатч-ветки `ref` — входы и шаги той версии файла, что на `ref`.
Регистрация на default branch нужна лишь для того, чтобы workflow вообще
значился в `gh workflow run`.)

## D-48 double-run gate (план 04-06)

Критерий приёмки фазы: «Матрица зелёная ДВАЖДЫ ПОДРЯД на нетронутой
сессии». Workflow `e2e-matrix` исполняет ДВА последовательных шага
прогона в одном job — без пересоздания сессии и любого вмешательства
между ними; суть проверки в том, что первый прогон не портит состояние
для второго. Любое падение любого прогона краснит job.

Процедура формального гейта (владелец, приёмка фазы):

1. Завершить сессию и перелогиниться через GDM. Раннер возвращается сам:
   user-юнит `goswitch-ci-runner` висит на `graphical-session.target`
   и стартует вместе с новой сессией (см. раздел systemd выше).
2. Дождаться, что раннер снова `online`, и запустить гейт с проверкой
   свежести:

   ```console
   $ gh api repos/Djarvur/goswitch/actions/runners --jq '.runners[] | {name, status}'
   $ gh workflow run e2e-matrix.yml --ref main \
       -f matrix=test/e2e/cases/matrix-v3.yaml -f fresh_session=true
   ```

3. Дождаться завершения (`gh run watch`) и сверить критерий успеха:
   ОБА шага прогона («matrix v3 run #1» и «matrix v3 run #2») зелёные
   в ОДНОМ запуске — conclusion `success`.

Свежесть не декларируется, а проверяется: при `fresh_session=true` job
начинается с loginctl-префлайта — возраст графической сессии раннера
сверяется с порогом 30 минут (константа шага `D48_MAX_SESSION_AGE_SEC`
в workflow); превышение = job падает с подсказкой «owner relogin
required (D-48)» → перелогиниться и повторить диспатч. Диспатч механики
двойного прогона на текущей сессии — та же команда без
`-f fresh_session=true`; формальный свежесессионный прогон — шаг
приёмки фазы (verify-work/04-07).

## Автономный ночной гейт (D-48 v2) (план 04-09, G-4-2)

Директива владельца 2026-09-17: «придумай, как тестировать это без
участия человека». Формальное доказательство D-48 — зелёный двойной
прогон матрицы на свежей сессии — становится НОЧНЫМ АВТОНОМНЫМ
ПРОГОНОМ. Цепочка:

```
root-таймер 03:50: systemctl restart gdm
  → GDM autologin входит владельца заново
  → свежая графическая сессия поднимает graphical-session.target
  → user-юниты стартуют сами: goswitch-ci-runner (раннер) +
    goswitch-d48-dispatch (диспетчер сессии)
  → диспетчер через ~2 мин запускает гейт:
    scripts/d48-nightly-dispatch.sh
    (= gh workflow run e2e-matrix fresh_session=true)
```

Человек не набирает команду и не перелогинивается никогда;
fresh_session-префлайт (loginctl ≤ 30 мин) остаётся МАШИННЫМ свидетелем
свежести — семантика доказательства не ослаблена, из цикла убран
именно человек.

> **Безопасность — осознанный tradeoff владельца.** Autologin на
> личной машине означает: любой у консоли входит в сессию владельца
> (T-04-09-01, accepted). Владелец применяет настройку ниже ЗНАЯ это;
> репозиторий возит только ДОКУМЕНТАЦИЮ — НИЧЕГО в репозитории не
> исполняет sudo и не применяет root-часть само. Все root-файлы из
> блоков 1–2 применяются руками владельца, один раз.

### Блок 1 — GDM autologin (root, вручную)

`/etc/gdm3/custom.yaml`, секция `[daemon]` (создать файл/секцию, если
нет):

```yaml
[daemon]
AutomaticLoginEnable=true
AutomaticLogin=<логин владельца>
```

Tradeoff — см. security-нотаут выше: включённый autologin входит в
сессию без пароля у физической консоли.

### Блок 2 — root-таймер пересоздания сессии (root, вручную)

`/etc/systemd/system/d48-nightly-relogin.service`:

```ini
[Unit]
Description=goswitch D-48 nightly: restart GDM for a fresh autologin session

[Service]
Type=oneshot
ExecStart=/usr/bin/systemctl restart gdm
```

`/etc/systemd/system/d48-nightly-relogin.timer`:

```ini
[Unit]
Description=goswitch D-48 nightly relogin schedule (03:50)

[Timer]
OnCalendar=*-*-* 03:50:00

[Install]
WantedBy=timers.target
```

Применение и проверка:

```console
$ sudo systemctl daemon-reload
$ sudo systemctl enable --now d48-nightly-relogin.timer
$ systemctl list-timers d48-nightly-relogin.timer
```

`restart gdm` разбирает текущую сессию; autologin возвращает
пользователя, user-юниты стартуют сами — в этом и есть механизм
«свежей сессии без человека».

### Блок 3 — systemd USER units (без root)

**(а) Раннер.** Существующий `goswitch-ci-runner.service` (раздел
«systemd user unit» выше) уже висит на `graphical-session.target`
(`After=`, `Restart=on-failure`) и переиспользует вручную
зарегистрированный каталог `~/actions-runner` — дополнительной
настройки не нужно, только включение, если ещё не включён:

```console
$ systemctl --user enable goswitch-ci-runner
```

**(б) Диспетчер сессии.** Файл
`~/.config/systemd/user/goswitch-d48-dispatch.service`:

```ini
[Unit]
Description=goswitch D-48 nightly dispatch (fresh session gate)
PartOf=graphical-session.target

[Service]
Type=oneshot
# Гвард ~2 мин: дать столу и раннеру устояться после входа.
ExecStartPre=/usr/bin/sleep 120
# Подставить ФАКТИЧЕСКИЙ путь рабочей копии репозитория на этой машине.
ExecStart=%h/goswitch/scripts/d48-nightly-dispatch.sh
# Опционально, файл вне git: строка D48_REF=<ветка> для до-мерж
# диспатчей фаза-ветки; без файла диспетчер идёт на main.
EnvironmentFile=-%h/.config/goswitch/d48-dispatch.env

[Install]
WantedBy=graphical-session.target
```

Применение и проверка:

```console
$ systemctl --user daemon-reload
$ systemctl --user enable goswitch-d48-dispatch
$ journalctl --user -u goswitch-d48-dispatch
```

### Ин-репо пояс (post-merge)

После мержа фазы в `main` workflow сам несёт ночной `schedule`
(cron `10 1 * * *` в **UTC**) — второй, ИЗБЫТОЧНЫЙ путь: GitHub
диспатчит определение default branch и без диспетчера сессии.
Concurrency-группа `e2e-gnome` (без cancel-in-progress) ставит
cron-прогон в очередь за прогоном диспетчера — два стенда на одном
столе по-прежнему невозможны. Мягкий режим префлайтов (idle-desktop и
fresh-session) делает неудачный слот нейтральным: `::warning::` +
пропуск прогонов + `exit 0` — без машинной настройки (блоки 1–3)
ночные прогоны НЕ краснеют. Часовой пояс: cron в UTC, root-таймер в
machine-local времени — при регионе, отличном от UTC+3, сдвинуть
cron так, чтобы слот оставался ~20 минут после автологина.

### Семантика доказательства

Формальное доказательство D-48 = АВТОМАТНЫЙ зелёный двойной прогон
(«matrix v3 run #1» + «matrix v3 run #2» в одном job, conclusion
`success`) на автологин-свежей сессии. fresh_session-префлайт
(loginctl ≤ 30 мин) остаётся машинным свидетелем свежести — по
директиве владельца 2026-09-17 из цикла убран человек, а не проверка.
До применения машинной настройки (блоки 1–3) scheduled-прогоны
завершаются нейтрально (мягкий режим) — без красного ночного шума;
реальное падение прогона при пройденных префлайтах краснеет при любом
триггере — ночной красный остаётся регрессионным сигналом.

## Обновление раннера

GitHub помечает устаревшие раннеры в Settings → Actions → Runners.
Обновление — тот же тарболл поверх рабочего каталога (конфигурация и
регистрация сохраняются):

```console
$ systemctl --user stop goswitch-ci-runner
$ cd ~/actions-runner
$ curl -o actions-runner-linux-x64.tar.gz -L <URL-нового-тарболла>
$ tar xzf actions-runner-linux-x64.tar.gz
$ systemctl --user start goswitch-ci-runner
```

## Снятие раннера с учёта

```console
$ systemctl --user disable --now goswitch-ci-runner
$ cd ~/actions-runner
$ TOKEN=$(gh api --method POST repos/Djarvur/goswitch/actions/runners/remove-token --jq .token)
$ ./config.sh remove --token "$TOKEN"
$ unset TOKEN
```

## Диагностика живого стола

Стенд печатает на живой сессии — окружение стола иногда деградирует.
Полный разбор в `test/e2e/README.md` (раздел «Матрица v1»); самое
частое:

- **Осиротевшее имя IBus** (след оборванного прогона; префлайт матрицы
  отказывается стартовать): лечится `ibus restart` — кейс ibus-restart
  доказывает безопасность.
- **Зависание моста gnome-shell** (обходы AT-SPI виснут до таймаута):
  перезапуск ВСЕЙ шины a11y — `kill <pid at-spi-bus-launcher> <pid её
  dbus-daemon>`; стек dbus-активируется заново, мосты пересобираются.
- **Один стенд на столе**: два параллельных стенда печатали бы в чужие
  окна — префлайт против живого `goswitchd`, concurrency-группа
  `e2e-gnome` в workflow.

## Безопасность

Раннер исполняет код репозитория на живом столе владельца — поэтому:

- **Только этот репозиторий**: раннер зарегистрирован в
  Djarvur/goswitch и не может подбираться чужими репозиториями аккаунта.
- **Runner-группа**: при желании сузить ещё сильнее — Settings → Actions
  → Runner groups: группа `e2e-gnome` с доступом ровно одного
  репозитория и ярлыком `gnome` (по умолчанию раннер попадает в группу
  Default; ограничение группы — рекомендация этого раздела).
- **Триггеры — workflow_dispatch + ночной schedule** (T-02-07-01 в
  силе): fork/push-триггеры отсутствуют — fork-PR физически не может
  запустить матрицу; диспатч — право ролей с write; schedule исполняет
  ТОЛЬКО определение default branch (ветка защищена, изменения через
  PR). Юнит диспетчера сессии (D-48 v2) исполняет только ин-репо
  скрипт с gh-гвардом; root-часть машинной настройки — исключительно
  ручное применение владельцем, репозиторий sudo не исполняет.
- **permissions: contents: read**, секретов workflow не использует.
- **concurrency e2e-gnome без cancel-in-progress**: один прогон за раз.
- **Registration-token** одноразовый, живёт минуты, в git не попадает
  (T-02-02-02b у Фазы 2 — та же дисциплина: минт в переменную, без
  логирования).
- Печать на живом столе — осознанное решение владельца (D-19), реестр
  угроз ведёт `SECURITY.md`.
