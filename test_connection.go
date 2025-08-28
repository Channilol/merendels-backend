package main

import (
	"log"
	"merendels-backend/config"
)

func main() {
	log.Println("Testing db connection")

	config.ConnectDatabase()

	if (config.DB != nil) {
		config.DB.Close()
		log.Println("Connection to db closed")
	}
}