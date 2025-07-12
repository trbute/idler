package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trbute/idler/server/internal/database"
)

func (cfg *ApiConfig) GetResourcesByNodeId(ctx context.Context, nodeID int32) ([]database.Resource, error) {
	cacheKey := fmt.Sprintf("resources:node:%d", nodeID)
	
	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var resources []database.Resource
		if json.Unmarshal([]byte(cached), &resources) == nil {
			return resources, nil
		}
	}
	
	resources, err := cfg.DB.GetResourcesByNodeId(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	
	if data, err := json.Marshal(resources); err == nil {
		cfg.Redis.Set(ctx, cacheKey, data, 24*time.Hour)
	}
	
	return resources, nil
}

func (cfg *ApiConfig) GetResourceNodeSpawnsByCoordinates(ctx context.Context, x, y int32) ([]database.ResourceNodeSpawn, error) {
	cacheKey := fmt.Sprintf("resource_nodes:%d:%d", x, y)
	
	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var nodes []database.ResourceNodeSpawn
		if json.Unmarshal([]byte(cached), &nodes) == nil {
			return nodes, nil
		}
	}
	
	nodes, err := cfg.DB.GetResourceNodeSpawnsByCoordinates(ctx, database.GetResourceNodeSpawnsByCoordinatesParams{
		PositionX: x,
		PositionY: y,
	})
	if err != nil {
		return nil, err
	}
	
	if data, err := json.Marshal(nodes); err == nil {
		cfg.Redis.Set(ctx, cacheKey, data, 24*time.Hour)
	}
	
	return nodes, nil
}

func (cfg *ApiConfig) GetResourceNodeSpawnById(ctx context.Context, spawnID int32) (database.ResourceNodeSpawn, error) {
	cacheKey := fmt.Sprintf("resource_node_spawn:%d", spawnID)
	
	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var spawn database.ResourceNodeSpawn
		if json.Unmarshal([]byte(cached), &spawn) == nil {
			return spawn, nil
		}
	}
	
	spawn, err := cfg.DB.GetResourceNodeSpawnById(ctx, spawnID)
	if err != nil {
		return database.ResourceNodeSpawn{}, err
	}
	
	if data, err := json.Marshal(spawn); err == nil {
		cfg.Redis.Set(ctx, cacheKey, data, 24*time.Hour)
	}
	
	return spawn, nil
}