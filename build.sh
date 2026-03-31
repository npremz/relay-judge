#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT_DIR="${1:-$ROOT_DIR/dist}"
BIN_NAME="relay-judge"
TARGET_GOOS="${GOOS:-$(go env GOOS)}"
TARGET_GOARCH="${GOARCH:-$(go env GOARCH)}"
BIN_EXT=""
ARCHIVE_NAME="$BIN_NAME-$TARGET_GOOS-$TARGET_GOARCH.tar.gz"
STAGING_DIR="$OUT_DIR/package-$TARGET_GOOS-$TARGET_GOARCH"
PACKAGE_DIR="$STAGING_DIR/$BIN_NAME-$TARGET_GOOS-$TARGET_GOARCH"

if [[ "$TARGET_GOOS" == "windows" ]]; then
  BIN_EXT=".exe"
fi

mkdir -p "$OUT_DIR"

CGO_ENABLED=0 GOOS="$TARGET_GOOS" GOARCH="$TARGET_GOARCH" go build -o "$OUT_DIR/$BIN_NAME$BIN_EXT" ./cmd/relay-judge
chmod 755 "$OUT_DIR/$BIN_NAME$BIN_EXT" || true
rm -rf "$OUT_DIR/subjects"
cp -R "$ROOT_DIR/subjects" "$OUT_DIR/subjects"

rm -rf "$STAGING_DIR"
mkdir -p "$PACKAGE_DIR"
cp "$OUT_DIR/$BIN_NAME$BIN_EXT" "$PACKAGE_DIR/$BIN_NAME$BIN_EXT"
cp -R "$ROOT_DIR/subjects" "$PACKAGE_DIR/subjects"
chmod 755 "$PACKAGE_DIR/$BIN_NAME$BIN_EXT" || true

rm -f "$OUT_DIR/$ARCHIVE_NAME"
tar -C "$STAGING_DIR" -czf "$OUT_DIR/$ARCHIVE_NAME" "$(basename "$PACKAGE_DIR")"
rm -rf "$STAGING_DIR"

echo "Built:"
echo "  target: $TARGET_GOOS/$TARGET_GOARCH"
echo "  $OUT_DIR/$BIN_NAME$BIN_EXT"
echo "  $OUT_DIR/subjects"
echo "  $OUT_DIR/$ARCHIVE_NAME"
