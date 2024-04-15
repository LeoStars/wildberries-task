package main

import (
	"encoding/json"
	c "github.com/LeoStars/wildberries-task/internal/cache"
	"github.com/LeoStars/wildberries-task/internal/database"
	"github.com/LeoStars/wildberries-task/internal/models"
	"github.com/LeoStars/wildberries-task/internal/streaming"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestDatabase(t *testing.T) {
	jsonFile, err := os.Open("models.json")
	if err != nil {
		t.Errorf("Не удалось открыть тестовый файл.")
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var order models.Order
	err = json.Unmarshal(byteValue, &order)
	if err != nil {
		t.Errorf("Не удалось считать данные о заказе. Неверный формат.")
	}

	nc, err := streaming.Connect()
	if err != nil {
		t.Errorf("Не удалось подключиться к серверу NATS.")
	}

	defer nc.Close()
	for i := 1; i <= 10; i++ {
		uid := uuid.New()
		order.OrderUid = strings.ReplaceAll(uid.String(), "-", "")
		err := database.WriteOrderToDB(order)
		if err != nil {
			t.Errorf("Could not get orders from database: %s", err)
		}
		orders := make(map[string]models.Order)
		cache := c.Cache{Orders: orders}
		cache.Set(order.OrderUid, order)

		err = nc.Publish("foo", byteValue)
		if err != nil {
			t.Errorf("Не удалось опубликовать данные в NATS-канал. Ошибка: %v", err)
		}

		err = database.DropOrderFromDB(order.OrderUid)

		if err != nil {
			t.Errorf("Could not remove orders from database: %s", err)
		}
	}
}
