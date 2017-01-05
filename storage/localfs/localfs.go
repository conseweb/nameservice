package localfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric/storage"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
)

var (
	log = logging.MustGetLogger("filesystem")
)

type Driver struct {
	chroot string
}

func NewDriver(rootPath string) (storage.StorageDriver, error) {
	var err error

	rootPath, err = filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	fs, err := os.Stat(rootPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(rootPath, 0644)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		log.Errorf("NewLocalDriver: %s, %v", rootPath, err)
		return nil, err
	} else {
		if !fs.IsDir() {
			log.Errorf("NewLocalDriver: %s is a directory.", rootPath)
			return nil, fmt.Errorf("root path should not be a directory.")
		}
	}

	return &Driver{
		chroot: rootPath,
	}, nil
}
func (d *Driver) Name() string {
	return "local filesystem"
}

func (d *Driver) GetContent(ctx context.Context, path string) ([]byte, error) {
	fpath, err := d.Abs(path)
	if err != nil {
		return nil, err
	}
	_ = fpath

	return nil, nil
}

func (d *Driver) PutContent(ctx context.Context, path string, content []byte) error {
	return nil
}

func (d *Driver) Reader(ctx context.Context, path string) (io.ReadCloser, error) {
	fpath, err := d.Abs(path)
	if err != nil {
		return nil, err
	}

	return os.Open(fpath)
}

func (d *Driver) Writer(ctx context.Context, path string, isAppend bool) (io.WriteCloser, error) {
	fpath, err := d.Abs(path)
	if err != nil {
		return nil, err
	}

	var flag int
	_, err = d.Stat(ctx, fpath)
	if err != nil && os.IsNotExist(err) {
		flag = os.O_CREATE | os.O_WRONLY
	} else if err != nil {
		return nil, err
	} else if isAppend {
		flag = os.O_WRONLY | os.O_APPEND
	} else {
		flag = os.O_APPEND
	}

	return os.OpenFile(fpath, flag, 0644)
}

func (d *Driver) Stat(ctx context.Context, path string) (storage.FileInfo, error) {
	fpath, err := d.Abs(path)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(fpath)
	if err != nil {
		return nil, err
	}

	return storage.NewOSFI(fi, d.chroot), nil
}

func (d *Driver) List(ctx context.Context, path string) ([]storage.FileInfo, error) {
	fpath, err := d.Abs(path)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(fpath)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return []storage.FileInfo{storage.NewOSFI(fi, d.chroot)}, nil
	}

	ret := []storage.FileInfo{}
	filepath.Walk(fpath, func(p string, info os.FileInfo, err error) error {
		ret = append(ret, storage.NewFI(strings.TrimPrefix(p, d.chroot), info.Size(), info.ModTime(), info.IsDir()))
		return err
	})

	return ret, nil
}

func (d *Driver) Mkdir(ctx context.Context, path string) error {
	fpath, err := d.Abs(path)
	if err != nil {
		return err
	}

	return os.MkdirAll(fpath, 0644)
}

func (d *Driver) Move(ctx context.Context, sourcePath string, destPath string) error {
	return os.Rename(sourcePath, destPath)
}

func (d *Driver) Delete(ctx context.Context, path string) error {
	fpath, err := d.Abs(path)
	if err != nil {
		return err
	}

	return os.RemoveAll(fpath)
}

func (d *Driver) Abs(path string) (string, error) {
	absPath, err := filepath.Abs(d.chroot + "/" + strings.TrimPrefix(path, "/"))
	if err != nil {
		log.Errorf("Abs %s, %s", path, err)
		return "", err
	}

	if !strings.HasPrefix(absPath, d.chroot) {
		return "", fmt.Errorf("%s not exists", path)
	}

	return absPath, nil
}
