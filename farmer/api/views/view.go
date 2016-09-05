package views

import (
	"bytes"
	"compress/gzip"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"

	logpkg "github.com/op/go-logging"
)

var (
	log = logpkg.MustGetLogger("farmer")
)

type Views struct {
	Names map[string]struct{}
	Index string
}

func New() *Views {
	v := &Views{make(map[string]struct{}), "index.html"}
	names := AssetNames()
	for _, n := range names {
		v.Names[n] = struct{}{}
	}
	return v
}

func (v *Views) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimLeft(r.URL.Path, "/")
	if name == "" {
		name = v.Index
	}
	if _, ok := v.Names[name]; !ok {
		name = v.Index
	}

	ext := path.Ext(name)
	ct := mime.TypeByExtension(ext)
	w.Header().Set("Content-Type", ct)

	data, err := Asset(name)
	if err != nil {
		log.Debugf("FromOwn, OpenFile name: %s, error: %s", name, err.Error())
		return
	}
	hdr := r.Header.Get("Accept-Encoding")
	if strings.Contains(hdr, "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		w.Write(data)
	} else {
		gz, err := gzip.NewReader(bytes.NewBuffer(data))
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		io.Copy(w, gz)
		gz.Close()
	}
}
