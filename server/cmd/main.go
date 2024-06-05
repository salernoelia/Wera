package main

import (
	"log"
	"net/http"

	"server/pkg/routers"

	"github.com/joho/godotenv"
)

func main() {
	 err := godotenv.Load()
	 if err != nil {
	         log.Fatal("Error loading .env file")
	 }
	


    router := routers.NewRouter()
    log.Println("Server is now accessible on the local network on port 8080")
    log.Fatal(http.ListenAndServe("0.0.0.0:8080", router))
}
