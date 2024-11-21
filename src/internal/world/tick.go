package world

import (
	"fmt"
	"time"
)

func (cfg *WorldConfig) ProcessTicks() {
	fmt.Println("starting tick")
	tick := time.NewTicker(time.Millisecond * 500)
	iter := 1
	for range tick.C {
		fmt.Printf("\ntick %v\n", iter)
		iter++
		for coord := range cfg.World.Grid {
			go cfg.processActions(coord)
		}
	}
}

func (cfg *WorldConfig) processActions(coord Coord) {
	chars := cfg.World.Grid[coord].Characters
	for _, char := range chars {
		switch char.ActionId {
		case 0:
			continue
		case 1:
			go cfg.processWoodCutting(coord, char.Name)
		}
	}
}

func (cfg *WorldConfig) processWoodCutting(coord Coord, charName string) {
	fmt.Printf("%v at %v, %v is choppin wood\n", charName, coord.PositionX, coord.PositionY)
}
