package streaming

import (
	"encoding/json"
	c "github.com/LeoStars/wildberries-task/internal/cache"
	"github.com/LeoStars/wildberries-task/internal/database"
	"github.com/LeoStars/wildberries-task/internal/models"
	"github.com/nats-io/nats.go"
)

func Connect() (*nats.Conn, error) {
	sc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return sc, err
	}
	return sc, nil
}

func Subscribe(sc *nats.Conn, cache *c.Cache) error {
	var data models.Order
	handler := func(msg *nats.Msg) {
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			return
		}
		err = database.WriteOrderToDB(data)
		if err != nil {
			return
		}
		cache.Set(data.OrderUid, data)
	}

	_, err := sc.Subscribe("foo", handler)

	if err != nil {
		return err
	}
	return nil
}
