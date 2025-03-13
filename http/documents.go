package http

import (
	"context"
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

type CreateDocumentRequest struct {
	Content string
}

func (r CreateDocumentRequest) Validate() error {
	if r.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

type CreateDocumentResponse struct {
	Document model.Document
}

// StatusCode implements the statusCodeGiver interface.
func (r *CreateDocumentResponse) StatusCode() int {
	return http.StatusCreated
}

type ListDocumentsResponse struct {
	Documents []model.Document
}

type GetDocumentResponse struct {
	Document model.Document
}

type UpdateDocumentRequest struct {
	Content string
}

func (r UpdateDocumentRequest) Validate() error {
	if r.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

type UpdateDocumentResponse struct {
	Document model.Document
}

type DeleteDocumentResponse struct {
	Success bool
}

// StatusCode implements the statusCodeGiver interface.
func (r *DeleteDocumentResponse) StatusCode() int {
	return http.StatusNoContent
}

func Documents(mux chi.Router, db documentCRUDer, ai embedder) {
	mux.Post("/documents", httph.JSONHandler(func(w http.ResponseWriter, r *http.Request, req CreateDocumentRequest) (*CreateDocumentResponse, error) {
		doc := model.Document{
			Content: req.Content,
		}

		chunks, err := model.CreateDocumentChunks(r.Context(), req.Content, ai.EmbedString)
		if err != nil {
			return nil, errors.Wrap(err, "error creating document chunks")
		}

		created, err := db.CreateDocument(r.Context(), doc, chunks)
		if err != nil {
			return nil, errors.Wrap(err, "error creating document")
		}

		return &CreateDocumentResponse{
			Document: created,
		}, nil
	}))

	mux.Get("/documents", httph.JSONHandler(func(w http.ResponseWriter, r *http.Request, _ any) (*ListDocumentsResponse, error) {
		docs, err := db.ListDocuments(r.Context())
		if err != nil {
			return nil, errors.Wrap(err, "error listing documents")
		}

		return &ListDocumentsResponse{
			Documents: docs,
		}, nil
	}))

	mux.Get("/documents/{id:[a-z0-9_]+}", httph.JSONHandler(func(w http.ResponseWriter, r *http.Request, _ any) (*GetDocumentResponse, error) {
		id := model.ID(chi.URLParam(r, "id"))

		doc, err := db.GetDocument(r.Context(), id)
		if err != nil {
			if errors.Is(err, model.ErrorDocumentNotFound) {
				return nil, httph.HTTPError{
					Code: http.StatusNotFound,
					Err:  errors.New("document not found"),
				}
			}
			return nil, errors.Wrap(err, "error getting document")
		}

		return &GetDocumentResponse{
			Document: doc,
		}, nil
	}))

	mux.Put("/documents/{id:[a-z0-9_]+}", httph.JSONHandler(func(w http.ResponseWriter, r *http.Request, req UpdateDocumentRequest) (*UpdateDocumentResponse, error) {
		id := model.ID(chi.URLParam(r, "id"))

		doc := model.Document{
			ID:      id,
			Content: req.Content,
		}

		chunks, err := model.CreateDocumentChunks(r.Context(), req.Content, ai.EmbedString)
		if err != nil {
			return nil, errors.Wrap(err, "error creating document chunks")
		}

		doc, err = db.UpdateDocument(r.Context(), doc, chunks)
		if err != nil {
			if errors.Is(err, model.ErrorDocumentNotFound) {
				return nil, httph.HTTPError{
					Code: http.StatusNotFound,
					Err:  errors.New("document not found"),
				}
			}
			return nil, errors.Wrap(err, "error updating document")
		}

		return &UpdateDocumentResponse{
			Document: doc,
		}, nil
	}))

	mux.Delete("/documents/{id:[a-z0-9_]+}", httph.JSONHandler(func(w http.ResponseWriter, r *http.Request, _ any) (*DeleteDocumentResponse, error) {
		id := model.ID(chi.URLParam(r, "id"))

		if err := db.DeleteDocument(r.Context(), id); err != nil {
			if errors.Is(err, model.ErrorDocumentNotFound) {
				return nil, httph.HTTPError{
					Code: http.StatusNotFound,
					Err:  errors.New("document not found"),
				}
			}
			return nil, errors.Wrap(err, "error deleting document")
		}

		// Return empty response with 204 No Content status code
		return &DeleteDocumentResponse{}, nil
	}))
}
