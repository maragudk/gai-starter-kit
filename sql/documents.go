package sql

import (
	"app/model"
	"context"

	"maragu.dev/errors"
	"maragu.dev/sqlh/sql"
)

// CreateDocument and add the chunks as well as the chunk embeddings.
func (d *Database) CreateDocument(ctx context.Context, doc model.Document, chunks []model.Chunk) (model.Document, error) {
	err := d.H.InTransaction(ctx, func(tx *sql.Tx) error {
		query := `
			insert into documents (content)
			values (?)
			returning *
		`
		err := tx.Get(ctx, &doc, query, doc.Content)
		if err != nil {
			return errors.Wrap(err, "error creating document")
		}

		if err := d.saveChunks(ctx, tx, doc.ID, chunks); err != nil {
			return errors.Wrap(err, "error saving chunks")
		}

		return nil
	})

	return doc, err
}

// saveChunks by deleting previous chunks and inserting new ones.
func (d *Database) saveChunks(ctx context.Context, tx *sql.Tx, docID model.ID, chunks []model.Chunk) error {
	query := `
		delete from chunks where documentID = ?
	`
	if err := tx.Exec(ctx, query, docID); err != nil {
		return errors.Wrap(err, "error deleting previous chunks")
	}

	for _, c := range chunks {
		query := `
			insert into chunks (documentID, "index", content)
			values (?, ?, ?)
			returning *
		`
		if err := tx.Get(ctx, &c, query, docID, c.Index, c.Content); err != nil {
			return errors.Wrap(err, "error creating chunk")
		}

		query = `
			insert into chunk_embeddings (chunkID, embedding)
			values (?, vec_quantize_binary(?))
		`
		if err := tx.Exec(ctx, query, c.ID, c.Embedding); err != nil {
			return errors.Wrap(err, "error creating chunk embedding")
		}
	}

	return nil
}

type ListDocumentsOptions struct {
	// Limit is the maximum number of documents to return.
	// Default is 100 if not specified.
	Limit int

	// Cursor is the ID of the last document seen.
	// If provided, the result will only include documents with IDs greater than the cursor.
	Cursor model.ID
}

func (d *Database) ListDocuments(ctx context.Context, opts ListDocumentsOptions) ([]model.Document, error) {
	if opts.Limit < 0 {
		panic("limit cannot be negative")
	}

	if opts.Limit == 0 {
		opts.Limit = 100
	}

	var query string
	var args []any

	if opts.Cursor != "" {
		query = `
			select id, created, updated, content
			from documents
			where id > ?
			order by id
			limit ?
		`
		args = []any{opts.Cursor, opts.Limit}
	} else {
		query = `
			select id, created, updated, content
			from documents
			order by id
			limit ?
		`
		args = []any{opts.Limit}
	}

	var docs []model.Document
	if err := d.H.Select(ctx, &docs, query, args...); err != nil {
		return nil, errors.Wrap(err, "error listing documents")
	}

	return docs, nil
}

func (d *Database) GetDocument(ctx context.Context, id model.ID) (model.Document, error) {
	query := `
		select id, created, updated, content
		from documents
		where id = ?
	`

	var doc model.Document
	if err := d.H.Get(ctx, &doc, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return doc, model.ErrorDocumentNotFound
		}
		return doc, errors.Wrap(err, "error getting document")
	}

	return doc, nil
}

func (d *Database) UpdateDocument(ctx context.Context, doc model.Document, chunks []model.Chunk) (model.Document, error) {
	err := d.H.InTransaction(ctx, func(tx *sql.Tx) error {
		var exists bool
		query := `
			select exists(select 1 from documents where id = ?)
		`
		if err := tx.Get(ctx, &exists, query, doc.ID); err != nil {
			return errors.Wrap(err, "error checking if document exists")
		}

		if !exists {
			return model.ErrorDocumentNotFound
		}

		query = `
			update documents
			set content = ?
			where id = ?
			returning *
		`

		if err := tx.Get(ctx, &doc, query, doc.Content, doc.ID); err != nil {
			return errors.Wrap(err, "error updating document")
		}

		if err := d.saveChunks(ctx, tx, doc.ID, chunks); err != nil {
			return errors.Wrap(err, "error saving chunks")
		}

		return nil
	})

	return doc, err
}

func (d *Database) DeleteDocument(ctx context.Context, id model.ID) error {
	return d.H.InTransaction(ctx, func(tx *sql.Tx) error {
		var exists bool
		query := `
			select exists(select 1 from documents where id = ?)
		`
		if err := tx.Get(ctx, &exists, query, id); err != nil {
			return errors.Wrap(err, "error checking if document exists")
		}

		if !exists {
			return model.ErrorDocumentNotFound
		}

		query = `
			delete from documents
			where id = ?
		`
		if err := tx.Exec(ctx, query, id); err != nil {
			return errors.Wrap(err, "error deleting document")
		}

		return nil
	})
}

func (d *Database) GetDocumentChunks(ctx context.Context, docID model.ID) ([]model.Chunk, error) {
	query := `
		select c.id, c.created, c.updated, c.documentID, c."index", c.content, e.embedding
		from chunks c
			join chunk_embeddings e on c.id = e.chunkID
		where c.documentID = ?
		order by c."index"
	`

	var chunks []model.Chunk
	err := d.H.Select(ctx, &chunks, query, docID)
	if err != nil {
		return nil, errors.Wrap(err, "error getting document chunks with embeddings")
	}

	return chunks, nil
}
