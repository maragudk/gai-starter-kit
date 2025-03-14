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

func Search(mux chi.Router, db searcher, ai embedder) {
	mux.Get("/search", httph.ErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		q := r.URL.Query().Get("q")
		embedding, err := ai.EmbedString(r.Context(), q)
		if err != nil {
			return errors.Wrap(err, "error embedding")
		}

		chunks, err := db.Search(r.Context(), q, embedding)
		if err != nil {
			return errors.Wrap(err, "error searching")
		}

		for _, chunk := range chunks {
			_, _ = w.Write([]byte("- [" + chunk.Content + "](/documents/" + string(chunk.ID) + ")\n"))
		}

		return nil
	}))
}
