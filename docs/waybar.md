# waybar

Add to `~/.config/waybar/config`:

```json
"custom/claude": {
  "exec": "claude-statusbar render --format waybar",
  "return-type": "json",
  "interval": 10,
  "on-click": "claude-statusbar detail --notify"
}
```

and put `"custom/claude"` in one of the `modules-*` arrays.

The JSON carries `text`, `tooltip`, `class` and `percentage`, so waybar shows the full
breakdown on hover with no extra work.

Style it by level:

```css
#custom-claude.warn   { color: #e5c07b; }
#custom-claude.crit   { color: #e06c75; }
#custom-claude.urgent { color: #e06c75; font-weight: bold; }
```

## Instant refresh (optional)

`interval` polling means up to that many seconds of lag. To redraw the moment Claude Code
records new numbers, give the module a signal:

```json
"custom/claude": {
  "exec": "claude-statusbar render --format waybar",
  "return-type": "json",
  "interval": 10,
  "signal": 12,
  "on-click": "claude-statusbar detail --notify"
}
```

and point the collector at it in `~/.claude/settings.json`:

```json
"statusLine": {
  "type": "command",
  "command": "claude-statusbar collect --refresh 'pkill -SIGRTMIN+12 waybar'"
}
```
