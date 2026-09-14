## Deferred Items

- closeEntrySurface presses ydotool key name "Escape" — resolves to the physical E key on ydotool 0.1.8
  status: open
  **What:** `test/e2e/case_m1.go` `closeEntrySurface` (surfaceShell branch) calls `pressKey(ctx, "Escape")`. Live probe 2026-09-15 (plan 02-05): ydotool 0.1.8's name table has no "Escape" entry and falls back to the first letter — the daemon observed keyval 0x65/keycode 18 (the E key), so the locked-session fallback types 'e' into the shell entry instead of dismissing it. The lowercase name "esc" is the one that maps to the real KEY_ESC (0xff1b/keycode 1), proven live in the same probe. Not fixed here: the shell-close path is 02-03 code outside 02-05's file scope and its fallback path was not exercised by this plan's cases. Same trap class as the 02-03 "space"→S finding; fix is a one-word change when the shell surface path is next touched (02-06 matrix runner is the natural moment).
