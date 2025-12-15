// SPDX-FileCopyrightText : © 2025 Galvanized Logic Inc.
// SPDX-License-Identifier: BSD-2-Clause

//go:build windows

package main

// windows specfic save location.

import (
	"os"
	"path"
)

// saveDir gives the save file location for Windows.
// - win  : C:\Users\[USER]\AppData\Local\purefreecell\*
func saveDir() string {
	return path.Join(os.Getenv("LOCALAPPDATA"), "purefreecell/")
}
