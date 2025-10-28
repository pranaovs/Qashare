package main

import (
	"log"

	"shared-expenses-app/db"
)

func main() {
	// Open database
	pool, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
}
