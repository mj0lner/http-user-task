package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"user-tasks/database"
	"user-tasks/handlers"
)

var JWTSecret []byte

func main() {
	database.ConnectDb()
	var secret = os.Getenv("JWT_SECRET")

	if secret == "" {
		log.Fatal("JWT_SECRET not set")
	}
	JWTSecret = []byte(secret)

	r := mux.NewRouter()
	fmt.Println("Server is running on :3000")

	// Регистрация пользователя
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	// Защищённые маршруты
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(handlers.AuthMiddleware)
	protected.HandleFunc("/users/{id}/status", handlers.GetUserStatus).Methods("GET")
	protected.HandleFunc("/users/leaderboard", handlers.GetLeaderboard).Methods("GET")
	protected.HandleFunc("/users/{id}/task/complete", handlers.CompleteTask).Methods("POST")
	protected.HandleFunc("/users/{id}/referrer", handlers.AddReferrer).Methods("POST")

	http.ListenAndServe(":3000", r)
}

// Обработчик для защищённого маршрута
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Welcome to the protected area!"})
}
