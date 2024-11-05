package main

import (
	"github.com/mmm-bank/card-manager-service/http"
	"github.com/mmm-bank/card-manager-service/storage"
	"log"
	"os"
)

func main() {
	addr := ":8080"
	s := storage.NewPostgresCards(os.Getenv("POSTGRES_URL"))
	server := http.NewCardService(s)

	log.Printf("Card manager server is running on port %s...", addr[1:])
	if err := http.CreateAndRunServer(server, addr); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}
