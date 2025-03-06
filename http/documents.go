package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"maragu.dev/httph"
)

type GetDocumentsResponse struct{}

func Documents(mux chi.Router) {
	mux.Get("/documents", httph.JSONHandler(func(w http.ResponseWriter, r *http.Request, _ any) (*GetDocumentsResponse, error) {
		return &GetDocumentsResponse{}, nil
	}))
}
