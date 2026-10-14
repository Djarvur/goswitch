---
status: investigating
trigger: "Phase-10 terminal word correction (double right Shift, EN<->RU) is byte-exact green on the VM stand but fails on the owner's host desktop in two modes: (A) wezterm — daemon self-aborts via verify-timeout (after terminal class verify_dead=true) and verify-mismatch, owner sees nothing; (B) GNOME Terminal (VTE) — daemon reports correction done changed=true (6 times), owner sees backspaces render but the committed replacement word never appears (final CommitText lost). Mode: find_root_cause_only. Host daemon is the owner's live input method — read-only host access only; VM stand experiments allowed with logging."
created: 2026-10-13T00:00:00+03:00
updated: 2026-10-14T12:58:00+03:00
---

## Current Focus

```yaml
reasoning_checkpoint:
  hypothesis: >
    Shared substrate: the host kernel's focus churn (7.2.9-2-t2 background-window
    defect) emits real input-context FocusOut/FocusIn pairs continuously and
    mid-flow, and the phase-10 terminal machinery's two classification signals
    do not survive it. MODE A: on a churn-wiped context, classifyTerminal()
    returns false in wezterm (verifyDead wiped by real FocusOuts at
    actor.go:900 — the ONLY source that can see wezterm, which is a11y-invisible:
    no identity, role query errors -> "not terminal"), so the gesture takes the
    GTK app-verification path (caps bit announced) that wezterm can never answer
    (no surrounding pushes, mutter never answers Require) -> 100 ms budget ->
    VerifyExpiry -> verify-timeout free-abort. The class learned by that very
    timeout is wiped by the next flap burst (journal: learned 16:20:00.334 ->
    wiped 16:20:05.478, flipCredit==0 so REAL FocusOuts), so on the host the
    owner can never get past the learning gesture; on the stand the class
    persists (stable context) and gesture 2 rides it to done+level-2 -> green.
    MODE B: (1) the same degradation misrouted 5 of 6 VTE gestures to the GTK
    level-1/selection paths (instant done rows 0.2-3 ms; executeLevel2 minimum
    is >=40 ms — only 16:20:47 at 145 ms was the true terminal branch), because
    the live AT-SPI role query answered non-terminal at those instants (it
    answered 60 at the 11 veto instants during typing; config has no
    correction.terminals on host AND stand; role errors are silent — the
    one-per-episode WARN was consumed 11:00:16 and no healthy autoConfirm ran
    since). (2) The visible commit loss on the host is the client-side
    commit-vs-queued-forwards reorder class documented live on the HOST
    2026-10-12 («ghbdtn» -> «ghприв») at 0 ms pacing; the 20 ms pacing quantum
    was sized and validated ONLY on the stand (vm-primary venue); the host
    still loses/vanishes the committed word ("conversion happens FIRST, then
    the word disappears" — owner's own word order = late erasure signature).
  confirming_evidence:
    - "Journal: terminal class verify_dead=true at 16:20:00.334 (VerifyExpiry) ->
      verify_dead=false at 16:20:05.478 with NO daemon flip between (flipCredit==0
      -> real FocusOuts — the four flap pairs 16:20:05.481-.569): the wipe is
      directly observed, not inferred."
    - "Journal latency forensics: 8 of 9 done rows all day are 0.2-3 ms after the
      action = level-1/selection GTK paths; only 16:20:47 (145 ms) matches
      executeLevel2's sleep geometry (6x20 ms + 20 ms + overhead)."
    - "Counters: ac_skip_terminal_veto=11 (role 60 answered during typing
      16:20:53-16:21:20) vs the instant done rows at 16:21:37/16:22:14/18/20
      (classification false seconds later) — same window, same app, flip-flop."
    - "Config: host AND stand cases run with no correction.terminals (CONFIG.md
      D7: absent key = classification not configured) — the identity source is
      dead by design on both; only role query + verifyDead remain."
    - "E1 (stand, mutating): term-word-vte with flip_after_correction=true —
      row 1 GREEN (commit + SetGlobalEngine 2 ms apart renders byte-exact);
      row 2's 'failure' is the expected mode-inversion artifact (its commit
      ghbdtn IS in the pty). The flip/re-mint does NOT swallow commits at idle."
    - "E2b (stand, mutating): term-word-vte PASS byte-exact under sustained
      2-of-4-vCPU stress-ng load — generic CPU load does not break the 20 ms
      pacing. The commit loss is not reproducible by idle or loaded stand runs."
    - "10-12 host live finding (pre-pacing): forwarded keys queue as GDK events
      while the commit lands synchronously — the reorder mechanism is real ON
      THE HOST; pacing fix verified only vm-primary (ADR-008 venue verdict)."
  falsification_test: >
    Mode A would be falsified if a flap-era FocusOut could be shown NOT to wipe
    verifyDead (code says it does; journal shows the transition), or if wezterm
    on the host pushed surrounding text (then the GTK verify path would be
    answerable and the timeout needs another cause). Mode B commit-loss would be
    falsified if a host -debug journal after one gesture showed the field's
    post-correction surrounding text CONTAINING the committed word and STAYING
    (then the loss is rendering-only) or if the drain-latency probe showed VTE
    drains forwarded keys well within 20 ms even under churn (then the reorder
    candidate falls and the flap-window commit drop becomes primary).
  fix_rationale: >
    (non-binding, plan-phase --gaps owns it) The root is environment-dependent
    state survival, not the write geometry: classification must not depend on
    signals the host churn destroys (persist verifyDead across synthetic-looking
    rapid pairs or re-arm it from the caps/push signature; make role-query
    failure LOUD or sticky per window instead of silent; or ship a
    correction.terminals default covering gnome-terminal/wezterm so identity
    carries the class), and the terminal write branch needs a host-validated
    pacing/drain strategy (drain-ack from the client or longer settle before the
    commit) rather than a stand-tuned constant.
  blind_spots: >
    (1) The host's IM-level churn density is unmeasurable from the daemon
    journal — the flap-window commit-drop candidate cannot be quantified
    read-only. (2) The exact sub-path of the five instant done rows
    (level-1-with-matching-push vs selection) is not derivable from INFO-level
    journal (the DEBUG level row would disambiguate; host runs INFO). (3) App
    attribution of the six done rows and of п.4's successful «ytn?» is
    hypothesis (the UAT text places п.1/п.2 only). (4) wezterm-nightly's
    surrounding-text behavior is code-analysis-based (phase-9); the nightly
    could push, which would make the 16:20:33 mismatch a wezterm row with a
    churn-diverged buffer instead of a VTE row.
  candidate_causes:
    - "code: classifyTerminal's behavioral source is wiped by ANY FocusOut
      (actor.go:900 unconditional setVerifyDead(false), even synthetic);
      role-query errors are silent after the first episode (warn-once never
      re-armed — no healthy autoConfirm all day)."
    - "config: correction.terminals absent on host (by-design D7 semantics) —
      terminal classification has no stable identity anchor on either machine."
    - "environment: host kernel 7.2.9-2-t2 focus churn — the direct driver of
      both the verifyDead wipes and the a11y/context instability; absent on the
      stock-kernel stand (proven: same code, same config semantics, green)."
    - "data/timing: 20 ms level2ForwardGap validated only on the stand; the
      host reorder class (2026-10-12 finding) is the same geometry at 0 ms."
  and_gate: >
    YES — mode B requires two contributing conditions simultaneously: the
    routing degradation (classification false at gesture time) AND the
    client-side commit loss (host-only delivery behavior). Mode A is
    single-cause (churn-wiped classification) but itself sits on the shared
    environment substrate.
```

hypothesis: CONFIRMED for mode A (journal-proven wipe + latency forensics); CONFIRMED for mode B routing; commit-loss bounded to host-client delivery with the reorder class leading (stand exonerations: flip, load).
test: complete (E1, E2b on stand; journal forensics on host; code reading).
expecting: n/a
next_action: return verdict (find_root_cause_only); no fix.

## Symptoms

expected: On the owner's host desktop, double right Shift in wezterm and GNOME Terminal converts the wrongly-typed word EN<->RU (as it does byte-exact on the VM stand, rows 19:17, 20:34, 21:17 on 2026-10-12, ADR-008).
actual: (A) wezterm: 2026-10-13 16:20:00 MSK action n=2 -> "terminal class","verify_dead":true -> "correction skipped","reason":"verify-timeout"; 16:20:33 action n=2 -> "correction skipped","reason":"verify-mismatch". Owner sees nothing happen. (B) GNOME Terminal (VTE): six "correction","outcome":"done","changed":true rows (16:20:12, 16:20:47, 16:21:37, 16:22:14, 16:22:18, 16:22:20), several with mode-switch + fresh engine creation; owner sees the old word deleted (backspace replay renders) but the committed replacement never appears.
errors: skip reasons verify-timeout / verify-mismatch; no error rows.
reproduction: owner's live desktop gesture (double right Shift) in wezterm / GNOME Terminal. Not reproducible on VM stand.
started: observed 2026-10-13 16:19-16:23 MSK host journal during owner UAT (10-UAT.md test 1, gaps G-10-1).

## Eliminated

- hypothesis: (prior phase-9 finding, 2026-10-12 probe file ~11:05) verify-timeout caused by app not providing surrounding text + mutter never answering RequireSurroundingText.
  evidence: that was the PRE-phase-10 app-verification path; phase 10 added the terminal branch with verify from the daemon's own stretch buffer (per code map). Yet the current 16:20:00 skip STILL says verify-timeout after terminal class verify_dead=true — so the dead/timeout now comes from the terminal-specific verify. Needs code reading to confirm which verify produced today's rows.
  timestamp: 2026-10-13 (carried as context, not yet confirmed for today's rows)
- hypothesis: the host config lacks correction.terminals and that alone is the stand/host divergence.
  evidence: the stand cases' own config doc (autocorrectConfigDoc, case_autocorrect.go:144) has NO terminals list either — stand and host run identical classification semantics; the divergence is the reliability of the sources (churn on host, stable on stand), not the config.
  timestamp: 2026-10-13
- hypothesis: the stand never exercises the focus-pair churn that wipes classification (verifyDead) — the stand is churn-free.
  evidence: ADR-008 venue-mechanics row: every ydotool call bounces IM focus producing focus_out/focus_in pairs — but on the stock kernel delivery stays healthy and the pairs do not land inside the critical windows; the host kernel defect produces the churn continuously and mid-flow. So the stand DOES emit pairs, but they are not destructive there.
  timestamp: 2026-10-13
- hypothesis (harness artifact, eliminated as product evidence): my first manual stand rounds showed keys traced by the daemon but never delivered to the VTE field (pty got only the newline, screen showed blank cells, a stray field with "x" answered the verify Require).
  evidence: the round lacked the case's focus gating (waitFocusInFrom) and the desktop carried stale surfaces; when the REAL case (term-word-vte) ran with its own gates on the same stand it PASSED byte-exact (bash/zsh/mid-line/clip-frozen rows green, 14:58). The traced-but-not-delivered signature is the TERM-01 stale-context mechanism reproduced by hand — contextually supporting, not causal for the host failures.
  timestamp: 2026-10-13

## Evidence

- timestamp: 2026-10-13 (pre-gathered by orchestrator)
  checked: host environment
  found: goswitchd active since 2026-10-13 00:19:41 MSK; ibus engine -> goswitch-ru; goswitchctl status corrections_done=9 corrections_skipped=2 skip_verify_mismatch=1 skip_verify_timeout=1 autocorrect_fired=0 autocorrect_abstained=44 ac_skip_terminal_veto=11 config_valid=true; wezterm nightly 20261010-083911-e2a98e1b with enable_wayland=true (text-input v3 path, NOT XIM).
  implication: daemon healthy, global engine ours; failures are in the terminal correction path, not registration.
- timestamp: 2026-10-13 (pre-gathered by orchestrator)
  checked: host journal focus churn
  found: pairs of focus_out/focus_in for the SAME engine 1-3 ms apart in bursts (16:20:05 four flips within 6 ms; also 16:19:48, 16:20:36/40, 16:21:18-26, 16:22:37). Host kernel 7.2.9-2-t2 has a known background-window keyboard defect; VM stand (stock 6.8.0) shows no churn.
  implication: prime suspect; correlation with skip/done rows is the first test.
- timestamp: 2026-10-13 (investigation)
  checked: full 12h host journal (msg histogram + all action/correction rows), ~/.config/goswitch/config.yaml, actor.go, appid.go, BuildPlan, both stand cases, ADR-008
  found: (a) host config has NO correction.terminals — by design (CONFIG.md D7: "absent key = classification not configured"), and the stand cases' config doc has none either; (b) 9 done rows all day — 8 are INSTANT (0.2-3 ms after action = level-1 Delete+Commit or selection commit; executeLevel2 for 1 rune costs >=40 ms due to 20 ms sleeps; the 6-rune row 16:20:47 took 145 ms); (c) only 16:20:47 routed terminal in VTE (role 60 answered); the 11 terminal-vetoes 16:20:53-16:21:20 prove role 60 DID answer during typing; (d) verifyDead learned true 16:20:00.334 was wiped false 16:20:05.478 by a REAL FocusOut (no flip in between, flipCredit==0); (e) role-query errors are SILENT today: the one-per-episode WARN was consumed 11:00:16 ("focused object role unknown") and no healthy autoConfirm ran after to close the episode (autocorrect_fired=0); (f) all 6 done rows are followed by mode flip + engine re-mint 1.8-2.6 ms after the commit; (g) both stand cases pin flip_after_correction=false — the owner's config has true; the stand never exercised the flip-after-commit sequence.
  implication: classification is churn-fragile on the host (mode A mechanism + VTE misroutes); the commit-loss has an untested stand/host config divergence (flip) and a load-sensitive pacing race (20 ms quantum).
- timestamp: 2026-10-13 (investigation)
  checked: goswitchctl status counters
  found: skip_verify_mismatch=1 skip_verify_timeout=1; ac_skip_terminal_veto=11; autocorrect_fired=0; corrections_done=9.
  implication: matches the journal rows 1:1; no hidden failures elsewhere.

## Resolution

root_cause: >
  MODE A (wezterm verify-timeout/verify-mismatch refusals): the host kernel's
  focus churn (7.2.9-2-t2) continuously emits REAL input-context FocusOut/FocusIn
  pairs, and classifyTerminal's only wezterm-visible source — the behavioral
  verifyDead class — is wiped by every one of them (internal/session/actor.go:900,
  unconditional setVerifyDead(false) in HandleLifecycle; journal-proven wipe
  16:20:00.334 -> 16:20:05.478 with flipCredit==0). A churn-wiped wezterm context
  takes the GTK app-verification path (caps bit announced, pending + Require)
  that wezterm can never satisfy (no surrounding pushes, mutter never answers) ->
  verify-timeout free-abort each time; the class the timeout itself teaches is
  wiped before the next gesture. The stand stays green because its stock kernel
  produces no churn, so the learned class persists and gesture 2 rides it.
  The 16:20:33 verify-mismatch is the same churn corrupting the verify inputs
  (a pushing client's fresh push not ending with the daemon's token+tail after
  flap-driven buffer/context resets) — the free-abort safety firing on
  churn-corrupted state.
  MODE B (VTE): two contributing causes (AND-gate fires). (1) Routing: the same
  churn makes classifyTerminal unreliable for VTE — role 60 answered during
  typing (11 terminal-vetoes) but non-terminal at 5 of 6 gesture instants (the
  instant done rows 0.2-3 ms are level-1/selection GTK paths; executeLevel2
  minimum is >=40 ms; only 16:20:47 at 145 ms took the terminal branch). With no
  correction.terminals on host or stand, only verifyDead (churn-wiped) and the
  live role query (churn-bounced a11y pair; errors silent after the 11:00:16
  WARN consumed the episode) remain — both churn-hostage. (2) Commit loss: the
  final CommitText does not survive to the host client's field while the deletion
  machinery renders — the same client-side commit-vs-queued-forwards reorder
  class documented live on the HOST on 2026-10-12 at 0 ms pacing; the 20 ms
  pacing quantum was sized and validated only on the stand. Stand experiments
  exonerate the alternative candidates: flip_after_correction=true (owner
  parity) is byte-exact green at idle (E1), and sustained 2/4-vCPU load stays
  byte-exact green (E2b) — the stand cannot reproduce the host's commit loss;
  the host-specific factor is the churn environment itself.
fix: (none — find_root_cause_only)
verification: >
  Mode A: direct journal evidence (verifyDead transition rows) + code path
  (actor.go:883-901, :1171-1191, :2507-2527) + stand/host config identity.
  Mode B routing: latency forensics on all 9 done rows + veto/flip-flop
  counters + silent-WARN episode reconstruction. Mode B commit loss: E1 (flip
  owner-parity on stand, row 1 green) + E2b (load on stand, all rows green) +
  the documented 2026-10-12 host reorder finding; final host-side
  disambiguation needs owner-assisted probes (listed).
files_changed: []
stand_mutations_log:
  - "rsync host tree (HEAD 7187887) -> stand mirror ~/goswitch; go build ->
    ~/go/bin/goswitchd; unit restarted (left ACTIVE)."
  - "E1: sed flip_after_correction false->true in mirror
    test/e2e/case_autocorrect.go (autocorrectConfigDoc), ran term-word-vte
    twice (14:58-15:02), then REVERTED the sed (verified)."
  - "E2b: stress-ng --cpu 2 120s during a pristine term-word-vte run (PASS)."
  - "Stand desktop: killed stale gnome-terminal-server/windows during manual-
    round debugging; re-applied the documented sources takeover
    ([('ibus','goswitch-en')]) after case-driven ibus restarts; ibus restart
    twice (case poisoned-name path + recovery). Unit daemon ACTIVE at close,
    global engine goswitch-en, mirror pristine."
host_mutations: "NONE — read-only (journalctl, goswitchctl status attempt,
  config read, ibus engine query); no injection, no windows, no restarts."

## E3 — Owner-assisted disambiguation probe (2026-10-13 18:33 MSK, host)

**Setup:** HEAD build with `-debug` (key tracing + surrounding-push rows) run as
transient user unit `goswitchd-debug-tmp`; release unit stopped for the duration
(~3 min); release v1.1.1-94 restored and verified after. Owner performed ONE
gesture in GNOME Terminal: typed `ghbdtn`, double right Shift.

**Trace (unit journal):**
- 18:33:07–08 keys g h b d t n (EN layout); 18:33:12 right Shift ×2 (0xffe2).
- 18:33:13.031 `action n:2` → **this instant took the TRUE terminal path**:
  `.176 correction done changed=true level=2 runes=6 source=ghbdtn result=привет latency_ms=144`
  (classification held — no churn wipe between typing and gesture this time).
- `.178 engine created goswitch/2 goswitch-ru` → `.184 switch_engine
  goswitch-ru` — the post-correction flip re-mints the engine object and
  switches the global engine **2–8 ms after the commit**.
- **`.192 surrounding push text="\n" cursor=0` — the field text is EMPTY 16 ms
  after the commit, before any further keystroke. «привет» is absent.**
- 18:33:17–19 owner typed `готово` + Enter; every push faithfully tracks each
  typed char (`г\n\n` → `готово\n\n`) — the surrounding-text channel works and
  would have shown «привет» if it had ever landed.

**Verdict (disambiguates B2):** the commit is **DROPPED, not late-erased**. All
six backspaces applied (source word gone), the CommitText never reached the
field. The dropping agent is the daemon's own post-commit flip: mode switch
re-mints the engine object and calls SetGlobalEngine within 2–8 ms of the
commit; under host focus churn the client's event loop processes the commit
late, after the context already switched to the fresh engine object, and GTK
discards the signal from the stale engine. Stand E1 (flip at idle) rendered
fine — flip adjacency alone is survivable; churn-delayed client processing is
the co-factor (AND-gate). Fix direction sharpens from "longer pacing" to
**delivery-order safety**: flip must not orphan an in-flight commit (defer
flip until a confirming push/keystroke, commit after flip, or reuse the engine
object) — pacing alone does not close this gap.

host_mutations: "temporary `-debug` transient unit (goswitchd-debug-tmp) for
~3 min with owner's consent; release unit restored and verified active;
debug binary removed from /tmp."

## Closure validation (G-10-1) — host-ревалидация (план 10-09, 2026-10-14)

**Установленная сборка (gap G-10-1 closure build):** v1.1.1-124-g7ace3fd
(HEAD 7ace3fd рабочего дерева, mise run ci зелёный перед установкой; оба корневых
фикса: 10-07 флап-стойкая классификация + identity-якорь [gnome-terminal, wezterm] +
ac_role_errors, 10-08 defer-until-confirmed доставка + stale-cache амендмент).
Датированный ряд установки: docs/adr/adr-008-spike-log.md, секция «Фаза 10
host-ревалидация G-10-1», ряд host-install-gap 2026-10-14 00:46.

**Фиксированный путь goswitchctl для ВСЕХ машинных проверок** (бинарник не на PATH;
резолв новейшим по mtime среди releases):
`/home/nil/.local/share/goswitch/releases/v1.1.1-124-g7ace3fd/goswitchctl`

**Откат по запросу владельца (назван ДО приёмки):**
`~/.local/share/goswitch/releases/v1.1.1/goswitchctl install`
(релизный каталог v1.1.1 на месте нетронутым — прецедент 10-06).

### Машинные предпроверки (read-only; 2026-10-14 00:46–00:49 MSK)

- unit: `systemctl --user is-active goswitchd` → **active**; ExecStart =
  /home/nil/.local/share/goswitch/releases/v1.1.1-124-g7ace3fd/goswitchd, старт
  00:46:57 MSK.
- Журнал поднятия (свежее, ошибок регистрации нет): config loaded → connected
  (ibus socket) → engine created goswitch-en → **component registered (engines:2)**
  → ctl service listening → engine reactivated.
- status по фикс. пути: парсится, **config_valid=true**, config_error= (пуст);
  счётчик role-ошибок из 10-07 присутствует: **ac_role_errors=0** — машинная
  видимость диагностики установлена (тихий эпизод 11:00:16 более не воспроизводим
  незаметно: каждая ошибка role-запроса эпизода считается, здоровый ответ обнуляет).
- autocorrect_fired=0 на момент установки — канонический негатив чист на старте;
  ac_skip_*-токены рендерятся только при ненулевых причинах (свежий счётчик — их
  нет; формат status 10-07, пиннут тестом ctlsvc).
- Host-мутации: ТОЛЬКО санкционированная установка (build из рабочего дерева →
  release-dir → goswitchctl install → рестарт unit'а). После неё — строго
  read-only: инъекций нет, окон не создавалось, чужие сервисы не перезапускались,
  ydotool не запускался. Жесты приёмки исполняет владелец (ниже).

### Re-acceptance владельца — инструкция (форма 10-UAT Current Test + флап-пункт)

Разрыв G-10-1 закрывается ТОЛЬКО ответом владельца — на живом столе, на
установленной gap-сборке:

1. **wezterm:** набрать «ghbdtn», двойной правый Shift → увидеть «привет».
2. **GNOME Terminal:** то же — набрать «ghbdtn», двойной правый Shift → «привет».
3. **Флап-бурст:** быстро пощёлкать фокус между окнами туда-сюда несколько раз и
   СРАЗУ выполнить жест в терминале — исправление должно сработать (проверка
   фикса 10-07 в её сценарии; воспроизводит журнальную тряску 16:20:05 — четыре
   пары за 6 мс).
4. **Середина фразы:** набрать слово в середине фразы и исправить жестом.
5. **Пунктуация:** «ytn?» + жест → «нет,».
6. **Молчание автокоррекции:** в терминалах во время всех ручных проб
   автокоррекция молчала (машинная проверка: `/home/nil/.local/share/goswitch/releases/v1.1.1-124-g7ace3fd/goswitchctl status`
   — autocorrect_fired не вырос).

Ответ владельца: **«approved»** или описание отклонений.
Откат по запросу владельца — немедленно: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install`.

### Журнальные ряды приёмки (снимаются ПОСЛЕ жестов владельца)

Статус: **ожидание жестов и вердикта владельца** (human-check; снимаются
`journalctl --user -u goswitchd` за окно приёмки). Критерии честных рядов:

- НЕТ skip-рядов **verify-timeout / verify-mismatch** (подписи отказов wezterm
  из Symptoms: 16:20:00 / 16:20:33);
- done-ряд несёт **changed=true** и НЕ сопровождается потерей: следующий
  surrounding-push трекает текст С исправленным словом (критерий E3 «не потерян» —
  пустой push `text="\n"` сразу после commit = провал фикса 10-08);
- флип-рубежи (mode / switch_engine) видны **СТРОГО ПОСЛЕ settlement-рубежей**
  (подтверждающий push / первая клавиша / grace 300 мс) — SetGlobalEngine между
  commit и settlement = E3-сигнатура = провал;
- при п.6 — счётчики: autocorrect_fired не вырос, ac_skip_terminal_veto
  задаокументирован.

**Вердикт владельца:** [ОЖИДАНИЕ — не синтезируется. Записывается дословно по
ответу владельца: «approved» → ряд закрытия G-10-1 (gap closed by owner
re-acceptance, дата); «rejected» с наблюдаемыми → честные журнальные ряды,
СТОП: никаких дальнейших правок кода вне нового цикла планирования (этот план
фиксирует свидетельства, не чинит), откат по запросу.]

Routing для верификатора: human-check маршрутизируется к владельцу через UAT на
end-of-phase (workflow.human_verify_mode=end-of-phase); автоматическое
исполнение/синтез вердикта запрещены — машинные проверки его не заменяют.


### Re-acceptance 2026-10-14 09:39–09:41 MSK — вердикт владельца: REJECTED (с наблюдаемыми)

Сборка v1.1.1-124-g7ace3fd (оба фикса 10-07+10-08). Результаты по пунктам
(дословно владельца):

1. **wezterm:** «набираю ghbdtn, жму double shift, вижу привет, и он удаляется».
2. **GNOME Terminal:** «тот же результат».
3. **Флап-бурст:** владелец выполнять отказался (skipped владельцем, не провален;
   в окне приёмки флап-тряска журнально и так отсутствует — см. ниже).
4. **Середина фразы:** «ghbdtn ghbdtn привет» — исправлено только последнее
   слово. Соответствует дизайну (пробег = токен у курсора + примыкающие руны;
   вся строка — Фаза 12 TERM-строка); рецепт середины: `ghbdtn ` + вручную
   набранное хвостовое слово текущей раскладкой, курсор в конце → жест правит
   только `ghbdtn`.
5. **Пунктуация:** «ytn?» → «нет,» — «работает».
6. **Молчание автокоррекции:** VERIFIED оркестратором (владелец: «проверь сам»):
   autocorrect_fired=0 за всю приёмку, abstained=42, ac_skip_terminal_veto=1.

Журнальные подписи жестов (journalctl --user -u goswitchd):

- 09:39:41.3167 `action n=2` → 09:39:41.3180 `correction done changed=true`
  (**+2,4 мс**) → 09:39:41.3181 `mode to ru` (**+0,03 мс после done**) →
  engine created /11 → focus_out/focus_in → 09:39:41.3266 `switch_engine`.
- 09:40:28.4678 `action n=2` → done changed=true (**+1,9 мс**) → mode to ru
  (+0,03 мс) → engine /15 → switch_engine.
- Эталон настоящего терминального прогона: 16:20:47 (145 мс). Оба жеста приёмки —
  **УРОВЕНЬ-1 (GTK-путь), терминальная ветка НЕ получила управления**: машина
  отложенного флипа (10-08) не армилась вовсе (мгновенный settleCorrectionFlip),
  чанкированный level-2 не запускался.
- НЕТ ни одного verify-timeout / verify-mismatch — прежние подписи отказов
  исчезли. Флап-пар в окне приёмки нет — отказ классификации воспроизвёлся и без
  churn.
- `ac_role_errors=47` (громкая диагностика 10-07 жива: эпизод 11:00:16 более
  незамечаемым не проходит), `config_valid=true`.

Интерпретация (вход следующего цикла планирования; этим планом код не правится):

1. **Классификация терминала на хосте не включается**: a11y-идентичность
   недоступна (role-unknown, ac_role_errors растёт), поведенческий verifyDead-класс
   не выучен к моменту жеста, дефолтный якорь correction.terminals остаётся без
   источника identity → жест уходит на уровень-1, где оба фикса 10-07/10-08
   физически не стоят (defer-until-confirmed armится только в executeLevel2).
2. **«Появляется и удаляется» = инверсия порядка delete-surrounding/commit на
   уровне-1 в хостовой среде** (класс переупорядочивания A2: вставка доезжает до
   клиента первой и видна, запаздывающее удаление стирает свежевставленное слово).
   Это та же видимая симптоматика, что исходная «конвертация, потом исчезает»,
   теперь с уровня-1.

**Вердикт:** REJECTED (п.1/п.2). G-10-1 остаётся открытым. СТОП без правок кода
вне нового цикла планирования; следующий цикл обязан реплеем этих журнальных
рядов доказать фикс ДО запуска владельцем (урок 2026-10-12). Откат по запросу
владельца: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install` (на
момент записи не запрашивался, сборка оставлена активной).

## Closure validation round-2 (G-10-1) — host-ревалидация (план 10-12, 2026-10-14)

Против блока **REJECTED 09:39** (выше). Урок 2026-10-12/09:39 исполнен до запуска
владельцем: реплей-гейт реальных рядов 09:39 (юнит-корпус ReplayHost0939*/TwoPhaseLevel1*)
зелёный на HEAD **до** стенда и установки; стенд не регрессировал (четыре терминальных
кейса байт-точные зелёные — секция «Фаза 10 re-pin round-2 G-10-1» в adr-008-spike-log.md).

**Установленная сборка (round-2 G-10-1 closure build):** v1.1.1-142-g2fbfee1
(HEAD 2fbfee1 рабочего дерева, mise run ci зелёный перед установкой; обе половины
round-2: 10-10 answer-only маршрут терминального контекста на ОБЕИХ ветках вердикта
HandleSurroundingText + WR-01 correction-busy + WR-02 epoch; 10-11 двухфазный уровень-1
— delete → подтверждающий рубеж → commit, фолбэк по бюджету 60 мс — + отложенный флип
уровня-1 на машине 10-08 + WR-03 same-context флап-дисциплина).
Датированный ряд установки: docs/adr/adr-008-spike-log.md, ряд host-install-gap-round2
2026-10-14 12:55 MSK.

**Фиксированный путь goswitchctl для ВСЕХ машинных проверок round-2** (бинарник не на
PATH; резолв новейшим по mtime среди releases):
`/home/nil/.local/share/goswitch/releases/v1.1.1-142-g2fbfee1/goswitchctl`

**Откат по запросу владельца (назван ДО приёмки):**
`~/.local/share/goswitch/releases/v1.1.1/goswitchctl install`
(релизный каталог v1.1.1 нетронут; прежний release-dir v1.1.1-124-g7ace3fd не
перезаписывался). Исполнение — немедленно по запросу.

### Машинные предпроверки (read-only; 2026-10-14 12:55–12:58 MSK)

- unit: `systemctl --user is-active goswitchd` → **active**; ExecStart =
  /home/nil/.local/share/goswitch/releases/v1.1.1-142-g2fbfee1/goswitchd, старт
  12:55:53 MSK (инсталлятор выполнил daemon-reload + restart сам).
- Журнал поднятия (ошибок регистрации нет): connected → ctl service listening →
  **component registered (engines:2)** → engine created goswitch-en/1, /2.
- status по фикс. пути: парсится, **config_valid=true**, config_error= (пуст);
  счётчики на нуле — чистая база приёмки: autocorrect_fired=0, autocorrect_abstained=0,
  corrections_done=0, ac_role_errors=0.
- Host-мутации: ТОЛЬКО санкционированная установка (build из рабочего дерева →
  НОВЫЙ release-dir → goswitchctl install; рестарт unit'а выполнен инсталлятором).
  После неё — строго read-only: инъекций нет, окон не создавалось, чужие сервисы
  не перезапускались, ydotool не запускался. Жесты приёмки исполняет владелец (ниже).

### Критерии журнальных рядов round-2 (сверка ПОСЛЕ жестов владельца)

Окно приёмки снимается `journalctl --user -u goswitchd` и сверяется:

- НЕТ skip-рядов **verify-timeout / verify-mismatch** (подписи отказов round-1:
  16:20:00 / 16:20:33; в блоке 09:39 их уже не было — требование сохраняется);
- **НЕТ инверсии класса A2** («вижу привет, и он удаляется» — дословные ряды
  владельца 09:39 в ОБОИХ терминалах): done-ряд несёт **changed=true**, и первый
  surrounding-push ПОСЛЕ коррекции трекает текст С исправленным словом — НЕ пустоту
  (`text="\n"` = сигнатура E3-потери) и НЕ старое слово (инверсия delete/commit);
  на уровне-1 двухфазность делает инверсию невоспроизводимой по построению (commit
  не идёт в провод одновременно с delete — подтверждающий рубеж), фолбэк по бюджету
  60 мс несёт громкий INFO-слаг «correction confirm timeout»;
- флип-рубежи (mode / switch_engine) видны **СТРОГО ПОСЛЕ settlement-рубежей**
  (подтверждающий push / первая клавиша / grace 300 мс) — SetGlobalEngine между
  commit и settlement = E3-сигнатура = провал; мгновенный «mode через 0,03 мс после
  done» (форма 09:39) больше невозможен — defer-until-confirmed армится на обеих
  ветках записи (терминальная executeLevel2 и level-1 flush);
- уровень коррекции: answer-only клиент (терминалы не пушат спонтанно) маршрутизируется
  в терминальную ветку (10-10) — терминальная геометрия executeLevel2 (латентность
  ≥40 мс); клиент с кэш-свидетельством идёт двухфазным уровнем-1 (10-11) — либо путь
  честен, инверсия и потеря исключены на обоих; Info-ряды демона без содержимого
  (D-20/D-21) — исходы/счётчики/латентности;
- при п.6 — счётчики: autocorrect_fired не вырос (машинная проверка по фикс. пути),
  ac_role_errors/ac_skip_terminal_veto задокументированы.

### Re-acceptance round-2 владельца — инструкция (форма 10-UAT Current Test, шесть пунктов)

Разрыв G-10-1 закрывается ТОЛЬКО ответом владельца — на живом столе, на
установленной сборке round-2 v1.1.1-142-g2fbfee1:

1. **wezterm:** набрать «ghbdtn», двойной правый Shift → увидеть «привет» —
   слово ОСТАЁТСЯ (не «появляется и удаляется» — дословный отказ 09:39).
2. **GNOME Terminal:** то же — «ghbdtn» + двойной правый Shift → «привет», слово остаётся.
3. **Флап-бурст:** быстро пощёлкать фокус между окнами туда-сюда несколько раз и
   СРАЗУ выполнить жест в терминале — исправление срабатывает (в 09:39 пункт был
   пропущен владельцем; флап-тряска в окне приёмки отсутствовала — отказ
   воспроизвёлся и без churn).
4. **Середина фразы:** набрать слово в середине фразы и исправить жестом (рецепт
   середины из 09:39: «ghbdtn » + хвостовое слово текущей раскладкой, курсор в конце
   — жест правит пробег у курсора; вся строка — Фаза 12).
5. **Пунктуация:** «ytn?» + жест → «нет,».
6. **Молчание автокоррекции:** в терминалах во время всех ручных проб автокоррекция
   молчала (машинная проверка:
   `/home/nil/.local/share/goswitch/releases/v1.1.1-142-g2fbfee1/goswitchctl status`
   — autocorrect_fired не вырос).

Ответ владельца: **«approved»** или описание отклонений.
Откат по запросу владельца — немедленно: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install`.

### Журнальные ряды приёмки round-2 (снимаются ПОСЛЕ жестов владельца)

Статус: **ОЖИДАНИЕ — жесты и вердикт владельца** (human-check; не синтезируется).
После жестов снять ряды за окно приёмки и сверить с критериями round-2 выше
(нет verify-timeout/verify-mismatch; нет инверсии — первый push трекает исправленное
слово; флипы строго после settlement). Записать ряды и вердикт ДОСЛОВНО:

- «approved» → ряд закрытия G-10-1 (gap closed by owner re-acceptance round-2, дата);
- «rejected» с наблюдаемыми → честные журнальные ряды, СТОП без правок кода вне
  нового цикла планирования (следующий цикл обязан реплеем НОВЫХ рядов доказать
  фикс до owner-запуска); откат по запросу владельца — немедленно.

**Вердикт владельца:** [ОЖИДАНИЕ — не синтезируется]

### Доставка инструкции round-2 (2026-10-14 13:00 MSK, план 10-12 Task 3)

Шестипунктовая инструкция re-acceptance round-2 (выше) доставлена владельцу вместе
с фиксированным путём goswitchctl для машинной проверки п.6
(`/home/nil/.local/share/goswitch/releases/v1.1.1-142-g2fbfee1/goswitchctl status`)
и путём отката (`~/.local/share/goswitch/releases/v1.1.1/goswitchctl install` —
исполнение немедленно по запросу). База приёмки чистая: счётчики демона на нуле с
момента установки 12:55 MSK (autocorrect_fired=0, corrections_done=0, ac_role_errors=0).

Статус на момент записи: **жесты владельцем НЕ исполнены, вердикт НЕ дан** —
журнальные ряды приёмки отсутствуют, записывать нечего (честное ОЖИДАНИЕ; вердикт
будет дописан дословно по ответу владельца в этот блок: «approved» → ряд закрытия
G-10-1 (gap closed by owner re-acceptance round-2, дата); «rejected» с наблюдаемыми →
честные ряды, СТОП без правок кода вне нового цикла планирования, откат по запросу).
Автоматический grep-гейт `approved|rejected` технически матчится с ТЕКСТОМ ИНСТРУКЦИИ
и ОЖИДАНИЯ — записанного вердикта НЕТ (прецедент 10-09 D4); routing: UAT end-of-phase.

Routing для верификатора: human-check round-2 маршрутизируется к владельцу через UAT
end-of-phase (workflow.human_verify_mode=end-of-phase); автоматическое исполнение/
синтез вердикта запрещены — машинные проверки его не заменяют.
