package main

import (
	"log"

	"github.com/VolticFroogo/TRHT-Webserver/db"
	"github.com/VolticFroogo/TRHT-Webserver/handler"
	"github.com/VolticFroogo/TRHT-Webserver/middleware/myJWT"
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
