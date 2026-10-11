-- gs-e2e-wezterm-xim.lua — the stand config of the spike's BASE transport
-- (plan 09-02): minimal on purpose — the owner's keys/plugins are NOT loaded
-- (Pitfall 7: ~/.wezterm.lua is never touched during the spike) — and
-- enable_wayland=false mirrors the OWNER's transport
-- (~/.wezterm.lua sets the same; the XIM-hypothesis precondition P0 pins).
local wezterm = require 'wezterm'
return { enable_wayland = false }
