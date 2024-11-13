package world

import (
	"time"

	"github.com/trbute/idler/internal/database"
)

type ServerConfig struct {
	DB       *database.Queries
	Platform string
	TickRate time.Duration
}

type item struct {
	id   int
	name string
}

type inventory struct {
	items []item
}

type user struct {
	username int
	actionId string
	inventory
}

type resourceNode struct {
	name     string
	resource item
}

type cell struct {
	xPosition     int
	yPosition     int
	users         []user
	resourceNodes []resourceNode
}

type world struct {
	grid  []cell
	items []item
}

func getWorld() *world {
	world := new(world)
	// worldbuilding logic
	return world
}
