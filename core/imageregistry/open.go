package imageregistry

import (
	"core/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ImageRegS struct {
	rootPath     string
	dataRootPath string
	blobsMap     map[string]*BlobS // key = blob SHA256
	imagesMap    map[string]*ImageS
}

func Open() *ImageRegS {

	dbPath := config.DataPath()

	ir := &ImageRegS{
		rootPath:     filepath.Join(dbPath, "image-registry", "docker", "registry", "v2"),
		dataRootPath: dbPath,
	}
	ir.scanBlobsDir()
	ir.scanImagesDir()
	return ir
}

func (ir *ImageRegS) GetUnusedBlobs() []*BlobS {

	res := []*BlobS{}
	for _, blob := range ir.blobsMap {
		if blob.UseCount > 0 {
			continue
		}

		res = append(res, blob)
	}
	return res
}

func (ir *ImageRegS) DeleteUnusedBlobs() []string {

	res := []string{}
	for _, blob := range ir.GetUnusedBlobs() {
		dir := filepath.Dir(blob.FilePath)
		fmt.Println("image-reg deleting blob: " + dir)
		res = append(res, filepath.Base(dir))
		os.RemoveAll(dir)
	}
	return res
}

func (ir *ImageRegS) DeleteImage(imageID string) {

	imageFilePath := filepath.Join(ir.dataRootPath, "image", imageID+".json")
	os.Remove(imageFilePath)

	tmp := strings.Split(imageID, ":")
	repoName := tmp[0]
	tag := tmp[1]

	dir := filepath.Join(ir.rootPath, "repositories", repoName, "_manifests", "tags", tag)
	fmt.Println("image-reg deleting image tag:", repoName, tag, "--", dir)
	os.RemoveAll(dir)
}
