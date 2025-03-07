package model

// Error is for errors in the business domain. See the constants below.
type Error string

const (
	ErrorDocumentNotFound = Error("DOCUMENT_NOT_FOUND")
)

func (e Error) Error() string {
	return string(e)
}
