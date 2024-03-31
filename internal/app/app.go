package app

import (
	"encoding/json"
	"fmt"
	"github.com/LeoStars/wildberries-task/internal/database"
	"github.com/LeoStars/wildberries-task/internal/models"
	"io/ioutil"
	"os"
)

func Run() {
	jsonFile, err := os.Open("models.json")
	if err != nil {
		fmt.Println(err)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var order models.Order
	err = json.Unmarshal(byteValue, &order)
	if err != nil {
		fmt.Println(err)
	}

	err = database.WriteOrderToDB(order)
	if err != nil {
		panic(err)
	}

	err = database.GetOrdersFromDB()
	if err != nil {
		panic(err)
	}

	err = database.DropOrderFromDB("b563feb7b284b6test")
	if err != nil {
		panic(err)
	}
}
