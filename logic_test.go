// SPDX-FileCopyrightText : © 2025 Galvanized Logic Inc.
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"testing"
)

var tlogic = &logic{} // global for testing.

// Tests that the first 1 million games have unique deals.
func TestRandom(t *testing.T) {
	var maxGame uint    // swap init order for faster or more complete test
	maxGame = 1_000_000 // slower: ~2.0sec :: expanded number of games.
	maxGame = 32_000    // faster: ~0.2sec :: original number of games.
	allGames := map[string]uint{}
	for seed := uint(0); seed < maxGame; seed++ {
		deal := shuffle(seed, deck)
		key := ""
		for i := range deal {
			key += deal[i].Sym
		}

		// ensure that the game deal is unique
		if v, ok := allGames[key]; ok {
			t.Fatalf("duplicate game %d %d", v, seed)
		} else {
			allGames[key] = seed
		}
	}
}

// go test -run Shuffle
func TestShuffle(t *testing.T) {
	for seed, game := range games {
		deal := shuffle(seed, deck)
		for i := range game {
			if game[i] != deal[i].Sym {
				dumpDeck(deal)
				t.Fatalf("seed %d card:%d expected:%s got:%s ", seed, i, game[i], deal[i].Sym)
			}
		}
	}
}

// go test -run Next
func TestNextInFoundation(t *testing.T) {
	tlogic.NewGame(0)
	if !tlogic.isNextInFoundation(CLB, InvalidCard, getCard(AC)) {
		t.Errorf("expected true")
	}
}

// Check the random algorithm against published deals for a given seed.
// eg: https://freecellgamesolutions.com/fcs/?game=999999
var games = map[uint][]string{
	1: []string{
		"JD", "2D", "9H", "JC", "5D", "7H", "7C", "5H",
		"KD", "KC", "9S", "5S", "AD", "QC", "KH", "3H",
		"2S", "KS", "9D", "QD", "JS", "AS", "AH", "3C",
		"4C", "5C", "TS", "QH", "4H", "AC", "4D", "7S",
		"3S", "TD", "4S", "TH", "8H", "2C", "JH", "7D",
		"6D", "8S", "8D", "QS", "6C", "3D", "8C", "TC",
		"6S", "9C", "2H", "6H",
	},
	2: []string{
		"QD", "QC", "KC", "3C", "4C", "2C", "KD", "5C",
		"4D", "JD", "JS", "6H", "QS", "6D", "2D", "9C",
		"TD", "JC", "8C", "6C", "8S", "4S", "5D", "QH",
		"7S", "9D", "KS", "7C", "6S", "4H", "AC", "8H",
		"AH", "9S", "TC", "2S", "3S", "TS", "9H", "2H",
		"3H", "AD", "7H", "3D", "5H", "8D", "KH", "7D",
		"AS", "5S", "TH", "JH",
	},
	11_982: []string{ // the unsolvable game from the original 32_000.
		"AH", "AS", "4H", "AC", "2D", "6S", "TS", "JS",
		"3D", "3H", "QS", "QC", "8S", "7H", "AD", "KS",
		"KD", "6H", "5S", "4D", "9H", "JH", "9S", "3C",
		"JC", "5D", "5C", "8C", "9D", "TD", "KH", "7C",
		"6C", "2C", "TH", "QH", "6D", "TC", "4S", "7S",
		"JD", "7D", "8H", "9C", "2H", "QD", "4C", "5H",
		"KC", "8D", "2S", "3S",
	},
	31_999: []string{
		"JD", "JH", "AD", "QH", "KH", "6S", "6D", "JC",
		"AC", "TH", "AS", "8H", "9D", "2H", "8D", "6H",
		"AH", "7H", "7C", "5D", "7S", "6C", "QC", "JS",
		"9C", "3D", "5C", "4C", "2S", "8S", "3C", "7D",
		"5H", "8C", "4H", "TD", "TS", "3H", "4S", "KC",
		"TC", "4D", "9S", "2C", "KD", "9H", "KS", "5S",
		"QS", "2D", "QD", "3S",
	},
	999_999: []string{
		"AH", "9S", "3D", "6C", "8D", "8H", "QS", "TS",
		"KD", "3C", "2D", "6D", "5H", "QD", "2S", "4D",
		"9D", "3S", "6H", "9H", "QC", "JH", "AS", "JS",
		"3H", "7H", "2H", "7S", "JC", "5D", "TD", "TH",
		"6S", "4S", "9C", "5C", "8C", "8S", "4C", "TC",
		"7C", "AC", "KH", "2C", "5S", "KS", "AD", "4H",
		"QH", "KC", "JD", "7D",
	},
}
