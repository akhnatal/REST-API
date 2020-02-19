package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

//The application we want to test
var a App

func TestMain(m *testing.M) {
	googleAPIKey, err := ioutil.ReadFile("apikey.txt") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	a = App{
		googleAPIKey: string(googleAPIKey),
	}
	a.Initialize(dbUser, dbPass, dbName)
	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

//Verify that the table we need for testing is available
func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

//Clean database
func clearTable() {
	a.DB.Exec("DELETE FROM orders")
	a.DB.Exec("ALTER TABLE orders AUTO_INCREMENT = 1")
}

//Query used to create the database table
const tableCreationQuery = `
CREATE TABLE IF NOT EXISTS orders
(
    id INT AUTO_INCREMENT PRIMARY KEY,
    distance INT NOT NULL,
	status CHAR(25) NOT NULL
)`

//Delete all records in the orders table and send a GET request to the /orders endpoint.
func TestEmptyTable(t *testing.T) {
	clearTable()
	req, _ := http.NewRequest("GET", "/orders?page=50&limit=10", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

//Test the HTTP response code
func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

//Manually add a new order to the database and, by accessing the correspondent endpoint,
//Check if the status code is 201 (resource created) and if the JSON response contains the correct information that was added
func TestCreateOrder(t *testing.T) {
	clearTable()

	payload := []byte(`{"origin": ["22.288017", "114.140835"],"destination": ["22.288039", "114.142345"]}`)

	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["status"] != "UNASSIGNED" {
		t.Errorf("Expected order status to be 'UNASSIGNED'. Got '%s'", m["status"])
	}
	// the distance is compared to 164.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["distance"] != 164.0 {
		t.Errorf("Expected order distance to be '164'. Got %v", m["distance"])
	}
	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected order ID to be '1'. Got %d", m["id"])
	}
}

//This test basically add a new order to the database and
//Check if the correct endpoint results in an HTTP response with status code 200 (success)
func TestListOrders(t *testing.T) {
	clearTable()
	addOrders(1)

	req, _ := http.NewRequest("GET", "/orders?page=0&limit=10", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

//Add a new order to the database for the tests
func addOrders(count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		statement := fmt.Sprintf("INSERT INTO orders(status, distance) VALUES('%s', %d)", basicStatus, 55)
		a.DB.Exec(statement)
	}
}

//Add a new user to the database and then we use the correct endpoint to update it
func TestTakeOrder(t *testing.T) {
	clearTable()
	addOrders(1)

	req, _ := http.NewRequest("GET", "/orders", nil)
	response := executeRequest(req)
	var originalOrder map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalOrder)

	payload := []byte(`{"status":"TAKEN"}`)

	req, _ = http.NewRequest("PATCH", "/orders/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalOrder["id"] {
		t.Errorf("Expected the id to remain the same (%d). Got %d", originalOrder["id"], m["id"])
	}

	if m["status"] == originalOrder["status"] {
		t.Errorf("Expected the status to change from %d to %d. Got %d", originalOrder["status"], m["status"], m["status"])
	}
}
