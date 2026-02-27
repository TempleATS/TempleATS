package storage

import (
	"context"
	"io"
	"net/http"
)

// Storage abstracts file upload and serving.
type Storage interface {
	// Save writes a file and returns its public URL path.
	Save(ctx context.Context, filename string, r io.Reader) (url string, err error)

	// Handler returns an http.Handler that serves uploaded files.
	// For local storage, this serves from disk.
	// For S3, this issues presigned URL redirects.
	Handler() http.Handler
}
