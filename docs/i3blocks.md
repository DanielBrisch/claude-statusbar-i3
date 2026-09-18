# i3blocks

Add to `~/.config/i3blocks/config`:

```ini
[claude]
command=claude-statusbar render --format i3blocks --markup pango --icon-size x-large
markup=pango
interval=10
```

Block order in the file is left-to-right on the bar, so put it wherever you want it to
appear. Reload with `i3-msg restart` — i3blocks only re-reads its config on start.

## Why markup=pango

At a typical bar size the icon comes out smaller than the digits beside it. Pango markup
lets the renderer wrap just the icon in a size tag, so it grows without touching the rest
of the block or any other block on your bar.

The property and the flag go together. `markup=pango` without `--markup pango` is
harmless; `--markup pango` without the property prints the tag literally on your bar:

```
<span size="x-large">✳</span> 63% 1h42 │ week 21% 4d
```

For a plain block, drop both. To tune the size, `--icon-size` takes any pango keyword:
`small`, `medium`, `large`, `x-large`, `xx-large`.

The countdown is computed at render time from the reset timestamp, so `interval=10` keeps
it ticking without Claude Code doing anything.

## Clicking does nothing

Deliberately. The block already carries both windows, so there is nothing a click could
usefully add. For the absolute reset times and the spend limit, run:

```sh
claude-statusbar detail
```

If you would rather have the old notification, wire it yourself — i3blocks exports
`$BLOCK_BUTTON`:

```ini
command=[ "$BLOCK_BUTTON" = 1 ] && claude-statusbar detail --notify; claude-statusbar render --format i3blocks --markup pango
```

## Instant refresh (optional)

`interval=10` means up to ten seconds of lag. If that bothers you, drive the block with a
signal instead:

```ini
[claude]
command=claude-statusbar render --format i3blocks --markup pango --icon-size x-large
markup=pango
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
