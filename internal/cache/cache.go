package cache

import (
	"github.com/LeoStars/wildberries-task/internal/models"
	"sync"
)

type Cache struct {
	sync.RWMutex
	Orders map[string]models.Order
}

func (c *Cache) Set(key string, value models.Order) {
	c.Lock()
	defer c.Unlock()
	c.Orders[key] = value
}

func (c *Cache) Get(key string) (models.Order, bool) {
	c.RLock()
	defer c.RUnlock()
	if item, ok := c.Orders[key]; ok {
		return item, true
	}
	return models.Order{}, false
}
