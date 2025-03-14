package http_test

import (
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

func TestSearch(t *testing.T) {
	t.Run("search documents", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		ai := aitest.NewClient(t)
		mux := chi.NewRouter()
		http.Search(mux, db, ai)

		// Create document with content and chunks
		doc := model.Document{Content: "This is a test document with searchable content"}
		chunks, err := doc.Chunk(context.Background(), ai.EmbedString)
		is.NotError(t, err)
		
		// Save document with chunks
		_, err = db.CreateDocument(t.Context(), doc, chunks)
		is.NotError(t, err)

		// Now search for content
		req := httptest.NewRequest("GET", "/search?q=searchable", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		is.Equal(t, stdhttp.StatusOK, w.Code)
		
		// The response should contain markdown links to the documents
		responseBody := w.Body.String()
		is.True(t, len(responseBody) > 0)
		is.True(t, strings.Contains(responseBody, "- ["))
		is.True(t, strings.Contains(responseBody, "](/documents/"))
	})
}