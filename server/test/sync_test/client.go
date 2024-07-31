package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Route represents a route information
type Route struct {
	Destination string `json:"destination"`
	NextHop     string `json:"next_hop"`
}

// RouteTable represents a local route table cache
type RouteTable struct {
	mu     sync.RWMutex
	routes map[string]Route
}

// NewRouteTable creates a new RouteTable
func NewRouteTable() *RouteTable {
	return &RouteTable{
		routes: make(map[string]Route),
	}
}

// GetRoute gets a route from the route table
func (rt *RouteTable) GetRoute(destination string) (Route, bool) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	route, exists := rt.routes[destination]
	return route, exists
}

// AddRoute adds a route to the route table
func (rt *RouteTable) AddRoute(route Route) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.routes[route.Destination] = route
}

// DataSyncManager manages data sync tasks
type DataSyncManager struct {
	rt *RouteTable
}

// NewDataSyncManager creates a new DataSyncManager
func NewDataSyncManager(rt *RouteTable) *DataSyncManager {
	return &DataSyncManager{
		rt: rt,
	}
}

// Start starts the data sync manager
func (dsm *DataSyncManager) Start() {
	go func() {
		for {
			dsm.syncData()
			time.Sleep(10 * time.Second) // Sync every 10 seconds
		}
	}()
}

// syncData performs the data sync task
func (dsm *DataSyncManager) syncData() {
	destination := "10.0.0.0/24"
	route, err := fetchRouteFromServer(destination)
	if err != nil {
		fmt.Printf("Failed to fetch route: %v\n", err)
		return
	}
	fmt.Printf("Fetched route: %+v\n", route)
	dsm.rt.AddRoute(route)
}

// fetchRouteFromServer fetches route information from the server
func fetchRouteFromServer(destination string) (Route, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/route?destination=%s", destination))
	if err != nil {
		return Route{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Route{}, fmt.Errorf("server returned status: %v", resp.Status)
	}

	var route Route
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		return Route{}, err
	}
	return route, nil
}

func main() {
	routeTable := NewRouteTable()
	dataSyncManager := NewDataSyncManager(routeTable)
	dataSyncManager.Start()

	// Simulate a request for a route
	destination := "10.0.0.0/24"
	route, exists := routeTable.GetRoute(destination)
	if !exists {
		fmt.Printf("Route for %s not found, waiting for sync\n", destination)
	} else {
		fmt.Printf("Route found: %+v\n", route)
	}

	// Wait for the data sync task to complete
	time.Sleep(15 * time.Second)

	// Check the route table again
	route, exists = routeTable.GetRoute(destination)
	if exists {
		fmt.Printf("Route found after sync: %+v\n", route)
	} else {
		fmt.Printf("Route for %s still not found\n", destination)
	}
}
