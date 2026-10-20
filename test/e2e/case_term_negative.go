package main

// The phase-10 frozen negative witnesses (plan 10-05, criterion 5): inside
// the VM stand, with the autocorrect layer ENABLED, vim (insert mode) and
// htop — both INSIDE gnome-terminal, the same terminal class — must stay
// SILENT under the wrong-layout word plus its separator: fired=0 and the
// terminal-veto abstentions grow (the D-55 veto of plan 10-03, now
// observable in the D-54 counters). The positive control types the same
// word into the same terminal class and corrects it with the manual Double
// — the silence is the VETO, not a deaf channel. The re-pin step runs the
// existing autocorrect-terminal-silent case in-process (its outcome is a
// journal row, never the case's verdict — the full wezterm re-pin stays
// with plan 10-06). vim/htop arm by the key-trace probe (the pty-probe
// arming would type into their buffers); vim's content oracle is the
// :wq-written buffer file (hex in the row), htop's honest observables are
// the counters plus the process being alive — a TUI has no text buffer for
// the content oracle, the class's content proof carries witness-vim.

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// paneShell is the pane command's shell prefix (the capturePane arg; the
// goconst floor).
const paneShell = "bash"

// vetoCountRe parses the D-54 counters' terminal-veto abstention token out
// of goswitchctl status.
var vetoCountRe = regexp.MustCompile(`ac_skip_terminal_veto=(\d+)`)

// acVetoCount reads the terminal-veto abstention total from goswitchctl
// status; an absent token is the zero (ctlsvc prints only the reasons that
// happened).
func acVetoCount(ctx context.Context, ctlBin string) (int, error) {
	out, _, err := runCtl(ctx, ctlBin, "status")
	if err != nil {
		return 0, fmt.Errorf("autocorrect status: %w", err)
	}
	m := vetoCountRe.FindStringSubmatch(out)
	if m == nil {
		return 0, nil
	}
	n, perr := strconv.Atoi(m[1])
	if perr != nil {
		return 0, fmt.Errorf("parse ac_skip_terminal_veto: %w", perr)
	}

	return n, nil
}

// vimPane builds the vim witness's pane program: no rcs, no swap, no
// viminfo, straight into insert mode (the pane command line embeds the
// fixture's ready path — the teardown kill marker).
func vimPane(ready, _ string) []string {
	return []string{paneShell, "-c", "echo ready > " + ready + "; vim -u NONE -N -n -i NONE +startinsert"}
}

// htopPane builds the htop witness's pane program.
func htopPane(ready, _ string) []string {
	return []string{paneShell, "-c", "echo ready > " + ready + "; exec htop"}
}

// termTypeWordSep types the wrong-layout word plus its separator (the
// boundary the veto answers) and gates the key-trace.
func (s *stand) termTypeWordSep(ctx context.Context, caseName string) error {
	base := s.countSub(keyTraceMark)
	if err := s.injectText(ctx, acWord); err != nil {
		return fmt.Errorf("%s: type %q: %w", caseName, acWord, err)
	}

	return s.waitForNew(ctx, keyTraceMark, base+minKeyEvents, keyWait)
}

// termSilenceOracle runs the shared silence oracle set of the witnesses:
// the settle, zero fired/correction records in the journal, the active-layer
// status with fired=0, and the veto delta — typing must have grown the
// terminal-veto abstentions (before→after).
func (s *stand) termSilenceOracle(ctx context.Context, caseName, ctlBin string) (int, error) {
	if err := sleepCtx(ctx, acTerminalSettle); err != nil {
		return 0, err
	}
	after, err := acVetoCount(ctx, ctlBin)
	if err != nil {
		return 0, err
	}
	if err := s.assertAcSilent(caseName); err != nil {
		return 0, err
	}
	if err := ctlStatusHas(ctx, ctlBin, "autocorrect_enabled=true", "autocorrect_fired=0"); err != nil {
		return 0, fmt.Errorf("%s: %w", caseName, err)
	}

	return after, nil
}

// termVimWriteQuit leaves insert mode and writes the buffer out through
// :wq (the typed command rides the same transit channel), then waits for
// the oracle file. The content is the buffer's ground truth.
func (s *stand) termVimWriteQuit(ctx context.Context, caseName, oraclePath string) (string, error) {
	if err := s.injectKeys(ctx, "esc"); err != nil {
		return "", fmt.Errorf("%s: esc: %w", caseName, err)
	}
	if err := s.injectText(ctx, ":wq "+oraclePath+"\n"); err != nil {
		return "", fmt.Errorf("%s: :wq: %w", caseName, err)
	}
	deadline := time.Now().Add(termCaptureWait)
	for {
		if data, err := os.ReadFile(oraclePath); err == nil {
			return string(data), nil
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("%s: vim oracle file %s never appeared within %v",
				caseName, oraclePath, termCaptureWait)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return "", err
		}
	}
}

// termWitnessVim proves the silence in vim's insert mode: the word plus its
// separator arms NOTHING (fired=0, veto abstentions grow) and the buffer
// holds the typed word verbatim (:wq oracle).
func (s *stand) termWitnessVim(ctx context.Context, caseName, ctlBin string) error {
	const rowTag = "witness-vim"
	oraclePath := filepath.Join(s.tmpDir, "vim-witness.txt")
	tf, err := s.openTermFixture(ctx, caseName, "vim", vimPane)
	if err != nil {
		return err
	}
	defer tf.close()
	if err := s.termArmGate(ctx, caseName); err != nil {
		return err
	}
	before, err := acVetoCount(ctx, ctlBin)
	if err != nil {
		return err
	}
	if err := s.termTypeWordSep(ctx, caseName); err != nil {
		return err
	}
	after, err := s.termSilenceOracle(ctx, caseName, ctlBin)
	if err != nil {
		return err
	}
	if after <= before {
		return fmt.Errorf("%s: %s: terminal-veto воздержания не выросли (%d→%d)", caseName, rowTag, before, after)
	}
	content, cerr := s.termVimWriteQuit(ctx, caseName, oraclePath)
	hexr := hex.EncodeToString([]byte(content))
	observed := fmt.Sprintf("venue: %s; fired=0; veto Δ=%d; buffer_hex=%s", venueMarker(ctx), after-before, hexr)
	cmd := "vim insert-mode + type ghbdtn␠ + esc + :wq" +
		" (counters + buffer-file oracle)"
	if cerr == nil && strings.TrimSpace(content) == corpusLatin && !strings.Contains(content, wordResultRU) {
		return termRow(rowTag, cmd, observed,
			"Verdict: pass (fired=0, terminal-veto воздержания >0, буфер не изменён)")
	}
	if rerr := termRow(rowTag, cmd, observed, "Verdict: FAIL (счётчики или буфер изменились)"); rerr != nil {
		return rerr
	}
	if cerr != nil {
		return cerr
	}

	return fmt.Errorf("%s: %s: буфер изменён: %q", caseName, rowTag, content)
}

// termWitnessHtop proves the silence in htop (a TUI inside the same
// terminal class): the same counter oracles plus the process being alive.
// The pty carries no honest content oracle for a TUI — the class's content
// proof carries witness-vim, and the row says so.
func (s *stand) termWitnessHtop(ctx context.Context, caseName, ctlBin string) error {
	const rowTag = "witness-htop"
	tf, err := s.openTermFixture(ctx, caseName, "htop", htopPane)
	if err != nil {
		return err
	}
	defer tf.close()
	if err := s.termArmGate(ctx, caseName); err != nil {
		return err
	}
	before, err := acVetoCount(ctx, ctlBin)
	if err != nil {
		return err
	}
	if err := s.termTypeWordSep(ctx, caseName); err != nil {
		return err
	}
	after, err := s.termSilenceOracle(ctx, caseName, ctlBin)
	if err != nil {
		return err
	}
	if after <= before {
		return fmt.Errorf("%s: %s: terminal-veto воздержания не выросли (%d→%d)", caseName, rowTag, before, after)
	}
	alive, aerr := runCmd(ctx, "pgrep", "-x", "htop")
	if aerr != nil || alive == "" {
		return fmt.Errorf("%s: %s: htop не жив после раунда", caseName, rowTag)
	}
	observed := fmt.Sprintf("venue: %s; fired=0; veto Δ=%d; htop жив (pid %s); "+
		"контент-свидетельство несёт witness-vim — у TUI нет текстового буфера",
		venueMarker(ctx), after-before, strings.Fields(alive)[0])
	cmd := "htop + type ghbdtn␠ (counters oracle + process alive)"
	if err := termRow(rowTag, cmd, observed,
		"Verdict: pass (fired=0, terminal-veto воздержания >0, процесс жив)"); err != nil {
		return err
	}
	// Best-effort quit (Esc closes any dialog the letters opened, q quits);
	// the teardown's marker kill is the fallback.
	_ = s.injectKeys(ctx, "esc", "esc")
	_ = s.injectText(ctx, "q")

	return nil
}

// termPositiveControl proves the manual channel in the SAME terminal class
// under the enabled layer: the word corrects through the write branch and
// the capture holds привет — the silence above is the veto, not deafness.
func (s *stand) termPositiveControl(ctx context.Context, caseName string) error {
	return s.termWordRowCore(ctx, caseName, "positive-control",
		capturePane("bash", "-c"), corpusLatin, corpusHexRU)
}

// termRepinRow runs the existing autocorrect-terminal-silent case
// in-process and records its outcome as a journal row — never the case's
// verdict (the plan's split: the VTE rows pin the terminal negative here,
// the full wezterm re-pin stays with plan 10-06).
func (s *stand) termRepinRow(ctx context.Context) {
	const rowTag = "autocorrect-terminal-silent-repin"
	cmd := "runAutocorrectTerminalSilent in-process (existing case)"
	err := runAutocorrectTerminalSilent(ctx, s)
	if err == nil {
		observed := "venue: " + venueMarker(ctx)
		_ = termRow(rowTag, cmd, observed, "Verdict: pass (существующий негатив зелёный на площадке)")

		return
	}
	// The blocked form: the error text rides the row; the case's verdict
	// rides its own witnesses.
	observed := "case error: " + boundRowText(err.Error())
	_ = termRow(rowTag, cmd, observed,
		"Verdict: blocked (полный wezterm re-pin остаётся за планом 10-06)")
}

// runTermNegativeWitnesses drives the frozen negative witnesses (criterion
// 5): the enabled layer, the observers, vim-insert and htop silent under
// the veto while the positive control corrects in the same class, then the
// re-pin row. A failure lands the honest-abort row and fails the case.
func runTermNegativeWitnesses(ctx context.Context, s *stand) error {
	const caseName = "term-negative-witnesses"
	err := runTermNegativeRounds(ctx, s)
	if err == nil {
		fmt.Printf("%s: vim-insert/htop silent under the veto, positive control green (rows in %s)\n",
			caseName, spikeJournal)

		return nil
	}
	obs := "case aborted: " + boundRowText(err.Error())
	if rerr := termRow(caseName, "терминальные негативные свидетели", obs, "Verdict: aborted"); rerr != nil {
		fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, rerr)
	}

	return err
}

// runTermNegativeRounds is the round program of the negative case
// (extracted so the honest-abort wrapper stays under the complexity
// budget).
func runTermNegativeRounds(ctx context.Context, s *stand) error {
	const caseName = "term-negative-witnesses"
	if err := startAutocorrectConfigDaemon(ctx, s, autocorrectConfigDoc); err != nil {
		return err
	}
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}
	if err := s.termObserverRound(ctx, caseName); err != nil {
		return err
	}
	if err := s.termWitnessVim(ctx, caseName, ctlBin); err != nil {
		return err
	}
	if err := s.termWitnessHtop(ctx, caseName, ctlBin); err != nil {
		return err
	}
	if err := s.termPositiveControl(ctx, caseName); err != nil {
		return err
	}
	s.termRepinRow(ctx)

	return nil
}
