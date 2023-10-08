package imageregistry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ImageS struct {
	GitRepo     string
	Tag         string
	ManifestSHA string
	ConfigSHA   string
	Layers      []*BlobS
}

func (ir *ImageRegS) scanImagesDir() {

	// blobsMap := ScanBlobsDir()

	res := map[string]*ImageS{}

	fp := filepath.Join(ir.rootPath, "repositories")
	repos, err := os.ReadDir(fp)
	if err != nil {
		panic(err)
	}

	for _, repo := range repos {
		if !repo.IsDir() {
			continue
		}

		for tagName, manifestSHA := range ir.tagMap(repo.Name()) {

			configSHA, layers := ir.openManifestBlob(manifestSHA)
			key := fmt.Sprintf("%s:%s", repo.Name(), tagName)
			res[key] = &ImageS{
				GitRepo:     repo.Name(),
				Tag:         tagName,
				ManifestSHA: manifestSHA,
				ConfigSHA:   configSHA,
				Layers:      layers,
			}
		}
	}

	ir.imagesMap = res
	// tlog.PrintJSON(res)
}

func (ir *ImageRegS) tagMap(repoName string) map[string]string { // tag => blob.imageManifest

	res := map[string]string{}

	fp := filepath.Join(ir.rootPath, "repositories", repoName, "_manifests", "tags")
	tags, err := os.ReadDir(fp)
	if err != nil {
		panic(err)
	}

	for _, tag := range tags {
		if !tag.IsDir() {
			continue
		}

		manifestLinkFilePath := filepath.Join(fp, tag.Name(), "current", "link")
		buf, _ := os.ReadFile(manifestLinkFilePath)
		res[tag.Name()] = strings.TrimPrefix(string(buf), "sha256:")
	}

	return res
}
