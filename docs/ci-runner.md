# CI-раннер e2e — self-hosted на ВМ goswitch-stand (план 02-07, D-19)

Владелец осознанно выбрал CI-контур уже в Фазе 2 (D-19): зелёная матрица
на раннере — часть phase gate. e2e-стенду нужна настоящая GNOME
Wayland-сессия с фокусом окон, эфемерная ubuntu-latest дать её не может.
Workflow `.github/workflows/e2e-matrix.yml` запускается вручную
(`workflow_dispatch`), прогоняет `mise run e2e-matrix` — ту же задачу,
что и локальный запуск (D-09), — и выгружает `e2e-report.txt` артефактом.

> **История размещения.** До 2026-10-23 раннер `green106` жил ВНУТРИ
> графической сессии владельца на его рабочем столе. Ночной schedule
> (cron) каждым прогоном подвергал живой стол AT-SPI-обходу
> (idle-desktop префлайт) — и дефект кеша at-spi2-core 2.52.0 на живом
> столе регулярно ронял gnome-shell целиком (5+ logout-ов сеанса
> владельца; диагноз и стек — в разделе «Диагностика»). 2026-10-23
> раннер перенесён на стенд-ВМ `goswitch-stand`, ночной schedule удалён
> из workflow (PR #22): прогоны — только ручной диспатч, печатают на
> столе ВМ, крах оболочки ВМ не задевает работу владельца.

Машина стенда — libvirt-ВМ «goswitch-stand» (Ubuntu 24.04, GNOME 46),
пользователь `stand`, GDM autologin включён (`/etc/gdm3/custom.conf`:
`AutomaticLoginEnable=true`, `AutomaticLogin=stand`) — сессия
самовосстанавливается после краха оболочки или ребута ВМ.

Инструкция воспроизводима без root на стороне раннера: всё живёт в
`$HOME` (каталог раннера, user unit, окружение сессии). Исключение —
пакеты стенда (sudo на ВМ есть) и gh CLI с admin-правом на репозитории
для минта токенов — годится любая машина с аутентифицированным gh
(например, хост владельца).

## Требования к машине

Из STACK (research Фазы 0, проверено на цели живьём):

| Требование                                    | Проверка                                                       |
|-----------------------------------------------|----------------------------------------------------------------|
| Живая GNOME Wayland-сессия (не headless)      | `echo $WAYLAND_DISPLAY` → `wayland-0`                          |
| GDM autologin (самовосстановление сессии ВМ)  | `grep AutomaticLogin /etc/gdm3/custom.conf`                    |
| Пользователь в группе `input`                 | `id -nG | grep -w input`                                       |
| `/dev/uinput` доступен на запись              | `test -w /dev/uinput && echo ok`                               |
| ydotool                                       | `ydotool --version` (0.1.8 — открывает `/dev/uinput` напрямую) |
| python3-gi + AT-SPI                           | `/usr/bin/python3 -c "import gi; gi.require_version('Atspi','2.0')"` |
| gir1.2-atspi-2.0, at-spi2-core                | `dpkg -l gir1.2-atspi-2.0 at-spi2-core`                        |
| systemd-coredump (стыковка крахов оболочки)   | `coredumpctl --version`                                        |
| Поверхности матрицы                           | `zenity --version`; `google-chrome --version`; `gnome-text-editor --version` |
| gedit — поверхность матрицы v3 (план 04-05)   | `gedit --version` (GTK3-поколение IM/AT-SPI; в дефолтный GNOME НЕ входит — `sudo apt install gedit`, репозиторий universe) |
| gh CLI с admin на репозитории (не на ВМ)      | `gh api repos/Djarvur/goswitch/actions/runners` → HTTP 200    |
| curl, tar, git                                | `command -v curl tar git`                                      |

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
    --labels gnome --name goswitch-stand --unattended
$ unset TOKEN
```

Ярлык `gnome` — обязательный: workflow ищет раннера по
`runs-on: [self-hosted, gnome]`. Дефолтные ярлыки (self-hosted, Linux,
X64) config.sh ставит сам.

## systemd user unit (графическая сессия)

Раннер обязан жить ВНУТРИ графической сессии стенда — иначе у стенда нет
ни Wayland-композитора, ни session-bus, ни IBus. Прецедент — юнит демона
`dist/systemd/user/goswitchd.service` (тот же паттерн «после
graphical-session.target»). Файл
`~/.config/systemd/user/goswitch-ci-runner.service`:

```ini
[Unit]
Description=goswitch CI runner (e2e-matrix, live GNOME session on goswitch-stand)
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
`$HOME`); выбранный на машине способ (импорт через GDM) зафиксирован
в SUMMARY плана.

### Включение и проверка

```console
$ systemctl --user daemon-reload
$ systemctl --user enable --now goswitch-ci-runner
$ systemctl --user is-active goswitch-ci-runner     # → active
$ gh api repos/Djarvur/goswitch/actions/runners --jq '.runners[] | {name, status, labels: [.labels[].name]}'
```

Ожидание: раннер `goswitch-stand` со статусом `online` и ярлыками
`self-hosted, Linux, X64, gnome`.

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

Процедура формального гейта (на ВМ goswitch-stand):

1. Пересоздать сессию стенда: на ВМ `sudo systemctl restart gdm` (или
   ребут ВМ) — autologin поднимет свежую сессию, user-юнит
   `goswitch-ci-runner` висит на `graphical-session.target` и стартует
   сам.
2. Дождаться, что раннер снова `online`, и запустить гейт с проверкой
   свежести:

   ```console
   $ gh api repos/Djarvur/goswitch/actions/runners --jq '.runners[] | {name, status}'
   $ gh workflow run e2e-matrix.yml --ref main \
       -f matrix=test/e2e/cases/matrix-v4.yaml -f fresh_session=true
   ```

3. Дождаться завершения (`gh run watch`) и сверить критерий успеха:
   ОБА шага прогона («matrix run #1» и «matrix run #2») зелёные
   в ОДНОМ запуске — conclusion `success`.

Свежесть не декларируется, а проверяется: при `fresh_session=true` job
начинается с loginctl-префлайта — возраст графической сессии раннера
сверяется с порогом 30 минут (константа шага `D48_MAX_SESSION_AGE_SEC`
в workflow); превышение = job падает с подсказкой «owner relogin
required (D-48)» → пересоздать сессию ВМ и повторить диспатч. Диспатч
механики двойного прогона на текущей сессии — та же команда без
`-f fresh_session=true`; формальный свежесессионный прогон — шаг
приёмки фазы (verify-work/04-07).

## Ночной автономный гейт — ВЫВЕДЕН из эксплуатации (2026-10-23)

Ночная цепочка (root-таймер 03:50 → GDM autologin → диспетчер сессии
`goswitch-d48-dispatch.service` → диспатч fresh_session-гейта; ин-репо
ночной `schedule` в workflow) выведена из эксплуатации решением владельца
2026-10-23 после серии ночных крахов gnome-shell на живом столе
(причина — в разделе «Диагностика»; PR #22 удалил cron, раннер перенесён
на ВМ). Машина владельца больше ничего не исполняет по расписанию:
root-таймера на хосте нет, диспетчер сессии не включён, schedule в
workflow отсутствует.

Возврат ночного прогона — только на ВМ goswitch-stand, только по явному
решению владельца, после зелёной серии ручных прогонов на ВМ (гейт
свежести и двойной прогон при этом сохраняются как есть).

## Матрицы: v1 → v4 (06-07)

Файлы матриц живут в `test/e2e/cases/`; каждая — superset предыдущей,
а ЗАМОРОЖЕННЫЕ файлы не меняются никогда (регресс-носитель):

| Матрица | План | Статус | Содержимое |
|---------|------|--------|------------|
| `matrix-v1.yaml` | 02-06 | frozen | 16 кейсов D-18 (слова/фразы/сбросы) |
| `matrix-v2.yaml` | 03-07 | frozen | superset v1: фразы, выделение, комбо, reload, MACR (21) |
| `matrix-v3.yaml` | 04-05 | frozen (гейт `git diff --exit-code`) | superset v2: gedit + chromium-x11 (32) |
| `matrix-v4.yaml` | 06-07 | ТЕКУЩАЯ, default диспатча | superset v3 байт-дословно + блок autocorrect: `autocorrect-fires` (исправление без жеста, `config_base: autocorrect` + keystroke/fresh-старты наблюдателя), `autocorrect-off` (дефолтный демон — ноль поведения), `autocorrect-oov` (двусторонний словарный промах → trigram-unsure, нетронуто) — 35 кейсов |

Правила v4:

- **Superset без ослабления**: строки v3 перенесены байт-дословно; две
  из них заморожены красными под two-source семантику сброса контекста
  (`word-mixed`, `phrase-mixed` — WINDOWS #12, вердикт владельца на
  verify-гейте фазы) и несутся как есть — выпадение ослабило бы матрицу.
- **Парольный негатив в матрицу НЕ включён честно**: у драйвера
  поверхностей нет password-шага (Tab-to-password в chromium — омнибокс-зона
  риска, см. заголовок v3); роль-негатив несёт автономный кейс
  `mise run e2e-autocorrect-password-silent` (GTK4-фикстура).
- **Дефолт диспатча — v4**; явное `-f matrix=…v3.yaml` запускает
  v3-регрессию (та же форма для v1/v2).
- **Двойной прогон D-48 сохранён** (см. раздел выше): оба шага
  гоняют ОДНУ выбранную матрицу.

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

Старый раннер хоста `green106` (остановлен и выключен 2026-10-23, файлы
в `~/actions-runner` на хосте сохранены): серверная регистрация остаётся
в статусе offline и заданий не получает; полное снятие — командой выше
на хосте, быстрый возврат — `systemctl --user enable --now
goswitch-ci-runner` на хосте (пока регистрация не снята).

## Диагностика живого стола

Стенд печатает на живой сессии ВМ — окружение стола иногда деградирует.
Полный разбор в `test/e2e/README.md` (раздел «Матрица v1»); самое
частое:

- **Осиротевшее имя IBus** (след оборванного прогона; префлайт матрицы
  отказывается стартовать): лечится `ibus restart` — кейс ibus-restart
  доказывает безопасность.
- **Зависание моста gnome-shell** (обходы AT-SPI виснут до таймаута):
  перезапуск ВСЕЙ шины a11y — `kill <pid at-spi-bus-launcher> <pid её
  dbus-daemon>`; стек dbus-активируется заново, мосты пересобираются.
- **Крах gnome-shell во время AT-SPI-обхода** (сегфолт в `libgobject`):
  известный дефект at-spi2-core 2.52.0 (Ubuntu 24.04, GNOME 46) —
  use-after-free в кеше моста atk-adaptor. Любой клиент, обошедший
  AT-SPI-дерево целиком (наш idle-desktop префлайт, экранный диктор),
  вызывает `GetItems` → `impl_GetItems` (cache-adaptor.c:327) →
  `spi_cache_foreach` (accessible-cache.c:423) → `g_object_ref` на
  уже освобождённый AtkObject → SIGSEGV всей оболочки. Кеш хранит сырые
  указатели без слабых ссылок (accessible-cache.c:112,
  `g_hash_table_new (g_direct_hash, g_direct_equal)`); объект,
  умерший без сигнала deregister, остаётся висячим указателем. Фикс
  апстрима — коммит d442ee182e «atk-adaptor: Have the SpiCache keep its
  own weak ref» (вошёл в 2.55.2/2.56, NEWS: «Attempt to fix a use after
  free in the atk bridge»); в 24.04 дефект остаётся. На стенде: дамп
  ловит systemd-coredump (`coredumpctl list`), сессию поднимает GDM
  autologin; бегущий job при этом закончится аварией (раннер умер
  вместе с сессией) — перезапустить диспатч. Живой дамп такого краха
  хоста (2026-10-21) разбран в
  `.planning/debug/2026-10-23-atspi-cache-uaf.md`.
- **Матрица v1–v4 на ВМ — фокус-гейты кейсов (открыто, 2026-10-24)**:
  конвейер CI на ВМ работает целиком (checkout, mise, сборка демона,
  префлайты, отчёт), но сами кейсы виснут на фокус-гейтах драйвера
  (waitZenityEntry и родственные): стол ВМ не выдаёт map-focus
  заспавненным окнам (та же находка плана 10-08 про терминалы, здесь
  подтверждена для zenity), pid-тычки `grab-input-pid` фокус не двигают
  (на хосте двигали), фреймовый grabFocus отдаёт `atspi_error`, клики
  по геометрии окна не приземляются. Матрица v1–v4 исторически
  пинилась ТОЛЬКО на живом столе хоста; на ВМ зелёной ещё не была.
  Открытая работа: генерализация клик-провокации фазы 10 на общие
  фокус-гейты драйвера (отдельная задача с собственным планом; на хосте
  изменения нейтральны — свидетель проходит до вмешательства).
- **Один стенд на столе**: два параллельных стенда печатали бы в чужие
  окна — префлайт против живого `goswitchd`, concurrency-группа
  `e2e-gnome` в workflow.

## Безопасность

Раннер исполняет код репозитория на живом столе — с 2026-10-23 это стол
ВМ goswitch-stand, а не рабочая машина владельца:

- **Только этот репозиторий**: раннер зарегистрирован в
  Djarvur/goswitch и не может подбираться чужими репозиториями аккаунта.
- **Runner-группа**: при желании сузить ещё сильнее — Settings → Actions
  → Runner groups: группа `e2e-gnome` с доступом ровно одного
  репозитория и ярлыком `gnome` (по умолчанию раннер попадает в группу
  Default; ограничение группы — рекомендация этого раздела).
- **Триггер — только workflow_dispatch** (T-02-07-01 в силе): fork/push-
  и schedule-триггеры отсутствуют — fork-PR физически не может запустить
  матрицу; диспатч — право ролей с write.
- **Autologin на ВМ** — осознанный tradeoff: физическая консоль ВМ
  (окно virt-manager на хосте) входит в сессию стенда без пароля;
  секретов в сессии стенда нет, рабочие данные владельца на ВМ не
  хранятся. Рабочая машина владельца autologin не использует.
- **permissions: contents: read**, секретов workflow не использует.
- **concurrency e2e-gnome без cancel-in-progress**: один прогон за раз.
- **Registration-token** одноразовый, живёт минуты, в git не попадает
  (T-02-02-02b у Фазы 2 — та же дисциплина: минт в переменную, без
  логирования).
- Печать на живом столе ВМ — решение владельца (D-19; перенос на ВМ —
  2026-10-23), реестр угроз ведёт `SECURITY.md`.
