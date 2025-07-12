package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trbute/idler/server/internal/database"
)

func (cfg *ApiConfig) GetItemById(ctx context.Context, itemID int32) (database.Item, error) {
	cacheKey := fmt.Sprintf("item:%d", itemID)
	
	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var item database.Item
		if json.Unmarshal([]byte(cached), &item) == nil {
			return item, nil
		}
	}
	
	item, err := cfg.DB.GetItemById(ctx, itemID)
	if err != nil {
		return database.Item{}, err
	}
	
	if data, err := json.Marshal(item); err == nil {
		cfg.Redis.Set(ctx, cacheKey, data, 24*time.Hour)
	}
	
	return item, nil
}

func (cfg *ApiConfig) GetItemByName(ctx context.Context, name string) (database.Item, error) {
	cacheKey := fmt.Sprintf("item:name:%s", name)
	
	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var item database.Item
		if json.Unmarshal([]byte(cached), &item) == nil {
			return item, nil
		}
	}
	
	item, err := cfg.DB.GetItemByName(ctx, name)
	if err != nil {
		return database.Item{}, err
	}
	
	if data, err := json.Marshal(item); err == nil {
		cfg.Redis.Set(ctx, cacheKey, data, 24*time.Hour)
	}
	
	return item, nil
}