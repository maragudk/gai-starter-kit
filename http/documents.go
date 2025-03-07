package http

import (
	"app/model"
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"maragu.dev/httph"
)

type documentCRUDer interface {
	CreateDocument(ctx context.Context, d model.Document) (model.Document, error)
	ListDocuments(ctx context.Context) ([]model.Document, error)
	GetDocument(ctx context.Context, id model.ID) (model.Document, error)
	UpdateDocument(ctx context.Context, id model.ID, d model.Document) (model.Document, error)
	DeleteDocument(ctx context.Context, id model.ID) error
}

func Documents(mux chi.Router, db documentCRUDer) {
	mux.Post("/documents", httph.JSONHandler(func(w http.ResponseWriter, r *http.Request, req CreateDocumentRequest) (*CreateDocumentResponse, error) {
		return &CreateDocumentResponse{}, nil
	}))

	mux.Get("/documents", httph.JSONHandler(func(w http.ResponseWriter, r *http.Request, _ any) (*ListDocumentsResponse, error) {
		return &ListDocumentsResponse{}, nil
	}))

	mux.Get("/documents/{id}", httph.JSONHandler(func(w http.ResponseWriter, r *http.Request, _ any) (*GetDocumentResponse, error) {
		return &GetDocumentResponse{}, nil
	}))

	mux.Put("/documents/{id}", httph.JSONHandler(func(w http.ResponseWriter, r *http.Request, req UpdateDocumentRequest) (*UpdateDocumentResponse, error) {
		return &UpdateDocumentResponse{}, nil
	}))

	mux.Delete("/documents/{id}", httph.JSONHandler(func(w http.ResponseWriter, r *http.Request, _ any) (*DeleteDocumentResponse, error) {
		return &DeleteDocumentResponse{}, nil
	}))
}
