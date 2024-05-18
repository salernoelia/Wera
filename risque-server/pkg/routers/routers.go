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
    router.HandleFunc("/meteoblue", handlers.FetchMeteoBlue).Methods("GET")
    router.HandleFunc("/speak", handlers.SpeakText).Methods("POST")
    router.HandleFunc("/weather", handlers.FetchAndSpeakWeatherData).Methods("GET")
    return router
}
