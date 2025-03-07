package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// setupRoutes for the server.
func (s *Server) setupRoutes() {
	s.mux.Group(func(r chi.Router) {
		r.Use(middleware.Compress(5))
		r.Use(middleware.RealIP)

		r.Group(func(r chi.Router) {
			r.Use(middleware.SetHeader("Content-Type", "application/json"))

			Documents(r, s.db)
		})
	})
}
