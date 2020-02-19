package app

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (a *App) InitializeRouter() {
	//Initialize the router
	a.Router = mux.NewRouter()
	//Wire up the routes
	a.InitializeRoutes()
}

func (a *App) InitializeRoutes() {
	a.Router.HandleFunc("/orders", a.listOrders).Methods("GET")
	a.Router.HandleFunc("/orders", a.placeOrder).Methods("POST")
	//{id:[0-9]+} part of the path indicates that Gorilla Mux should treat process a URL only if the id is a number
	a.Router.HandleFunc("/orders/{id:[0-9]+}", a.takeOrder).Methods("PATCH")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}
