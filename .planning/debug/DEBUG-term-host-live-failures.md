---
status: investigating
trigger: "Phase-10 terminal word correction (double right Shift, EN<->RU) is byte-exact green on the VM stand but fails on the owner's host desktop in two modes: (A) wezterm — daemon self-aborts via verify-timeout (after terminal class verify_dead=true) and verify-mismatch, owner sees nothing; (B) GNOME Terminal (VTE) — daemon reports correction done changed=true (6 times), owner sees backspaces render but the committed replacement word never appears (final CommitText lost). Mode: find_root_cause_only. Host daemon is the owner's live input method — read-only host access only; VM stand experiments allowed with logging."
created: 2026-10-13T00:00:00+03:00
updated: 2026-10-17T00:54:00+03:00
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

### Re-acceptance round-2, 2026-10-14 19:21 MSK — вердикт владельца: REJECTED («поведение в wezterm НЕ изменилось»)

Сборка v1.1.1-142-g2fbfee1 (PID 373321, старт 12:55:53, exe = releases/v1.1.1-142-g2fbfee1/goswitchd — идентичность проверена по /proc). Жест владельца в wezterm:

- 19:21:26.732 `action n=2` → **19:21:26.876 `correction done changed=true` (+144 мс = ТЕРМИНАЛЬНАЯ ВЕТКА level-2**; вчера тот же жест — 2 мс уровень-1; эталон терминального 145 мс) → **19:21:26.980 `mode to ru` (+104 мс после done = отложенный флип** освобождён подтверждением/клавишей, НЕ мгновенно) → engine /7 → focus-пары → switch_engine.
- 19:21:28.746 `autocorrect skipped reason=terminal-veto` — **выученный класс терминала жив на хосте впервые** (10-07 флап-стойкое обучение работает); ac_skip_terminal_veto=1 в status.
- Работает строго round-2 код: маршрут answer-only (10-10) и отложенный флип (10-08/10-11) на хосте подтверждены журналом.

**Видимый результат НЕ изменился (владелец дословно: «поведение в wezterm НЕ изменилось»)** — слово по-прежнему пропадает после коррекции.

**Интерпретация (вход цикла №3): потеря сместилась на доставку пакета level-2.** Демон теперь делает всё верно (терминальная ветка, отложенный флип, выученный класс), но пакет «6×BackSpace-forward + commit» с темпом 20 мс wezterm под хостовой задержкой обрабатывает с перестановкой: вставка обгоняет часть форвардов — слово видно, запаздывающие удаления съедают его (класс A2 внутри level-2; на стенде темп 20 мс байт-точен, хост невоспроизводим — доказано E1/E2b). Верификация-после-вставки слепа: answer-only клиент не пушит состояние поля сам. Кандидаты цикла №3: двухфазная дисциплина для самих чанков терминальной записи (delete-чанк → доказательство через Require-ответ, который клиент ДАЁТ → следующий чанк), либо уход от форвард-бекспейсов к delete_surrounding_text там, где клиент его поддерживает (wezterm 0x29 отвечает на Require — поддержку DeleteSurroundingText проверить на стенде), либо комбинированная стратегия. Требование реплея рядов сохраняется; новая подпись для реплея: done +144 мс, флип +104 мс, terminal-veto +2 с, ноль verify-строк.

**Вердикт:** REJECTED (раунд 2). G-10-1 открыт; СТОП без правок кода вне нового цикла планирования. Сборка оставлена активной (машинное поведение корректно; видимый дефект тот же, что в v1.1.1 — регресса нет). Откат по запросу: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install`.

### Дополнение round-2, 2026-10-14 19:28 MSK — ТРОЙНОЙ жест (n:3) впервые дошёл до результата на хосте; вставка УДЕРЖАЛАСЬ

Владелец протестировал рецепт «середины фразы», но нажал ТРОЙНОЕ переключение: `action n=3` в 19:28:41.197 (строка «ghbdtn test», вся набрана в EN после 19:28:36 combo → en). Результат владельца: «ghbdtn test превращается в привет еуые» — ФРАЗОВАЯ коррекция конвертировала всю строку EN→RU (ghbdtn→привет, test→еуые — все буквы test имеют табличные позиции) и — ГЛАВНОЕ — **результат удержался на экране, потери не было**.

Журнальная подпись: action n=3 → done changed=true **+43 мс** (двухфазный уровень-1, бюджет 60 мс — удаление доказано ДО вставки) → mode to ru +40 мс после done (отложенный флип). Ноль verify-timeout/mismatch/skip-строк.

**Значение:** (1) 10-11 двухфазная доставка уровня-1 ПОДТВЕРЖДЕНА на хосте живым рядом владельца — инверсия delete/commit на уровне-1 устранена в бою. (2) Фразовый путь (Triple, scope Фазы 12 TERM-строка) впервые довёл результат до экрана wezterm на хосте — feasibility-сигнал для Фазы 12 существенно улучшился (раньше фраза умирала в verify-timeout, 10:00:00 2026-10-12). (3) Остаётся сломанным ТОЛЬКО путь слова через терминальную ветку level-2 (двойной жест, 144 мс, слово исчезает — блок 19:21): темпированный burst 20 мс без доказательства доставки. Цикл №3 сужается до: доказуемая доставка чанков level-2 (двухфазность по Require-ответам ИЛИ delete_surrounding_text).

Замечание к рецептам: рецепт «середины» в UAT был двусмысленным (не указан жест; «test» в русской раскладке невозможен — даёт «еуые»). UAT-формулировка исправлена: двойной Shift + русский хвост.

### Коррекция свидетельств round-2, 2026-10-14 ~19:45 MSK — владелец: «результат НЕ остался»; тест фразы был в ZCode, не в терминале

Поправка владельца к блоку 19:28: «результат остался на экране — нет, не остался. я делал эту проверку в zcode, а не в терминале». Следствия:

1. **Отзываю вывод «вставка удержалась»**: фразовая коррекция в ZCode (НЕ терминальный контекст) тоже потеряла результат. Двухфазный уровень-1 (delete → доказательство → commit, 43 мс) на хосте результат НЕ спас — потеря доставки шире терминальной ветки.
2. **Дефект, который надо решить (мандат владельца, дословно): «корректировка по двойному шифту в wezterm срабатывает, я вижу исправленное слово, после чего оно удаляется. эту проблему надо решить»** — приоритет №1, «почини мне переключение в wezterm для начала, потом поговорим про все остальное» (под «переключением» владелец имел в виду коррекцию переключённого слова).
3. ZCode-потеря фразы — отдельный симптом, в цикл №3 НЕ входит (после wezterm).

Рабочая гипотеза №1 для цикла №3 (уточнится диагностикой): терминальная вставка = пакет «N×BackSpace-форвард (темп 20 мс) + commit»; commit клиент применяет мгновенно, форварды застревают в очереди клиента под хостовой задержкой и доезжают ПОСЛЕ вставки, стирая её. Кандидаты фикса: DeleteSurroundingText вместо форвард-бекспейсов (один упорядоченный IM-канал с commit — поддерживает ли wezterm-nightly, проверить на стенде), почанковая двухфазность по Require-ответам, разбор взаимодействия с отложенным флипом. Реплей-подписи для гейтов: 19:21:26 (done +144 мс, флип +104 мс) и 19:36:00 (done +76 мс).

## Closure validation round-4 (G-10-1) — host-ревалидация (план 10-18, 2026-10-16/17)

Против блоков **19:21:26 / 19:36:00** (потеря сместилась на доставку пакета level-2:
перестановка «вставка обгоняет запаздывающие BackSpace-форварды» под хостовой задержкой
wezterm) и бисекции 2026-10-16 (DST — compositor-killer: Require-рефлексия для
wezterm-класса мертва, форварды — единственный write-примитив). Порядок раунда исполнен
дословно (урок 2026-10-12/09:39): реплей-гейт (ряд repin-round4 21:32–21:38 UTC, spike-log)
→ стенд re-pin ×4 + живой drain-ряд (repin-round4-stand 21:39–21:48 UTC) → установка.

**Установленная сборка (round-4 G-10-1 closure build):** v1.1.1-180-gefac750
(HEAD efac750 рабочего дерева; машина round-4 10-17: drain-settle 150 мс перед коммитом
терминальной ветки при освобождённом мьютексе, флип вне окна — терминальный слот без
grace и verify-релизов, delivery-метки INFO, abort/commit-дисциплина «свободный стоп
только до первого форварда»; гвард 10-16: DST только при свидетельстве материализованности,
счётчик corr_dst_guarded). Датированный ряд установки: docs/adr/adr-008-spike-log.md,
ряд host-install-gap-round4 2026-10-17 00:51 MSK.

**Фиксированный путь goswitchctl для ВСЕХ машинных проверок round-4** (бинарник не на
PATH; резолв новейшим по mtime среди releases):
`/home/nil/.local/share/goswitch/releases/v1.1.1-180-gefac750/goswitchctl`

**Откат по запросу владельца (назван ДО приёмки):**
`~/.local/share/goswitch/releases/v1.1.1/goswitchctl install`
(релизный каталог v1.1.1 нетронут — проверено test -x и mtime Oct 8; прежние release-dir
v1.1.1-124-g7ace3fd / v1.1.1-142-g2fbfee1 не перезаписывались). Исполнение — немедленно
по запросу.

### Машинные предпроверки (read-only; 2026-10-17 00:51–00:53 MSK)

- unit: `systemctl --user is-active goswitchd` → **active**; ExecStart =
  /home/nil/.local/share/goswitch/releases/v1.1.1-180-gefac750/goswitchd, старт
  00:51:27 MSK (инсталлятор выполнил daemon-reload + enable + restart сам);
  **/proc/2670465/exe → тот же release-dir** — идентичность по /proc сверена.
- Журнал поднятия (ошибок регистрации нет): config loaded → connected → ctl service
  listening → **component registered (engines:2)** → engine created ×2 → engine
  reactivated goswitch-en.
- status по фикс. пути: парсится, **config_valid=true**, config_error= (пуст); счётчики
  на нуле — чистая база приёмки: corrections_done=0, corrections_skipped=0,
  autocorrect_fired=0, autocorrect_abstained=0, ac_role_errors=0; стартовое
  **corr_dst_guarded=0** (счётчик гварда 10-16 присутствует в установленной сборке —
  задокументирован).
- Host-мутации: ТОЛЬКО санкционированная установка (build из рабочего дерева → НОВЫЙ
  release-dir → goswitchctl install; рестарт unit'а выполнен инсталлятором). После неё —
  строго read-only: инъекций нет, окон не создавалось, чужие сервисы не перезапускались,
  ydotool не запускался, -debug-юнитов нет (клавиши владельца не инструментируются —
  T-10-18-01). Жесты приёмки исполняет владелец (ниже).

### Критерии журнальных рядов round-4 (сверка ПОСЛЕ жестов владельца)

Окно приёмки снимается `journalctl --user -u goswitchd` (production INFO-метки —
штатная наблюдаемость, не -debug) и сверяется:

- (а) НЕТ skip-рядов **verify-timeout / verify-mismatch** (подписи отказов round-1:
  16:20:00 / 16:20:33; в раундах 2-3 их не было — требование сохраняется);
- (б) delivery-метки **присутствуют и упорядочены**: `chunks-sent` → `commit-sent`;
  done-ряд несёт **changed=true**;
- (в) флип-рубежи (mode / switch_engine) — **СТРОГО ПОСЛЕ release-причины**
  (flip-released с причиной push/keystroke: следующая клавиша / подтверждающий пуш);
  **SetGlobalEngine между коммитом и release-причиной = провал** (терминальный слот
  round-4 не имеет grace- и verify-релизов — мгновенный «mode через 0,03 мс» формы 09:39
  и «mode +104 мс от verify» формы 19:21 оба невозможны по построению);
- (г) НЕТ крахов/перезапусков gnome-shell в окне приёмки (гвард 10-16 в установленной
  сборке; **крах/перезапуск shell = честный провал безопасности** с STOP и разбором —
  форма 10-13);
- (д) autocorrect_fired **не вырос** (машинная проверка по фикс. пути);
- (е) **absent-метки/тишина в живом ряду = честный провал фикса**, никогда не пропуск
  (T-10-18-06).

### Re-acceptance round-4 владельца — инструкция (гейт ОДИН + опциональные пункты)

Разрыв G-10-1 закрывается ТОЛЬКО ответом владельца — на живом столе, на установленной
сборке round-4 v1.1.1-180-gefac750. Мандат: «корректировка по двойному шифту в wezterm…
эту проблему надо решить», «потом поговорим про все остальное».

**ГЕЙТ (один вопрос):**
1. **wezterm:** набрать «ghbdtn», двойной правый Shift → «привет» — и слово
   **ОСТАЁТСЯ** на экране (не «появляется и удаляется» — дословный отказ 19:21).

**Остальные пункты — по желанию владельца, всё записывается:**

2. GNOME Terminal: то же — «ghbdtn» + двойной правый Shift → «привет», слово остаётся.
3. Флап-бурст: быстро пощёлкать фокус между окнами и СРАЗУ жест в терминале.
4. Середина фразы: «ghbdtn » + хвостовое слово текущей раскладкой, курсор в конце →
   двойной Shift правит пробег у курсора (вся строка — Фаза 12).
5. Пунктуация: «ytn?» + жест → «нет,».
6. Молчание автокоррекции (машинная проверка оркестратора по фикс. пути:
   `/home/nil/.local/share/goswitch/releases/v1.1.1-180-gefac750/goswitchctl status` —
   autocorrect_fired не вырос).

**ЯВНО предъявлено владельцу (прозрачность, не прятать):**

- **(А) Суженная семантика флипа:** после коррекции режим переключается на
  **СЛЕДУЮЩЕМ нажатии клавиши** (или подтверждающем пуше клиента), НЕ мгновенно —
  первая клавиша после жеста уже печатает в НОВОМ режиме. Правило владельца
  2026-09-28 «режим следует за скриптом результата» ОТЛОЖЕНО, НЕ отменено: его
  вердикт по этой семантике валидирует сужение; отдельное несогласие фиксируется
  отдельным наблюдением вердикта (release-сайты правятся только новым циклом).
- **(Б) Честный стек доказательств:** машинного чтения поля хоста НЕ существует
  (Require-рефлексия мертва для wezterm-класса — бисекция 2026-10-16; -debug-юниты
  запрещены — клавиши владельца не трейсится). Персистентность слова доказывается
  СТОПКОЙ: (1) стенд pty-drain под явной матрицей pacing (ряды repin-round4-stand:
  каждый 0x7f строго прежде байтов слова при каждой комбинации) + (2) хостовые
  журнальные delivery-метки (исходы/порядок/латентности без содержимого) + (3) финальный
  жест владельца. Машинные проверки его не заменяют.

Ответ владельца: **«approved»** или описание отклонений.
Откат по запросу владельца — немедленно: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install`.

### Журнальные ряды приёмки round-4 (снимаются ПОСЛЕ жестов владельца)

Статус: **ОЖИДАНИЕ — жесты и вердикт владельца** (human-check; не синтезируется).
После жестов снять ряды за окно приёмки и сверить с критериями (а)-(е) выше; absent-метки
и тишина — честный провал фикса (критерий (е)). Записать ряды и вердикт ДОСЛОВНО:

- «approved» → ряд закрытия G-10-1 (gap closed by owner re-acceptance round-4, дата)
  в этом блоке, spike-log и UAT;
- «rejected» с наблюдаемыми → честные журнальные ряды, **СТОП без правок кода вне нового
  цикла планирования** (следующий цикл обязан реплеем НОВЫХ рядов доказать фикс до
  owner-запуска); откат по запросу владельца — немедленно.

**Вердикт владельца:** [ОЖИДАНИЕ — не синтезируется]

Routing для верификатора: human-check round-4 маршрутизируется к владельцу через UAT
end-of-phase (workflow.human_verify_mode=end-of-phase); автоматическое исполнение/
синтез вердикта запрещены — машинные проверки его не заменяют. Автоматический
grep-гейт `approved|rejected` технически матчится с ТЕКСТОМ ИНСТРУКЦИИ и ОЖИДАНИЯ —
записанного вердикта НЕТ (прецеденты 10-09 D4 / 10-12 13:00).

### Доставка инструкции round-4 (2026-10-17 00:54 MSK, план 10-18 Task 3)

Инструкция round-4 (гейт: слово в wezterm ОСТАЁТСЯ; опциональные пункты; суженная
семантика флипа и честный стек доказательств предъявлены явно) доставлена владельцу
вместе с фикс. путём goswitchctl для машинной проверки п.6 и путём отката. База приёмки
чистая: счётчики демона на нуле с момента установки 00:51 MSK. Статус на момент записи:
**жесты владельцем НЕ исполнены, вердикт НЕ дан** — журнальные ряды приёмки отсутствуют,
записывать нечего (честное ОЖИДАНИЕ; вердикт будет дописан дословно по ответу владельца
в этот блок).

### Round-4 addendum: ревью-фиксы CR-01/CR-02 (2026-10-17 01:56 MSK, fix-сессия 10-17)

Fresh code review round-3+4 дельты (10-REVIEW, 2026-10-16 22:15 UTC, delta ab24c4c..HEAD)
нашёл два БЛОКЕРА reentrancy drain-settle окна (общий корень: busy-гвард покрывал
вторую коррекцию и lifecycle, но не собственные эмиссии ввода/флипов посреди окна) и
три дешёвых находки. Все пять исправлены в fix-сессии строго по находкам, TDD
(RED-evidence files evidence-20261017-10-17-red-{cr01,cr02,wr07,wr08}.json — все
RED_EVIDENCE_OK на неизменённом HEAD, GREEN после фиксов):

- **CR-01 (a0969fd)** — жестовый flipTo внутри drain-окна исполнял SetGlobalEngine
  между форвардами и ещё не отправленным коммитом (E3-класс: «удаление без замены»).
  Фикс: при поднятом level2Busy flipTo паркует флип в существующий слот
  отложенного флипа с терминальной дисциплиной релиза (следующая клавиша /
  подтверждающий пуш; не гейтится flip_after_correction — жест это явная команда);
  пост-коррекционное армирование не вытесняет занятый жестом слот. Ячейка (α)
  корпуса: combo в окне → ноль SetGlobalEngine/mode-записей до коммита, метка
  flip-parked, релиз строго после коммита.
- **CR-02 (0d6e74c)** — клавиши, обработанные в незалоченном окне, ложились на
  провод РАНЬШЕ коммита (детерминированная инверсия «xпривет » вместо «привет x»).
  Фикс: typed-layout ввод (printables + BackSpace — ровно то, что трекает буфер)
  паркуется на время машины, обрабатывается штатно после сеттла, транзиты
  реэмиттятся демоном (вердикт прибытия consumed); порядок провода = deletes →
  commit → held keys, FIFO; ожидание bounded (settle + cap×gap + квант слэка —
  никогда не вечность, после границы громкая деградация, клавиши не теряются).
  Жесты/модификаторы/reset-клавиши не парятся (их семантика — на мгновении
  прибытия; зафиксированный busy-отказ combo и парковка CR-01 совместны).
- **WR-06 (e9bdc5f)** — чтение terminalDrainSettle перенесено под мьютекс
  (дисциплина поля WR-01-класса).
- **WR-07 (2166c64)** — ProbeForward получил nil-гард ErrProbeNoEngine —
  fail-closed контракт семейства пяти вербов выровнен.
- **WR-08 (29bbde6)** — cause-поиск оракула drain-кейса срезан к сегменту текущего
  раунда (хвост от flip-armed): чужой ряд больше не может закрыть текущий.

Гейты fix-сессии: `mise run ci` зелёный на КОЖДОМ коммите; регрессионный пандус
round-4 (ReplayHost0939 / TwoPhaseLevel1 / FlapWindowContextIdentity /
Level1ConfirmFlush / TestDstGuard / ReplayWezterm1921 / ReplayHost1936 + новые
ячейки α/β) зелёный под -race (23 пина); go.mod/go.sum заморожены (tidy-diff,
diff ab24c4c..HEAD пуст); go vet ./... чисто. Суженные семантики round-4 не
ослаблены: flip_after_correction=false байт-стабилен, D-55 шов един,
fail-closed non-terminal диверсии и WR-01/02/03 пины на месте (полный пандус).

**Установленная сборка (round-4 review-fix build):** v1.1.1-190-g29bbde6
(HEAD 29bbde6 рабочего дерева = round-4 машина 10-17 + гвард 10-16 + все пять
ревью-фиксов выше). Установка 2026-10-17 01:56:12 MSK санкционированным
процессом: build из рабочего дерева → НОВЫЙ release-dir
`/home/nil/.local/share/goswitch/releases/v1.1.1-190-g29bbde6/` → его
goswitchctl install (daemon-reload + enable + restart инсталлятором).

**Фиксированный путь goswitchctl для ВСЕХ машинных проверок round-4 (актуален):**
`/home/nil/.local/share/goswitch/releases/v1.1.1-190-g29bbde6/goswitchctl`

**Машинные предпроверки (read-only; 2026-10-17 01:56 MSK):**

- unit: `systemctl --user is-active goswitchd` → **active**; ExecStart =
  /home/nil/.local/share/goswitch/releases/v1.1.1-190-g29bbde6/goswitchd,
  старт 01:56:12 MSK; **/proc/2958787/exe → тот же release-dir** — идентичность
  по /proc сверена.
- Журнал поднятия (ошибок регистрации нет): config loaded → ctl service
  listening → connected → engine created → **component registered (engines:2)**
  → engine reactivated goswitch-en. (Два WARN в окне рестарта — гаснущее
  соединение СТАРОГО демона v1.1.1 при остановке юнита и известная деградация
  «tray indicator attach failed: no StatusNotifierWatcher»; к регистрации
  отношения не имеют.)
- status по фикс. пути: парсится, **config_valid=true**, config_error= (пуст);
  счётчики на нуле — чистая база приёмки: corrections_done=0,
  corrections_skipped=0, autocorrect_fired=0, autocorrect_abstained=0,
  ac_role_errors=0, corr_dst_guarded=0.
- Host-мутации: ТОЛЬКО санкционированная установка (рестарт unit'а выполнен
  инсталлятором). После неё — строго read-only. Жесты приёмки исполняет
  владелец.

**Инструкция re-acceptance владельца — ОБНОВЛЕНА (тот же ОДИН гейт, новая сборка):**

ГЕЙТ (один вопрос), на сборке **v1.1.1-190-g29bbde6**:
1. **wezterm:** набрать «ghbdtn», двойной правый Shift → «привет» — и слово
   **ОСТАЁТСЯ** на экране (не «появляется и удаляется» — дословный отказ 19:21).

Остальные пункты — по желанию владельца, всё записывается (без изменений к
инструкции round-4): 2. GNOME Terminal — то же; 3. флап-бурст; 4. середина
фразы; 5. «ytn?» + жест → «нет,»; 6. автокоррекция молчала (машинно:
`/home/nil/.local/share/goswitch/releases/v1.1.1-190-g29bbde6/goswitchctl status`
— autocorrect_fired не вырос).

**ЯВНО предъявлено владельцу (не прятать):**

- **(А) Суженная семантика флипа — с round-4 review-fix уточнением:** после
  коррекции режим переключается на **СЛЕДУЮЩЕМ нажатии клавиши** (или
  подтверждающем пуше), НЕ мгновенно — первая клавиша печатает уже в НОВОМ
  режиме. НОВОЕ в этой сборке: жест переключения режима, нажатый ВО ВРЕМЯ
  доставки коррекции, больше не рвёт её — режим переключается СРАЗУ ПОСЛЕ
  того, как исправление легло в поле (жест не теряется, «удаление без
  замены» и инверсия набора исключены по построению). Правило владельца
  2026-09-28 «режим следует за скриптом результата» отложено, не отменено.
- **(Б) Честный стек доказательств — без изменений:** машинного чтения поля
  хоста нет; персистентность доказывается стопкой (стенд pty-drain + журнальные
  delivery-метки + финальный жест владельца). К стопке добавлены юнит-ячейки
  (α)/(β) реплей-корпуса: жест и набор ПОСЛЕ жеста внутри окна доставки
  воспроизведены на машинах и зеленые на этой сборке.

Ответ владельца: **«approved»** или описание отклонений.
Откат по запросу владельца — немедленно:
`~/.local/share/goswitch/releases/v1.1.1/goswitchctl install` (релизный каталог
v1.1.1 нетронут; прежние release-dir не перезаписывались).

Критерии журнальных рядов приёмки — прежние round-4 (а)-(е), дополнение: в окне
приёмки допустима метка **flip-parked** (жест во время доставки — теперь
корректная геометрия CR-01, а не провал); её release-рубеж по-прежнему строго
после коммита и по честной причине.

### Re-acceptance round-4, 2026-10-17 21:56 MSK — REJECTED; гипотеза владельца «повторное удаление» сильнее «запоздалого дренирования»

Сборка v1.1.1-190-g29bbde6 (машина round-4 + фиксы CR-01/CR-02). Владелец дал ВИДЕО (Screencast from 2026-10-17 21-55-52.mp4): ghbdtn → двойной шифт → «привет» заменил слово → ~1 с спустя строка ПУСТАЯ (стёрты и ghbdtn-последствия, и привет — 12 символов суммарно).

**Журнал жеста (21:56:01.636 action n=2):** terminal class verify_dead=true (10-07 жив) → 01.917 chunks-sent forwards=6 pacing_ms=120 → commit-sent latency_ms=281 → done changed=true → flip-armed. **РОВНО ОДИН пакет из 6 форвардов, второй эмиссии НЕТ.** Флип так и не релизнулся (не было следующей клавиши/пуша) — SetGlobalEngine в окне vanish НЕ было (пересоздания движка нет). Через ~3 с после коммита (21:56:04.9) — пара focus_out/focus_in same-engine на контексте wezterm.

**Дискриминация гипотез (владелец vs оркестратор):**
- «Запоздалый дрейн» (оркестратор, отозвано): FIFO дал бы «ghbdtnпривет» (коммит вперёд) или уцелевший «ghbdtn» (форварды вперёд) — кадр показывает ЧИСТУЮ строку: не подходит.
- **«Повторное удаление» (владелец) — подходит точно:** 6+6=12 стёртых символов при 6 отправленных ⇒ второй прогон пакета на клиенте. Кандидат-механизм: replay очереди форвардов при сбросе IM-контекста (фокус-пара 21:56:04.9, ~3 с после коммита) либо отрисованный-и-сброшенный preedit. Различимые только инструментированным жестом.

**Следствие:** фикс-drain-окна любой длины не лечат повторный прогон; требуется per-event видимость доставки. План round-5: (1) CONTENT-FREE инструментированный демон (счётчики/тайминги эмиссий и пересбросов, НОЛЬ содержимого клавиш — не кейлоггер, D-20/D-21) на ОДИН жест владельца — локализовать, кто повторяет пакет (клиент-replay на контекст-сбросе vs что-то ещё); (2) по вердикту — обход: не эмитить форварды до устойчивости контекста (quiescence-гейт после focus-пар), либо одношаговая замена слова через… (в поиске).

**Вердикт:** REJECTED (Test 5). G-10-1 открыт. Сборка 190 оставлена активной (коррекция работает, слово теряется — как и в 111-й; регресса печатной дисциплины CR-01/02 не замечено в этом жесте — окно дренирования отработало, повтор не на эмиссии).

### Round-5 instrument build, 2026-10-18 — контентно-свободные emission-seq метки + host selfdrive драйвер (без живых прогонов)

Коммиты `6ad0850` + `78179ee` (ветка gsd/phase-10-term-slovo). НИКАКИХ живых запусков: хост-юнит v1.1.1 не тронут, инъекций не было, окон не открывали.

**(1) Инструмент демона (внутренние ряды, D-20/D-21 абсолют):** каждая строка `correction delivery` несёт `emission-seq` (монотонный счётчик терминальных бёрстов) и `t_rel_ms` (смещение от якоря бёрста) — добавлены В КОНЕЦ строки, замороженные grep-формы (например `"stage":"flip-released","reason":"keystroke"`) сохраняют соседство атрибутов. Новые ряды `stage=burst-window phase=open/close` скобками обрамляют проводное окно эмиссии (open с planned forwards; close с фактическими — при free-abort `forwards=0 outcome=aborted`, при коммите `forwards=6 outcome=committed`). Отложенный флип штампует СВОЙ бёрст (seq/start в deferredFlip) — поздний flip-released не едет на чужом seq. Юнит-корпус: `internal/session/deliverymarks_test.go` — порядок скобок, seq на каждой строке, монотонность через раунды, флип-ряды, контентно-свободная булавка (в строках delivery нет ни ghbdtn, ни привет).

**(2) Host selfdrive драйвер (`test/e2e -selfdrive`, mise `test-selfdrive` = DRY-RUN only):** гоняет ВЫДЕЛЕННУЮ пейн-панель, которую владелец открывает вручную. Владелец печатает в панели ОДНУ команду: `script -q /tmp/gs-pane.log`. Дисциплина маркера: перед КАЖДЫМ бёрстом — `ydotool type .`, пауза 300 мс, проверка новых байтов лога; маркер не дошёл → громкий skip-ряд и АБОРТ батареи (вслепую настоящий код не печатается никогда). Затем BackSpace стирает маркер, `ghbdtn` + двойной Shift_R (те же keyvals, что в e2e-корпусе), пауза settle 3 с, парс НОВЫХ байтов лога. Классификатор хореографии: E6→T6 = чисто; E6→T6→E6 = ВТОРОЕ УДАЛЕНИЕ (дефект воспроизведён); no-correction / erase-only / empty / unexpected. Батарея: 1 probe + 3 baseline + 3 timing (пауза набор→жест 5 с); churn фокуса только НАБЛЮДАЕТСЯ (journalctl -u goswitchd lifecycle-ряды в окне прогона; без -debug их нет — честная note в ряду). Датированные вердикт-ряды: stdout + /tmp/gs-selfdrive-verdicts.log, ТОЛЬКО счётчики.

**Дискриминация по построению:** демон говорит один бёрст (один emission-seq, одни скобки burst-window), а pty-лог показывает два прогона удаления ⇒ replay на клиенте доказан; в журнале два seq ⇒ баг демона доказан. Юнит-корпус парсера: `test/e2e/selfdrive_test.go` (race-green, в ci).

**Инструкция оркестратору (позже, под его контролем):**
1. Владелец открывает выделенную панель wezterm и печатает там: `script -q /tmp/gs-pane.log`
2. Оркестратор запускает: `go run ./test/e2e -selfdrive -pane-log /tmp/gs-pane.log` (опционально `-baseline-runs/-timing-runs/-timing-wait/-settle/-pacing/-selfdrive-out`)
3. Веридикт-ряды: stdout + `/tmp/gs-selfdrive-verdicts.log`; juxtaposition с журналом демона (`journalctl --user -u goswitchd | grep 'emission-seq'` — на СЛЕДУЮЩЕЙ сборке с инструментом; текущая стабильная v1.1.1 инструмента НЕ содержит — установка отдельным шагом оркестратора).

### Round-5 самопилот, ночь 19 окт (00:38–01:23 MSK) — инжектор доставляет; найдена НОВАЯ причина отказа: панельный контекст attached к xkb МИНУС goswitch (демон видит Shift-модификаторы, не видит букв)

Починено по пути: (1) device-attach settle 2 c после uinject Open — mutter подключает новое evdev-устройство асинхронно; без паузы первый маркер теряется (00:38–01:13: все первые маркеры MISSING); (2) маркер-бюджет 300 мс → 1500 мс (эхо через демон+скрипт медленнее); (3) маркер-цифра «7» (раскладко-независимый); (4) нормализация режима — ПРЯМОЙ `ibus engine goswitch-en` (gsettings current не двигает ibus на этом стеке — пин 23:44/23:47); (5) selftype-режим (-selftype: драйвер сам печатает корпус+жест в сфокусированную панель, дампит delivery-метки демона, вердикт персистентности — одно слово владельца); (6) uinject таблица = весь a-z + пробел + пунктуация (pane-type канал для видимых сообщений в панели); (7) ydotool выброшен.

**Батарея прошла ЦЕЛИКОМ (settle починил доставку)** — 7 прогонов, но все `no-correction`: события прогона «T1 E7 T6» (маркер, Ctrl-U стёр маркер+хвост предыдущего прогона, корпус). Журнал демона за окно батареи: КАЖДЫЙ двойной шифт = `action n:1` + `mode to ru/en` (ЧИСТЫЙ ФЛИП — буфер пуст), autocorrect-строк НОЛЬ (демон не видел НИ ОДНОЙ буквы корпуса). Демон видит Shift-модификаторы, НЕ видит буквенных клавиш ⇒ панельный text-input контекст attached к обычному xkb, МИМО goswitch-моста. Причина detached: создание/уничтожение виртуального uinput-устройства (input-device churn) сбрасывает IM-контекст панели в xkb (фаза-9 «map-time focus yank» семья). ibus engine re-assert (`ibus engine goswitch-en`) перед батареей НЕ пере-bridged контекст (автокоррект-строк после него ноль).

**Состояние:** сборка v1.1.1--b2a1b1f + driver HEAD (самопилот); дефект G-10-1 (видео: delete→insert→delete) НЕ воспроизведён автопилотом — воспроизводится ДРУГОЕ состояние (панель мимо демона). Оба состояния реальны и, вероятно, связаны: churn IM-контекстов (uinput/hotplug) — кандидат №1 в ежедневный дефект владельца: контекст панели периодически пересоздаётся, и после пересоздания поведение меняется.

**Round-5 остов:** (α) локализовать detached/re-attached переходы (DEBUG-уровень демона: логировать ProcessKeyEvent-приходы и focus-пересоздания контекста — content-free); (β) зафиксировать, КТО владеет контекстом после churn (mutter text-input → ibus мост vs xkb); (γ) дизайн фикса: re-attach гейт (не слать коррекцию, пока контекст не подтверждён) или пере-bridging по событию устройства.

Мой долг перед владельцем (дословно из его сообщений): «мне надоело работать для тебя тестировщиком», «давай ты на этот раз все сделаешь как надо, или я тебя просто удалю» — протокол selftype (панель + одна команда + одно слово) ровно это и даёт. НЕ дёргать владельца до воспроизводимого самопилота.

### Selftype 2026-10-19 19:03 MSK — машина round-4 отработала на хосте ЦЕЛИКОМ (n:2, terminal class, burst, commit, done); слово «стало и исчезло» ⇒ потеря окончательно локализована в клиентской доставке

Владелец прогнал selftype (видео 19-03-44): корпус «ghbdtn» лёг ПОЛНОСТЬЮ (settle починил потерю первой буквы), жест → **полная машина**: action n:2 → terminal class verify_dead=true → burst-window open forwards=6 → chunks-sent pacing 120 → commit-sent +275 мс → done changed=true. Вердикт владельца: **«стало и исчезло»**.

⇒ ТЕРМИНАЛЬНАЯ ВЕТКА РАБОТАЕТ НА ХОСТЕ; машина доставки round-4 (дренаж, отложенный флип) не спасает — потеря = клиентская перестановка «клавиши-удаления vs вставка»: ключевые события (ForwardKeyEvent) стоят в очереди wezterm на ~секунды, commit (text-input) применяется мгновенно; никакой демонтовский тайминг не лечит разную скорость каналов (доказано 4 раундами: 20 мс / 120 мс / 150 мс окно / отложенный флип).

**Round-6 направление (единственная конструкция, неподвержённая очередям): удаление протокольным сообщением DeleteSurroundingText вместо клавиш-форвардов** — тот же text-input канал, что и commit, порядок гарантирован протоколом. Гард mutter-краша (10-16: materialized-surrounding evidence) обязателен — краш только на девственных контекстах, коррекция идёт по тёплым. Неизвестный: поддерживает ли wezterm-nightly zwp delete_surrounding_text — стенд-проба (сплит-вербы 10-13 уже в дереве) отвечает машинно.

Протокол selftype подтверждён рабочим: одна команда владельцем, одно слово ответом — остальное машиной.

### Closure validation round-6 (G-10-1) — парные форварды: установка + selftype-гейт, 2026-10-19

Round-6 pivot (мандат владельца; sourcedig `.planning/debug/2026-10-19-wezterm-repeat-sourcedig.md`,
0c2fd6e): корень «появилось и исчезло» — КЛИЕНТСКИЙ автоповтор wezterm, взводимый press-only
форвардами BackSpace (window.rs:1194-1206; снятие только release'ом той же клавиши); через
GNOME repeat-delay 500 мс планировщик синтезирует вторую волну удалений ~30 Гц. Фикс —
амендмент (7) ADR-008: каждый BackSpace терминальной записи едет парой press+release
(fed2a98 RED → 33f32c3 GREEN; ADR-008 + re-pin стенд-ряды 8bd4edc). DST-ветка закрыта
независимо зондом 10-13 (4f940b0: wezterm игнорирует DST, тёплая VTE DST роняет mutter 46.2) —
маршрутизация остаётся на форвардах, теперь парных. Инференс round-5 «форварды стоят в очереди
секунды» ОТЗВАН.

**Машинные предпроверки (read-only; 2026-10-19 21:18–21:20 MSK):**

- Гейты ДО установки: mise run ci зелёный (build+vet+lint+test -race, 21 пакет); пандус реплея
  зелёный под -race (ReplayHost0939 ×4 / TwoPhaseLevel1 ×4 / FlapWindowContextIdentity ×2 /
  Level1ConfirmFlush / DstGuard ×3 / новый TestActor_Level2ForwardsComeInPressReleasePairs);
  go.mod/go.sum заморожены (tidy-diff чист); стенд re-pin — четыре замороженных кейса зелёные
  (repin-round6-pairs 17:53–18:17 UTC, pty байт-точность сохранена: пара = один 0x7f).
- Установка (единственная host-мутация): build из рабочего дерева (HEAD 8bd4edc =
  v1.1.1-226-g8bd4edc) → НОВЫЙ release-dir ~/.local/share/goswitch/releases/v1.1.1-226-g8bd4edc
  → его goswitchctl install: component content verified, registry visible, unit daemon-reload +
  enable + restart, sources wrapped, engine activated goswitch-en.
- unit: active; ExecStart = releases/v1.1.1-226-g8bd4edc/goswitchd (старт 21:18:54 MSK);
  **/proc/2096528/exe → тот же release-dir** (идентичность по /proc сверена); журнал поднятия:
  component registered (engines:2) → engine reactivated goswitch-en, ошибок регистрации нет.
- status по фикс. пути (/home/nil/.local/share/goswitch/releases/v1.1.1-226-g8bd4edc/goswitchctl):
  парсится, config_valid=true, config_error= (пуст); счётчики на нуле — чистая база приёмки:
  corrections_done=0, corrections_skipped=0, autocorrect_fired=0, autocorrect_abstained=0,
  ac_role_errors=0, corr_dst_guarded=0. Замечание: version=dev — деревостроенная сборка без
  goreleaser-стемпов (прецедент gap-сборок; идентичность несёт release-dir + /proc/exe).
- Откат по запросу владельца (назван ДО приёмки): ~/.local/share/goswitch/releases/v1.1.1/
  goswitchctl install (релизный каталог v1.1.1 нетронут; предыдущие gap-сборки не перезаписаны).
- Host после установки — строго read-only: инъекций нет, окон не создавалось, чужие сервисы
  не перезапускались. Дальше — протокол selftype (панель + одна команда + одно слово), НИКАКИХ
  дёрганий владельца сверх его одного слова.

**Критерии журнальных рядов round-6 (сверка ПОСЛЕ жеста/самтайпа):**

- (а) НЕТ skip-рядов verify-timeout / verify-mismatch;
- (б) delivery-метки присутствуют и упорядочены: burst-window open → chunks-sent → commit-sent
  → burst-window close; НОВОЕ round-6: chunks-sent несёт forwards=12 keys=6 (пары: эмиссий вдвое
  против ключей), pacing_ms ≈ 240 (12×20 мс) — рост против 120.round-4 ЭТО ОЖИДАЕМ,
  не регресс; done-ряд changed=true;
- (в) флип-рубежи строго ПОСЛЕ release-причины (SetGlobalEngine между коммитом и причиной =
  провал — форма round-4);
- (г) нет крахов/перезапусков gnome-shell;
- (д) autocorrect_fired не вырос; (е) absent-метки/тишина = честный провал, никогда не пропуск.

**Selftype round-6 — инструкция владельца (одна команда, одно слово):**

1. Открыть ПУСТУЮ панель wezterm (та же, что в 19:03), сфокусировать, руки убрать.
2. Напечатать в ней ОДНУ команду:
   `cd /home/nil/DiskD/W/Djarvur/goswitch && go run ./test/e2e -selftype`
3. Драйвер сам: нормализует режим → 2 с settle → печатает «ghbdtn» → двойной Right Shift →
   печатает delivery-метки демона. Владелец смотрит на слово и отвечает ОДНИМ словом:
   **ОСТАЛОСЬ** или **ИСЧЕЗЛО**.
Ожидание по sourcedig (§6.4): release гасит взвод повтора — слово остаётся при ЛЮБОМ gsettings
repeat-delay; «исчезло» после ≥500 мс покоя = повтор всё ещё жив (новая охота).

Жест-эквивалент без самтайпа (по желанию): мышь в wezterm, «ghbdtn», двойной правый Shift,
смотреть ≥2 с (взвод стрелял в last_press+500 мс — теперь release пришёл на +20 мс).

### ЗАКРЫТИЕ G-10-1 — 2026-10-20, вердикт владельца: «ОСТАЛОСЬ»

Selftype-прогон на сборке v1.1.1-226-g8bd4edc (парные press+release форварды, амендмент (7) ADR-008): слово «привет» после двойного правого Shift ОСТАЛОСЬ на экране. Владелец: «осталось» — закрывающий оракул G-10-1 выдан.

Корневая цепочка шести раундов (полная летопись — в этом файле):
1. 10-07: флап-стойкая классификация (verifyDead переживает пары фокуса);
2. 10-10: маршрутизация answer-only (жест попадает в терминальную ветку без a11y-identity);
3. 10-11/10-17: доставка с дренированием и отложенным флипом;
4. 10-16: гвард чужого mutter-краша (DST-путь закрыт: no-op в wezterm + сессия-киллер на VTE);
5. round-6: парные форварды — корень по исходникам (press-only взводил клиентский автоповтор wezterm, window.rs:1194-1206; повтор через GNOME repeat-delay 500 мс съедал слово), release гасит взвод.

Предсказание sourcedig §6.4 сбылось дословно. Утро 2026-10-20: три живых_correction владельца (слова 7 и 11 букв) — полный цикл машины, устоялись (журнал 10:15:29/10:15:57).

Откат-путь прежний: ~/.local/share/goswitch/releases/v1.1.1/goswitchctl install. Хвосты для следующих фаз: ZCode-потеря фразы (вне скоупа по мандату), Тест 2 (формальная классификация края TERM-02), WR-05 advisory, OQ5, WINDOWS unrun-verify.
