package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	"googlemaps.github.io/maps"

	"context"
)

//This structure provides references to the router and the database
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname string) {
	//Create database connection
	connectionString := fmt.Sprintf("%s:%s@/%s", user, password, dbname)

	var err error
	a.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	//Initialize the router
	a.Router = mux.NewRouter()
	//Wire up the routes
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/orders", a.listOrders).Methods("GET")
	a.Router.HandleFunc("/orders", a.placeOrder).Methods("POST")
	//{id:[0-9]+} part of the path indicates that Gorilla Mux should treat process a URL only if the id is a number
	a.Router.HandleFunc("/orders/{id:[0-9]+}", a.takeOrder).Methods("PATCH")
}

func (a *App) placeOrder(w http.ResponseWriter, r *http.Request) {
	var c Coordinate
	var o Order

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload, Coordinates should be array of two strings")
		return
	}

	//Use Google Distance Matrix API to compute distance using coordinates
	o.Distance = computeDistance(c.Origin, c.Destination)
	o.Status = basicStatus

	defer r.Body.Close()

	if err := o.placeOrder(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, o)
}

func (a *App) takeOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var o Order

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&o); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}

	defer r.Body.Close()

	o.ID = id
	var success Order
	success.Status = "SUCCESS"

	if err := o.takeOrder(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, success)
}

func (a *App) listOrders(w http.ResponseWriter, r *http.Request) {

	limit, errl := strconv.Atoi(r.FormValue("limit"))
	page, errp := strconv.Atoi(r.FormValue("page"))

	if errl != nil {
		respondWithError(w, http.StatusBadRequest, "Limit should be a valid integer.")
		return
	}
	if errp != nil {
		respondWithError(w, http.StatusBadRequest, "Page should be a valid integer.")
		return
	}

	orders, err := listOrders(a.DB, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, orders)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func computeDistance(origin []string, destination []string) int {

	c, err := maps.NewClient(maps.WithAPIKey(googleAPIKey))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
	r := &maps.DistanceMatrixRequest{
		Origins:      []string{origin[0] + " " + origin[1]},
		Destinations: []string{destination[0] + " " + destination[1]},
	}

	route, err := c.DistanceMatrix(context.Background(), r)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	return route.Rows[0].Elements[0].Distance.Meters
}
