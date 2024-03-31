package database

import (
	"database/sql"
	"fmt"
	"github.com/LeoStars/wildberries-task/internal/models"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5433
	user     = "postgres"
	password = "postgres"
	dbname   = "wildberries"
)

func WriteOrderToDB(order models.Order) error {
	databaseInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", databaseInfo)
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	delivery := order.Delivery
	_, err = tx.Exec("INSERT INTO delivery (name, phone, zip, city, address, region, email) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7)", delivery.Name, delivery.Phone, delivery.Zip, delivery.City,
		delivery.Address, delivery.Region, delivery.Email)
	if err != nil {
		return err
	}

	payment := order.Payment
	_, err = tx.Exec("INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, "+
		"bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		payment.Transaction, payment.RequestId, payment.Currency, payment.Provider, payment.Amount, payment.PaymentDt,
		payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, "+
		"customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard, payment, delivery)"+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, (SELECT MAX(id) FROM payment),"+
		"(SELECT MAX(id) FROM delivery))", order.OrderUid, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId, order.DateCreated,
		order.OofShard)
	if err != nil {
		return err
	}

	items := order.Items
	for _, item := range items {
		_, err = tx.Exec("INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size, total_price, "+
			"nm_id, brand, status, order_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)", item.ChrtId,
			item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmId,
			item.Brand, item.Status, order.OrderUid)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		panic(err)
	}
	return nil
}

func GetOrdersFromDB() error {
	databaseInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", databaseInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM orders")
	if err != nil {
		return err
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		var order models.Order
		var delivery models.Delivery
		var payment models.Payment
		var deliveryId int
		var paymentId int
		items := make([]models.Item, 0)

		err := rows.Scan(&order.OrderUid, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerId, &order.DeliveryService, &order.Shardkey,
			&order.SmId, &order.DateCreated, &order.OofShard, &deliveryId, &paymentId)
		if err != nil {
			return err
		}

		rowDelivery, err := db.Query("SELECT * FROM delivery WHERE id = $1", deliveryId)
		if err != nil {
			return err
		}
		rowDelivery.Next()
		err = rowDelivery.Scan(&deliveryId, &delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City,
			&delivery.Address, &delivery.Region, &delivery.Email)
		if err != nil {
			return err
		}
		rowDelivery.Close()

		rowPayment, err := db.Query("SELECT * FROM payment WHERE id = $1", paymentId)
		if err != nil {
			return err
		}
		rowPayment.Next()
		err = rowPayment.Scan(&payment.RequestId, &payment.Currency, &payment.Provider, &payment.Amount,
			&payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
			&paymentId, &payment.Transaction)
		if err != nil {
			return err
		}
		rowPayment.Close()

		rowsItems, err := db.Query("SELECT * FROM items WHERE order_id = $1", order.OrderUid)
		if err != nil {
			return err
		}
		for rowsItems.Next() {
			var item models.Item
			var itemId int
			var orderId string
			err = rowsItems.Scan(&itemId, &item.ChrtId, &item.TrackNumber, &item.Price, &item.Rid,
				&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmId,
				&item.Brand, &item.Status, &orderId)
			if err != nil {
				return err
			}
			items = append(items, item)
		}
		rowsItems.Close()
		order.Delivery = delivery
		order.Payment = payment
		order.Items = items
		orders = append(orders, order)
	}

	fmt.Println("Successfully got orders!")
	fmt.Println(orders)
	return nil
}

func DropOrderFromDB(orderId string) error {
	databaseInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", databaseInfo)
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	rows, err := db.Query("SELECT * FROM orders WHERE order_uid = $1", orderId)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		var deliveryId int
		var paymentId int

		err := rows.Scan(&order.OrderUid, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerId, &order.DeliveryService, &order.Shardkey,
			&order.SmId, &order.DateCreated, &order.OofShard, &deliveryId, &paymentId)
		if err != nil {
			return err
		}

		_, err = tx.Exec("DELETE FROM items WHERE order_id = $1", orderId)
		if err != nil {
			return err
		}

		_, err = tx.Exec("DELETE FROM orders WHERE order_uid = $1", orderId)
		if err != nil {
			return err
		}

		_, err = tx.Exec("DELETE FROM delivery WHERE id = $1", deliveryId)
		if err != nil {
			return err
		}

		_, err = tx.Exec("DELETE FROM payment WHERE id = $1", paymentId)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
	return nil
}
