package routers

import (
	"server/pkg/handlers"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
    router := mux.NewRouter()
    router.HandleFunc("/cityclimate", handlers.FetchCityClimate).Methods("GET")
    router.HandleFunc("/meteoblue", handlers.FetchMeteoBlue).Methods("GET")
    router.HandleFunc("/llmtest", handlers.TestLLM).Methods("GET")
    router.HandleFunc("/speak", handlers.SpeakText).Methods("POST")
    router.HandleFunc("/weather", handlers.FetchAndSpeakWeatherData).Methods("GET")
    return router
}
