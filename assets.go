package main

import _ "embed"

// appIconPNG is the icon the About dialog shows. mygo build reads the same
// PNG to make the app's own icon, and resourcesExclude in mygo.json keeps it
// out of the app's resources, so the binary carries it instead of a file
// beside the app.
//
//go:embed resources/icon.png
var appIconPNG []byte
