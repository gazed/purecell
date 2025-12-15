// SPDX-FileCopyrightText : © 2025 Galvanized Logic Inc.
// SPDX-License-Identifier: BSD-2-Clause

//go:build darwin || ios

package main

// apple (macos, ios) specfic save location.

import (
	"os"
	"path"
)

// directoryLocation gives the save file location for macos and ios.
func saveDir() string {
	return path.Join(os.Getenv("HOME"),
		"/Library/Application Support/com.galvanizedlogic.purefreecell/")
}
