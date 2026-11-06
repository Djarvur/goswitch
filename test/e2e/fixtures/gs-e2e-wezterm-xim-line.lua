-- gs-e2e-wezterm-xim-line.lua — the stand config of the phase-12 line
-- sel-probe's XIM cells (plan 12-01): the xim body (enable_wayland=false —
-- the OWNER's transport since 2026-10-24, Pitfall 11: a plan without an
-- xim row is forbidden) plus the CopyMode Shift+Home select recipe
-- byte-identical to the wayland fixture (the Multiple sequence: CopyMode
-- leaves the selection ACTIVE after Close; copy_on_select (default) puts
-- it in PRIMARY — read by wl-paste --primary, the X11-PRIMARY → Wayland
-- bridge is cell (b)'s question, A3). Stand-owned file: the owner's
-- ~/.wezterm.lua is NEVER touched (Pitfall 7).
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
return { enable_wayland = false, keys = keys }
