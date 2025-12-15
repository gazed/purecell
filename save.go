// SPDX-FileCopyrightText : © 2025 Galvanized Logic Inc.
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"log/slog"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

// Save persists any game state that needs to be remembered between one
// game session and the next. Save needs to be public and visible for
// the encoding package.
type Save struct {
	file string // Save file name.

	// data saved to disk.
	Seed    uint `yaml:"seed"` // current game.
	Full    bool `yaml:"full"` // true if game is fullscreen.
	Display struct {
		Wx int `yaml:"wx"`
		Wy int `yaml:"wy"`
		Ww int `yaml:"ww"`
		Wh int `yaml:"wh"`
	} `yaml:"display,flow"` // last window location
	Scores map[uint]uint `yaml:"scores"` // high scores for completed games
}

// newSave creates default persistent application state. The directory
// is platform specific, eg: save_windows.go
// The default starting seed is 000001.
func newSave(dir, fname string) *Save {
	s := &Save{Seed: 1, Scores: map[uint]uint{}}
	s.file = savePath(dir, fname) //
	return s
}

// savePath returns the full path to the save file.
// The save directory is created if it does not exist.
func savePath(dir, fname string) string {
	if err := os.MkdirAll(dir, 0755); err != nil {
		dir = ""
	}
	return path.Join(dir, fname)
}

// persistWindow saves the new window location and size, while preserving
// the other information.
func (s *Save) persistWindow(x, y, w, h int) {
	s.Display.Wx, s.Display.Wy = x, y
	s.Display.Ww, s.Display.Wh = w, h
	s.persist()
}

// persistSeed saves the game number while preserving
// the other information.
func (s *Save) persistSeed(seed uint) {
	s.Seed = seed
	s.persist()
}

// persistFullscreen save the full screen preference while preserving
// the other information.
func (s *Save) persistFullScreen(fullScreen bool) {
	s.Full = fullScreen
	s.persist()
}

// persist is called to record any user preferences. This is expected
// to be called when a user preference changes.
func (s *Save) persist() {
	if data, err := yaml.Marshal(&s); err == nil {
		if err = os.WriteFile(s.file, data, 0644); err != nil {
			slog.Debug("save game state", "error", err)
		}
	} else {
		slog.Debug("encode game state", "error", err)
	}
}

// restore reads persisted information from disk.
// It handles the case where a previous restore file doesn't exist.
func (s *Save) restore() {
	if dbytes, err := os.ReadFile(s.file); err == nil {
		if err = yaml.Unmarshal(dbytes, s); err != nil {
			slog.Debug("restore game state", "error", err)
		}
	}
}
