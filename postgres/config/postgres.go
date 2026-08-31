package config

import (
	"log"
	"os"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	once     sync.Once
	dbResult *gorm.DB
	dbErr    error
)

func GetPostgres() (*gorm.DB, error) {
	once.Do(func() {
		dsn := os.Getenv("DNS")

		log.Println("Conectando a PostgreSQL...")

		dbResult, dbErr = gorm.Open(
			postgres.Open(dsn),
			&gorm.Config{},
		)

		if dbErr != nil {
			log.Println("Error al conectar a PostgreSQL:", dbErr)
			return
		}

		sqlDB, err := dbResult.DB()
		if err != nil {
			dbErr = err
			return
		}
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(10)
		log.Println("Conexión a PostgreSQL inicializada correctamente")
	})

	return dbResult, dbErr
}
