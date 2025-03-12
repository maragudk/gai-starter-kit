// Package model has domain models used throughout the application.
package model

import (
	"bytes"
	"encoding/binary"

	"maragu.dev/gai"
)

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

// QuantizeEmbedding to a binary vector and serialize to byte slice representation.
// Can then be inserted directly into a bit vector in the database.
// See also [sqlitevec.SerializeEmbedding].
func QuantizeEmbedding[T gai.VectorComponent](embedding []T) []byte {
	var quantized []uint8
	for _, vc := range embedding {
		b := 0
		if vc > 0 {
			b = 1
		}
		quantized = append(quantized, uint8(b))
	}

	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, quantized)
	return buf.Bytes()
}
