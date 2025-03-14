package sql_test

import (
	"testing"

	"maragu.dev/is"

	"app/aitest"
	"app/model"
	"app/sqltest"
)

func TestDatabase_Search(t *testing.T) {
	t.Run("finds document with exact text match", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		ai := aitest.NewClient(t)

		// Create document with specific content
		doc := model.Document{
			Content: "This is a test document about artificial intelligence",
		}

		chunks, err := doc.Chunk(t.Context(), ai.EmbedString)
		is.NotError(t, err)

		doc, err = db.CreateDocument(t.Context(), doc, chunks)
		is.NotError(t, err)

		// Search for exact text
		query := "artificial intelligence"

		// Make sure the embedding doesn't match
		embedding, err := ai.EmbedString(t.Context(), "it's a pony")
		is.NotError(t, err)

		chunks, err = db.Search(t.Context(), query, embedding)
		is.NotError(t, err)
		is.Equal(t, 1, len(chunks))

		is.Equal(t, doc.Content, chunks[0].Content)
	})

	t.Run("finds document with semantic similarity", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		ai := aitest.NewClient(t)

		// Create document with specific content
		doc := model.Document{
			Content: "Five big sheep dancing joyfully to disco music",
		}

		chunks, err := doc.Chunk(t.Context(), ai.EmbedString)
		is.NotError(t, err)

		doc, err = db.CreateDocument(t.Context(), doc, chunks)
		is.NotError(t, err)

		// Search for semantically similar text
		query := "Some animals partying"
		embedding, err := ai.EmbedString(t.Context(), query)
		is.NotError(t, err)

		chunks, err = db.Search(t.Context(), query, embedding)
		is.NotError(t, err)
		is.Equal(t, 1, len(chunks))

		is.Equal(t, doc.Content, chunks[0].Content)
	})

	t.Run("returns empty results for non-matching query", func(t *testing.T) {
		db := sqltest.NewDatabase(t)
		ai := aitest.NewClient(t)

		// Create document
		doc := model.Document{
			Content: "This is about programming languages",
		}

		chunks, err := doc.Chunk(t.Context(), ai.EmbedString)
		is.NotError(t, err)

		_, err = db.CreateDocument(t.Context(), doc, chunks)
		is.NotError(t, err)

		// Search for unrelated content
		query := "quantum physics theories"
		embedding, err := ai.EmbedString(t.Context(), query)
		is.NotError(t, err)

		results, err := db.Search(t.Context(), query, embedding)
		is.NotError(t, err)
		is.Equal(t, 0, len(results))
	})
}
