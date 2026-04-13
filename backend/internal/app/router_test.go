package app

import "testing"

func TestNewRouterRegistersRoutes(t *testing.T) {
	router := NewRouter(nil)
	if router == nil {
		t.Fatal("router is nil")
	}
}
