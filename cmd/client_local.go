package main

import (
	"flag"
	"github.com/google/uuid"
	"shoot-game/pkg/backend"
)

// Starts a local instance of the game with bots.

func main() {
	botNums := flag.Int("bots", 1, "设置对战机器人数量.")
	flag.Parse()

	currentPlayer := backend.Player{
		Name:           "Ipad",
		Icon:           'A',
		IdentifierBase: backend.IdentifierBase{UUID: uuid.New()},
		CurrentPosition: backend.Coordinate{
			X: -1,
			Y: -5,
		},
	}
	game := backend.NewGame()
	game.AddEntity(&currentPlayer)

	view := fro
}
