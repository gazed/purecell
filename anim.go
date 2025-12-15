// SPDX-FileCopyrightText : © 2025 Galvanized Logic Inc.
// SPDX-License-Identifier: BSD-2-Clause

package main

// anim.go applies animation effects to the models and ui.

import (
	"math"
	"time"
)

// Animation is a programatically controlled cut scene.
type Animation interface {

	// Run updates the Animation, returning the updated Animation.
	// delta is the elapsed time since the last Run.
	Run(delta time.Duration) Animation
}

// =============================================================================

// animation programatically controls a cut scene over a set period of time.
type animation struct {
	elapsed  time.Duration   // elapsed run time
	duration time.Duration   // total animation time in milliseconds
	intro    func()          // one time on start if not nil.
	during   func(t float64) // pass in lerp ratio.
	outro    func()          // one time on finish if not nil.
	next     Animation       // a followup animation.
}

// Run implements the Animation interface.
func (a *animation) Run(delta time.Duration) Animation {
	if a == nil {
		return nil // no animation
	}
	if a.elapsed == 0 && a.intro != nil {
		a.intro() // run once at start
	}

	// run animation
	a.elapsed += delta
	fract := min(1.0, float64(a.elapsed)/float64(a.duration))
	if a.elapsed < a.duration {
		if a.during != nil {
			a.during(fract)
		}
		return a
	}

	// animation is finished
	if a.outro != nil {
		a.outro() // run once at end.
	}

	// return the next animation if there is one.
	if a.next != nil {
		return a.next
	}
	return nil
}

// =============================================================================
// game animations

type move struct {
	from uint
	to   uint
}

// move one or more cards from one board position to another,
// ie: move a group of cards in the cascade to a new board position.
func animateCardMoves(gm *game, from [52]uint) Animation {
	a := &animation{elapsed: 0, duration: 200 * time.Millisecond, next: nil}

	// on start: find out which cards have moved.
	prev := from // copy array by value.
	moves := map[uint]move{}
	a.intro = func() {
		for i, bid := range gm.logic.board {
			cid := uint(i)
			switch {
			case bid >= HIDDEN_CARD:
				// don't animate existing foundation cards during gameplay.
			case prev[cid] >= HIDDEN_CARD && bid != prev[cid]:
				// animate foundation cards when changing to new game.
				moves[cid] = move{
					from: prev[cid] - HIDDEN_CARD,
					to:   bid,
				}
			case bid != prev[cid]:
				// regular card move
				moves[cid] = move{
					from: prev[cid],
					to:   bid,
				}
			}
		}
	}

	// during: move the cards from a to b.
	a.during = func(t float64) {

		// used to lift the card above the other cards while moving.
		sint := math.Sin(t * math.Pi) // 0 to 1.0 back to 0
		lift := 0.05 + 0.3*sint

		// move each card that changed.
		for cid, move := range moves {
			sax, say, saz := placeCard(move.from)
			sbx, sby, sbz := placeCard(move.to)
			sx := lerp(sax, sbx, t)
			sy := lerp(say, sby, t)
			sz := lerp(saz, sbz, t) + lift
			gm.cards[cid].SetAt(sx, sy, sz)
		}
	}

	// on end: redraw the latest board.
	a.outro = func() {
		gm.redrawBoard()

		// check if any cards can be auto moved to the foundation.
		// if so, then immediately run as the next animation.
		if gm.logic.AutoMoveCard() {
			gm.updateInfo()
			a.next = animateCardMoves(gm, gm.logic.PreviousBoard())

			// speed up sequential moves.
			an := a.next.(*animation)
			maxspeed := 90 * time.Millisecond
			slowdown := time.Duration(float64(a.duration) * 0.80)
			an.duration = max(maxspeed, slowdown)
		}
	}
	return a
}

// a very subdued "tada!" animation when the game is won.
func animateGameComplete(gm *game) Animation {
	a := &animation{elapsed: 0, duration: 5000 * time.Millisecond}
	r, g, b := gameColor(gm.save.Seed)

	// fade between regular background and end game background.
	a.during = func(t float64) {
		sint := math.Sin(t * math.Pi)        // 0 to 1.0 back to 0
		gm.board.SetColor(r, g, b, 1.0-sint) // 1 to 0.0 back to 1
	}

	// reset the regular background
	a.outro = func() {
		gm.board.SetColor(r, g, b, 1.0)
	}
	return a
}

// ============================================================================
// utility methods

// Precise method, which guarantees v = v1 when t = 1.
// This method is monotonic only when v0 * v1 < 0.
// Lerping between same values might not produce the same value
// from https://en.wikipedia.org/wiki/Linear_interpolation
func lerp(v0, v1, t float64) float64 {
	return (1-t)*v0 + t*v1
}
