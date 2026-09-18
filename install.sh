#!/bin/sh
set -eu

REPO="${REPO:-DanielBrisch/claude-statusbar-i3}"
BIN="claude-statusbar"
BINDIR="${BINDIR:-$HOME/.local/bin}"
VERSION="${VERSION:-latest}"

die() { printf 'install: %s\n' "$*" >&2; exit 1; }

need() { command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"; }

need uname
need tar
if command -v curl >/dev/null 2>&1; then
  fetch() { curl -fsSL "$1" -o "$2"; }
  fetch_stdout() { curl -fsSL "$1"; }
elif command -v wget >/dev/null 2>&1; then
  fetch() { wget -qO "$2" "$1"; }
  fetch_stdout() { wget -qO- "$1"; }
else
  die "need curl or wget"
fi

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux|darwin) ;;
  *) die "unsupported OS: $os" ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) die "unsupported architecture: $arch" ;;
esac

if [ "$VERSION" = latest ]; then
  VERSION=$(fetch_stdout "https://api.github.com/repos/$REPO/releases/latest" \
    | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
    | head -n 1)
  [ -n "$VERSION" ] || die "could not resolve the latest release; set VERSION=vX.Y.Z"
fi

tarball="${BIN}_${VERSION#v}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/$VERSION"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

printf 'install: fetching %s %s (%s/%s)\n' "$BIN" "$VERSION" "$os" "$arch"
fetch "$base/$tarball" "$tmp/$tarball" || die "download failed: $base/$tarball"
fetch "$base/checksums.txt" "$tmp/checksums.txt" || die "download failed: $base/checksums.txt"

if command -v sha256sum >/dev/null 2>&1; then
  sha256() { sha256sum "$1" | cut -d' ' -f1; }
elif command -v shasum >/dev/null 2>&1; then
  sha256() { shasum -a 256 "$1" | cut -d' ' -f1; }
else
  die "need sha256sum or shasum to verify the download"
fi

want=$(grep " $tarball\$" "$tmp/checksums.txt" | cut -d' ' -f1)
[ -n "$want" ] || die "$tarball is not listed in checksums.txt"
got=$(sha256 "$tmp/$tarball")
[ "$want" = "$got" ] || die "checksum mismatch for $tarball (want $want, got $got)"

tar -xzf "$tmp/$tarball" -C "$tmp"
[ -f "$tmp/$BIN" ] || die "$BIN not found inside $tarball"

mkdir -p "$BINDIR"
install -m 0755 "$tmp/$BIN" "$BINDIR/$BIN" 2>/dev/null \
  || { cp "$tmp/$BIN" "$BINDIR/$BIN" && chmod 0755 "$BINDIR/$BIN"; }

printf 'install: %s -> %s\n' "$BIN" "$BINDIR/$BIN"

case ":$PATH:" in
  *":$BINDIR:"*) ;;
  *) printf 'install: %s is not on your PATH; add it to your shell profile\n' "$BINDIR" ;;
esac

printf '\nNext:\n  %s init\n' "$BIN"
