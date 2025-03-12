package http_test

import (
	"bytes"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"maragu.dev/is"

	"app/http"
	"app/model"
	"app/sqltest"
)

func TestDocuments(t *testing.T) {
	t.Run("create document", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		mux := chi.NewRouter()
		http.Documents(mux, db)

		reqBody := http.CreateDocumentRequest{
			Content: "Test document",
		}
		reqBodyBytes, err := json.Marshal(reqBody)
		is.NotError(t, err)

		req := httptest.NewRequest("POST", "/documents", bytes.NewReader(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusCreated, w.Code)

		var resp http.CreateDocumentResponse
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		is.NotError(t, err)
		is.True(t, resp.Document.ID != "")
		is.True(t, !resp.Document.Created.T.IsZero())
		is.True(t, !resp.Document.Updated.T.IsZero())
		is.Equal(t, reqBody.Content, resp.Document.Content)
	})

	t.Run("list documents", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		mux := chi.NewRouter()
		http.Documents(mux, db)

		// First create a document
		doc := model.Document{Content: "Test document"}
		createdDoc, err := db.CreateDocument(t.Context(), doc, nil)
		is.NotError(t, err)

		// Now list documents
		req := httptest.NewRequest("GET", "/documents", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusOK, w.Code)

		var resp http.ListDocumentsResponse
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		is.NotError(t, err)
		is.Equal(t, 1, len(resp.Documents))
		is.Equal(t, createdDoc.ID, resp.Documents[0].ID)
		is.Equal(t, createdDoc.Content, resp.Documents[0].Content)
	})

	t.Run("get document", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		mux := chi.NewRouter()
		http.Documents(mux, db)

		// First create a document
		doc := model.Document{Content: "Test document"}
		createdDoc, err := db.CreateDocument(t.Context(), doc, nil)
		is.NotError(t, err)

		// Now get the document
		req := httptest.NewRequest("GET", "/documents/"+string(createdDoc.ID), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusOK, w.Code)

		var resp http.GetDocumentResponse
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		is.NotError(t, err)
		is.Equal(t, createdDoc.ID, resp.Document.ID)
		is.Equal(t, createdDoc.Content, resp.Document.Content)
	})

	t.Run("update document", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		mux := chi.NewRouter()
		http.Documents(mux, db)

		// First create a document
		doc := model.Document{Content: "Test document"}
		createdDoc, err := db.CreateDocument(t.Context(), doc, nil)
		is.NotError(t, err)

		// Now update the document
		reqBody := http.UpdateDocumentRequest{
			Content: "Updated content",
		}
		reqBodyBytes, err := json.Marshal(reqBody)
		is.NotError(t, err)

		req := httptest.NewRequest("PUT", "/documents/"+string(createdDoc.ID), bytes.NewReader(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusOK, w.Code)

		var resp http.UpdateDocumentResponse
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		is.NotError(t, err)
		is.Equal(t, createdDoc.ID, resp.Document.ID)
		is.True(t, createdDoc.Content != resp.Document.Content)
		is.Equal(t, reqBody.Content, resp.Document.Content)
	})

	t.Run("delete document", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		mux := chi.NewRouter()
		http.Documents(mux, db)

		// First create a document
		doc := model.Document{Content: "Test document"}
		createdDoc, err := db.CreateDocument(t.Context(), doc, nil)
		is.NotError(t, err)

		// Now delete the document
		req := httptest.NewRequest("DELETE", "/documents/"+string(createdDoc.ID), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusNoContent, w.Code)

		// With a 204 No Content, the framework may still serialize the response
		// but the client should ignore the body

		// Verify it's deleted
		_, err = db.GetDocument(t.Context(), createdDoc.ID)
		is.True(t, err != nil)
	})

	t.Run("invalid document ID format", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		mux := chi.NewRouter()
		http.Documents(mux, db)

		// Use an ID with invalid characters (uppercase and hyphen not allowed)
		req := httptest.NewRequest("GET", "/documents/INVALID-ID", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusNotFound, w.Code) // Should return 404 for route not found
	})
}
