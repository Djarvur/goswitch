#!/usr/bin/env bash
# d48-nightly-dispatch.sh — диспатчер формального гейта D-48 (G-4-2, план 04-09).
#
# Запускается systemd user-юнитом goswitch-d48-dispatch в свежей
# автологин-сессии — человек команду не набирает никогда (директива
# владельца 2026-09-17: убрать человека из цикла приёмки). Цепочка
# машинной настройки — autologin GDM, root-таймер, оба user-юнита —
# документирована в docs/ci-runner.md, раздел
# «Автономный ночной гейт (D-48 v2)»; в репозитории НИЧЕГО не исполняет
# sudo и не применяет root-часть сама.

set -euo pipefail

# До мержа фазы override-файл (~/.config/goswitch/d48-dispatch.env, вне
# git) задаёт фаза-ветку; дефолт main — пост-мерж состояние.
ref="${D48_REF:-main}"

# Гвард аутентификации: без gh или без токена диспатч невозможен —
# именованная ошибка в stderr вместо трассировки gh (юнит стартует без
# терминала, читать её некому кроме journalctl).
if ! command -v gh >/dev/null 2>&1; then
  echo "d48-dispatch: gh CLI not found in PATH — install gh first" >&2
  exit 1
fi
if ! gh auth status >/dev/null 2>&1; then
  echo "d48-dispatch: gh is not authenticated — run: gh auth login" >&2
  exit 1
fi

# --repo обязателен: юнит стартует с произвольным cwd, а gh не резолвит
# репозиторий из $HOME.
exec gh workflow run e2e-matrix.yml --repo Djarvur/goswitch --ref "$ref" \
  -f matrix=test/e2e/cases/matrix-v3.yaml -f fresh_session=true
