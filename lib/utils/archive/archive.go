package archive

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Compress(input string, output OutputOpt, opts ...CompressorOpt) error {
	if output == nil {
		return errors.New("missing output")
	}

	// set output
	file, err := output()
	if err != nil {
		return err
	}
	defer file.Close()

	// apply options
	for _, opt := range opts {
		file, err = opt(file)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	// add tar writer
	archive := tar.NewWriter(file)
	defer archive.Close()

	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(input); err != nil {
		return err
	}

	return filepath.Walk(input, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		// generate tar header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		header.Name = filepath.Clean(strings.TrimPrefix(file, input))

		// write header
		if err := archive.WriteHeader(header); err != nil {
			return err
		}

		// open file
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data
		if _, err := io.Copy(archive, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()
		return nil
	})
}

func Uncompress(file io.Reader, output string, opts ...UncompressorOpt) error {
	var err error
	for _, opt := range opts {
		file, err = opt(file)
		if err != nil {
			return err
		}
	}
	archive := tar.NewReader(file)

	for {
		header, err := archive.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(output, header.Name)

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, archive); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
