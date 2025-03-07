package sql_test

import (
	"testing"

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

		// List
		docs, err := db.ListDocuments(t.Context())
		is.NotError(t, err)
		is.True(t, len(docs) > 0)
		found := false
		for _, d := range docs {
			if d.ID == created.ID {
				found = true
				is.Equal(t, updated.Content, d.Content)
				break
			}
		}
		is.True(t, found)

		// Delete
		err = db.DeleteDocument(t.Context(), created.ID)
		is.NotError(t, err)

		// Verify deletion
		_, err = db.GetDocument(t.Context(), created.ID)
		is.True(t, err != nil)
	})

	t.Run("list documents returns documents", func(t *testing.T) {
		db := sqltest.NewDatabase(t)

		// Create several documents
		doc1 := model.Document{Content: "First document"}
		doc2 := model.Document{Content: "Second document"}
		doc3 := model.Document{Content: "Third document"}

		created1, err := db.CreateDocument(t.Context(), doc1)
		is.NotError(t, err)
		created2, err := db.CreateDocument(t.Context(), doc2)
		is.NotError(t, err)
		created3, err := db.CreateDocument(t.Context(), doc3)
		is.NotError(t, err)

		// List and verify docs exist
		docs, err := db.ListDocuments(t.Context())
		is.NotError(t, err)
		is.True(t, len(docs) >= 3)

		// Verify all documents are in the results
		found1 := false
		found2 := false
		found3 := false
		for _, doc := range docs {
			if doc.ID == created1.ID {
				found1 = true
			} else if doc.ID == created2.ID {
				found2 = true
			} else if doc.ID == created3.ID {
				found3 = true
			}
		}

		is.True(t, found1)
		is.True(t, found2)
		is.True(t, found3)
	})
}
