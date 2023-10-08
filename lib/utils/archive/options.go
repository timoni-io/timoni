package archive

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
)

type NopWriterCloser struct {
	io.Writer
}

func (NopWriterCloser) Close() error {
	return nil
}

type OutputOpt func() (io.WriteCloser, error)

func WriteFile(output string) OutputOpt {
	return func() (io.WriteCloser, error) {
		// Create dirs to output file
		err := os.MkdirAll(filepath.Dir(output), 0766)
		if err != nil {
			return nil, err
		}

		// Create/Open file for writing
		return os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	}
}

func Writer(w io.Writer) OutputOpt {
	return func() (io.WriteCloser, error) {
		return NopWriterCloser{Writer: w}, nil
	}
}

type CompressorOpt func(file io.Writer) (io.WriteCloser, error)

func CompressLZ4(file io.Writer) (io.WriteCloser, error) {
	return lz4.NewWriter(file), nil
}

func CompressGZIP(file io.Writer) (io.WriteCloser, error) {
	return gzip.NewWriter(file), nil
}

func CompressZSTD(file io.Writer) (io.WriteCloser, error) {
	return zstd.NewWriter(file)
}

type UncompressorOpt func(file io.Reader) (io.Reader, error)

func UncompressLZ4(file io.Reader) (io.Reader, error) {
	return lz4.NewReader(file), nil
}

func UncompressGZIP(file io.Reader) (io.Reader, error) {
	return gzip.NewReader(file)
}

func UncompressZSTD(file io.Reader) (io.Reader, error) {
	return zstd.NewReader(file)
}
