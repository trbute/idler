package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trbute/idler/server/internal/database"
)

func (cfg *ApiConfig) GetResourceNodeById(ctx context.Context, nodeID int32) (database.ResourceNode, error) {
	cacheKey := fmt.Sprintf("resource_node:%d", nodeID)
	
	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var node database.ResourceNode
		if json.Unmarshal([]byte(cached), &node) == nil {
			return node, nil
		}
	}
	
	node, err := cfg.DB.GetResourceNodeById(ctx, nodeID)
	if err != nil {
		return database.ResourceNode{}, err
	}
	
	if data, err := json.Marshal(node); err == nil {
		cfg.Redis.Set(ctx, cacheKey, data, 24*time.Hour)
	}
	
	return node, nil
}