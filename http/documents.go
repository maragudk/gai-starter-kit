package http

import (
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"maragu.dev/errors"
	"maragu.dev/httph"

	"app/model"
)

type documentCRUDer interface {
	CreateDocument(ctx context.Context, d model.Document, chunks []model.Chunk) (model.Document, error)
	ListDocuments(ctx context.Context) ([]model.Document, error)
	GetDocument(ctx context.Context, id model.ID) (model.Document, error)
	UpdateDocument(ctx context.Context, d model.Document, chunks []model.Chunk) (model.Document, error)
	DeleteDocument(ctx context.Context, id model.ID) error
}

type embedder interface {
	EmbedString(ctx context.Context, s string) ([]byte, error)
}

func Documents(mux chi.Router, db documentCRUDer, ai embedder) {
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
			return errors.Wrap(err, "error creating document chunks")
		}

		if _, err := db.CreateDocument(r.Context(), doc, chunks); err != nil {
			return errors.Wrap(err, "error creating document")
		}

		w.WriteHeader(http.StatusCreated)

		return nil
	}))

	mux.Get("/documents", httph.ErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		docs, err := db.ListDocuments(r.Context())
		if err != nil {
			return errors.Wrap(err, "error listing documents")
		}

		for _, doc := range docs {
			w.Write([]byte("- [" + doc.ID + "](/documents/" + doc.ID + ")\n"))
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
			return errors.Wrap(err, "error getting document")
		}

		w.Write([]byte(doc.Content))

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
			return errors.Wrap(err, "error creating document chunks")
		}

		doc, err = db.UpdateDocument(r.Context(), doc, chunks)
		if err != nil {
			if errors.Is(err, model.ErrorDocumentNotFound) {
				return httph.HTTPError{
					Code: http.StatusNotFound,
					Err:  errors.New("document not found"),
				}
			}
			return errors.Wrap(err, "error updating document")
		}

		w.Write(body)

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
			return errors.Wrap(err, "error deleting document")
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}))
}