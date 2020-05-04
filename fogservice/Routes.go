package main

import (
	"github.com/gorilla/mux"

	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
	Queries     []string
}

type Routes []Route

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
		//Queries(route.Queries)
	}

	return router
}

var routes = Routes{
	Route{
		Name:        "ping",
		Method:      "GET",
		Pattern:     "/ping",
		HandlerFunc: Ping,
		Queries:     []string{},
	},
	Route{
		Name:        "getResources",
		Method:      "GET",
		Pattern:     "/getResources",
		HandlerFunc: GetResources,
		Queries:     []string{},
	},
}
