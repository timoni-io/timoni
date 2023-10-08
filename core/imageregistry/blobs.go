package imageregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type BlobS struct {
	FilePath string
	SHA256   string
	Type     string
	UseCount int // ile razy dany blob jest uzyty
	Size     int
}

func (ir *ImageRegS) scanBlobsDir() {

	fp := filepath.Join(ir.rootPath, "blobs", "sha256")
	items, err := os.ReadDir(fp)
	if err != nil {
		panic(err)
	}

	res := map[string]*BlobS{}
	for _, item := range items {
		if !item.IsDir() {
			continue
		}

		fp2 := filepath.Join(fp, item.Name())
		items2, err := os.ReadDir(fp2)
		if err != nil {
			panic(err)
		}

		for _, item2 := range items2 {
			if !item2.IsDir() {
				continue
			}
			blobFilePath := filepath.Join(fp2, item2.Name(), "data")
			res[item2.Name()] = &BlobS{
				FilePath: blobFilePath,
				SHA256:   item2.Name(),
				Type:     getBlobType(blobFilePath),
				UseCount: 0,
			}
		}
	}
	// tlog.PrintJSON(res)
	ir.blobsMap = res
}

var (
	buf1 = []byte(`"layers": [`)
	buf2 = []byte(`"mediaType": "application/vnd.docker.container.image.v1+json"`)
)

func getBlobType(filePath string) string {

	r, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer r.Close()

	var buf [1]byte
	_, err = io.ReadFull(r, buf[:])
	if err != nil {
		panic(err)
	}

	if buf[0] == '{' {
		buf, _ := os.ReadFile(filePath)
		if bytes.Contains(buf, buf1) && bytes.Contains(buf, buf2) {
			// fmt.Println(filePath)
			return "manifest"

		} else {
			// fmt.Println(filePath)
			return "config"
		}

	}

	return "data"
}

type rawManifestS struct {
	Config rawManifestDataS
	Layers []rawManifestDataS
}

type rawManifestDataS struct {
	MediaType string
	Size      int
	Digest    string
}

func (ir *ImageRegS) openManifestBlob(sha string) (configSHA string, layers []*BlobS) {

	blob := ir.blobsMap[sha]
	if blob == nil {
		fmt.Println("blob not found", sha)
		return "", nil
	}

	if blob.Type != "manifest" {
		fmt.Println("blob wrong type", blob.Type)
		return "", nil
	}

	blob.UseCount++

	buf, _ := os.ReadFile(blob.FilePath)
	data := new(rawManifestS)
	json.Unmarshal(buf, data)

	configSHA = strings.TrimPrefix(data.Config.Digest, "sha256:")
	ir.blobsMap[configSHA].Size = data.Config.Size
	ir.blobsMap[configSHA].UseCount++

	layers = []*BlobS{}
	for _, layer := range data.Layers {

		sha := strings.TrimPrefix(layer.Digest, "sha256:")
		blobData := ir.blobsMap[sha]

		if blobData == nil {
			fmt.Println("blob not found", sha)
			return "", nil
		}

		if blobData.Type != "data" {
			fmt.Println("blob wrong type", blobData.Type)
			return "", nil
		}

		blobData.Size = layer.Size
		blobData.UseCount++
		layers = append(layers, blobData)
	}

	return
}
