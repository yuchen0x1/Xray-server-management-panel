#!/bin/sh
set -eu

REPOSITORY=$1
TAG=$2
ARCH=$3
VARIANT=${4:-}
case "$REPOSITORY" in
    *[!a-zA-Z0-9_./-]* | "") echo 'Invalid release repository' >&2; exit 1 ;;
esac
printf '%s\n' "$TAG" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'
case "$ARCH/$VARIANT" in
    amd64/* | 386/* | arm64/*) PLATFORM=$ARCH; XRAY_NAME=$ARCH; MTG_NAME=$ARCH ;;
    arm/v6) PLATFORM=armv6; XRAY_NAME=arm32; MTG_NAME=arm ;;
    arm/v7) PLATFORM=armv7; XRAY_NAME=arm32; MTG_NAME=arm ;;
    *) echo "Unsupported release architecture: $ARCH/$VARIANT" >&2; exit 1 ;;
esac

ARCHIVE="x-ui-linux-$PLATFORM.tar.gz"
BASE="https://github.com/$REPOSITORY/releases/download/$TAG"
curl -fSL --retry 5 --retry-all-errors --connect-timeout 20 --max-time 300 -o "$ARCHIVE" "$BASE/$ARCHIVE"
curl -fSL --retry 5 --retry-all-errors --connect-timeout 20 --max-time 60 -o SHA256SUMS "$BASE/SHA256SUMS"
awk -v name="$ARCHIVE" '{ filename = $2; sub(/^\*/, "", filename); if (filename == name) { print; count++ } } END { if (count != 1) exit 1 }' SHA256SUMS > selected-checksum
sha256sum -c selected-checksum
mkdir build
tar -xzf "$ARCHIVE" -C build --strip-components=1
cp build/x-ui.sh ./x-ui.sh
if [ "$PLATFORM" != "$XRAY_NAME" ]; then
    mv "build/bin/xray-linux-$PLATFORM" "build/bin/xray-linux-$XRAY_NAME"
fi
if [ "$PLATFORM" != "$MTG_NAME" ]; then
    mv "build/bin/mtg-linux-$PLATFORM" "build/bin/mtg-linux-$MTG_NAME"
fi
test -s build/x-ui
test -s "build/bin/xray-linux-$XRAY_NAME"
test -s "build/bin/mtg-linux-$MTG_NAME"
rm "$ARCHIVE" SHA256SUMS selected-checksum
