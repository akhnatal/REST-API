package mysql

import (
	"database/sql"
	"fmt"
	"log"
)

func (a *App) InitializeDB(user, password, dbname string) {
	//Create database connection
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, dbHost, dbPort, dbname)

	var err error
	a.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}
}
