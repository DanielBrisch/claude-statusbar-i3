# Contributing

Thanks for taking a look.

## Branches

- **`main`** — released code. Only ever updated by a pull request from `develop`.
- **`develop`** — integration branch. Everything new lands here first.

Work happens on a branch off `develop`:

```sh
git switch develop
git pull
git switch -c feat/my-change
```

Open the pull request against `develop`. Direct pushes to `main` and `develop` are not
the workflow, even when you have the rights — a pull request is what gives CI a chance to
run and the change a chance to be read.

## Before you open a PR

```sh
gofmt -l .        # must print nothing
go vet ./...
go test -race ./...
```

CI runs the same three on Linux and macOS, plus `shellcheck install.sh`.

## Tests come first

This project is written test-first, and the tests are the specification. A pull request
that changes behaviour without a test that fails before the change will be sent back.

Two conventions worth knowing:

- `internal/render/testdata/*.txt` are golden files. Regenerate with
  `UPDATE_GOLDEN=1 go test ./internal/render/` and read the diff before committing it.
- `internal/usage` must not import any other package in this repo. It is the domain: it
  takes a snapshot and decides freshness and thresholds, nothing else.

## Commit messages

Conventional commits, in English, lowercase subject:

```
feat(render): add a polybar output format
fix(state): break updated_at ties by write order
```

## Layout

One type per file, with its methods beside it. Interfaces live alone in a file named
after them. The only package-level functions are constructors — everything else is a
method on the type it is about.

```
internal/usage       the domain: Limit, Thresholds, Level, TimeLeft, Snapshot
internal/statusline  Parser and the wire payload Claude Code sends on stdin
internal/state       Store, the on-disk DTOs, a file lock and an atomic writer
internal/render      Renderer + one Formatter per bar, behind a Registry
internal/notify      Notifier and its backends
internal/setup       Installer for settings.json, and the per-bar Snippet
internal/cli         App, one Command per file, flags and path resolution
```

Dependencies point one way: `cli` → `render`/`state`/`setup`/`notify` → `usage`.
`usage` imports nothing of ours.

## Adding a bar

Four things, none of which touch existing code:

1. A type in `internal/render` implementing `Formatter` — one file, named after the bar.
2. An entry in `internal/render/registry.go`.
3. A case in `internal/setup/snippet.go` so `init` can print its config.
4. A page under `docs/`.

Keep each bar's idioms inside its own formatter. `Output` carries an `ExitCode` precisely
so a convention like i3blocks' exit 33 stays in `i3blocks.go` instead of leaking into the
CLI.
