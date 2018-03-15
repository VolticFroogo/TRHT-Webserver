package main

import (
	"log"

	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/db"
	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/handler"
	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/middleware/myJWT"
)

func main() {
	if err := db.InitDB(); err != nil {
		log.Printf("Error initializing database: %v", err)
		return
	}

	if err := myJWT.InitKeys(); err != nil {
		log.Printf("Error initializing JWT keys: %v", err)
		return
	}

	handler.Start()
}
