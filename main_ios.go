// SPDX-FileCopyrightText : © 2025 Galvanized Logic Inc.
// SPDX-License-Identifier: BSD-2-Clause

//go:build ios || (ios && debug)

package main

// main_ios.go turns on console logging for any ios build.

import (
	"io"
	"log/slog"

	"github.com/gazed/vu"
)

// override the default setLogging to dump debugging logs directly
// to the console.
func init() {
	setLogging = func(w io.Writer) {
		slog.SetDefault(slog.New(slog.NewTextHandler(vu.ConsoleWriter(), &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

	// override hasNumberpad to false as there is no nice way
	// to enter digits on ios.
	numberpadExists = false
}
