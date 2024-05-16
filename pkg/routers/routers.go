package routers

import (
	"risque-server/pkg/handlers"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
    router := mux.NewRouter()
    router.HandleFunc("/users", handlers.CreateUser).Methods("POST")
    router.HandleFunc("/users", handlers.GetUsers).Methods("GET")
    router.HandleFunc("/cityclimate", handlers.FetchCityClimate).Methods("GET")
    return router
}
