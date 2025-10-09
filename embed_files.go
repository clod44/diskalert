package main

import (
	"embed"
)

//go:embed public/*
var Assets embed.FS
