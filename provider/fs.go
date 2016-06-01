package provider

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/lestrrat/go-pdebug"
)

// NewFS creates a new Provider that looks for JSON documents
// from the local file system. Documents are only searched
// within `root`
func NewFS(root string) *FS {
	return &FS{
		mp:   NewMap(),
		Root: root,
	}
}

// Get fetches the document specified by the `key` argument.
// Everything other than .Path is ignored. Note that
// once a document is read, it WILL be cached for the duration
// of this object, unless you call `Purge`
func (fp *FS) Get(key *url.URL) (out interface{}, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("provider.FS.Get(%s)", key.String()).BindError(&err)
		defer g.End()
	}

	if strings.ToLower(key.Scheme) != "file" {
		return nil, errors.New("unsupported scheme '" + key.Scheme + "'")
	}

	// Everything other than "Path" is ignored
	path := filepath.Clean(filepath.Join(fp.Root, key.Path))

	mpkey := &url.URL{Path: path}
	if x, err := fp.mp.Get(mpkey); err == nil {
		return x, nil
	}

	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, errors.New("target is not a file")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var x interface{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&x); err != nil {
		return nil, err
	}

	fp.mp.Set(path, x)

	return x, nil
}