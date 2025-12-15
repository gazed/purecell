// SPDX-FileCopyrightText : © 2025 Galvanized Logic Inc.
// SPDX-License-Identifier: BSD-2-Clause

package main

// logic.go contains the game rules and game state.

import (
	"fmt"
	"log/slog"
	"slices"
)

const (
	// card color
	BLK uint = 0
	RED uint = 1

	// card suit
	CLB uint = 0
	DMD uint = 1
	HRT uint = 2
	SPD uint = 3

	// card rank
	ACES uint = 0
	TWOS uint = 1
	THRE uint = 2
	FOUR uint = 3
	FIVE uint = 4
	SIXS uint = 5
	SEVN uint = 6
	EGHT uint = 7
	NINE uint = 8
	TENS uint = 9
	JACK uint = 10
	QUEN uint = 11
	KING uint = 12

	// Card IDs
	AC uint = 0
	AD uint = 1
	AH uint = 2
	AS uint = 3
	C2 uint = 4
	D2 uint = 5
	H2 uint = 6
	S2 uint = 7
	C3 uint = 8
	D3 uint = 9
	H3 uint = 10
	S3 uint = 11
	C4 uint = 12
	D4 uint = 13
	H4 uint = 14
	S4 uint = 15
	C5 uint = 16
	D5 uint = 17
	H5 uint = 18
	S5 uint = 19
	C6 uint = 20
	D6 uint = 21
	H6 uint = 22
	S6 uint = 23
	C7 uint = 24
	D7 uint = 25
	H7 uint = 26
	S7 uint = 27
	C8 uint = 28
	D8 uint = 29
	H8 uint = 30
	S8 uint = 31
	C9 uint = 32
	D9 uint = 33
	H9 uint = 34
	S9 uint = 35
	TC uint = 36
	TD uint = 37
	TH uint = 38
	TS uint = 39
	JC uint = 40
	JD uint = 41
	JH uint = 42
	JS uint = 43
	QC uint = 44
	QD uint = 45
	QH uint = 46
	QS uint = 47
	KC uint = 48
	KD uint = 49
	KH uint = 50
	KS uint = 51

	// board positions
	FC uint = 4 // club foundation are built up ACE to KING
	FD uint = 5 // diamond foundation
	FH uint = 6 // heart foundation
	FS uint = 7 // spade foundation

	// hide cards using an invalid board location
	// By convention HIDDEN_CARD is only used to hide foundation cards,
	// and is added to the existing foundation board ID.
	HIDDEN_CARD uint = 9999 // used to hide buried foundation cards.
	NO_CARD     uint = 999  // used for empty slots

	// empty piles are indicated by 100+pileID
	EMPTY_PILE1  uint = uint(100)
	EMPTY_PILE16 uint = uint(115)

	// Each visible card has a board ID.
	// 0:167 gives 168 spots for 1 top row plus 20 cascade rows.
	MAX_BOARD_ID uint = 167

	// 1 million games starting at game 0.
	MAX_SEED uint = 999_999
)

// Deck is a sorted deck of playing cards.
// This remains constant and is used to create shuffled decks of cards.
var deck = [52]Card{
	{ID: AC, Suit: CLB, Rank: ACES, Color: BLK, Sym: "AC"},
	{ID: AD, Suit: DMD, Rank: ACES, Color: RED, Sym: "AD"},
	{ID: AH, Suit: HRT, Rank: ACES, Color: RED, Sym: "AH"},
	{ID: AS, Suit: SPD, Rank: ACES, Color: BLK, Sym: "AS"},
	{ID: C2, Suit: CLB, Rank: TWOS, Color: BLK, Sym: "2C"},
	{ID: D2, Suit: DMD, Rank: TWOS, Color: RED, Sym: "2D"},
	{ID: H2, Suit: HRT, Rank: TWOS, Color: RED, Sym: "2H"},
	{ID: S2, Suit: SPD, Rank: TWOS, Color: BLK, Sym: "2S"},
	{ID: C3, Suit: CLB, Rank: THRE, Color: BLK, Sym: "3C"},
	{ID: D3, Suit: DMD, Rank: THRE, Color: RED, Sym: "3D"},
	{ID: H3, Suit: HRT, Rank: THRE, Color: RED, Sym: "3H"},
	{ID: S3, Suit: SPD, Rank: THRE, Color: BLK, Sym: "3S"},
	{ID: C4, Suit: CLB, Rank: FOUR, Color: BLK, Sym: "4C"},
	{ID: D4, Suit: DMD, Rank: FOUR, Color: RED, Sym: "4D"},
	{ID: H4, Suit: HRT, Rank: FOUR, Color: RED, Sym: "4H"},
	{ID: S4, Suit: SPD, Rank: FOUR, Color: BLK, Sym: "4S"},
	{ID: C5, Suit: CLB, Rank: FIVE, Color: BLK, Sym: "5C"},
	{ID: D5, Suit: DMD, Rank: FIVE, Color: RED, Sym: "5D"},
	{ID: H5, Suit: HRT, Rank: FIVE, Color: RED, Sym: "5H"},
	{ID: S5, Suit: SPD, Rank: FIVE, Color: BLK, Sym: "5S"},
	{ID: C6, Suit: CLB, Rank: SIXS, Color: BLK, Sym: "6C"},
	{ID: D6, Suit: DMD, Rank: SIXS, Color: RED, Sym: "6D"},
	{ID: H6, Suit: HRT, Rank: SIXS, Color: RED, Sym: "6H"},
	{ID: S6, Suit: SPD, Rank: SIXS, Color: BLK, Sym: "6S"},
	{ID: C7, Suit: CLB, Rank: SEVN, Color: BLK, Sym: "7C"},
	{ID: D7, Suit: DMD, Rank: SEVN, Color: RED, Sym: "7D"},
	{ID: H7, Suit: HRT, Rank: SEVN, Color: RED, Sym: "7H"},
	{ID: S7, Suit: SPD, Rank: SEVN, Color: BLK, Sym: "7S"},
	{ID: C8, Suit: CLB, Rank: EGHT, Color: BLK, Sym: "8C"},
	{ID: D8, Suit: DMD, Rank: EGHT, Color: RED, Sym: "8D"},
	{ID: H8, Suit: HRT, Rank: EGHT, Color: RED, Sym: "8H"},
	{ID: S8, Suit: SPD, Rank: EGHT, Color: BLK, Sym: "8S"},
	{ID: C9, Suit: CLB, Rank: NINE, Color: BLK, Sym: "9C"},
	{ID: D9, Suit: DMD, Rank: NINE, Color: RED, Sym: "9D"},
	{ID: H9, Suit: HRT, Rank: NINE, Color: RED, Sym: "9H"},
	{ID: S9, Suit: SPD, Rank: NINE, Color: BLK, Sym: "9S"},
	{ID: TC, Suit: CLB, Rank: TENS, Color: BLK, Sym: "TC"},
	{ID: TD, Suit: DMD, Rank: TENS, Color: RED, Sym: "TD"},
	{ID: TH, Suit: HRT, Rank: TENS, Color: RED, Sym: "TH"},
	{ID: TS, Suit: SPD, Rank: TENS, Color: BLK, Sym: "TS"},
	{ID: JC, Suit: CLB, Rank: JACK, Color: BLK, Sym: "JC"},
	{ID: JD, Suit: DMD, Rank: JACK, Color: RED, Sym: "JD"},
	{ID: JH, Suit: HRT, Rank: JACK, Color: RED, Sym: "JH"},
	{ID: JS, Suit: SPD, Rank: JACK, Color: BLK, Sym: "JS"},
	{ID: QC, Suit: CLB, Rank: QUEN, Color: BLK, Sym: "QC"},
	{ID: QD, Suit: DMD, Rank: QUEN, Color: RED, Sym: "QD"},
	{ID: QH, Suit: HRT, Rank: QUEN, Color: RED, Sym: "QH"},
	{ID: QS, Suit: SPD, Rank: QUEN, Color: BLK, Sym: "QS"},
	{ID: KC, Suit: CLB, Rank: KING, Color: BLK, Sym: "KC"},
	{ID: KD, Suit: DMD, Rank: KING, Color: RED, Sym: "KD"},
	{ID: KH, Suit: HRT, Rank: KING, Color: RED, Sym: "KH"},
	{ID: KS, Suit: SPD, Rank: KING, Color: BLK, Sym: "KS"},
}

// InvalidCard used for debugging error cases.
var InvalidCard Card = Card{ID: NO_CARD, Sym: "--"}

// -----------------------------------------------------------------------------
// logic for Freecell controls the game rules and the
// positioning of the cards.
type logic struct {
	selected uint     // currently selected card 0-51.
	gameSeed uint     // unique game ID.
	deal     [52]Card // a shuffled standard playing deck of cards.

	// Track game state by mapping each card to a board location.
	// This encapsulates game state in a compact structure.
	// Empty spots are marked with NO_CARD.
	//   freecells    0,1,2,3 - empty, or a single card.
	//   foundations  4,5,6,7 - empty, or the foundation top card.
	//   cascade 1    8,16,24,...,160 -- space for 20 cards in a cascade.
	//   cascade 2    9,17,25,...,161
	//   cascade 3   10,18,26,...,162
	//   cascade 4   11,19,27,...,163
	//   cascade 5   12,20,28,...,164
	//   cascade 6   13,21,29,...,165
	//   cascade 7   14,22,30,...,166
	//   cascade 8   15,23,31,...,167
	board [52]uint // board locations for each card ID.

	// track player moves by saving board state after each move.
	// Add a player move each time a card is placed.
	// Get the previous game state each player undo.
	// Moves moves
	moves *moves // stack of board positions
}

// Start a new game of freecell based on the given game number seed.
// Initializes the game cards from the given seed.
// Expected to be called by the UI layer.
func (l *logic) NewGame(seed uint) {
	l.gameSeed = seed  // remember the game number for the UI.
	l.moves = &moves{} //
	l.clearSelected()  // start with nothing selected.

	// put the shuffled cards into the cascades.
	l.deal = shuffle(seed, deck)
	for cid := AC; cid <= KS; cid++ {
		l.board[l.deal[cid].ID] = cid + 8
	}

	// save the initial board position.
	l.moves.reset()
	l.moves.record(l.board)
}

// Ordered list of unsolvable freecell games.
// From: https://cards.fandom.com/wiki/FreeCell#Unsolvable_Combinations
var UnsolvableGames = []uint{
	11_982, 146_692, 186_216, 455_889,
	495_505, 512_118, 517_776, 781_948,
}

// IsGameSolvable returns true if the given game seed can be solved.
func (l *logic) IsGameSolvable(gameSeed uint) bool {
	_, found := slices.BinarySearch(UnsolvableGames, gameSeed)
	return !found
}

// IsGameWon returns true when all the kings are on the foundation piles.
func (l *logic) IsGameWon() bool {
	return l.board[KC] == FC && l.board[KD] == FD &&
		l.board[KH] == FH && l.board[KS] == FS
}

// Return the current number of moves. This is like keeping score.
// It is calculated as the number of available undos plus 2 times
// the number of undos that have been done (since each undo reduces
// the number of available undos)
// Don't count the initial board position.
func (l *logic) MoveCount() int {
	if l.moves.count() > 0 {
		return l.moves.count() - 1
	}
	return 0
}

// GetSelected returns the selected card and its cascade sequence.
// An empty vector is returned if nothing is selected.
// If selected is valid, and there is a sequence, then the sequence
// will be valid as well. A valid sequence means there are enough free spots
// to move it and that the sequence extends to the end of the cascade.
func (l *logic) GetSelected() (v []uint) {
	if !l.isSelectionActive() {
		return v
	}
	v = append(v, uint(l.selected)) // return at least the selected card.

	// return the selected card and its cascade sequence if one is available.
	maxCascade := 10     // prevent infinite loops if state is bad.
	cardID := l.selected // start at the selected card
	boardPosition := l.board[l.selected]
	if l.isCascade(boardPosition) {
		nextCardID := l.cardAt(boardPosition + 8)
		for nextCardID != NO_CARD && l.nextInSequence(getCard(cardID), getCard(nextCardID)) && len(v) < maxCascade {
			cardID = nextCardID
			boardPosition = l.board[cardID]
			nextCardID = l.cardAt(boardPosition + 8)
			v = append(v, uint(cardID))
		}
	}
	return v
}

// Undo the most recent move.
// Triggered the UI due to user action.
func (l *logic) Undo() {
	l.clearSelected()        // clear any picked cards
	l.board = l.moves.undo() // reset the board to the previous game state.
}

// Board returns the board positions for each card.
func (l *logic) Board() [52]uint { return l.board }

// PreviousBoard returns the previous board positions for each card.
func (l *logic) PreviousBoard() [52]uint {
	mv := l.moves
	if len(mv.stack) > 1 {
		return mv.stack[len(mv.stack)-2] // previous board.
	}
	return mv.stack[len(mv.stack)-1] // current board
}

// Interact handles a user action, either picking a card or placing a card.
// - pick: AC:KS for a card, EMPTY_PILE1:EMPTY_PILE16 for empty piles
//
// return true if one more cards was moved to a new location.
func (l *logic) Interact(pick uint) bool {
	if !l.canInteract(pick) {
		previousPick := l.selected
		l.clearSelected() // clear picked card...

		// try to select a new card if its not the same card.
		if pick != previousPick {
			if isCard(pick) && l.canInteract(pick) {
				l.selected = pick
			}
		}
		return false // no card was moved
	}

	// attempt to place the selected cards onto the picked card.
	// CanInteract has already validated the move.
	if l.isSelectionActive() {
		s := getCard(l.selected) // single selection, or top card in selected sequence.
		seq := l.GetSelected()   // selection sequence.
		l.clearSelected()        // clear selection.

		// selection sequence will be size 1 if there is only 1 card selected.
		switch {
		case pick >= EMPTY_PILE1 && pick <= EMPTY_PILE16:
			// place the picked card on an empty pile.
			// Note the UI communicates negative IDs for empty piles.
			pileID := pick - EMPTY_PILE1 // convert UI pick to pileID

			switch {
			case l.isFreecell(pileID) && len(seq) == 1:
				// place a single card in an empty freecell
				if l.emptyPile(pileID) {
					l.board[s.ID] = pileID
					l.moves.record(l.board)
					return true
				}

			case l.isFoundation(pileID) && len(seq) == 1:
				// place a single card on an empty foundation
				if s.Suit == pileID-4 { // pile must match card suit
					// if foundation pile is empty and the card is an ACE
					// of the suit for that foundation pile.
					if l.emptyPile(pileID) && s.Rank == ACES {
						l.board[s.ID] = pileID
						l.moves.record(l.board)
						return true
					}
				}

			case pileID >= 8 && pileID <= 15:
				// try placing a card or card sequence on an empty cascade
				// need to double check that the stack size is valid since the
				// empty cascade is being consumed by the move.
				if l.emptyPile(pileID) {
					if len(seq) > l.movableStackSize(true) {
						slog.Error("aborting sequence move")
						return false // ABORT move
					}
					l.board[seq[0]] = pileID
					for i := 1; i < len(seq); i++ {
						l.board[seq[i]] = l.board[seq[i-1]] + 8
					}
					l.moves.record(l.board)
					return true
				}
			}

		case l.isCard(pick):
			// place the picked card on the selected card.
			// canInteract has already validated the move.
			p := getCard(pick)
			boardPick := l.board[p.ID]

			switch {
			case l.isFoundation(boardPick) && len(seq) == 1:
				// for foundation cards, bury the previous top card
				// and make the picked card the top of the foundation pile.
				if s.Rank == p.Rank+1 {
					// hide the existing top foundation card.
					// selected card is the new foundation top.
					l.board[p.ID] = l.board[p.ID] + HIDDEN_CARD
					l.board[s.ID] = boardPick
					l.moves.record(l.board)
					return true
				}

			case l.isCascade(boardPick):
				// place a card or sequence of cards on a cascade.
				if l.nextInSequence(p, s) {
					// move selected card onto the picked card
					l.board[seq[0]] = l.board[p.ID] + 8

					// move the rest of the sequence, if there is a sequence.
					for i := 1; i < len(seq); i++ {
						l.board[seq[i]] = l.board[seq[i-1]] + 8
					}
					l.moves.record(l.board)
					return true
				}
			}
		}
		return false // no card was moved.
	}

	// there is no picked card, and the interaction is valid,
	// so assign a new picked card.
	if isCard(pick) {
		l.selected = pick
	}
	return false // no card was moved.
}

// Trys to move cards safely to the foundation.
// Returns true if one or more cards were moved.
// check if a card should be moved to the foundation.
//   - Aces are always moved up.
//   - 2's and up are only moved if previous rank are all up.
//
// Only moves one card at a time to let the UI control the flow.
// Returns true if a card was auto moved.
func (l *logic) AutoMoveCard() bool {

	// ignore auto moves until player has made the first move.
	if l.moves.count() < 2 {
		return false
	}

	// get the current top foundation cards. They may be empty.
	fc := getCard(l.cardAt(FC))
	fd := getCard(l.cardAt(FD))
	fh := getCard(l.cardAt(FH))
	fs := getCard(l.cardAt(FS))
	minRank := -1 // meaning one of the foundations is empty
	if fc.ID != NO_CARD && fd.ID != NO_CARD &&
		fh.ID != NO_CARD && fs.ID != NO_CARD {
		minRank = min(int(fc.Rank), int(fd.Rank), int(fh.Rank), int(fs.Rank))
	}

	// all selectable cards are candidates, some of these may be empty.
	candidates := []Card{
		getCard(l.cardAt(0)), // freecell cards
		getCard(l.cardAt(1)),
		getCard(l.cardAt(2)),
		getCard(l.cardAt(3)),
		l.lastInCascade(0), // cascade cards
		l.lastInCascade(1),
		l.lastInCascade(2),
		l.lastInCascade(3),
		l.lastInCascade(4),
		l.lastInCascade(5),
		l.lastInCascade(6),
		l.lastInCascade(7),
	}

	// check the 12 candidate cards
	// "hide" buried foundation cards.
	for _, c := range candidates {
		if c.ID == NO_CARD {
			continue // ignore empty piles
		}

		// can only move up if all of the previous ranks are up.
		if int(c.Rank) != minRank+1 && int(c.Rank) != minRank+2 {
			continue // ignore cards that can't move up.
		}

		// check if the card is next in the foundation.
		boardID := c.Suit + 4
		switch c.Suit {
		case CLB:
			if l.isNextInFoundation(c.Suit, fc, c) {
				if fc.ID != NO_CARD {
					// hide current top foundation card.
					l.board[fc.ID] = l.board[fc.ID] + HIDDEN_CARD
				}

				// move the candidate to the foundation.
				l.board[c.ID] = boardID
				l.moves.record(l.board)
				if l.isSelected(c.ID) {
					l.clearSelected()
				}
				return true
			}
		case DMD:
			if l.isNextInFoundation(c.Suit, fd, c) {
				if fd.ID != NO_CARD {
					// hide current top foundation card.
					l.board[fd.ID] = l.board[fd.ID] + HIDDEN_CARD
				}

				// move the candidate to the foundation.
				l.board[c.ID] = boardID
				l.moves.record(l.board)
				if l.isSelected(c.ID) {
					l.clearSelected()
				}
				return true
			}
		case HRT:
			if l.isNextInFoundation(c.Suit, fh, c) {
				if fh.ID != NO_CARD {
					// hide current top foundation card.
					l.board[fh.ID] = l.board[fh.ID] + HIDDEN_CARD
				}

				// move the candidate to the foundation.
				l.board[c.ID] = boardID
				l.moves.record(l.board)
				if l.isSelected(c.ID) {
					l.clearSelected()
				}
				return true
			}
		case SPD:
			if l.isNextInFoundation(c.Suit, fs, c) {
				if fs.ID != NO_CARD {
					// hide current top foundation card.
					l.board[fs.ID] = l.board[fs.ID] + HIDDEN_CARD
				}

				// move the candidate to the foundation.
				l.board[c.ID] = boardID
				l.moves.record(l.board)
				if l.isSelected(c.ID) {
					l.clearSelected()
				}
				return true
			}
		}
	}
	return false // no cards moved
}

// get the card at the given board location.
// Return NO_CARD if there is nothing there.
// location: 0-169 possible board locations for a card.
func (l *logic) cardAt(boardPosition uint) uint {
	for cid := AC; cid <= KS; cid++ {
		if l.board[cid] == boardPosition {
			return cid
		}
	}
	return NO_CARD // no card at location.
}

// isLastInCascade returns true if the given card is the
// last card in a cascade.
func (l *logic) isLastInCascade(cardID uint) bool {
	boardLocation := l.board[cardID]
	if boardLocation >= 8 && boardLocation <= MAX_BOARD_ID {
		nextInCascade := boardLocation + 8
		return l.cardAt(nextInCascade) == NO_CARD
	}
	return false // not in a cascade
}

// lastInCascade uses the cascadeID (0-7) to return the cardID of the
// last card in the indicated cascade.
func (l *logic) lastInCascade(cascadeID uint) (card Card) {
	for cid := AC; cid <= KS; cid++ {
		boardLocation := l.board[cid]
		if l.isLastInCascade(cid) && (cascadeID == boardLocation%8) {
			return deck[cid]
		}
	}
	return InvalidCard // cascades can be empty
}

// emptyPile returns true if there is no card in the
// indicated pile. Note that a cascade is empty if there
// is no card in the top spot.
// pileID: 0-15 one of the following board piles:
// - Freecell   : 0,1,2,3
// - Foundation : 4,5,6,7
// - Cascade    : 8,9,10,11,12,13,14,15
func (l *logic) emptyPile(pileID uint) bool {
	if pileID >= 0 && pileID <= 15 {
		for cid := AC; cid <= KS; cid++ {
			if l.board[cid] == pileID {
				return false
			}
		}
		return true
	}

	// developer error: should not reach here.
	slog.Error("invalid pile ID", "pileID", pileID)
	return false
}

// emptyFreeCells returns the number of empty free cells.
func (l *logic) emptyFreeCells() int {
	piles := []uint{0, 1, 2, 3}
	return l.countEmptyCells(piles)
}

// emptyCascades returns the number of empty cascade piles
func (l *logic) emptyCascades() int {
	piles := []uint{8, 9, 10, 11, 12, 13, 14, 15}
	return l.countEmptyCells(piles)
}

// countEmptyCells returns the number of empty piles.
func (l *logic) countEmptyCells(piles []uint) int {
	empty := 0
	for _, pileID := range piles {
		if l.emptyPile(pileID) {
			empty++
		}
	}
	return empty
}

// nextInSequence returns true if a can be placed on b in cascade,
// ie: returns true if Card b is 1 rank less than card a and is the opposite suit.
func (l *logic) nextInSequence(a, b Card) bool {
	return (b.Rank == (a.Rank - 1)) && b.Color != a.Color
}

// Card and Board position validation utilities.
func (l *logic) isCard(cardID uint) bool        { return cardID >= AC && cardID <= KS }
func (l *logic) isCascade(boardID uint) bool    { return boardID >= 8 && boardID <= MAX_BOARD_ID }
func (l *logic) isFoundation(boardID uint) bool { return boardID >= 4 && boardID <= 7 }
func (l *logic) isFreecell(boardID uint) bool   { return boardID >= 0 && boardID <= 3 }

// isNextInFoundation returns true if Card b is the next
// card that should be placed in the foundation pile for the given suit.
func (l *logic) isNextInFoundation(suit uint, a, b Card) bool {
	if suit > SPD {
		slog.Error("isNextInFoundation invalid suit")
		return false
	}
	onEmpty := a.ID == NO_CARD && b.Suit == suit && b.Rank == ACES
	onCard := a.ID != NO_CARD && b.Suit == suit && b.Rank == a.Rank+1
	return onEmpty || onCard
}

// getSequence attempts to return a valid cascade sequence for the given cardID.
// Returns empty vector if there is no valid cascade sequence.
// The sequence must end with the last card in the cascade.
// There must be enough free cells for the sequence size.
// Expected to be used to validate user picks.
func (l *logic) getSequence(cardID uint) (v []uint) {
	boardPosition := l.board[cardID]
	if l.isCascade(boardPosition) {
		v = append(v, cardID)
		nextCardID := l.cardAt(boardPosition + 8)
		for nextCardID != NO_CARD && l.nextInSequence(getCard(cardID), getCard(nextCardID)) {
			if len(v) >= 13 {
				slog.Error("getSequence loop safety trigger")
				break // prevent infinite loops in case of programming error.
			}
			v = append(v, nextCardID)
			boardPosition = l.board[nextCardID]
			cardID = nextCardID
			nextCardID = l.cardAt(boardPosition + 8)
		}

		// the last card of the sequence must be the last card in the cascade
		lastCard := v[len(v)-1]
		if l.cardAt(l.board[lastCard]+8) != NO_CARD {
			v = []uint{} // not a valid sequence.
			return v
		}

		// check the users desired stack size against the max allowed.
		needsEmptyCascade := !l.canMoveToCascade(v[0])
		if len(v) > l.movableStackSize(needsEmptyCascade) {
			v = []uint{} // not enough spots to move sequence.
		}
	} else if l.isFreecell(boardPosition) {
		v = append(v, cardID)
	}
	return v
}

// canMoveToCascade checks the last card of each cascade to see if
// the given card can be placed on it.
func (l *logic) canMoveToCascade(cardID uint) bool {
	c := getCard(cardID)
	for cascadeID := uint(0); cascadeID < 8; cascadeID++ {
		lastCardInCascade := l.lastInCascade(cascadeID)
		if lastCardInCascade.ID != NO_CARD {
			if l.nextInSequence(getCard(lastCardInCascade.ID), c) {
				return true
			}
		}
	}
	return false
}

// movableStackSize returns the maximum size of a movable card stack.
// Implies that the stack is being moved somewhere... either onto a card
// in another card or to an empty cascade. Based on logic from
// https://boardgames.stackexchange.com/questions/45155/freecell-how-many-cards-can-be-moved-at-once
//
// Currently choosing the more conservative max 1 empty cascade movable
// stack size rather than the pow(2, emptyCascadeCount)
// The formula has to adapt if the stack is being moved onto another non-empty cascade
// or if it is being moved to an empty cascade, reducing the movable stack size.
func (l *logic) movableStackSize(isEmptyCascadeUsed bool) int {
	emptyCascades := l.emptyCascades()
	if emptyCascades <= 0 {
		return l.emptyFreeCells() + 1
	}
	if isEmptyCascadeUsed {
		emptyCascades -= 1
	}
	if emptyCascades > 0 {
		extraCascades := emptyCascades - 1
		return 2 * (l.emptyFreeCells() + 1 + extraCascades)
	}
	return l.emptyFreeCells() + 1
}

// isSelected returns true if the indicated card has been selected
// for a move. This can include the cards in a cascade sequence.
// Expected to be used by the UI to highlight selected cards.
func (l *logic) isSelected(cardID uint) bool {
	cards := l.GetSelected()
	for _, cid := range cards {
		if cid == cardID {
			return true
		}
	}
	return false
}
func (l *logic) clearSelected()          { l.selected = NO_CARD }
func (l *logic) isSelectionActive() bool { return l.isCard(l.selected) }

// canInteract returns true for cards or piles that are a valid
// for a possible user move... either picking a card, or placing a card.
// * pick : 1:51 for a card, EMPTY_PILE1:EMPTY_PILE16 for empty piles
func (l *logic) canInteract(pick uint) bool {
	// check valid locations to place the selected card or cards.
	// When selection is active then "pick" is where the cards are going.
	if l.isSelectionActive() {
		return l.canPlaceCard(pick)
	}

	// nothing selected, so check if card can be selected.
	return l.canSelectCard(pick)
}

// canPlaceCard returns true if the picked card can be placed
// on another card or empty pile.
func (l *logic) canPlaceCard(pick uint) bool {
	selects := l.GetSelected()

	// consider the empty piles
	if pick >= EMPTY_PILE1 && pick <= EMPTY_PILE16 {
		s := getCard(selects[0])
		pileID := pick - EMPTY_PILE1

		// always valid to place a card on an empty freecell.
		if l.isFreecell(pileID) && len(selects) == 1 {
			return l.emptyPile(pileID)
		}

		// check placing a card on an empty foundation.
		// The card must be an ACE matching the foundation suit.
		if l.isFoundation(pileID) && len(selects) == 1 {
			return (s.Suit == pileID-4) && s.Rank == ACES
		}

		// always valid to place a card on an empty cascade.
		if pileID >= 8 && pileID <= 15 {
			return l.emptyPile(pileID)
		}

		// should not reach here.
		slog.Error("invalid card pick", "pick", pick)
		return false
	}

	// the user picked a card in order to place the
	// selected cards on the picked card.
	cardID := uint(pick)
	if l.isCard(cardID) {
		p := getCard(cardID)
		s := getCard(selects[0])
		boardPick := l.board[cardID]

		// if card is on a foundation pile, then it must be the next highest
		// card rank and the same suit. Only valid for single selected card.
		if l.isFoundation(boardPick) && len(selects) == 1 {
			suit := boardPick - 4
			return l.isNextInFoundation(suit, p, s)
		}

		// attempt to put the picked card onto the selected card.
		// The pick card must be the last in the cascade and it must be
		// the next highest rank and the opposite color from the top selected card.
		if l.isCascade(boardPick) {
			if l.isLastInCascade(cardID) {
				return l.nextInSequence(p, s)
			}
			return false
		}

		// a picked card can't interact with cards in the freecells.
		return false
	}

	// dev error: should never reach here
	slog.Error("invalid canPlaceCard pick", "pick", pick)
	return false
}

// canSelectCard returns true if the given board location has a selectable card.
// Can only pick the cards, not the empty piles.
// FUTURE: indicate when there are no available moves.
func (l *logic) canSelectCard(pick uint) bool {
	if !isCard(pick) {
		return false
	}
	boardPick := l.board[pick] // board location of the picked card.

	// foundation cards can never be picked up.
	// FUTURE: make this an option. Some implementations allow cards to
	//         be moved from the foundation back to the cascade.
	if l.isFoundation(boardPick) {
		return false
	}

	// check that the pick can be placed somewhere.
	if l.isCascade(boardPick) || l.isFreecell(boardPick) {
		seq := l.getSequence(pick)
		if len(seq) <= 0 {
			return false
		}
		c := getCard(seq[0]) // top card in picked sequence.

		// check valid moves for single selections
		if len(seq) == 1 {
			if l.emptyFreeCells() > 0 {
				return true // a single card can be moved to an empty cell.
			}

			// check if the card can be moved to a foundation pile.
			foundationPileID := c.Suit + 4
			if l.emptyPile(foundationPileID) && c.Rank == ACES {
				return true
			}
			topCard := getCard(l.cardAt(foundationPileID))
			if l.isNextInFoundation(c.Suit, topCard, c) {
				return true
			}
		}
		if l.emptyCascades() > 0 {
			return true // a valid sequence can be moved to an empty cascade
		}

		// check the last card of each cascade to see if the first
		// card in the sequence one can be placed on it.
		return l.canMoveToCascade(seq[0])
	}
	return false
}

// shuffle the deck based on the given seed.
func shuffle(seed uint, ordered [52]Card) (shuffled [52]Card) {
	deck := [52]uint{} // deck of 52 unique cards
	deal := [52]uint{} // ids of shuffled cards.

	// initialize the deck and deal.
	for cid := AC; cid <= KS; cid++ {
		deck[cid] = cid
		deal[cid] = NO_CARD
	}

	// shuffle
	dealt := 0            // cards dealt.
	remainder := uint(52) // remaining cards be dealt
	srand(seed)           // seed the random number generator.
	for i := 0; i < len(deck); i++ {
		j := randClassic() % remainder // choose a random card
		deal[dealt] = deck[j]          // deal the random card
		dealt += 1
		remainder -= 1
		deck[j] = deck[remainder] // remove dealt card.
	}

	// create and return the shuffled deck of cards.
	for i := 0; i < len(deal); i++ {
		shuffled[i] = ordered[deal[i]]
	}
	return shuffled
}

// -----------------------------------------------------------------------------
// Card represents a standard playing card.
// It mainly holds suit, rank, and color information.
// The card suit and rank are determined by ID where the
// card id is from 0 to 51. See Card::cardSym below.
type Card struct {
	ID    uint   // unique card id: 0 to 51
	Suit  uint   // 0-3  :: club, diamond, heart, spade.
	Rank  uint   // 0-12 :: ace, 2, 3,..., 10, J, Q, K.
	Color uint   // 0-1  :: black, red
	Sym   string // human readable unique ID.
}

// getCard returns (a copy of) the requested card (by value)
func getCard(cardID uint) Card {
	if isCard(cardID) {
		return deck[cardID]
	}
	return InvalidCard
}

// Return true if the card id is valid.
func isCard(cardID uint) bool { return cardID >= AC && cardID <= KS }

// -----------------------------------------------------------------------------
// moves records player moves, allowing undos.
// Records the board position of each card after each move.
// FUTURE: support Redos.
type moves struct {
	stack [][52]uint // each move is the board position of each card.
	undos int        // count number of player undos
}

// record the current board position.
// Array's are passed by value, so this is copy.
func (mv *moves) record(move [52]uint) {
	mv.stack = append(mv.stack, move) // push
}

// undo updates gamestate to the previous move.
// Always keep the initial game state where moves.size() == 1
func (mv *moves) undo() (previousBoard [52]uint) {
	if len(mv.stack) > 1 {
		mv.stack = mv.stack[:len(mv.stack)-1] // pop
		mv.undos += 1
	}
	return mv.stack[len(mv.stack)-1]
}

// reset clears all moves and resets move counters
func (mv *moves) reset() {
	mv.stack = [][52]uint{}
	mv.undos = 0
}

// count returns the number of moves.  This is the number of game moves
// plus twice the undo's since each undo removes a game move.
func (mv *moves) count() int {
	return len(mv.stack) + mv.undos*2
}

//--------------------------------------------------------------------------------------------------
// Reproduce the classic microsoft rand() function.
// From: https://rosettacode.org/wiki/Linear_congruential_generator#C++
//
// These are the original microsoft solitaire games for a given seed.
// There were originally 32,000 games. There is a testcase to check that
// the randomness supports 1_000_000 unique games.

var rseed uint = 0 // global seed
const RAND_MAX_32 = ((1 << 31) - 1)

// set the random number seed.
func srand(x uint) { rseed = x }

func randClassic() uint {
	rseed = (rseed*214013 + 2531011) & RAND_MAX_32
	return rseed >> 16
}

//--------------------------------------------------------------------------------------------------
// DEBUG utilities

// dumpDeck is only used for debugging.
func dumpDeck(deckOfCards [52]Card) {
	for cid, c := range deckOfCards {
		fmt.Printf("%s ", c.Sym)
		if (cid+1)%8 == 0 {
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n")
}

// dumpBoard is only used for debugging.
func dumpBoard(board [52]uint) {
	last := uint(0)
	for _, bid := range board {
		if bid < MAX_BOARD_ID && bid > last {
			last = bid
		}
	}
	for bid := range last + 1 {

		// get the card at the given board position.
		c := InvalidCard
		for cid := AC; cid <= KS; cid++ {
			if board[cid] == uint(bid) {
				c = deck[cid]
			}
		}
		fmt.Printf("%s ", c.Sym)
		if (bid+1)%8 == 0 {
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n")
}
