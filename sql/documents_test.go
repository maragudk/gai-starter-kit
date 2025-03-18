package sql_test

import (
	"testing"
	"time"

	"maragu.dev/is"

	"app/aitest"
	"app/model"
	"app/sql"
	"app/sqltest"
)

func TestDocuments_CRUD(t *testing.T) {
	t.Run("create, read, update, delete", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		ai := aitest.NewClient(t)

		// Create
		doc := model.Document{
			Content: "Test document content",
		}

		chunks, err := doc.Chunk(t.Context(), ai.EmbedString)
		is.NotError(t, err)

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

		chunks, err = doc.Chunk(t.Context(), ai.EmbedString)
		is.NotError(t, err)

		updated, err := db.UpdateDocument(t.Context(), doc, chunks)
		is.NotError(t, err)
		is.Equal(t, created.ID, updated.ID)
		is.Equal(t, created.Created, updated.Created)
		is.Equal(t, doc.Content, updated.Content)

		// List
		docs, err := db.ListDocuments(t.Context(), sql.ListDocumentsOptions{})
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

		// List documents with default options
		docs, err := db.ListDocuments(t.Context(), sql.ListDocumentsOptions{})
		is.NotError(t, err)
		is.Equal(t, 3, len(docs))

		// Since IDs are auto-generated and we don't know their exact order,
		// just verify we have all expected documents regardless of order
		contentMap := make(map[string]bool)
		for _, doc := range docs {
			contentMap[doc.Content] = true
		}
		
		is.Equal(t, 3, len(contentMap))
		is.True(t, contentMap["Document 1"])
		is.True(t, contentMap["Document 2"])
		is.True(t, contentMap["Document 3"])
	})

	t.Run("pagination", func(t *testing.T) {
		db := sqltest.NewDatabase(t)

		// Create 5 documents with known contents
		docContents := []string{
			"Paginated Document A",
			"Paginated Document B",
			"Paginated Document C",
			"Paginated Document D",
			"Paginated Document E",
		}
		
		for _, content := range docContents {
			doc := model.Document{Content: content}
			_, err := db.CreateDocument(t.Context(), doc, nil)
			is.NotError(t, err)
		}

		// Test default limit (100)
		allDocs, err := db.ListDocuments(t.Context(), sql.ListDocumentsOptions{})
		is.NotError(t, err)
		is.Equal(t, 5, len(allDocs))
		
		// Verify all contents are present (regardless of order)
		contentMap := make(map[string]bool)
		for _, doc := range allDocs {
			contentMap[doc.Content] = true
		}
		for _, content := range docContents {
			is.True(t, contentMap[content])
		}

		// Test with limit
		limitDocs, err := db.ListDocuments(t.Context(), sql.ListDocumentsOptions{Limit: 2})
		is.NotError(t, err)
		is.Equal(t, 2, len(limitDocs))

		// Test cursor-based pagination with three pages
		firstPageDocs, err := db.ListDocuments(t.Context(), sql.ListDocumentsOptions{Limit: 2})
		is.NotError(t, err)
		is.Equal(t, 2, len(firstPageDocs))

		// Get second page using cursor from first page
		secondPageDocs, err := db.ListDocuments(t.Context(), sql.ListDocumentsOptions{
			Limit:  2,
			Cursor: firstPageDocs[len(firstPageDocs)-1].ID,
		})
		is.NotError(t, err)
		is.Equal(t, 2, len(secondPageDocs))

		// Ensure no overlap between pages
		firstPageIDs := make(map[model.ID]bool)
		for _, doc := range firstPageDocs {
			firstPageIDs[doc.ID] = true
		}
		
		for _, doc := range secondPageDocs {
			// Second page should not contain IDs from first page
			is.True(t, !firstPageIDs[doc.ID])
		}

		// Get third page
		thirdPageDocs, err := db.ListDocuments(t.Context(), sql.ListDocumentsOptions{
			Limit:  2,
			Cursor: secondPageDocs[len(secondPageDocs)-1].ID,
		})
		is.NotError(t, err)
		// Third page should have exactly 1 document
		is.Equal(t, 1, len(thirdPageDocs))
		
		// Ensure third page has unique IDs
		for _, doc := range thirdPageDocs {
			// Third page should not contain IDs from first page
			is.True(t, !firstPageIDs[doc.ID])
		}
		
		secondPageIDs := make(map[model.ID]bool)
		for _, doc := range secondPageDocs {
			secondPageIDs[doc.ID] = true
		}
		
		for _, doc := range thirdPageDocs {
			// Third page should not contain IDs from second page
			is.True(t, !secondPageIDs[doc.ID])
		}
		
		// Verify all unique IDs total 5 documents
		allIDs := make(map[model.ID]bool)
		for _, doc := range firstPageDocs {
			allIDs[doc.ID] = true
		}
		for _, doc := range secondPageDocs {
			allIDs[doc.ID] = true
		}
		for _, doc := range thirdPageDocs {
			allIDs[doc.ID] = true
		}
		
		// All pages combined should contain all 5 document IDs
		is.Equal(t, 5, len(allIDs))
	})
}