package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type Order struct {
	ID       int    `json:"id,omitempty"`
	Distance int    `json:"distance,omitempty"`
	Status   string `json:"status,omitempty"`
}

type Coordinate struct {
	Origin      []string `json:"origin,omitempty"`
	Destination []string `json:"destination,omitempty"`
}

var basicStatus = "UNASSIGNED"

func (o *Order) placeOrder(db *sql.DB) error {
	statement := fmt.Sprintf("INSERT INTO orders(status, distance) VALUES('%s', '%d')", o.Status, o.Distance)

	_, err := db.Exec(statement)
	if err != nil {
		return err
	}
	err = db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&o.ID)
	if err != nil {
		return err
	}
	return nil
}

//First SQL request to verify order exist and is not already taken,
//Second SQL request to update the order's status if it's not already taken
func (o *Order) takeOrder(db *sql.DB) error {
	statement := fmt.Sprintf("SELECT status FROM orders WHERE id =%d;", o.ID)

	var oTest Order

	err1 := db.QueryRow(statement).Scan(&oTest.Status)

	if err1 != nil {
		err1 := errors.New("Order doesn't exist")
		return err1
	}

	if oTest.Status == "TAKEN" {
		err := errors.New("Order already TAKEN")
		return err
	}

	statement = fmt.Sprintf("UPDATE orders SET status='%s' WHERE id=%d", o.Status, o.ID)
	_, err := db.Exec(statement)
	return err
}

//Fetches records from the orders table and limits the number of records based on the limit value passed by parameter
//Page parameter determines how many records are skipped
func listOrders(db *sql.DB, page, limit int) ([]Order, error) {
	statement := fmt.Sprintf("SELECT id, status, distance FROM orders LIMIT %d OFFSET %d", limit, page)
	rows, err := db.Query(statement)

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
