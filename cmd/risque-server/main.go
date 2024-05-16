package main

import (
	"log"
	"net/http"
	"risque-server/pkg/routers"
)

func main() {
    router := routers.NewRouter()
    log.Println("Server is running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}
