package storage

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Local stores files on disk and serves them via http.FileServer.
type Local struct {
	Dir    string
	Prefix string // URL prefix, e.g. "/uploads/"
}

func NewLocal(dir, prefix string) (*Local, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &Local{Dir: dir, Prefix: prefix}, nil
}

func (l *Local) Save(_ context.Context, filename string, r io.Reader) (string, error) {
	dest := filepath.Join(l.Dir, filename)
	f, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		os.Remove(dest)
		return "", err
	}
	return l.Prefix + filename, nil
}

func (l *Local) Handler() http.Handler {
	return http.StripPrefix(l.Prefix, http.FileServer(http.Dir(l.Dir)))
}
