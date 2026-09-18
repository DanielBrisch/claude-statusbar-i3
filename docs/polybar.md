# polybar

Add to `~/.config/polybar/config.ini`:

```ini
[module/claude]
type = custom/script
exec = claude-statusbar render --format polybar
interval = 10
format = <label>
```

and add `claude` to one of the `modules-*` lines.

The output already carries polybar's colour tags and a left-click action that fires the
breakdown notification, so no `click-left` is needed.

If you would rather control the colour from polybar itself, render plain text and wrap it
yourself:

```ini
exec = claude-statusbar render --format plain
```
