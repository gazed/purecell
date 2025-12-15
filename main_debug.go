// SPDX-FileCopyrightText : © 2025 Galvanized Logic Inc.
// SPDX-License-Identifier: BSD-2-Clause

//go:build debug

package main

// main_debug.go turns on debug logs when building with
// "go build -tags debug"

import (
	"io"
	"log/slog"
	"os"
)

// override the default setLogging to dump debugging logs directly
// to the console.
func init() {
	setLogging = func(w io.Writer) {
		// used to find loading and startup issues.
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}
}
