package main

type Order struct {
	ID       int    `json:"id,omitempty"`
	Distance int    `json:"distance,omitempty"`
	Status   string `json:"status"`
}

type Coordinate struct {
	Origin      []string `json:"origin"`
	Destination []string `json:"destination"`
}

var basicStatus = "UNASSIGNED"
