package ws

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
		Name:        "updateFogNodePings",
		Method:      "POST",
		Pattern:     "/updateFogNodePings",
		HandlerFunc: UpdateFogNodePings,
		Queries:     []string{},
	},
	Route{
		Name:        "updateFogNodeResources",
		Method:      "POST",
		Pattern:     "/updateFogNodeResources",
		HandlerFunc: UpdateFogNodeResources,
		Queries:     []string{},
	},
	Route{
		Name:        "getFogNodeIPs",
		Method:      "GET",
		Pattern:     "/getFogNodeIPs",
		HandlerFunc: GetFogNodeIPs,
		Queries:     []string{},
	},
}
