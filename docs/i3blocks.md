# i3blocks

Add to `~/.config/i3blocks/config`:

```ini
[claude]
command=claude-statusbar render --format i3blocks
interval=10
```

Block order in the file is left-to-right on the bar, so put it wherever you want it to
appear. Reload with `i3-msg restart`.

The countdown is computed at render time from the reset timestamp, so `interval=10` keeps
it ticking without Claude Code doing anything.

## Click

Left-click sends the breakdown as a desktop notification (`dunstify`, falling back to
`notify-send`). No extra configuration — the block reads `$BLOCK_BUTTON` itself.

## Instant refresh (optional)

`interval=10` means up to ten seconds of lag. If that bothers you, drive the block with a
signal instead:

```ini
[claude]
command=claude-statusbar render --format i3blocks
interval=10
signal=12
```

and have the collector poke it. Edit the `statusLine` command in `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "claude-statusbar collect --refresh 'pkill -SIGRTMIN+12 i3blocks'"
  }
}
```

Now the block redraws the moment Claude Code records new numbers, and `interval=10` is
just the fallback that keeps the countdown ticking.

Pick a signal number nothing else uses — grep the rest of your i3blocks config for
`signal=` first. If the `pkill` fails, `collect` reports it on stderr and carries on, so a
wrong number never breaks Claude Code's status line.

## i3blocks 1.4 vs 1.5

This uses the plain text protocol (`full_text` / `short_text` / `color` on separate
lines, exit 33 for urgent), which both versions accept.
