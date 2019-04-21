package main

import (
	"fmt"
	"io/ioutil"
)

const (
	dbName = "challenge_lalamove"
	dbUser = "tester"
	dbPass = "test"
	dbHost = "localhost"
	dbPort = "33066"
)

//Entry point for the app
func main() {

	googleAPIKey, err := ioutil.ReadFile("apikey.txt") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	a := App{}
	//Create database connection  and wire up the routes
	a.Initialize(dbUser, dbPass, dbName)
	//Start the application
	a.Run(":8080")
}
