package main

import (
	"backend/internal/database"
	"backend/internal/handlers"
	"github.com/rs/cors"

	"database/sql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

var db *sql.DB

func main() {
	var err error
	db, err = database.Open()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer database.Close(db)

	authHandler := handlers.NewAuthHandler(db)

	r := mux.NewRouter()
	r.HandleFunc("/signup", authHandler.SignUp).Methods("POST")
	r.HandleFunc("/signin", authHandler.SignIn).Methods("POST")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
