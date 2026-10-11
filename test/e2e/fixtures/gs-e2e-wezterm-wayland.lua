-- gs-e2e-wezterm-wayland.lua — the stand config of the spike's
-- native-Wayland recipe trial (plan 09-02): enable_wayland=true is the
-- recipe under test (P0' re-probe: WAYLAND_DISPLAY must appear in the
-- process environ), and the CopyMode Shift+Home binding is the select
-- recipe (the Multiple sequence: CopyMode leaves the selection ACTIVE
-- after Close — the read channel the wl-paste --primary probe measures).
-- Stand-owned file: the owner's ~/.wezterm.lua is NEVER touched (Pitfall 7).
local wezterm = require 'wezterm'
local keys = {}
table.insert(keys, {
  key = 'Home', mods = 'SHIFT',
  action = wezterm.action.Multiple{
    wezterm.action.ActivateCopyMode,
    wezterm.action.CopyMode { SetSelectionMode = 'Cell' },
    wezterm.action.CopyMode 'MoveToStartOfLineContent',
    wezterm.action.CopyMode 'Close',
  },
})
return { enable_wayland = true, keys = keys }
