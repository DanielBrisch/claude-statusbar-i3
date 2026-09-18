# claude-usage-status-i3

Claude Code's `/usage` numbers on your status bar.

```
… TEMP 52°C │ session 63% 1h42 │ week 21% 4d │ 🔊 40% …
              └── 5-hour window   └── weekly window
                  63% used,           21% used,
                  resets in 1h42      resets in 4 days
```

Left-click the block for the full breakdown:

```
5h window     63%  resets Sep 18 13:42 (1h42)
Weekly        21%  resets Sep 22 16:00 (4d)
as of 11:58
```

Both figures are account-wide. Per-session numbers — model, cost, context — are recorded
but deliberately kept out of the bar and the breakdown: with several Claude Code sessions
open, they describe whichever one wrote last, which is not the one you are looking at.
Reach them through `--format json` or a `--template` if you want them anyway.

Works with **i3blocks**, **waybar**, **polybar**, or anything that can run a command
(`--format json` / `--format plain`).

## How it works

Claude Code hands its [status line](https://code.claude.com/docs/en/statusline) command a
JSON payload on stdin, and that payload already contains the rate limits:

```json
"rate_limits": {
  "five_hour": { "used_percentage": 23.5, "resets_at": 1738425600 },
  "seven_day": { "used_percentage": 41.2, "resets_at": 1738857600 }
}
```

`claude-statusbar collect` sits in that slot, records the numbers to a small state file,
and `claude-statusbar render` reads them back for your bar.

No API calls. No credentials. Nothing reads your OAuth token or your transcripts.

```
Claude Code ──stdin──> collect ──> ~/.local/state/claude-statusbar/state.json ──> render ──> your bar
```

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/DanielBrisch/claude-usage-status-i3/main/install.sh | sh
```

Or, with a Go toolchain:

```sh
go install github.com/DanielBrisch/claude-usage-status-i3/cmd/claude-statusbar@latest
```

The installer honours `BINDIR` (default `~/.local/bin`) and `VERSION` (default: latest release).

The repository is `claude-usage-status-i3`; the command it installs is `claude-statusbar`.

## Setup

```sh
claude-statusbar init --bar i3blocks
```

That registers the collector in `~/.claude/settings.json` (backing up whatever was there)
and prints the snippet for your bar. It refuses to clobber an existing `statusLine` —
chain yours instead:

```sh
claude-statusbar init --passthrough '~/.claude/my-statusline.sh'
```

Per-bar instructions: [i3blocks](docs/i3blocks.md) · [waybar](docs/waybar.md) · [polybar](docs/polybar.md)

Then send one message in Claude Code so the status line fires, and the block appears.

By default the bar polls the state file on its own `interval`. If you want it to redraw
the instant Claude Code records new numbers, give `collect` a `--refresh` command — see
the per-bar docs.

## Commands

| | |
|---|---|
| `collect` | Reads the statusLine payload on stdin and records it. `--print compact\|none`, `--passthrough CMD`, `--refresh CMD` |
| `render` | Renders for your bar. `--format i3blocks\|waybar\|polybar\|plain\|json` |
| `detail` | Prints the full breakdown. `--notify` sends it as a desktop notification |
| `init` | Registers the collector and prints the bar snippet |
| `doctor` | Explains why the block might be empty |

### Appearance

```sh
claude-statusbar render --label session --weekly-label week --warn 60 --crit 85 --urgent 95
claude-statusbar render --color-warn '#E5C07B' --color-crit '#E06C75'
```

Colours follow the worst live window. At `--urgent` the i3blocks format also exits 33,
which marks the block urgent.

Or lay it out yourself:

```sh
claude-statusbar render --template '{session_pct} ({session_reset}) · week {weekly_pct}'
```

Placeholders: `{label}` `{weekly_label}` `{session_pct}` `{session_reset}` `{weekly_pct}`
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
- **i3bar has no hover events**, so under i3blocks the breakdown is on left-click. Waybar
  gets a real tooltip.

## State file

`$XDG_STATE_HOME/claude-statusbar/state.json`, falling back to
`~/.local/state/claude-statusbar/state.json`. Written atomically under `flock`, because
every open Claude Code session runs the collector. Rate limits are account-wide, so the
freshest write wins; per-session cost and context are kept separately and pruned after 30
minutes idle.

## License

MIT
