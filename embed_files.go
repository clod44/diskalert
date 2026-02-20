package main

import (
	"embed"
)

//go:embed public/* templates/index.html
var Assets embed.FS