package http

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"maragu.dev/errors"
	"maragu.dev/httph"

	"app/model"
	"app/sql"
)

type documentCRUDer interface {
	CreateDocument(ctx context.Context, d model.Document, chunks []model.Chunk) (model.Document, error)
	ListDocuments(ctx context.Context, opts sql.ListDocumentsOptions) ([]model.Document, error)
	GetDocument(ctx context.Context, id model.ID) (model.Document, error)
	UpdateDocument(ctx context.Context, d model.Document, chunks []model.Chunk) (model.Document, error)
	DeleteDocument(ctx context.Context, id model.ID) error
}

type embedder interface {
	EmbedString(ctx context.Context, s string) ([]byte, error)
}

func Documents(mux chi.Router, db documentCRUDer, ai embedder, log *slog.Logger) {
	mux.Post("/documents", httph.ErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return httph.HTTPError{Code: http.StatusBadRequest, Err: errors.Wrap(err, "error reading request body")}
		}

		doc := model.Document{
			Content: string(body),
		}

		chunks, err := doc.Chunk(r.Context(), ai.EmbedString)
		if err != nil {
			log.Info("Error creating document chunks", "error", err)
			return httph.HTTPError{Err: errors.Wrap(err, "error creating document chunks"), Code: http.StatusBadGateway}
		}

		if _, err := db.CreateDocument(r.Context(), doc, chunks); err != nil {
			log.Info("Error creating document", "error", err)
			return errors.Wrap(err, "error creating document")
		}

		w.WriteHeader(http.StatusCreated)

		return nil
	}))

	mux.Get("/documents", httph.ErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		limitStr := r.URL.Query().Get("limit")
		cursor := model.ID(r.URL.Query().Get("cursor"))

		var limit int
		if limitStr != "" {
			n, err := strconv.Atoi(limitStr)
			if err != nil {
				return httph.HTTPError{Err: errors.Wrap(err, "invalid limit"), Code: http.StatusBadRequest}
			}
			limit = n
		}

		docs, err := db.ListDocuments(r.Context(), sql.ListDocumentsOptions{
			Limit:  limit,
			Cursor: cursor,
		})
		if err != nil {
			log.Info("Error listing documents", "error", err)
			return errors.Wrap(err, "error listing documents")
		}

		// Write the document list as markdown links
		for _, doc := range docs {
			_, _ = w.Write([]byte("- [" + string(doc.ID) + "](/documents/" + string(doc.ID) + ")\n"))
		}

		// If we have documents and there might be more, include pagination hint
		if len(docs) > 0 && len(docs) == limit {
			lastID := docs[len(docs)-1].ID
			_, _ = w.Write([]byte("\n[Next Page](/documents?cursor=" + string(lastID) + "&limit=" + limitStr + ")\n"))
		}

		return nil
	}))

	mux.Get("/documents/{id:[a-z0-9_]+}", httph.ErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := model.ID(chi.URLParam(r, "id"))

		doc, err := db.GetDocument(r.Context(), id)
		if err != nil {
			if errors.Is(err, model.ErrorDocumentNotFound) {
				return httph.HTTPError{
					Code: http.StatusNotFound,
					Err:  errors.New("document not found"),
				}
			}

			log.Info("Error getting document", "error", err)
			return errors.Wrap(err, "error getting document")
		}

		_, _ = w.Write([]byte(doc.Content))

		return nil
	}))

	mux.Put("/documents/{id:[a-z0-9_]+}", httph.ErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := model.ID(chi.URLParam(r, "id"))

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return httph.HTTPError{Code: http.StatusBadRequest, Err: errors.Wrap(err, "error reading request body")}
		}

		doc := model.Document{
			ID:      id,
			Content: string(body),
		}

		chunks, err := doc.Chunk(r.Context(), ai.EmbedString)
		if err != nil {
			log.Info("Error creating document chunks", "error", err)
			return httph.HTTPError{Err: errors.Wrap(err, "error creating document chunks"), Code: http.StatusBadGateway}
		}

		if _, err = db.UpdateDocument(r.Context(), doc, chunks); err != nil {
			if errors.Is(err, model.ErrorDocumentNotFound) {
				return httph.HTTPError{
					Code: http.StatusNotFound,
					Err:  errors.New("document not found"),
				}
			}

			log.Info("Error updating document", "error", err)
			return errors.Wrap(err, "error updating document")
		}

		_, _ = w.Write(body)

		return nil
	}))

	mux.Delete("/documents/{id:[a-z0-9_]+}", httph.ErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := model.ID(chi.URLParam(r, "id"))

		if err := db.DeleteDocument(r.Context(), id); err != nil {
			if errors.Is(err, model.ErrorDocumentNotFound) {
				return httph.HTTPError{
					Code: http.StatusNotFound,
					Err:  errors.New("document not found"),
				}
			}

			log.Info("Error deleting document", "error", err)
			return errors.Wrap(err, "error deleting document")
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}))
}
