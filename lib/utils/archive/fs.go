package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"lib/tlog"
	"lib/utils/maps"
	"strings"

	"github.com/pierrec/lz4"
)

type EmbedLZ4 struct{}
type EmbedGzip struct{}
type EmbedCompression interface {
	EmbedLZ4 | EmbedGzip
}

// embed compressed tar file
type EmbedTar[C EmbedCompression] []byte

type TarReader struct {
	*tar.Reader
	Close func() error
}

func (tarFS EmbedTar[C]) reader() (*TarReader, error) {
	var data io.Reader = bytes.NewReader(tarFS)

	close := func() error { return nil }

	// Select uncompressor
	switch any((*C)(nil)).(type) {
	case *EmbedLZ4:
		data = lz4.NewReader(data)
	case *EmbedGzip:
		r, err := gzip.NewReader(data)
		if err != nil {
			return nil, err
		}
		close = r.Close
		data = r
	}

	return &TarReader{
		tar.NewReader(data),
		close,
	}, nil
}

// example: decode.Decode(Files().Get("01-namespace.yaml"))
func (tarFS EmbedTar[C]) Files(dir, overlay string) (maps.Maper[string, io.Reader], error) {
	reader, err := tarFS.reader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	files := maps.NewOrdered[string, io.Reader](nil, nil)
	for {
		fi, err := reader.Next()
		switch err {
		case nil:
		case io.EOF:
			return files, nil
		default:
			return nil, err
		}

		// Ignore dirs, links ...
		if fi.Typeflag != tar.TypeReg {
			continue
		}

		// skip files without correct prefix
		if dir != "." && !strings.HasPrefix(fi.Name, dir) {
			continue
		}

		fi.Name = strings.TrimPrefix(fi.Name, dir)

		// Skip files existing in map
		isOverlay := false
		if strings.HasPrefix(fi.Name, overlay) {
			fi.Name = strings.TrimPrefix(fi.Name, overlay)
			isOverlay = true
		}

		if files.Exists(fi.Name) && !isOverlay {
			tlog.Debug("exists", fi.Name)
			continue
		}

		// copy uncompressed file data
		buf := &bytes.Buffer{}
		io.Copy(buf, reader)
		files.Set(fi.Name, buf)
	}
}
