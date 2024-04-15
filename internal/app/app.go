package app

import (
	c "github.com/LeoStars/wildberries-task/internal/cache"
	"github.com/LeoStars/wildberries-task/internal/database"
	"github.com/LeoStars/wildberries-task/internal/models"
	"github.com/LeoStars/wildberries-task/internal/server"
	"github.com/LeoStars/wildberries-task/internal/streaming"
	"log"
	"net/http"
)

func Run() {
	orders := make(map[string]models.Order)
	cache := c.Cache{Orders: orders}

	databaseOrders, err := database.GetOrdersFromDB()
	if err != nil {
		log.Fatalf("Could not get orders from database: %s", err)
	}
	for _, databaseOrder := range databaseOrders {
		cache.Set(databaseOrder.OrderUid, databaseOrder)
	}

	nc, err := streaming.Connect()
	if err != nil {
		log.Fatalf("Could not connect to NATS: %s", err)
	}

	defer nc.Close()

	err = streaming.Subscribe(nc, &cache)
	if err != nil {
		log.Fatalf("Could not subscribe to NATS: %s", err)
	}

	server.HttpHandlersStart(&cache)
	defer log.Fatal(http.ListenAndServe(":8080", nil))
}
