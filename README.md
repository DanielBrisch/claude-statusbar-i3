# claude-statusbar-i3

Claude Code's `/usage` numbers on your status bar, so you know whether you can keep going
without opening Claude and asking.

```
… TEMP 52°C │ ✳ session 63% 1h42 │ week 21% 4d │ 🔊 40% …
                  └── 5-hour window   └── weekly window
                      63% used,           21% used,
                      resets in 1h42      resets in 4 days
```

A single Go binary with no dependencies, for **i3blocks**, **waybar**, **polybar**, or
anything that can run a command. It reads the rate limits Claude Code already hands to its
[status line](https://code.claude.com/docs/en/statusline) command, so there are no API
calls and it never touches your credentials.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/DanielBrisch/claude-statusbar-i3/main/install.sh | sh
```

Or, with a Go toolchain:

```sh
go install github.com/DanielBrisch/claude-statusbar-i3/cmd/claude-statusbar@latest
```

The repository is `claude-statusbar-i3`; the command it installs is `claude-statusbar`.

## Setup

```sh
claude-statusbar init --bar i3blocks
```

That registers the collector in `~/.claude/settings.json`, backing up whatever was there,
and prints the snippet for your bar. Send one message in Claude Code so the status line
fires, and the block appears.

If the block stays empty, `claude-statusbar doctor` says why.

## More

[Configuration and flags](docs/configuration.md) · [i3blocks](docs/i3blocks.md) ·
[waybar](docs/waybar.md) · [polybar](docs/polybar.md) · [Contributing](CONTRIBUTING.md)

## License

MIT
