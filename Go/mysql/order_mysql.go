package mysql

import (
	"database/sql"
	"errors"
	"fmt"
)

func placeOrder(o Order, db *sql.DB) (Order, error) {
	query := fmt.Sprintf("INSERT INTO orders(status, distance) VALUES('%s', '%d')", o.Status, o.Distance)

	_, err := db.Exec(query)
	if err != nil {
		return o, err
	}
	err = db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&o.ID)
	if err != nil {
		return o, err
	}
	return o, nil
}

//First SQL request to verify order exist and is not already taken,
//Second SQL request to update the order's status if it's not already taken
func takeOrder(o Order, db *sql.DB) error {
	query := fmt.Sprintf("SELECT status FROM orders WHERE id =%d;", o.ID)

	var oTest Order

	errl := db.QueryRow(query).Scan(&oTest.Status)

	if errl != nil {
		return errors.New("Order doesn't exist")
	}

	if oTest.Status == "TAKEN" {
		return errors.New("Order already TAKEN")
	}

	query = fmt.Sprintf("UPDATE orders SET status='%s' WHERE id=%d", o.Status, o.ID)
	_, err := db.Exec(query)
	return err
}

//Fetches records from the orders table and limits the number of records based on the limit value passed by parameter
//Page parameter determines how many records are skipped
func listOrders(db *sql.DB, page, limit int) ([]Order, error) {
	query := fmt.Sprintf("SELECT id, status, distance FROM orders LIMIT %d OFFSET %d", limit, page)
	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	orders := []Order{}

	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.Status, &o.Distance); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}
