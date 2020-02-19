package main

import (
	"fmt"
	"io/ioutil"
)

const (
	dbName = "db_sam"
	dbUser = "tester"
	dbPass = "test"
	//dbHost = "db"
	dbHost = "192.168.99.100"
	dbPort = "3306"
)

//Entry point for the app
func main() {

	googleAPIKey, err := ioutil.ReadFile("apikey.txt") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	a := App{
		googleAPIKey: string(googleAPIKey),
	}
	//Create database connection  and wire up the routes
	a.Initialize(dbUser, dbPass, dbName)
	//Start the application
	a.Run(":8080")
}
