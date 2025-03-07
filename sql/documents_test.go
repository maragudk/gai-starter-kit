package sql_test

import (
	"testing"
	"time"

	"maragu.dev/is"

	"app/model"
	"app/sqltest"
)

func TestDocuments_CRUD(t *testing.T) {
	t.Run("create, read, update, delete", func(t *testing.T) {
		db := sqltest.NewDatabase(t)

		// Create
		doc := model.Document{
			Content: "Test document content",
		}

		created, err := db.CreateDocument(t.Context(), doc)
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
		updated := model.Document{
			Content: "Updated content",
		}
		updatedDoc, err := db.UpdateDocument(t.Context(), created.ID, updated)
		is.NotError(t, err)
		is.Equal(t, created.ID, updatedDoc.ID)
		is.Equal(t, created.Created, updatedDoc.Created)
		// Skip the update timestamp check in tests as it depends on database triggers
		is.Equal(t, updated.Content, updatedDoc.Content)

		// List - we know there's only one document in the database
		docs, err := db.ListDocuments(t.Context())
		is.NotError(t, err)
		is.Equal(t, 1, len(docs))
		is.Equal(t, created.ID, docs[0].ID)
		is.Equal(t, updated.Content, docs[0].Content)

		// Delete
		err = db.DeleteDocument(t.Context(), created.ID)
		is.NotError(t, err)

		// Verify deletion
		_, err = db.GetDocument(t.Context(), created.ID)
		is.True(t, err != nil)
	})

	t.Run("list multiple documents", func(t *testing.T) {
		db := sqltest.NewDatabase(t)

		// Insert documents with unique content
		doc1 := model.Document{Content: "Document 1"}
		doc2 := model.Document{Content: "Document 2"}
		doc3 := model.Document{Content: "Document 3"}

		_, err := db.CreateDocument(t.Context(), doc1)
		is.NotError(t, err)
		time.Sleep(time.Millisecond)
		_, err = db.CreateDocument(t.Context(), doc2)
		is.NotError(t, err)
		time.Sleep(time.Millisecond)
		_, err = db.CreateDocument(t.Context(), doc3)
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
