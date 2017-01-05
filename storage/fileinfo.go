package storage

import (
	"os"
	"strings"
	"time"
)

type FileInfo interface {
	Path() string
	Size() int64
	ModTime() time.Time
	IsDir() bool
}

type FileInfoField struct {
	FilePath    string    `json:"path"`
	FileSize    int64     `json:"size"`
	LastModTime time.Time `json:"modtime,omitempty"`
	Dir         bool      `json:"isdir"`
}

func (fi FileInfoField) Path() string {
	return fi.FilePath
}

func (fi FileInfoField) Size() int64 {
	return fi.FileSize
}

func (fi FileInfoField) IsDir() bool {
	return fi.Dir
}

func (fi FileInfoField) ModTime() time.Time {
	return fi.LastModTime
}

var _ FileInfo = FileInfoField{}
var _ FileInfo = &FileInfoField{}

func NewFI(name string, size int64, mod time.Time, isdir bool) *FileInfoField {
	return &FileInfoField{
		FilePath:    name,
		FileSize:    size,
		LastModTime: mod,
		Dir:         isdir,
	}
}

func NewOSFI(fi os.FileInfo, chroot string) *FileInfoField {
	return &FileInfoField{
		FilePath:    strings.TrimPrefix(fi.Name(), chroot),
		FileSize:    fi.Size(),
		LastModTime: fi.ModTime(),
		Dir:         fi.IsDir(),
	}
}
