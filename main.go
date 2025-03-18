package main

import (
	"frappuccino/db"
	"log"
)

func main() {
	_, err := db.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
}
