package http_test

import (
	"bytes"
	"context"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"maragu.dev/is"

	"app/aitest"
	"app/http"
	"app/model"
	"app/sqltest"
)

func TestDocuments(t *testing.T) {
	t.Run("create document", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		ai := aitest.NewClient(t)
		mux := chi.NewRouter()
		http.Documents(mux, db, ai)

		content := "Test document"
		reqBodyBytes := []byte(content)

		req := httptest.NewRequest("POST", "/documents", bytes.NewReader(reqBodyBytes))
		req.Header.Set("Content-Type", "text/markdown")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusCreated, w.Code)
		is.Equal(t, 0, w.Body.Len()) // No response body expected
	})

	t.Run("list documents", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		ai := aitest.NewClient(t)
		mux := chi.NewRouter()
		http.Documents(mux, db, ai)

		// First create a document
		doc := model.Document{Content: "Test document"}
		createdDoc, err := db.CreateDocument(t.Context(), doc, nil)
		is.NotError(t, err)

		// Now list documents
		req := httptest.NewRequest("GET", "/documents", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusOK, w.Code)
		
		// Check that the response contains a markdown link to the document
		expectedLink := "- [" + string(createdDoc.ID) + "](/documents/" + string(createdDoc.ID) + ")\n"
		is.Equal(t, expectedLink, w.Body.String())
	})

	t.Run("get document", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		ai := aitest.NewClient(t)
		mux := chi.NewRouter()
		http.Documents(mux, db, ai)

		// First create a document
		doc := model.Document{Content: "Test document"}
		createdDoc, err := db.CreateDocument(t.Context(), doc, nil)
		is.NotError(t, err)

		// Now get the document
		req := httptest.NewRequest("GET", "/documents/"+string(createdDoc.ID), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusOK, w.Code)
		
		// The response should be the document content as plain text
		is.Equal(t, createdDoc.Content, w.Body.String())
	})

	t.Run("update document", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		ai := aitest.NewClient(t)
		mux := chi.NewRouter()
		http.Documents(mux, db, ai)

		// First create a document
		doc := model.Document{Content: "Test document"}
		createdDoc, err := db.CreateDocument(t.Context(), doc, nil)
		is.NotError(t, err)

		// Now update the document
		updatedContent := "Updated content"
		reqBodyBytes := []byte(updatedContent)

		req := httptest.NewRequest("PUT", "/documents/"+string(createdDoc.ID), bytes.NewReader(reqBodyBytes))
		req.Header.Set("Content-Type", "text/markdown")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusOK, w.Code)
		
		// The response should contain the updated content directly
		is.Equal(t, updatedContent, w.Body.String())
	})

	t.Run("delete document", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		ai := aitest.NewClient(t)
		mux := chi.NewRouter()
		http.Documents(mux, db, ai)

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
		ai := aitest.NewClient(t)
		mux := chi.NewRouter()
		http.Documents(mux, db, ai)

		// Use an ID with invalid characters (uppercase and hyphen not allowed)
		req := httptest.NewRequest("GET", "/documents/INVALID-ID", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusNotFound, w.Code) // Should return 404 for route not found
	})

	t.Run("create document with chunks", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		ai := aitest.NewClient(t)
		mux := chi.NewRouter()
		http.Documents(mux, db, ai)

		// Create a document with content that should be chunked
		content := "This is paragraph one.\n\nThis is paragraph two.\n\nThis is paragraph three."
		reqBodyBytes := []byte(content)

		req := httptest.NewRequest("POST", "/documents", bytes.NewReader(reqBodyBytes))
		req.Header.Set("Content-Type", "text/markdown")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusCreated, w.Code)

		// List documents to get the ID of the created document
		reqList := httptest.NewRequest("GET", "/documents", nil)
		wList := httptest.NewRecorder()
		mux.ServeHTTP(wList, reqList)
		
		// Extract document ID from the list response
		responseLine := wList.Body.String()
		// Parse the document ID from the Markdown link format: "- [ID](/documents/ID)"
		startPos := strings.Index(responseLine, "[") + 1
		endPos := strings.Index(responseLine, "]")
		id := model.ID(responseLine[startPos:endPos])
		
		// Verify chunks were created
		chunks, err := db.GetDocumentChunks(context.Background(), id)
		is.NotError(t, err)
		is.True(t, len(chunks) > 0) // Should have at least one chunk
	})
}