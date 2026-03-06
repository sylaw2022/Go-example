package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sylaw/fullstack-app/internal/api/handlers"
	"github.com/sylaw/fullstack-app/internal/service"
)

// SetupRouter initializes and returns a chi router with all registered routes
func SetupRouter(userService service.UserService) *chi.Mux {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Simple CORS implementation for development
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Register routes
	r.Get("/health", handlers.HealthCheck)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		userHandler := handlers.NewUserHandler(userService)

		r.Route("/users", func(r chi.Router) {
			r.Get("/", userHandler.GetAll)
			r.Get("/{id}", userHandler.GetByID)
		})
	})

	return r
}
