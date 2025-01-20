package main

import (
	"testing"

	"github.com/go-chi/chi"
	"github.com/taker0084/Go-BookingApp/internal/config"
)

func TestRoutes(t *testing.T){
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
		//do nothing
	default:
		t.Errorf("type is not *chi.Mux, type is %T", v)
	}

}