// Package model has domain models used throughout the application.
package model

type ID string

type Document struct {
	ID      ID
	Created Time
	Updated Time
	Content string
}

type Chunk struct {
	ID         ID
	Created    Time
	Updated    Time
	DocumentID ID `db:"documentID"`
	Index      int
	Content    string
	Embedding  []byte
}
