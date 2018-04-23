package main

// **********
// Taken from
// https://github.com/calmh/randomart
// and modded a bit to allow direct reading

// Package randomart generates OpenSSH style randomart images.

import (
	"bytes"
)

// Dimensions of the generated image.
const (
	XDim = 17
	YDim = 9
)

const (
	start = -1
	end   = -2
)

// Board is a generated randomart board.
type Board struct {
	Tiles    [YDim][XDim]int8
	title    string
	subtitle string
}

// Generate creates a Board to represent the given data by applying the drunken
// bishop algorithm.
func Generate(data []byte, title string) Board {
	return GenerateSubtitled(data, title, "")
}

func GenerateSubtitled(data []byte, title, subtitle string) Board {
	board := Board{title: title, subtitle: subtitle}
	var x, y int
	x = XDim / 2
	y = YDim / 2
	board.Tiles[y][x] = start

	for _, b := range data {
		for s := uint(0); s < 8; s += 2 {
			d := (b >> s) & 3
			switch d {
			case 0, 1:
				// Up
				if y > 0 {
					y--
				}
			case 2, 3:
				// Down
				if y < YDim-1 {
					y++
				}
			}
			switch d {
			case 0, 2:
				// Left
				if x > 0 {
					x--
				}
			case 1, 3:
				// Right
				if x < XDim-1 {
					x++
				}
			}
			if board.Tiles[y][x] >= 0 {
				board.Tiles[y][x]++
			}
		}
	}
	if board.Tiles[YDim/2][XDim/2] == 0 {
		board.Tiles[YDim/2][XDim/2] = start
	}
	board.Tiles[y][x] = end
	return board
}

// Returns the string representation of the Board, using the OpenSSH ASCII art
// character set.
func (board Board) String() string {
	var chars = []string{
		" ", ".", "o", "+", "=",
		"*", "B", "O", "X", "@",
		"%", "&", "#", "/", "^",
	}
	var buf bytes.Buffer

	if len(board.title) > 15 {
		board.title = board.title[:15]
	}

	writeTitle(&buf, board.title)

	for _, row := range board.Tiles {
		buf.WriteString("|")
		for _, c := range row {
			var s string
			if c == start {
				s = "S"
			} else if c == end {
				s = "E"
			} else if int(c) < len(chars) {
				s = chars[c]
			} else {
				s = chars[len(chars)-1]
			}
			buf.WriteString(s)
		}
		buf.WriteString("|\n")
	}

	writeTitle(&buf, board.subtitle)

	return buf.String()
}

func writeTitle(buf *bytes.Buffer, title string) {
	if title != "" {
		extraChars := len(title) + 2 - XDim
		if extraChars > 0 {
			title = title[:XDim-extraChars]
		}
		title = "[" + title + "]"
	}

	leftLen := (XDim - len(title)) / 2
	rightLen := XDim - len(title) - leftLen

	buf.WriteString("+")
	for i := 0; i < leftLen; i++ {
		buf.WriteString("-")
	}

	buf.WriteString(title)

	for i := 0; i < rightLen; i++ {
		buf.WriteString("-")
	}
	buf.WriteString("+\n")
}
