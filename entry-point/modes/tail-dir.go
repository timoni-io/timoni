package modes

import (
	"bufio"
	"encoding/binary"
	"entry-point/global"
	"errors"
	"fmt"
	"io"
	"lib/tlog"
	"lib/utils/set"
	"os"
	"path/filepath"
	"time"
)

var (
	errorOpen   int = 1
	errorSeek   int = 1
	errorNotEOF int = 1
)

var tailFileMap = set.NewSafe[string](nil)

// TailDir parses log files from directory
func TailDir() {
	if global.ReadLogFilesFromDir == "" {
		return
	}
	for {
		time.Sleep(3 * time.Second)

		if _, err := os.Stat(global.ReadLogFilesFromDir); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(global.ReadLogFilesFromDir, 0755)
			if err != nil {
				tlog.Error(err.Error())
			}
		}

		files, err := os.ReadDir(global.ReadLogFilesFromDir)
		if err != nil {
			tlog.Error(err.Error())
			continue
		}

		for _, f := range files {
			go tailFile(filepath.Join(global.ReadLogFilesFromDir, f.Name()))
		}
	}
}

func tailFile(filePath string) {

	// zabezpieczenie, aby nie otwierac dwa razy tego samego pliku
	if tailFileMap.Exists(filePath) {
		return
	}
	tailFileMap.Add(filePath)

	// odczytuje zawartosc pliku
	var file *os.File
	var err error

	details := map[string]string{
		"log-file": filePath,
		"log-dir":  filepath.Dir(filePath),
	}

	for {
		tlog.Info("opening log file: "+filePath, details)
		file, err = os.Open(filePath)
		if err != nil {
			tlog.Error(fmt.Sprintf("[%d] Failed to open file: %s", errorOpen, filePath), details)
			time.Sleep(time.Duration(errorOpen) * time.Second)
			errorOpen++
			continue
		}
		errorOpen = 1
		defer file.Close()

		//get lines already in the file
		fs := bufio.NewScanner(file)
		fs.Split(bufio.ScanLines)

		for fs.Scan() {
			// write to websocket
			binary.Write(os.Stdout, binary.LittleEndian, fs.Bytes())
		}

		// get consecutive lines from the file
		offset, err := file.Seek(0, io.SeekEnd)
		if err != nil {
			tlog.Error(fmt.Sprintf("[%s]"+err.Error(), fmt.Sprint(errorSeek)), details)
			time.Sleep(time.Second * time.Duration(errorSeek))
			errorSeek++
			continue
		}
		errorSeek = 1

		buffer := make([]byte, 1024)
		for {
			readBytes, err := file.ReadAt(buffer, offset)
			if err != nil {
				if err != io.EOF {
					tlog.Error(fmt.Sprintf("[%s]Error reading lines: "+err.Error(), fmt.Sprint(errorNotEOF)), details)
					time.Sleep(time.Second * time.Duration(errorNotEOF))
					errorNotEOF++
					break
				}
				errorNotEOF = 1
			}
			if readBytes == 0 {
				time.Sleep(time.Second)
				continue
			}
			offset += int64(readBytes)
			// write to websocket
			binary.Write(os.Stdout, binary.LittleEndian, buffer[:readBytes])
		}
	}
}
