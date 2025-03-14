package http

import (
	"app/model"
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"maragu.dev/errors"
	"maragu.dev/httph"
)

type searcher interface {
	Search(ctx context.Context, query string, embedding []byte) ([]model.Chunk, error)
}

type SearchResponse struct {
	Chunks []model.Chunk
}

func Search(mux chi.Router, db searcher, ai embedder) {
	mux.Get("/search", httph.JSONHandler(func(w http.ResponseWriter, r *http.Request, _ any) (*SearchResponse, error) {
		q := r.URL.Query().Get("q")
		embedding, err := ai.EmbedString(r.Context(), q)
		if err != nil {
			return nil, errors.Wrap(err, "error embedding")
		}

		chunks, err := db.Search(r.Context(), q, embedding)
		if err != nil {
			return nil, errors.Wrap(err, "error searching")
		}

		return &SearchResponse{Chunks: chunks}, nil
	}))
}
