// SPDX-FileCopyrightText : © 2025 Galvanized Logic Inc.
// SPDX-License-Identifier: BSD-2-Clause

package main

// main.go initializes the game logic and starts the game engine.

import (
	"embed"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/gazed/vu"
	"github.com/gazed/vu/load"
)

// application build version set at build time using ldflags -X main.Version="x.x.x"
var Version = "x.x.x" // default if not set by build.

// setLogging logs to the data directory info.log file.
// setLogging can be overridden by debug or platform builds.
var setLogging func(w io.Writer) = func(w io.Writer) {
	slog.SetDefault(slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

// numberpadExists is true if the platform allows the player to type digits.
// This is needed for editing the game seed.
var numberpadExists = true // true for macos, windows. ios overrides to false.

// defaultSize returns reasonable screen size that works for macos and windows.
// This is over-written in the save file once the player resizes or repositions
// the window. The game prefers tall and narrow windows, ie: 9:16
var defaultSize func() (x, y, w, h int) = func() (x, y, w, h int) {
	// return 100, 100, 900, 1600 // 9x16 - ie: iphone
	// return 100, 100, 1500, 2000 // 3x4 - ie: ipad 13"
	return 100, 100, 1200, 1800 // 2x3 - ie: ipad mini, ipad 11"
}

// Game startup initializes the game systems and starts the
// game engine loop.
func main() {

	// initialize logging. Overwrite log file each run.
	logfile := savePath(saveDir(), "info.log")                  // create dir if necessary
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE, 0666) // overwrite previous log file
	if err != nil {
		slog.Error("log file open", "err", err)
		return
	}
	setLogging(f)
	defer f.Close()

	// override vu.load.ReadFile function to use embedded resources.
	load.ReadFile = embeddedReadFile

	// restore persistent game data, if any.
	launch := &launcher{}
	launch.save = newSave(saveDir(), "freecell.save")
	launch.save.restore()
	slog.Info("starting game", "seed", launch.save.Seed)

	// use default window size if there was no save data.
	// tall and narrow dimensions are preferred.
	firstLaunch := launch.save.Display.Ww == 0
	if firstLaunch {
		x, y, w, h := defaultSize()
		launch.save.persistWindow(x, y, w, h)
	}

	// set the window to the saved dimensions.
	dsp := launch.save.Display
	launch.wx, launch.wy = dsp.Wx, dsp.Wy
	launch.ww, launch.wh = dsp.Ww, dsp.Wh

	// initialize engine.
	eng, err := vu.NewEngine(
		vu.Windowed(),
		vu.Title("Pure Freecell"),
		vu.Size(int32(launch.wx), int32(launch.wy), int32(launch.ww), int32(launch.wh)),
		vu.Background(0.01, 0.01, 0.01, 1.0),
	)
	if err != nil {
		slog.Error("engine start", "err", err)
		return
	}

	// start the engine loop that calls Update.
	eng.SetResizeListener(launch) // get window resize callbacks.
	eng.Run(launch, launch)       // does not return
}

// -----------------------------------------------------------------------------
// launcher combines the game logic with the game save state.
type launcher struct {
	game           *game // rules and state.
	save           *Save // saved game state
	wx, wy, ww, wh int   // initial screen position
}

// Load is the application one time startup callback to create initial assets.
// It is called after the window has been initialized.
func (launch *launcher) Load(eng *vu.Engine) error {

	// update the saved screen size now that the display is available.
	x, y, w, h := eng.WindowSize()
	launch.wx, launch.wy = int(x), int(y)
	launch.ww, launch.wh = int(w), int(h)
	launch.save.persistWindow(int(x), int(y), int(w), int(h))
	slog.Info("window size", "x", x, "y", y, "w", w, "h", h)

	// restore full screen based on the game save.
	if launch.save.Full {
		eng.ToggleFullscreen()
	}

	// create the game controller
	launch.game = createGame(eng, launch.ww, launch.wh, launch.save)
	return nil
}

// Update is the application engine callback each game "tick".
func (launch *launcher) Update(eng *vu.Engine, in *vu.Input, delta time.Duration) {
	launch.game.Update(eng, in, delta)
}

// Resize is called by the engine when the window size changes.
func (launch *launcher) Resize(windowLeft, windowTop int32, windowWidth, windowHeight uint32) {
	wx, wy, ww, wh := int(windowLeft), int(windowTop), int(windowWidth), int(windowHeight)
	launch.game.Resize(wx, wy, ww, wh)
}

// -----------------------------------------------------------------------------
// bundle the game assets with the binary.
// NOTE: shaders need both *.shd and *.spv files.
//
//go:embed assets/images/*.png
//go:embed assets/models/*.glb
//go:embed assets/shaders/*.s*
//go:embed assets/fonts/*.ttf
var assets embed.FS

// embeddedReadFile used to override vu.load.ReadFile
func embeddedReadFile(filepath string) ([]byte, error) { return assets.ReadFile(filepath) }
