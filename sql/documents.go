package sql

import (
	"app/model"
	"context"

	"maragu.dev/errors"
	"maragu.dev/sqlh/sql"
)

func (d *Database) CreateDocument(ctx context.Context, doc model.Document) (model.Document, error) {
	query := `
		insert into documents (content)
		values (?)
		returning id, created, updated, content
	`

	var result model.Document
	err := d.H.Get(ctx, &result, query, doc.Content)
	if err != nil {
		return model.Document{}, errors.Wrap(err, "error creating document")
	}

	return result, nil
}

func (d *Database) ListDocuments(ctx context.Context) ([]model.Document, error) {
	query := `
		select id, created, updated, content
		from documents
		order by created desc
	`

	var results []model.Document
	err := d.H.Select(ctx, &results, query)
	if err != nil {
		return nil, errors.Wrap(err, "error listing documents")
	}

	return results, nil
}

func (d *Database) GetDocument(ctx context.Context, id model.ID) (model.Document, error) {
	query := `
		select id, created, updated, content
		from documents
		where id = ?
	`

	var result model.Document
	err := d.H.Get(ctx, &result, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Document{}, model.ErrorDocumentNotFound
		}
		return model.Document{}, errors.Wrap(err, "error getting document")
	}

	return result, nil
}

func (d *Database) UpdateDocument(ctx context.Context, id model.ID, doc model.Document) (model.Document, error) {
	var result model.Document

	err := d.H.InTransaction(ctx, func(tx *sql.Tx) error {
		// First check if the document exists
		var exists bool
		checkQuery := `
			select exists(select 1 from documents where id = ?)
		`
		if err := tx.Get(ctx, &exists, checkQuery, id); err != nil {
			return errors.Wrap(err, "error checking if document exists")
		}

		if !exists {
			return model.ErrorDocumentNotFound
		}

		// Then update the document
		updateQuery := `
			update documents
			set content = ?
			where id = ?
			returning id, created, updated, content
		`

		if err := tx.Get(ctx, &result, updateQuery, doc.Content, id); err != nil {
			return errors.Wrap(err, "error updating document")
		}

		return nil
	})

	if err != nil {
		return model.Document{}, err
	}

	return result, nil
}

func (d *Database) DeleteDocument(ctx context.Context, id model.ID) error {
	return d.H.InTransaction(ctx, func(tx *sql.Tx) error {
		// First check if the document exists
		var exists bool
		checkQuery := `
			select exists(select 1 from documents where id = ?)
		`
		if err := tx.Get(ctx, &exists, checkQuery, id); err != nil {
			return errors.Wrap(err, "error checking if document exists")
		}

		if !exists {
			return model.ErrorDocumentNotFound
		}

		// Then delete the document
		deleteQuery := `
			delete from documents
			where id = ?
		`

		if err := tx.Exec(ctx, deleteQuery, id); err != nil {
			return errors.Wrap(err, "error deleting document")
		}

		return nil
	})
}
