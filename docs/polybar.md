# polybar

Add to `~/.config/polybar/config.ini`:

```ini
[module/claude]
type = custom/script
exec = claude-statusbar render --format polybar --icon-font 2
interval = 10
format = <label>
```

and add `claude` to one of the `modules-*` lines.

The output already carries polybar's colour tags and a left-click action that fires the
breakdown notification, so no `click-left` is needed.

## Making the icon bigger

At bar sizes the icon lands smaller than the digits beside it. Polybar has no pango
markup, so `--markup pango` does nothing here — its mechanism is a font switch, and
`--icon-font N` wraps just the icon in `%{TN}`.

The index is 1-based over the fonts you declared on the bar, so `--icon-font 2` selects
`font-1`:

```ini
[bar/main]
font-0 = "DejaVu Sans Mono:size=10;2"
font-1 = "DejaVu Sans Mono:size=16;3"
```

Leave the flag off and the icon is drawn at text size, which is the safe default: a
`%{T2}` pointing at a font you never declared is a glyph polybar cannot draw.

## Colour

The module inherits your bar's foreground and stays that way at every usage level. To
have it change, pass the colours you want:

```ini
exec = claude-statusbar render --format polybar --icon-font 2 --color-warn '#E5C07B' --color-crit '#E06C75'
```

If you would rather control the colour from polybar itself, render plain text and wrap it
yourself:

```ini
exec = claude-statusbar render --format plain
```
