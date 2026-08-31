package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tcero76/business-engine/postgres/config"
	"github.com/tcero76/business-engine/postgres/services"
)

func main() {
	fmt.Println("business-engine-service started")

	db, err := config.GetPostgres()
	if err != nil {
		log.Fatal("Error de conexión a la BD", err.Error())
	}
	_ = services.NewInventoryService(db)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	fmt.Println("business-engine-service started")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
