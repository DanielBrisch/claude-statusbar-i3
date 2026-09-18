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
- `internal/usage` must not import `internal/render` or `internal/statusline`. It is the
  domain: it takes a snapshot and decides freshness and thresholds, nothing else.

## Commit messages

Conventional commits, in English, lowercase subject:

```
feat(render): add a polybar output format
fix(state): break updated_at ties by write order
```

## Adding a bar

A new bar is a function in `internal/render`, a case in the `render` subcommand, a
snippet in `internal/cli/snippet.go`, and a page under `docs/`. Keep each bar's idioms
inside its own renderer — the domain does not know that bars exist.
