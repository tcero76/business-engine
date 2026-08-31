package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tcero76/borrar/postgres/config"
	"github.com/tcero76/borrar/postgres/services"
)

func main() {
	fmt.Println("borrar-service started")

	db, err := config.GetPostgres()
	if err != nil {
		log.Fatal("Error de conexión a la BD", err.Error())
	}
	_ = services.NewInventoryService(db)
	db.Debug().Select(1)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
}
