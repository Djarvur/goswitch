---
phase: quick-261108-qwa
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - .goreleaser.yaml
  - README.md
  - README.en.md
autonomous: true
requirements:
  - QUICK-261108-QWA
estimate:
  tokens: 24000
  raw_tokens: 18000
  tasks: 2
  confidence: med

must_haves:
  truths:
    - В .goreleaser.yaml ОБА билда (goswitchd и goswitchctl) несут список архитектур [amd64, arm64]; список goos остаётся [linux]; никаких других архитектур и никаких новых ключей не добавлено (A7 «+ nothing else»)
    - Оба комментария конфига (заголовочный discretion-блок A7 и инлайн-комментарий у списка goarch первого билда) фиксируют исполнение: arm64 добавлен по запросу владельца 2026-11-08
    - Локальный снапшот `mise exec -- goreleaser release --snapshot --clean` зелёный и производит два архива — goswitch_<version>_linux_amd64.tar.gz и goswitch_<version>_linux_arm64.tar.gz; каждый архив несёт ОБА бинарника; artifacts.json фиксирует goarch arm64 у обоих бинарников
    - README.md: пункт раздела «Требования» называет обе архитектуры (linux/amd64 или linux/arm64); шаг 1 канала А называет оба имени архива; пример tar -xzf использует маску goswitch_*_linux_*.tar.gz
    - README.en.md несёт зеркальные правки в тех же трёх местах (Requirements / Channel A шаг 1 / пример tar)
    - Весь прочий текст обоих README нетронут — правки строго в трёх местах на файл
  artifacts:
    - .goreleaser.yaml
    - README.md
    - README.en.md
  key_links:
    - Имена архивов в README ↔ фактические имена goreleaser default name_template (goswitch_<version>_linux_<arch>.tar.gz)
    - .goreleaser.yaml goarch ↔ сборка релиза: конфиг — ЕДИНСТВЕННЫЙ источник архитектур (release.yml гоняет `goreleaser release --clean` без матрицы, других упоминаний архитектур в репо нет)
---

<objective>
Релизы goswitch начинают публиковать бинарники под linux/arm64 в дополнение к linux/amd64 — запрос владельца 2026-11-08 («публикуются ли у нас бинарники под линукс, amd64 и arm64? если нет — надо начать публиковать!»). Это И ЕСТЬ тот самый оговорённый запрос, который комментарий A7 в .goreleaser.yaml ждал с Фазы 4: «add arm64 here (+ nothing else) on owner request».

Три точечные правки:
1. .goreleaser.yaml — goarch: [amd64, arm64] на ОБЕИХ сборках (goswitchd, goswitchctl); оба комментария (заголовочный discretion-блок + инлайн) переведены в исполненное состояние с датой запроса;
2. README.md — требования, имя скачиваемого архива, пример tar -xzf;
3. README.en.md — зеркально те же три места.

Purpose: arm64-пользователи (ARM-ноутбуки/серверы/SBC с GNOME Wayland) получают релизный канал; RU/EN документация честно называет обе архитектуры.
Output: один конфиг-файл релиза + два README. Код демона, go.mod, CI-workflow, прочая документация — нетронуты.
</objective>

<execution_context>
@/home/nil/DiskD/W/Djarvur/goswitch/.zcode/gsd-core/workflows/execute-plan.md
@/home/nil/DiskD/W/Djarvur/goswitch/.zcode/gsd-core/templates/summary.md
</execution_context>

<context>
@.goreleaser.yaml
@README.md
@README.en.md
@mise.toml

Проверенные факты (живые пробы на этом хосте, 2026-11-08):
- `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ./...` и `go vet ./...` зелёные — код чистый Go (stdlib + godbus), кросс-компиляция arm64 с хоста amd64 работает, cgo не нужен.
- goreleaser v2 (пин 2.18.1 в mise.toml [tools]) с default name_template даёт архивы goswitch_<version>_linux_amd64.tar.gz и goswitch_<version>_linux_arm64.tar.gz; каждый пер-архитектурный архив несёт ОБА бинарника (goswitchd + goswitchctl) — эта семантика остаётся.
- Скоуп ровно linux amd64 + arm64 — A7 «+ nothing else»: никаких darwin/windows/386/armv7/riscv64.
- release.yml уже гоняет `mise exec -- goreleaser release --clean` без собственной матрицы архитектур — .goreleaser.yaml единственный источник; CI менять не нужно.
- .goreleaser-dist/ — gitignored вывод сборки (предыдущий снапшот-прогон существует; --clean его затирает — это штатно).
- Директива 6 (проверки только на ВМ) касается живых product-проверок (e2e, инжект, a11y); `goreleaser release --snapshot` и `mise run ci` — сборочно-тестовые проверки класса `go build`, на хосте легальны.
</context>

<tasks>

<task type="auto">
  <name>Task 1: .goreleaser.yaml — добавить arm64 в оба билда и зафиксировать исполненный запрос A7 в комментариях</name>
  <files>.goreleaser.yaml</files>
  <action>
Три правки в файле .goreleaser.yaml, БОЛЬШЕ НИЧЕГО:

1. Билд goswitchd (первый элемент builds): список goarch с одной архитектурой заменить на `goarch: [amd64, arm64]`. Инлайн-комментарий у этого списка переписать в исполненном виде: arm64 добавлен по запросу владельца 2026-11-08 (A7 — ровно arm64, ничего сверх; предыдущая оговорка «+arm64 on owner request» исполнена).

2. Билд goswitchctl (второй элемент builds): тот же список `goarch: [amd64, arm64]` (комментария у него нет и не появляется).

3. Заголовочный discretion-блок (комментарий перед `version: 2`, пункт про linux-архитектуры): переписать первый пункт блока из обещания («add arm64 here (+ nothing else) on owner request») в исполненное состояние — linux/amd64 + linux/arm64; arm64 добавлен запросом владельца 2026-11-08; за пределами этих двух архитектур расширения нет (A7 «+ nothing else»).

Не трогать: goos: [linux] у обоих билдов; секции archives/checksum/changelog/release; dist: .goreleaser-dist; НЕ добавлять новые ключи (env, builds-матрицы, name_template, overrides) — кросс-компиляция arm64 собирается без специального env (чистый Go, cgo при кросс-сборке выключается сам).
  </action>
  <verify>
    <automated>mise exec -- goreleaser check && mise exec -- goreleaser release --snapshot --clean && ls .goreleaser-dist/goswitch_*_linux_amd64.tar.gz .goreleaser-dist/goswitch_*_linux_arm64.tar.gz && test "$(grep -c '"goarch":"arm64"' .goreleaser-dist/artifacts.json)" -ge 3 && tar -tzf .goreleaser-dist/goswitch_*_linux_arm64.tar.gz | grep -c goswitchctl</automated>
  </verify>
  <done>
goreleaser check зелёный; снапшот-релиз собирает 4 бинарника (2 билда × 2 архитектуры) и 2 архива с каноническими именами; arm64-архив содержит оба бинарника (goswitchd + goswitchctl); artifacts.json фиксирует goarch arm64 не менее чем в трёх записях (оба бинарника + архив); оба комментария конфига называют дату исполнения запроса владельца 2026-11-08; в diff файла — только списки goarch и два комментария.
  </done>
</task>

<task type="auto">
  <name>Task 2: README.md и README.en.md — обе архитектуры в требованиях, имени архива и примере tar</name>
  <files>README.md, README.en.md</files>
  <action>
Ровно три места в каждом файле, зеркально:

README.md (русский):
1. Раздел «Требования», третий пункт: пару «linux/amd64» с квалификацией в единственном числе заменить на «linux/amd64 или linux/arm64» с квалификацией во множественном числе (архитектуры релизных бинарников). Пара архитектур держится на одной строке — перенос строки её не разрывает.
2. Раздел «Установка», канал А («релизный архив»), шаг 1: перед именем архива добавить выбор по архитектуре процессора и назвать ОБА имени — `goswitch_<version>_linux_amd64.tar.gz` или `goswitch_<version>_linux_arm64.tar.gz`; каждое имя файла целиком на одной строке; хвост шага про содержимое архива (оба бинарника и checksums.txt) не трогать.
3. Там же, шаг 2, код-блок: строку распаковки с маской, зашивающей одну архитектуру, заменить на маску `goswitch_*_linux_*.tar.gz` — одна команда tar -xzf работает для любого из двух скачанных архивов.

README.en.md (английский) — те же три места зеркально:
1. Раздел Requirements, третий пункт: «linux/amd64 or linux/arm64» + квалификация во множественном числе (the release binaries' architectures); пара на одной строке.
2. Раздел Install, Channel A, шаг 1: выбор архива по архитектуре процессора + оба имени файла, каждый целиком на одной строке.
3. Код-блок шага 2: та же универсальная маска `goswitch_*_linux_*.tar.gz`.

Ограничения: сохранить принятую жёсткую разбивку строк (~74 символа) обоих файлов; имён файлов, пары архитектур и код-блока перенос не разрывает; весь прочий текст обоих README нетронут; язык каждого файла сохраняется своим (русский/английский), формулировки рядом с правками не переписываются.
  </action>
  <verify>
    <automated>grep -qF "linux/amd64 или linux/arm64" README.md && grep -qF "_linux_arm64.tar.gz" README.md && test "$(grep -cF "goswitch_*_linux_*.tar.gz" README.md)" -eq 1 && grep -qF "linux/amd64 or linux/arm64" README.en.md && grep -qF "_linux_arm64.tar.gz" README.en.md && test "$(grep -cF "goswitch_*_linux_*.tar.gz" README.en.md)" -eq 1 && ! grep -qF "goswitch_*_linux_amd64.tar.gz" README.md && ! grep -qF "goswitch_*_linux_amd64.tar.gz" README.en.md && ! grep -qF "архитектура релизных бинарников" README.md && ! grep -qF "the release binaries' architecture" README.en.md && mise run ci</automated>
  </verify>
  <done>
Оба README в требованиях называют linux/amd64 и linux/arm64; шаг скачивания называет оба имени архива (каждое целиком на одной строке); пример tar -xzf один и универсален для обеих архитектур; квалификации в единственном числе и старая маска одной архитектуры из обоих файлов исчезли; зелёная итерация — mise run ci (build + vet + lint + test -race) без ошибок.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| GitHub Releases → пользователь | Пользователь скачивает релизный архив по инструкциям README и исполняет бинарники у себя в сессии |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-quick-qwa-01 | Tampering | Релизные ассеты: новый arm64-архив + бинарники | medium | mitigate | Целостность остаётся на существующем канале checksum: блок checksum.sha256 (checksums.txt, T-04-03-01) автоматически покрывает ОБА новых архива — Task 1 ничего не меняет в provenance; Task 1 verify проверяет, что arm64-архив и оба arm64-бинарника реально создаются |
| T-quick-qwa-02 | Spoofing | README download-инструкции (RU/EN) | low | accept | Инструкции ведут на официальную страницу Releases через TLS; имена файлов совпадают с default name_template goreleaser; новых зеркал/URL в скоупе нет |
| T-quick-qwa-SC | Tampering | Новые пакеты/инструменты | low | accept | Новых зависимостей НЕТ: goreleaser 2.18.1 уже запинен в mise.toml [tools] с Фазы 4, go.mod не меняется; пакетный аудит не требуется |

ASVS level 1; block_on=high — критичных/высоких угроз нет.
</threat_model>

<verification>
1. `mise exec -- goreleaser check` — конфиг валиден после правок.
2. `mise exec -- goreleaser release --snapshot --clean` — полный локальный прогон без публикации: 4 бинарника (goswitchd/goswitchctl × amd64/arm64), 2 архива с каноническими именами `goswitch_<version>_linux_{amd64,arm64}.tar.gz`, каждый архив несёт оба бинарника, checksums.txt покрывает оба архива.
3. grep-гейты Task 2 на оба README (обе архитектуры в требованиях, оба имени архива, универсальная маска tar, отсутствие старых форм единственной архитектуры).
4. `mise run ci` — зелёная итерация по конвенции 2 (build + vet + golangci-lint + test -race).
5. Живая публикация arm64 произойдёт ближайшим тегом релиза через release.yml (в скоуп этой задачи НЕ входит — тег не ставится).
</verification>

<success_criteria>
- .goreleaser.yaml собирает обе архитектуры: снапшот-прогон зелёный, arm64-архив и arm64-бинарники существуют, каждый архив содержит goswitchd + goswitchctl.
- Комментарии конфига фиксируют: arm64 добавлен по запросу владельца 2026-11-08, A7 «+ nothing else» — других архитектур нет.
- README.md и README.en.md: требования — обе архитектуры; канал А — оба имени архива; пример tar универсален; старые формы одной архитектуры исчезли.
- mise run ci зелёный; diff строго в трёх файлах (.goreleaser.yaml, README.md, README.en.md).
</success_criteria>

<output>
Create `.planning/quick/261108-qwa-relizy-dobavit-linux-arm64-k-reliznym-bi/261108-qwa-SUMMARY.md` when done
</output>
