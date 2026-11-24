#!/bin/sh
set -eu

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "==> Building ZPad with MyGo..."
go tool mygo build

ARCH="$(uname -m)"
if [ "$ARCH" = "arm64" ]; then
  TARGET_DIR="build/darwin-arm64"
  SWIFT_TARGET="arm64-apple-macosx12.0"
else
  TARGET_DIR="build/darwin-amd64"
  SWIFT_TARGET="x86_64-apple-macosx12.0"
fi

APP="$TARGET_DIR/ZPad.app"
APPEX="$APP/Contents/PlugIns/ZPadFinderSync.appex"
SDK="$(xcrun --show-sdk-path)"
VERSION="$(/usr/bin/defaults read "$ROOT/$APP/Contents/Info.plist" CFBundleShortVersionString 2>/dev/null || echo "0.5.2")"

echo "==> Building ZPadFinderSync.appex ($SWIFT_TARGET)..."
rm -rf "$APPEX"
mkdir -p "$APPEX/Contents/MacOS" "$APPEX/Contents/Resources"

swiftc \
  -parse-as-library \
  -application-extension \
  -O \
  -target "$SWIFT_TARGET" \
  -sdk "$SDK" \
  -module-name ZPadFinderSync \
  -framework Cocoa \
  -framework FinderSync \
  -Xlinker -e -Xlinker _NSExtensionMain \
  -o "$APPEX/Contents/MacOS/ZPadFinderSync" \
  "$ROOT/macos/FinderSync/FinderSync.swift"

cp "$ROOT/macos/FinderSync/Info.plist" "$APPEX/Contents/Info.plist"
/usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString $VERSION" "$APPEX/Contents/Info.plist" >/dev/null
/usr/libexec/PlistBuddy -c "Set :CFBundleVersion $VERSION" "$APPEX/Contents/Info.plist" >/dev/null

if [ -f "$APP/Contents/Resources/AppIcon.icns" ]; then
  cp "$APP/Contents/Resources/AppIcon.icns" "$APPEX/Contents/Resources/AppIcon.icns"
fi

echo "==> Registering zpad:// URL scheme in ZPad.app Info.plist..."
APP_PLIST="$ROOT/$APP/Contents/Info.plist"
/usr/libexec/PlistBuddy -c "Delete :CFBundleURLTypes" "$APP_PLIST" >/dev/null 2>&1 || true
/usr/libexec/PlistBuddy -c "Add :CFBundleURLTypes array" "$APP_PLIST"
/usr/libexec/PlistBuddy -c "Add :CFBundleURLTypes:0 dict" "$APP_PLIST"
/usr/libexec/PlistBuddy -c "Add :CFBundleURLTypes:0:CFBundleURLName string com.xiaoz.zpad" "$APP_PLIST"
/usr/libexec/PlistBuddy -c "Add :CFBundleURLTypes:0:CFBundleURLSchemes array" "$APP_PLIST"
/usr/libexec/PlistBuddy -c "Add :CFBundleURLTypes:0:CFBundleURLSchemes:0 string zpad" "$APP_PLIST"

echo "==> Signing ZPadFinderSync.appex and ZPad.app..."
codesign --force --sign - --entitlements "$ROOT/macos/FinderSync/entitlements.plist" "$APPEX"
codesign --force --sign - "$APP"

if [ "${1:-}" = "--install" ]; then
  echo "==> Installing to /Applications/ZPad.app..."
  rm -rf /Applications/ZPad.app
  ditto "$APP" /Applications/ZPad.app
  /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -f /Applications/ZPad.app
  pluginkit -a /Applications/ZPad.app/Contents/PlugIns/ZPadFinderSync.appex || true
  killall ZPadFinderSync 2>/dev/null || true
  echo "==> Installed /Applications/ZPad.app with ZPadFinderSync.appex."
fi
