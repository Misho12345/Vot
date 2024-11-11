package main

import (
	"backend/internal/database"
	"backend/internal/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"log"
	"net/http"
)

func main() {
	db, err := database.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	authHandler := handlers.NewAuthHandler(db)
	roomHandler := handlers.NewRoomHandler(10)
	roomHandler.Start()

	r := mux.NewRouter()

	r.HandleFunc("/api/auth/signup", authHandler.SignUp).Methods("POST")
	r.HandleFunc("/api/auth/signin", authHandler.SignIn).Methods("POST")
	r.HandleFunc("/api/ws", roomHandler.HandleWebSocket).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"Authorization"},
		AllowCredentials: true,
		Debug:            true,
	})

	handler := c.Handler(r)

	log.Printf("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
