package storage

import (
	"io"

	"golang.org/x/net/context"
)

type StorageDriver interface {
	BaseFile

	GetContent(ctx context.Context, path string) ([]byte, error)
	PutContent(ctx context.Context, path string, content []byte) error
	Reader(ctx context.Context, path string) (io.ReadCloser, error)
	Writer(ctx context.Context, path string, append bool) (io.WriteCloser, error)

	List(ctx context.Context, path string) ([]FileInfo, error)
	Mkdir(ctx context.Context, path string) error
}

type BaseFile interface {
	Name() string
	Stat(ctx context.Context, path string) (FileInfo, error)
	Move(ctx context.Context, sourcePath string, destPath string) error
	Delete(ctx context.Context, path string) error
}

type File interface {
	BaseFile
	GetContent(ctx context.Context, path string) ([]byte, error)
	PutContent(ctx context.Context, path string, content []byte) error
	Reader(ctx context.Context, path string) (io.ReadCloser, error)
	Writer(ctx context.Context, path string, append bool) (io.WriteCloser, error)
}

type Dir interface {
	BaseFile
	List(ctx context.Context, path string) ([]string, error)
}
