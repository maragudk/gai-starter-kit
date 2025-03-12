package sql_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"maragu.dev/is"

	"app/aitest"
	"app/model"
	"app/sqltest"
)

func TestDocuments_CRUD(t *testing.T) {
	t.Run("create, read, update, delete", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		c := aitest.NewClient(t)

		// Create
		doc := model.Document{
			Content: "Test document content",
		}

		chunks := stringToChunks(t, c.EmbedString, doc.Content)

		created, err := db.CreateDocument(t.Context(), doc, chunks)
		is.NotError(t, err)
		is.True(t, created.ID != "")
		is.True(t, !created.Created.T.IsZero())
		is.True(t, !created.Updated.T.IsZero())
		is.Equal(t, doc.Content, created.Content)

		// Get
		retrieved, err := db.GetDocument(t.Context(), created.ID)
		is.NotError(t, err)
		is.Equal(t, created.ID, retrieved.ID)
		is.Equal(t, created.Created, retrieved.Created)
		is.Equal(t, created.Updated, retrieved.Updated)
		is.Equal(t, created.Content, retrieved.Content)

		// Update
		doc.ID = created.ID
		doc.Content = "Updated content"

		chunks = stringToChunks(t, c.EmbedString, doc.Content)

		updated, err := db.UpdateDocument(t.Context(), doc, chunks)
		is.NotError(t, err)
		is.Equal(t, created.ID, updated.ID)
		is.Equal(t, created.Created, updated.Created)
		is.Equal(t, doc.Content, updated.Content)

		// List
		docs, err := db.ListDocuments(t.Context())
		is.NotError(t, err)
		is.Equal(t, 1, len(docs))
		is.Equal(t, created.ID, docs[0].ID)

		// Delete
		err = db.DeleteDocument(t.Context(), created.ID)
		is.NotError(t, err)

		// Verify deletion
		_, err = db.GetDocument(t.Context(), created.ID)
		is.Error(t, model.ErrorDocumentNotFound, err)
	})

	t.Run("list multiple documents", func(t *testing.T) {
		db := sqltest.NewDatabase(t)

		// Insert documents with unique content
		doc1 := model.Document{Content: "Document 1"}
		doc2 := model.Document{Content: "Document 2"}
		doc3 := model.Document{Content: "Document 3"}

		_, err := db.CreateDocument(t.Context(), doc1, nil)
		is.NotError(t, err)
		time.Sleep(time.Millisecond)
		_, err = db.CreateDocument(t.Context(), doc2, nil)
		is.NotError(t, err)
		time.Sleep(time.Millisecond)
		_, err = db.CreateDocument(t.Context(), doc3, nil)
		is.NotError(t, err)

		// List documents
		docs, err := db.ListDocuments(t.Context())
		is.NotError(t, err)
		is.Equal(t, 3, len(docs))

		is.Equal(t, "Document 3", docs[0].Content)
		is.Equal(t, "Document 2", docs[1].Content)
		is.Equal(t, "Document 1", docs[2].Content)
	})
}

// stringToChunks based on whitespace.
func stringToChunks(t *testing.T, embedder func(context.Context, string) ([]byte, error), content string) []model.Chunk {
	t.Helper()

	parts := strings.Split(content, " ")
	var chunks []model.Chunk
	for i, part := range parts {
		chunks = append(chunks, model.Chunk{
			Index:     i,
			Content:   part,
			Embedding: must(embedder(t.Context(), part)),
		})
	}
	return chunks
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
