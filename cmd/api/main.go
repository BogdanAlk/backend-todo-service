package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BogdanAlk/backend-todo-service/internal/db"
	"github.com/BogdanAlk/backend-todo-service/internal/httpapi"
	"github.com/BogdanAlk/backend-todo-service/internal/tasks"
	"github.com/BogdanAlk/backend-todo-service/internal/users"
)

func main() {
	// App port
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// Root context
	ctx := context.Background()

	// Init database
	database, err := db.New(ctx)
	if err != nil {
		log.Fatalf("db init failed: %v", err)
	}
	defer database.Pool.Close()

	// Repositories & handlers
	usersRepo := users.NewRepository(database.Pool)
	authHandler := httpapi.NewAuthHandler(usersRepo)

	tasksRepo := tasks.NewRepository(database.Pool)
	tasksHandler := httpapi.NewTasksHandler(tasksRepo)
	// Router
	mux := http.NewServeMux()
	// Auth
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/auth/register", authHandler.Register)

	// Tasks (JWT protected)
	mux.Handle("/tasks", httpapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tasksHandler.List(w, r)
		case http.MethodPost:
			tasksHandler.Create(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/tasks/", httpapi.AuthMiddleware(http.HandlerFunc(tasksHandler.GetUpdateDelete)))
	// Protected test endpoint
	mux.Handle("/me", httpapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := httpapi.GetUserID(r)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf("user_id=%d", id)))
	})))
	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// DB health check
	mux.HandleFunc("/db-health", func(w http.ResponseWriter, r *http.Request) {
		pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := database.Pool.Ping(pingCtx); err != nil {
			http.Error(w, "db not ok", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("db ok"))
	})
	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
