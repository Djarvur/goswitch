-- gs-e2e-wezterm-stand.lua — the stand config of the phase-10 live word
-- case (plan 10-06): minimal on purpose — the owner's ~/.wezterm.lua is
-- NEVER touched (Pitfall 7), and NO enable_wayland override: the nightly
-- build rides its native Wayland default (the 2026-10-12 verdict — the
-- phase-9 xim/wayland fixture flags are historical instruments of a build
-- that is gone). The caps re-pin row of the case observes what THIS
-- build announces; the routing is the live classification.
return {}
