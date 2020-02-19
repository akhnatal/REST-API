package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	"googlemaps.github.io/maps"

	"context"
)

//This structure provides references to the router, the database and the google client
type App struct {
	Router       *mux.Router
	DB           *sql.DB
	clientGoogle *maps.Client
	googleAPIKey string
}

func (a *App) Initialize(user, password, dbname string) {
	//Create database connection
	a.InitializeDB(user, password, dbname)
	//Initialize the router
	a.InitializeRouter()

	var err error
	//Initialize a client that can request Google Maps API
	a.clientGoogle, err = maps.NewClient(maps.WithAPIKey(a.googleAPIKey))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) placeOrder(w http.ResponseWriter, r *http.Request) {
	var c Coordinate
	var o Order

	decoder := json.NewDecoder(r.Body)
	// Verify
	if err := decoder.Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload, Coordinates should be array of two strings")
		return
	}

	defer r.Body.Close()

	// Verify Coordinates are an array of 2 strings
	if len(c.Origin) != 2 || len(c.Destination) != 2 {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload, Coordinates in request must be an array of exactly two strings")
		return
	}

	// Verify Coordinates can be cast to float, to ensure they conatin digits
	for i := 0; i < len(c.Origin); i++ {
		if _, err := strconv.ParseFloat(c.Origin[i], 64); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload, Origin's string is not a number")
			return
		}

		if _, err := strconv.ParseFloat(c.Destination[i], 64); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload, Destination's string is not a number")
			return
		}
	}

	//Use Google Distance Matrix API to compute distance using coordinates
	value, err := a.computeDistance(c.Origin, c.Destination)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	o.Distance = value
	o.Status = basicStatus

	// Place the order in the database
	var errl error
	o, errl = placeOrder(o, a.DB)
	if errl != nil {
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

	// Update the order's status in the database
	if err := takeOrder(o, a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, success)
}

func (a *App) listOrders(w http.ResponseWriter, r *http.Request) {
	//Verify that page qnd limit are digits and page is not zero
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Limit should be a valid integer.")
		return
	}

	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Page should be a valid integer.")
		return
	}
	if page == 0 {
		respondWithError(w, http.StatusBadRequest, "Page should start with 1.")
		return
	}

	// Ask the list of order to the database
	orders, err := listOrders(a.DB, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, orders)
}

func (a *App) computeDistance(origin []string, destination []string) (int, error) {

	// Prepare the request for the google maps api
	r := &maps.DistanceMatrixRequest{
		Origins:      []string{origin[0] + " " + origin[1]},
		Destinations: []string{destination[0] + " " + destination[1]},
	}

	// Call the DistanceMatrix API
	route, err := a.clientGoogle.DistanceMatrix(context.Background(), r)
	if err != nil {
		return -1, err
	}

	// Extract the distance in meters and verify it's defferent from zero
	dist := route.Rows[0].Elements[0].Distance.Meters
	if dist == 0 {
		return -1, errors.New("Google Maps Distance Matrix API : Coordinates are wrong")
	}

	return dist, nil
}
