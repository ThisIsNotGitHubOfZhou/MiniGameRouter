package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Route represents a route information
type Route struct {
	Destination string `json:"destination"`
	NextHop     string `json:"next_hop"`
}

// RouteTable represents a route table
type RouteTable struct {
	routes map[string]Route
}

// NewRouteTable creates a new RouteTable
func NewRouteTable() *RouteTable {
	return &RouteTable{
		routes: make(map[string]Route),
	}
}

// AddRoute adds a route to the route table
func (rt *RouteTable) AddRoute(route Route) {
	rt.routes[route.Destination] = route
}

// GetRouteHandler handles the request for getting a route
func (rt *RouteTable) GetRouteHandler(w http.ResponseWriter, r *http.Request) {
	destination := r.URL.Query().Get("destination")
	route, exists := rt.routes[destination]
	if !exists {
		http.Error(w, "Route not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(route)
}

func main() {
	routeTable := NewRouteTable()
	routeTable.AddRoute(Route{Destination: "10.0.0.0/24", NextHop: "192.168.1.1"})

	http.HandleFunc("/route", routeTable.GetRouteHandler)
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
