# 001 — Form, not season

**Decision:** the app represents *recent form*. Any stat defaults to a rolling window (last N GWs / matches), never whole-season totals. No season toggle unless the product owner asks.

**Why:** explicit product direction: the site exists to inform pre-deadline decisions from current form.

**How to apply:** make new metrics window-scoped like the existing ones (`SP_WINDOW_MATCHES`, `form.gw_window`, must-have `form_window`). Set-piece goal conversion counts goals inside the window, which was accepted over season totals. A season toggle was proposed for shot zones and explicitly dropped.
