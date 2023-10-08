package utils

import (
	"fmt"
	"io"
	"os"
)

func CopyFile(source, destination string, perm ...os.FileMode) (int64, error) {
	stat, err := os.Stat(source)
	if err != nil {
		return 0, err
	}

	if !stat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", source)
	}

	in, err := os.Open(source)
	if err != nil {
		return 0, err
	}
	defer in.Close()

	out, err := os.Create(destination)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	n, err := io.Copy(out, in)
	if err != nil {
		return n, err
	}

	if len(perm) > 0 {
		if err = os.Chmod(destination, perm[0]); err != nil {
			return n, err
		}
	}
	return n, nil
}
