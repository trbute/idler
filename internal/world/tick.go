package world

import "time"

func (cfg *ServerConfig) ProcessTicks() {
	// world := getWorld()
	for range time.Tick(cfg.TickRate) {
		// modify world state per tick
	}
}
