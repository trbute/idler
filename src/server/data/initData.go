package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/trbute/idler/server/internal/database"
)

type DataConfig struct {
	DB *database.Queries
}

type Action struct {
	Name string `json:"name"`
}

type Item struct {
	Name   string `json:"name"`
	Weight int    `json:"weight"`
}

type ResourceNode struct {
	Name       string `json:"name"`
	ActionName string `json:"action_name"`
	Tier       int    `json:"tier"`
	Drops      []Drop `json:"drops"`
}

type Drop struct {
	Name   string `json:"name"`
	Chance int    `json:"chance"`
}

type Grid struct {
	PositionX     int      `json:"position_x"`
	PositionY     int      `json:"position_y"`
	ResourceNodes []string `json:"resource_nodes"`
}

type Version struct {
	Value string `json:"value"`
}

func (cfg *DataConfig) InitData() {
	version := Version{}
	cfg.loadJSONData("data/json/version.json", &version)

	dbVersion, err := cfg.DB.GetVersion(context.Background())
	if err != nil {
		fmt.Println("Error getting version", err)
		panic(err)
	}

	if version.Value == dbVersion.Value {
		fmt.Println("World up to date")
		return
	}

	Actions := []Action{}
	cfg.loadJSONData("data/json/actions.json", &Actions)
	cfg.StoreActions(Actions)

	Items := []Item{}
	cfg.loadJSONData("data/json/items.json", &Items)
	cfg.StoreItems(Items)

	ResourceNodes := []ResourceNode{}
	cfg.loadJSONData("data/json/resource_nodes.json", &ResourceNodes)
	cfg.StoreResourceNodes(ResourceNodes)

	GridItems := []Grid{}
	cfg.loadJSONData("data/json/grid.json", &GridItems)
	cfg.StoreGridItems(GridItems)

	cfg.DB.UpdateVersion(context.Background(), version.Value)
}

func (cfg *DataConfig) loadJSONData(path string, v interface{}) {
	jsonFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteValue, v)
}

func (cfg *DataConfig) StoreActions(actions []Action) {
	for _, action := range actions {
		cfg.DB.CreateAction(context.Background(), action.Name)
	}
}

func (cfg *DataConfig) StoreItems(items []Item) {
	for _, item := range items {
		cfg.DB.CreateItem(context.Background(), database.CreateItemParams{
			Name:   item.Name,
			Weight: int32(item.Weight),
		})
	}
}

func (cfg *DataConfig) StoreResourceNodes(resourceNodes []ResourceNode) {
	for _, resourceNode := range resourceNodes {
		action, err := cfg.DB.GetActionByName(context.Background(), resourceNode.ActionName)
		if err != nil {
			panic(err)
		}

		cfg.DB.CreateResourceNode(context.Background(), database.CreateResourceNodeParams{
			Name:     resourceNode.Name,
			ActionID: action.ID,
			Tier:     int32(resourceNode.Tier),
		})
		resourceNodeRecord, err := cfg.DB.GetResourceNodeByName(
			context.Background(),
			resourceNode.Name,
		)
		if err != nil {
			panic(err)
		}

		for _, drop := range resourceNode.Drops {
			item, err := cfg.DB.GetItemByName(context.Background(), drop.Name)
			if err != nil {
				panic(err)
			}

			cfg.DB.CreateResource(context.Background(), database.CreateResourceParams{
				ResourceNodeID: resourceNodeRecord.ID,
				ItemID:         item.ID,
				DropChance:     int32(drop.Chance),
			})
		}
	}
}

func (cfg *DataConfig) StoreGridItems(gridItems []Grid) {
	for _, gridItem := range gridItems {
		cfg.DB.CreateGridItem(context.Background(), database.CreateGridItemParams{
			PositionX: int32(gridItem.PositionX),
			PositionY: int32(gridItem.PositionY),
		})
		for _, resourceNode := range gridItem.ResourceNodes {
			resourceNode, err := cfg.DB.GetResourceNodeByName(context.Background(), resourceNode)
			if err != nil {
				panic(err)
			}

			cfg.DB.CreateResourceNodeSpawn(
				context.Background(),
				database.CreateResourceNodeSpawnParams{
					NodeID:    resourceNode.ID,
					PositionX: int32(gridItem.PositionX),
					PositionY: int32(gridItem.PositionY),
				},
			)
		}
	}
}
