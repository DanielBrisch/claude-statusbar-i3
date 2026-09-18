# Configuration

Everything beyond `init`. Per-bar setup lives in
[i3blocks](i3blocks.md) · [waybar](waybar.md) · [polybar](polybar.md).

## Commands

| | |
|---|---|
| `collect` | Reads the statusLine payload on stdin and records it. `--print compact\|none`, `--passthrough CMD`, `--refresh CMD` |
| `render` | Renders for your bar. `--format i3blocks\|waybar\|polybar\|plain\|json` |
| `detail` | Prints the full breakdown. `--notify` sends it as a desktop notification |
| `init` | Registers the collector and prints the bar snippet |
| `doctor` | Explains why the block might be empty |

### Appearance

**The block takes no colour of its own.** It prints text and lets your bar draw it in
whatever colour the rest of your status line uses. A tool that decides your bar is red
today is a tool fighting your theme, so colour is something you ask for:

```sh
claude-statusbar render --color-warn '#E5C07B' --color-crit '#E06C75'
```

Then the block turns yellow past `--warn` and red past `--crit`, following whichever live
window is worst. The thresholds move independently of the colours:

```sh
claude-statusbar render --warn 60 --crit 85 --urgent 95
```

`--urgent` on its own only labels the level. Add `--urgent-exit` to make the i3blocks
format exit 33 past that point, which is how i3bar is told to mark a block urgent — it
recolours the block from your bar's `urgent_workspace` palette, so it is opt-in for the
same reason the colours are.

Under waybar nothing is opt-in: the module always reports `class` as `ok`, `warn`, `crit`
or `urgent`, and your CSS decides whether that means anything.

Or lay it out yourself:

```sh
claude-statusbar render --template '{session_pct} ({session_reset}) · week {weekly_pct}'
```

The block opens with two separate things, and each has its own flag:

```sh
claude-statusbar render --icon '✳' --label session --weekly-label week
```

`--icon` is the glyph that says which tool the block is about; `--label` is the word that
says which window the first pair of figures belongs to, the same job `--weekly-label` does
for the second. Either can be emptied: `--icon ''` drops the glyph, `--label ''` leaves the
icon alone in front of the figures.

They are separate because only the icon is ever resized. Enlarging `✳ session` as one
string would blow up the word too.

The default icon is U+2733, written without a variation selector so it renders from your
monospace font rather than the colour emoji font and takes the block's colour. If your bar
font lacks it, pass any glyph you like.

At bar sizes that icon lands smaller than the digits next to it. Where the bar parses
pango markup, `--markup pango` wraps just the icon in a size tag so it grows on its own:

```sh
claude-statusbar render --format i3blocks --markup pango --icon-size x-large
```

Your bar has to be told to parse it — under i3blocks that is `markup=pango` on the block,
see [docs/i3blocks.md](docs/i3blocks.md). Without that the tag shows up literally.
`--format plain` and `--format json` ignore the flag entirely.

Polybar parses no pango, so it has its own knob: `--icon-font N` draws the icon with the
bar's Nth font. See [polybar.md](polybar.md).

Placeholders: `{icon}` `{label}` `{weekly_label}` `{session_pct}` `{session_reset}` `{weekly_pct}`
`{weekly_reset}` `{spend_pct}` `{spend_reset}` `{model}` `{cost}` `{context_pct}`.
Expired or missing windows render as `—`.

## What it does not do

These are design limits, not bugs:

- **The numbers only move while Claude Code is running.** Between sessions the weekly
  figure is the last one observed — `detail` always states the observation time. Usage
  from claude.ai or another machine shows up on your next local session.
- **API key, Bedrock and Vertex logins get no `rate_limits`** from Claude Code, so the
  block stays empty. `doctor` will tell you that is what happened.
- **`statusLine` is a single slot** in `settings.json`. Use `--passthrough` to keep yours.
- **i3bar has no hover events.** Waybar gets a real tooltip and polybar a click action;
  under i3blocks the block is all you get, which is why it carries both windows itself.
  `claude-statusbar detail` is always there when you want the absolute reset times.

## State file

`$XDG_STATE_HOME/claude-statusbar/state.json`, falling back to
`~/.local/state/claude-statusbar/state.json`. Written atomically under `flock`, because
every open Claude Code session runs the collector. Rate limits are account-wide, so the
freshest write wins; per-session cost and context are kept separately and pruned after 30
minutes idle.
