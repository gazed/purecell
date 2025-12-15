// SPDX-FileCopyrightText : © 2025 Galvanized Logic Inc.
// SPDX-License-Identifier: BSD-2-Clause

package main

// game.go creates and controls the 2D and 3D visible elements.

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log/slog"
	"math"
	"math/rand"
	"time"

	"github.com/gazed/vu"
	"github.com/gazed/vu/load"
	"github.com/gazed/vu/math/lin"
)

// game runs the freecell game, creating the visible models and
// using the logic update the game based on user actions.
type game struct {
	eng        *vu.Engine
	mx, my     int       // mouse positions
	dx, dy     int       // mouse delta
	ww, wh     int       // window dimensions
	save       *Save     // saved game data.
	logic      *logic    // game rules.
	state      int       // player action states.
	gameOver   bool      // game has been won
	seedSelect []int32   // captures the game select key presses.
	seedDial   int       // the game select speed dial progress.
	seed01     float64   // 0:1 random value based on seed
	gameStart  time.Time // used to track time since start.

	// 3D game models.
	scene *vu.Entity   // 3D root
	light *vu.Entity   // scene light
	cards []*vu.Entity // 3D deck cards
	piles []*vu.Entity // 3D placeholders for empty card piles.
	board *vu.Entity   // 3D background for the play surface.

	// 2D game UI.
	ui         *vu.Entity // 2D root
	undoButton *vu.Entity //
	prevButton *vu.Entity //
	nextButton *vu.Entity //
	seedButton *vu.Entity //
	unsolvable *vu.Entity // marks games that can't be won.
	scoreIcon  *vu.Entity // game score and previous highscore

	// game UI text
	text     *image.NRGBA // the text image update texture.
	number   *vu.Entity   // text display for the game seed.
	scores   *vu.Entity   // text display for the game score.
	infoInit bool         // set true after the first update.

	// animation: moving a card, or end game celebration.
	anim Animation // nil if no animation running.
}

const (
	// states are used to switch between actions
	PlayState   = 0 // playing the current game seed.
	SelectState = 1 // selecting a new game seed using digits.
	DialState   = 2 // selecting a new game seed using hold and press.

	// size of the cards.
	cardScale      = 0.06 // chosen by what looks good.
	cardWidth      = 11.4 // meters (from blender model)
	cardHeight     = 17.8 // meters (from blender model)
	halfCardWidth  = cardWidth * 0.5
	halfCardHeight = cardHeight * 0.5
	cardZ          = 0.0

	// size of UI text
	txtWidth, txtHeight = 192.0, 192.0

	// button press hold delay is the time needed to consider
	// a long press as a deliberate hold.
	holdDelay = 0.75 // seconds.
)

// createGame is called once on startup.
// Use seed 25904 (easy game) for testing.
func createGame(eng *vu.Engine, ww, wh int, save *Save) *game {
	gm := &game{eng: eng, ww: ww, wh: wh, save: save}
	gm.logic = &logic{}

	// load 2D assets
	eng.ImportAssets("icon.shd", "tint.shd")                          // shaders
	eng.ImportAssets("crown.png", "next.png", "prev.png", "undo.png") // buttons
	eng.ImportAssets("seed.png", "unsolvable.png")                    // more buttons
	eng.ImportAssets("48:hack.ttf")                                   // fonts

	// create the 2D UI
	gm.ui = eng.AddScene(vu.Scene2D)
	gm.undoButton = gm.ui.AddModel("shd:tint", "msh:icon", "tex:color:undo")
	gm.prevButton = gm.ui.AddModel("shd:tint", "msh:icon", "tex:color:prev")
	gm.nextButton = gm.ui.AddModel("shd:tint", "msh:icon", "tex:color:next")
	gm.seedButton = gm.ui.AddModel("shd:tint", "msh:icon", "tex:color:seed")
	gm.undoButton.SetColor(1, 1, 1, 1).SetLayer(1)
	gm.prevButton.SetColor(1, 1, 1, 1).SetLayer(1)
	gm.nextButton.SetColor(1, 1, 1, 1).SetLayer(1)
	gm.seedButton.SetColor(1, 1, 1, 1).SetLayer(1)
	gm.scoreIcon = gm.ui.AddModel("shd:icon", "msh:icon", "tex:color:crown").SetLayer(1)
	gm.unsolvable = gm.ui.AddModel("shd:icon", "msh:icon", "tex:color:unsolvable").SetLayer(3)

	// create the UI text using double buffered text.
	gm.text = image.NewNRGBA(image.Rect(0, 0, txtWidth, txtHeight))
	gm.scores = gm.ui.AddModel("shd:tint", "msh:icon", "fnt:hack24")
	gm.scores.SetColor(0, 0, 0, 1).SetLayer(2)
	gm.scores.AddUpdatableTexture(gm.eng, "scores", gm.text)
	gm.number = gm.ui.AddModel("shd:tint", "msh:icon", "fnt:hack48")
	gm.number.AddUpdatableTexture(gm.eng, "number", gm.text)
	gm.number.SetColor(0, 0, 0, 1).SetLayer(2)

	// load the 3D assets
	eng.ImportAssets("card.shd", "tex3D.shd", "board.shd")   // shaders
	eng.ImportAssets("card.glb")                             // card model
	eng.ImportAssets("FC.png", "FD.png", "FH.png", "FS.png") // textures
	eng.ImportAssets("empty.png")                            // more textures

	// creates card assets: card0 to card51, an empty pile,
	// and the foundation empty piles.
	gm.createCardAssets()

	// create the 3D scene
	gm.scene = eng.AddScene(vu.Scene3D)
	gm.light = gm.scene.AddLight(vu.DirectionalLight).SetAt(-1, -1, -2)
	gm.light.SetLight(1.0, 1.0, 1.0, 1.0) // RGB 0:1, Intensity 0:up

	// place a 3D board quad behind the cards.
	gm.board = gm.scene.AddModel("shd:board", "msh:quad")
	gm.board.SetColor(0, 0, 0, 1)
	gm.board.SetModelUniform("args4", []float32{float32(gm.ww), float32(gm.wh), 0.0, 0.0})

	// create 16 empty card pile spots. Textures created in game::createCardAssets
	pileTextures := []string{
		"card52", "card52", "card52", "card52", "card53", "card54", "card55", "card56",
		"card52", "card52", "card52", "card52", "card52", "card52", "card52", "card52",
	}
	gm.piles = make([]*vu.Entity, 16)
	for pid := range gm.piles {
		tex := pileTextures[pid]
		emptyPile := gm.scene.AddModel("shd:tex3D", "msh:card", "tex:color:"+tex)
		emptyPile.SetScale(cardScale, cardScale, 0.0)
		if pid >= int(FC) && pid <= int(FS) {
			emptyPile.SetScale(cardScale*1.05, cardScale*1.05, 0.0)
		}
		gm.piles[pid] = emptyPile
	}

	// create the cards.
	gm.cards = make([]*vu.Entity, KS+1)
	for cid := AC; cid <= KS; cid++ {
		tex := fmt.Sprintf("card%d", cid)
		card := gm.scene.AddModel("shd:card", "msh:card", "tex:color:"+tex)
		card.SetScale(cardScale, cardScale, cardScale).SetColor(1, 1, 1, 1)
		gm.cards[cid] = card
	}

	// fresh deal based on the current seed.
	gm.resetBoard()
	return gm
}

// Resize updates the window dimensions needed for ray picking.
func (gm *game) Resize(wx, wy, ww, wh int) {
	gm.ww, gm.wh = ww, wh

	// only need to save changes to the non-fullscreen location and size.
	if !(wx == 0 && wy == 0) &&
		!(wx == gm.save.Display.Wx && wy == gm.save.Display.Wy &&
			ww == gm.save.Display.Ww && wh == gm.save.Display.Wh) {
		gm.save.persistWindow(wx, wy, ww, wh)
	}

	// place the background to cover the app window behind the cards.
	fw, fh := float64(ww), float64(wh)
	gm.board.SetScale(fw, fh, 0.0).SetAt(0, 0, cardZ-0.5)
	gm.board.SetModelUniform("args4", []float32{float32(gm.ww), float32(gm.wh), 0.0, 0.0})

	// place the UI elements.
	// button sizes scale based on the available display width
	cx, cy := fw*0.5, fh*0.5           // center pixel location.
	xmin, _ := cx-fw*0.5, cy-fh*0.5    // top left pixel location.
	xmax, ymax := cx+fw*0.5, cy+fh*0.5 // bottom right pixel location.

	// buttons are a fraction of available width
	buttonSize := min(fw*0.4, 160.0)
	pixelGap := 40.0
	gm.undoButton.SetScale(buttonSize, buttonSize, 0).SetAt(xmin+0.5*buttonSize+pixelGap, ymax-buttonSize, 0)
	gm.prevButton.SetScale(buttonSize*0.5, buttonSize, 0).SetAt(xmax-2.75*buttonSize-pixelGap, ymax-buttonSize, 0)
	gm.nextButton.SetScale(buttonSize*0.5, buttonSize, 0).SetAt(xmax-0.25*buttonSize-pixelGap, ymax-buttonSize, 0)
	gm.seedButton.SetScale(buttonSize*2.0, buttonSize, 0).SetAt(xmax-1.5*buttonSize-pixelGap, ymax-buttonSize, 0)

	// place the score icon and text.
	textSize := buttonSize * 1.2
	sx, sy := cx, ymax-buttonSize*1.2
	gm.scoreIcon.SetScale(buttonSize*1.4, buttonSize*1.4, 0).SetAt(sx-buttonSize, sy, 0)
	gm.unsolvable.SetScale(buttonSize*1.4, buttonSize*1.4, 0).SetAt(sx-buttonSize, sy, 0)
	gm.unsolvable.Cull(true) // only shown if game is unsolvable.
	sx -= buttonSize * 0.68
	sy += buttonSize * 0.4
	gm.scores.SetAt(sx, sy, 0).SetScale(textSize, textSize, 0)

	// place the game ID text.
	textSize *= 1.5 // game ID is a bit larger.
	sx, sy, _ = gm.seedButton.At()
	sx += buttonSize * 0.08
	sy += buttonSize * 0.65
	gm.number.SetAt(sx, sy, 0).SetScale(textSize, textSize, 0)

	// reset the card piles
	for pid := range uint(16) {
		x, y, z := placePile(pid)
		gm.piles[pid].SetAt(x, y, z)
	}

	// handle different aspect ratios by adjusting the camera position.
	// Needed to handle fixed screen sizes like ipad 3:4 and iphone 9:16.
	// Note: heuristic works ok for most reasonable screen ratios.
	// The board height is ignored for the distance calculation.
	camHeight := -2.5 * fh / fw
	camDistance := gm.camToBoardDistance(10.5, 0.0, 90.0, fw/fh)
	gm.scene.Cam().SetAt(0.0, camHeight, camDistance)
}

// placePile positions the empty card piles.
func placePile(boardID uint) (x, y, z float64) {
	x, y, z = placeCard(boardID) // same x,y
	z = cardZ - 0.001            // behind all the other cards.
	return x, y, z
}

// placeCard returns the card position for a given board location.
// cards are in columns
func placeCard(boardID uint) (x, y, z float64) {
	xgap, ygap, zgap := 0.75, 0.96, 0.001
	xoff, yoff, zoff := -3.5, 0.0, cardZ
	if boardID > MAX_BOARD_ID {
		if boardID > HIDDEN_CARD {
			// hidden foundation card.
			boardID = boardID - HIDDEN_CARD
			zoff = zoff - 0.1
		} else {
			slog.Error("unexpected board location", "boardID", boardID)
			return 0, 0, 0
		}
	}
	row, col := float64(boardID/8), float64(boardID%8)

	// the cascade starts in the row 1, and the subsequent
	// rows are overlapped.
	if row > 0 {
		yoff -= 0.8
		ygap = 0.4
	}

	// calculate the card position.
	x = (xoff + col) * xgap // start left and go right for each col.
	y = yoff - row*ygap     // start top and go lower for each row.
	z = zoff + row*zgap     // start back and come closer for each row.
	return x, y, z
}

// ============================================================================
// Get the camera distance based on the target width and height. See:
// https://forum.defold.com/t/3d-adjust-camera-position-in-front-of-some-objects-on-window-resize-solved/71171/6
func (gm *game) camDistanceToTarget(FOVRadians, size float64) float64 {
	return (size * 0.5) / math.Tan(FOVRadians*0.5)
}

// horizontalFOV is returned in Radians.
func (gm *game) horizontalFOV(verticalFOV, aspectRatio float64) float64 {
	return 2.0 * math.Atan(math.Tan(verticalFOV*0.5)*aspectRatio)
}

// verticalFOV is in degrees.
func (gm *game) camToBoardDistance(width, height, verticalFOV, aspectRatio float64) float64 {
	horizontalFOV := gm.horizontalFOV(verticalFOV, aspectRatio)
	distanceForWidth := gm.camDistanceToTarget(horizontalFOV, width)
	distanceForHeight := gm.camDistanceToTarget(verticalFOV, height)
	return max(distanceForWidth, distanceForHeight)
}

func (gm *game) camVerticalDistanceToTargetTop(FOV, size float64) float64 {
	return (size * 0.5) * lin.Deg(math.Tan(lin.Rad(FOV)*0.5))
}

// ============================================================================
// Update is the application engine callback called once per
// engine tick where delta is the elapsed time since the last call.
func (gm *game) Update(eng *vu.Engine, in *vu.Input, delta time.Duration) {

	// check for serious problems.
	if eng.LoadErrors() {
		slog.Error("stopping due to asset loading errors")
		eng.Shutdown()
		return
	}

	// update user mouse moves.
	gm.dx, gm.dy = gm.mx-int(in.Mx), gm.my-int(in.My)
	gm.mx, gm.my = int(in.Mx), int(in.My)

	// update background shader
	timer := time.Since(gm.gameStart)
	ticker := timer.Seconds()
	gm.board.SetModelUniform("args4", []float32{float32(gm.ww), float32(gm.wh), float32(ticker), float32(gm.seed01)})

	// highlight buttons if over.
	gm.handleHover(gm.mx, gm.my)

	// handle one time key presses.
	for press := range in.Pressed {
		switch press {
		case vu.KQ: // quit game
			eng.Shutdown() // game is saved in main.
		case vu.KF11, vu.KF:
			// F11 is the standard window key for toggling fullscreen.
			// F is also commonly used.
			// macos Ctrl-Cmd-F is handled automatically by the macos window manager.
			eng.ToggleFullscreen()
			gm.save.Full = !gm.save.Full
			gm.save.persistFullScreen(gm.save.Full)

		case vu.KT: // play the end game effect.
			gm.anim = animateGameComplete(gm)
		}
	}

	// finish ongoing animations, ignoring user input until
	// the animation completes.
	if gm.anim != nil {
		gm.anim = gm.anim.Run(delta) // returns nil when complete.
		return
	}

	// Actions depend on game state
	switch gm.state {
	case SelectState:
		// select new game by typing a 6 digit game number
		gm.runSelect(eng, in, delta)
	case DialState:
		// select new game by holding down on the prev/next buttons.
		gm.runSpeedDial(eng, in, delta)
	case PlayState:
		// regular game play
		for press := range in.Pressed {
			switch {
			case press == vu.KML || press == vu.TOUCH:
				gm.handleButtonClick(gm.mx, gm.my)
				gm.handleCardClick()
			}
		}

		// react to continuous press events.
		for press, startPress := range in.Down {
			switch {
			case press == vu.KML || press == vu.TOUCH:
				timeDown := time.Now().Sub(startPress)
				gm.handleButtonHold(gm.mx, gm.my, timeDown)
			}
		}
		if gm.state == SelectState {
			gm.updateGameSeed("------")
			return // start running SelectState next update
		}
	default:
		slog.Debug("invalid game state", "state", gm.state)
	}

	// check if the game has finished.
	if !gm.gameOver {
		gm.gameOver = gm.logic.IsGameWon()
		if gm.gameOver {
			score := uint(gm.logic.MoveCount())
			slog.Info("game complete", "seed", gm.save.Seed, "score", score)

			// update the best score.
			if bestScore, ok := gm.save.Scores[gm.save.Seed]; ok {
				if score < bestScore {
					gm.save.Scores[gm.save.Seed] = score
					gm.save.persist()
				}
			} else {
				gm.save.Scores[gm.save.Seed] = score
				gm.save.persist()
			}
			gm.updateInfo()
			gm.anim = animateGameComplete(gm)
		}
	}

	// wait for the font to load before the initial text update.
	// Afterwards only need to update if it changes.
	if !gm.infoInit {
		gm.infoInit = gm.updateInfo()
	}
}

// reset the game to the default deal.
func (gm *game) resetBoard() {
	previousBoard := gm.logic.Board()
	gm.logic.NewGame(gm.save.Seed)
	gm.unsolvable.Cull(gm.logic.IsGameSolvable(gm.save.Seed))
	gm.gameStart = time.Now()
	gm.gameOver = false

	// generate a color for the board shader.
	r, g, b := gameColor(gm.save.Seed)
	gm.board.SetColor(r, g, b, 1.0)

	// generate a random faction based on the seed.
	gm.seed01 = gameSeedToFrac(gm.save.Seed)

	// update the stats
	gm.updateInfo()

	// animate the cards to the new positions.
	gm.anim = animateCardMoves(gm, previousBoard)
}

// redrawBoard redraws the current board state.
func (gm *game) redrawBoard() {
	gm.updateInfo() // update score.

	// place the cards.
	for cid, bid := range gm.logic.Board() {
		gm.cards[cid].SetColor(1, 1, 1, 1)
		gm.cards[cid].Cull(false)
		if bid >= HIDDEN_CARD {
			gm.cards[cid].Cull(true)
		} else {
			x, y, z := placeCard(bid)
			gm.cards[cid].SetAt(x, y, z)
		}
	}

	// highlight any selected cards.
	selected := gm.logic.GetSelected()
	sr, sg, sb := 1.0, 0.8, 0.0
	for _, cid := range selected {
		gm.cards[cid].SetColor(sr, sg, sb, 1)
	}
}

// updateInfo updates the game text.
func (gm *game) updateInfo() bool {
	line := 56.0 // pixel spacing between text lines.

	// get the scores
	score := fmt.Sprintf("%03d", gm.logic.MoveCount())
	prevScore := "---"
	if ps, ok := gm.save.Scores[gm.save.Seed]; ok {
		prevScore = fmt.Sprintf("%03d", ps)
	}

	// update the game score and seed
	draw.Draw(gm.text, gm.text.Bounds(), image.Transparent, image.Point{}, draw.Src)
	e1 := gm.scores.WriteImageText("hack48", score, 0, int(line*0), gm.text)
	e2 := gm.scores.WriteImageText("hack48", prevScore, 0, int(line*1.34), gm.text)
	gm.scores.UpdateTexture(gm.eng, gm.text)
	e3 := gm.updateGameSeed(fmt.Sprintf("%06d", gm.save.Seed))

	// return true if all the info was updated.
	// Expect false if the font is not yet loaded.
	return e1 == nil && e2 == nil && e3 == nil
}

// update the game seed
func (gm *game) updateGameSeed(gameSeed string) (err error) {
	draw.Draw(gm.text, gm.text.Bounds(), image.Transparent, image.Point{}, draw.Src)
	err = gm.number.WriteImageText("hack48", gameSeed, 0, 0, gm.text)
	gm.number.UpdateTexture(gm.eng, gm.text)
	return err
}

// process a player click.
func (gm *game) handleCardClick() {
	pick := gm.hitCard(gm.scene.Cam(), gm.ww, gm.wh, gm.mx, gm.my)
	switch {
	case pick >= EMPTY_PILE1 && pick <= EMPTY_PILE16:
		if gm.logic.Interact(pick) {
			gm.anim = animateCardMoves(gm, gm.logic.PreviousBoard())
			return
		}
		gm.redrawBoard()
	case pick >= AC && pick <= KS:
		if gm.logic.Interact(pick) {
			gm.anim = animateCardMoves(gm, gm.logic.PreviousBoard())
			return
		}
		gm.redrawBoard()
	case pick >= HIDDEN_CARD:
		gm.logic.clearSelected() // remove selection.
		gm.redrawBoard()
	default:
		slog.Error("not possible: dev error")
	}
}

// handleButtonClick checks for a player button click
// and calls the appropriate action if a button was clicked.
func (gm *game) handleButtonClick(mx, my int) {
	buttons := map[string]*vu.Entity{
		"undo": gm.undoButton,
		"prev": gm.prevButton,
		"next": gm.nextButton,
		"seed": gm.seedButton,
	}
	for name, button := range buttons {
		if !gm.overButton(button, mx, my) {
			continue // not over this button
		}

		// find which button was clicked.
		switch name {
		case "prev":
			if gm.save.Seed > 0 {
				gm.save.Seed = gm.save.Seed - 1
				gm.save.persistSeed(gm.save.Seed)
				gm.resetBoard()
			}
		case "next":
			if gm.save.Seed < MAX_SEED {
				gm.save.Seed = gm.save.Seed + 1
				gm.save.persistSeed(gm.save.Seed)
				gm.resetBoard()
			}
		case "seed":
			if numberpadExists {
				gm.state = SelectState
			}
		case "undo":
			if !gm.gameOver {
				gm.logic.Undo()
				gm.redrawBoard()
			}
		}
		break // done since buttons don't overlap.
	}
}

// return true if the mouse is over the given button.
func (gm *game) overButton(button *vu.Entity, mx, my int) bool {
	px, py := float64(mx), float64(my)
	sx, sy, _ := button.Scale()
	cx, cy, _ := button.At()
	hx, hy := sx*0.5, sy*0.5
	return px > cx-hx && px < cx+hx && py > cy-hy && py < cy+hy
}

// click and hold on the prev/next buttons to enter
// a mode to quickly change the game seed using only a mouse press.
func (gm *game) handleButtonHold(mx, my int, pressed time.Duration) {
	if gm.overButton(gm.prevButton, mx, my) && pressed.Seconds() > holdDelay {
		gm.seedDial = int(gm.save.Seed)
		gm.state = DialState // start decrementing the game seed.
	}
	if gm.overButton(gm.nextButton, mx, my) && pressed.Seconds() > holdDelay {
		gm.seedDial = int(gm.save.Seed)
		gm.state = DialState // start incrementing the game seed.
	}
}

// handleHover highlights buttons when the mouse is over them.
func (gm *game) handleHover(mx, my int) {
	buttons := map[string]*vu.Entity{
		"undo": gm.undoButton,
		"prev": gm.prevButton,
		"next": gm.nextButton,
	}
	if numberpadExists {
		buttons["seed"] = gm.seedButton
	}

	// set default button color
	for _, button := range buttons {
		button.SetColor(1, 1, 1, 1)
	}

	// highlight color if over a button.
	px, py := float64(mx), float64(my)
	for _, button := range buttons {
		sx, sy, _ := button.Scale()
		cx, cy, _ := button.At()
		hx, hy := sx*0.5, sy*0.5
		if px < cx-hx || py < cy-hy || px > cx+hx || py > cy+hy {
			continue // not over this button
		}

		// set hover color
		button.SetColor(1, 1, 0, 1)
		break // can only be over one button.
	}
}

// -------------------------------------------------------------------------
// runSelect: if game select is active, then collect 5 system digits and
// start that game
func (gm *game) runSelect(eng *vu.Engine, in *vu.Input, delta time.Duration) {
	for press := range in.Pressed {
		switch press {
		case vu.K0, vu.K1, vu.K2, vu.K3, vu.K4, vu.K5, vu.K6, vu.K7, vu.K8, vu.K9,
			vu.KP0, vu.KP1, vu.KP2, vu.KP3, vu.KP4, vu.KP5, vu.KP6, vu.KP7, vu.KP8, vu.KP9:
			gm.seedSelect = append(gm.seedSelect, press)
			seedStr, seed := parseSelectKeys(gm.seedSelect)
			gm.updateGameSeed(seedStr)

			// finish game select when there are 6 digits.
			if len(gm.seedSelect) == 6 {
				gm.save.persistSeed(seed)
				gm.resetBoard()
				gm.seedSelect = gm.seedSelect[:0]
				gm.state = gm.state &^ SelectState // exit select state
			}
		default:
			// any non-numeric key exits select state
			gm.seedSelect = gm.seedSelect[:0]
			gm.state = gm.state &^ SelectState // exit select state
			gm.redrawBoard()
		}
	}
}

// -------------------------------------------------------------------------
// runSpeedDial: if game speed dial is active, then churn the game seed
// until the button is released.
func (gm *game) runSpeedDial(eng *vu.Engine, in *vu.Input, delta time.Duration) {

	// update user mouse moves.
	ax, ay := math.Abs(float64(gm.dx)), math.Abs(float64(gm.dy))
	gm.mx, gm.my = int(in.Mx), int(in.My)

	// exit speed dial select if the button press is released.
	_, ok1 := in.Down[vu.KML]
	_, ok2 := in.Down[vu.TOUCH]
	if !ok1 && !ok2 {
		gm.save.persistSeed(uint(gm.seedDial))
		gm.resetBoard()
		gm.state = gm.state &^ DialState // exit dial state
	}

	// react to continuous press events.
	for press, _ := range in.Down {
		switch {
		case press == vu.KML || press == vu.TOUCH:
			switch {
			case gm.overButton(gm.prevButton, gm.mx, gm.my):
				gm.speedDial(ax, ay, -1)
			case gm.overButton(gm.nextButton, gm.mx, gm.my):
				gm.speedDial(ax, ay, 1)
			}
		default:
			// exit if any other key is pressed.
			gm.save.persistSeed(uint(gm.seedDial))
			gm.resetBoard()
			gm.state = gm.state &^ DialState // exit dial state
		}
	}
}

// speedDial handles rapidly incrementing or decrementing the game seed
// while in DialState.
// dir is 1 or -1 for increment and decrement
func (gm *game) speedDial(ax, ay float64, dir int) {
	exp := 2.5
	gm.seedDial = gm.seedDial + dir*int(math.Pow(ay, exp)) + dir*int(ax)
	if gm.seedDial <= 0 {
		gm.seedDial = 0
	}
	if gm.seedDial >= int(MAX_SEED) {
		gm.seedDial = int(MAX_SEED)
	}
	gm.updateGameSeed(fmt.Sprintf("%06d", gm.seedDial))
	if gm.seedDial == 0 || gm.seedDial == int(MAX_SEED) {
		gm.save.persistSeed(uint(gm.seedDial))
		gm.resetBoard()
		gm.state = gm.state &^ DialState // exit dial state
	}
}

// -------------------------------------------------------------------------

// createCardAssets by merging each card face with a common card back.
func (gm *game) createCardAssets() {

	// load the UV template for all cards.
	uvImg := getNRGBA("cardBase.png")

	// card front images are imported as image data and used to
	// create individual card UV textures.
	cardFaceNames := []string{
		"AC.png", "AD.png", "AH.png", "AS.png",
		"2C.png", "2D.png", "2H.png", "2S.png",
		"3C.png", "3D.png", "3H.png", "3S.png",
		"4C.png", "4D.png", "4H.png", "4S.png",
		"5C.png", "5D.png", "5H.png", "5S.png",
		"6C.png", "6D.png", "6H.png", "6S.png",
		"7C.png", "7D.png", "7H.png", "7S.png",
		"8C.png", "8D.png", "8H.png", "8S.png",
		"9C.png", "9D.png", "9H.png", "9S.png",
		"TC.png", "TD.png", "TH.png", "TS.png",
		"JC.png", "JD.png", "JH.png", "JS.png",
		"QC.png", "QD.png", "QH.png", "QS.png",
		"KC.png", "KD.png", "KH.png", "KS.png",

		// empty card piles
		"empty.png",

		// empty foundation piles.
		"FC.png", "FD.png", "FH.png", "FS.png",
	}

	// create card assets by combining the UV template with the card faces.
	cardAssets := []*load.ImageData{}
	copyPoint := image.Point{1, 174}
	for _, faceName := range cardFaceNames {

		// create new card UV image for each face.
		base := image.NewNRGBA(uvImg.Bounds())
		draw.Draw(base, uvImg.Bounds(), uvImg, image.ZP, draw.Src)
		faceImg := getNRGBA(faceName) // load the card face image.

		// combine the two into the final card UV texture.
		copyRect := image.Rectangle{copyPoint, copyPoint.Add(faceImg.Bounds().Size())}
		draw.Draw(base, copyRect, faceImg, image.ZP, draw.Src)

		// turn the image back into the engine image data.
		idata := &load.ImageData{}
		idata.Opaque = false
		idata.Width = uint32(base.Bounds().Size().X)
		idata.Height = uint32(base.Bounds().Size().Y)
		idata.Pixels = []byte(base.Pix)
		cardAssets = append(cardAssets, idata)
	}

	// upload all the card uv images into texture assets.
	gm.eng.MakeTextures("card", cardAssets)
}

// hitCard takes advantage that all the cards are facing the player
// along the Z axis. Converting the card corner world coordinates
// into screen coordinates gives a simple check with the mouse.
// The closer card is the picked card.
func (gm *game) hitCard(cam *vu.Camera, ww, wh, mx, my int) (cid uint) {
	// card corner offsets in world coordinates.
	hx, hy := halfCardWidth*cardScale, halfCardHeight*cardScale
	hitCard, hitZ := HIDDEN_CARD, -100.0 // no card hit

	// check the empty piles.
	for pid := uint(0); pid < 16; pid++ {
		wx, wy, wz := gm.piles[pid].At()

		// get the corner pixel coordinates.
		xtop, ytop := cam.Screen(wx-hx, wy+hy, wz, ww, wh)
		xbot, ybot := cam.Screen(wx+hx, wy-hy, wz, ww, wh)
		if mx < xtop || mx > xbot || my < ytop || my > ybot {
			continue // did not hit this card.
		}

		// card hit, pick the card if it is closer.
		if wz > hitZ {
			hitCard, hitZ = pid+100, wz
		}
	}

	// test the visible cards
	board := gm.logic.Board()
	for cid := AC; cid <= KS; cid++ {
		if board[cid] >= HIDDEN_CARD {
			continue // can't interact with hidden cards.
		}
		wx, wy, wz := gm.cards[cid].At()

		// get the corner pixel coordinates.
		xtop, ytop := cam.Screen(wx-hx, wy+hy, wz, ww, wh)
		xbot, ybot := cam.Screen(wx+hx, wy-hy, wz, ww, wh)
		if mx < xtop || mx > xbot || my < ytop || my > ybot {
			continue // did not hit this card.
		}

		// card hit, pick the card if it is closer.
		if wz > hitZ {
			hitCard, hitZ = cid, wz
		}
	}
	return hitCard
}

// getNRGBA loads a png image and returns an image.NRGBA.
func getNRGBA(name string) *image.NRGBA {
	cardData, err := load.DataBytes(name)
	if err != nil {
		slog.Error("missing cardUV.png asset")
		return image.NewNRGBA(image.Rect(0, 0, 0, 0))
	}
	imgData, err := png.Decode(bytes.NewReader(cardData))
	switch t := imgData.(type) {
	case *image.NRGBA:
		return t
	}
	slog.Error("invalid cardUV.png asset")
	return image.NewNRGBA(image.Rect(0, 0, 0, 0))
}

// parseSelectKeys turns a slice of numeric key presses into a number
// and a display string. Expects only digit keys.
func parseSelectKeys(keys []int32) (display string, number uint) {
	pre, num := "", ""
	for cnt := 0; cnt < 6-len(keys); cnt++ {
		pre = "_" + pre
	}
	for _, key := range keys {
		switch key {
		case vu.K0, vu.KP0:
			num += "0"
		case vu.K1, vu.KP1:
			num += "1"
		case vu.K2, vu.KP2:
			num += "2"
		case vu.K3, vu.KP3:
			num += "3"
		case vu.K4, vu.KP4:
			num += "4"
		case vu.K5, vu.KP5:
			num += "5"
		case vu.K6, vu.KP6:
			num += "6"
		case vu.K7, vu.KP7:
			num += "7"
		case vu.K8, vu.KP8:
			num += "8"
		case vu.K9, vu.KP9:
			num += "9"
		default:
			// only ever expecting digits.
		}
	}
	fmt.Sscanf(num, "%d", &number)
	return pre + num, number
}

// gameColor creates a random RGB base color on a seed.
// Use HSL to get random colors in a desired range.
// * hue        = 260-360, 0-60  : purple, red, yellow
// * saturation = 0:100 percentage, ie: 40-90%
// * lightness  = 0:100 percentage, ie: 40-70%
func gameColor(seed uint) (r, g, b float64) {
	rng := rand.New(rand.NewSource(int64(seed)))
	H := rng.Float64() * 360.0   // full range for hue.
	S := 0.9                     // lots of color saturation
	L := rng.Float64()*0.5 + 0.2 // 0.2 to 0.7 for some random lightness.
	r, g, b = HSLtoRGB(H, S, L)
	return r, g, b
}

// HSLtoRGB converts color space values.
// h is 0 to 360, S, L are percentages.
func HSLtoRGB(h, s, l float64) (r, g, b float64) {
	c := (1.0 - math.Abs(2.0*l-1.0)) * s
	x := c * (1.0 - math.Abs(math.Mod(h/60.0, 2)-1.0))
	switch {
	case 0 <= h && h < 60:
		r, g, b = c, x, 0
	case 60 <= h && h < 120:
		r, g, b = x, c, 0
	case 120 <= h && h < 180:
		r, g, b = 0, c, x
	case 180 <= h && h < 240:
		r, g, b = 0, x, c
	case 240 <= h && h < 300:
		r, g, b = x, 0, c
	case 300 <= h && h < 360:
		r, g, b = c, 0, x
	}
	m := l - c*0.5
	return r + m, g + m, b + m
}

// gameSeedToFrac generates a random value from the seed.
// The value is in the range [0..1).
func gameSeedToFrac(seed uint) (random float64) {
	rng := rand.New(rand.NewSource(int64(seed)))
	return rng.Float64()
}
