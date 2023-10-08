package db

import (
	"bufio"
	"bytes"
	"context"
	"core/config"
	"core/db2"
	"encoding/json"
	"fmt"
	"io"
	"lib/utils"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	log "lib/tlog"
)

const (
	Ready    = "ready"
	Building = "building"
	Failed   = "fail"
)

type ImageS struct {
	ID                  string
	ParentImageID       string
	SourceGit           SourceGitS
	DockerFileContent   string
	DockerFilePath      string
	UserID              string
	UserEmail           string
	TimeCreation        time.Time 
	TimeBegin           time.Time 
	TimeEnd             time.Time 
	BuildID             string    
	BuildStatus         string    // pending, building, ready, fail
	BuildRootPath       string
	Digest              string
	SizeMBWithParents   int    // MB
	SizeMBWithoutParent int    // MB
	EnvThatOrdered      string 
	EntryPointVersion   string
}

// ImageBuildBlueprintS ...
type ImageBuildBlueprintS struct {
	ImageID           string
	BuildID           string
	BuildRootPath     string
	SourceGit         SourceGitS
	DockerFilePath    string
	DockerFileContent string
	ImageBuilderID    string
	Status            string
	EntryPointGitTag  string
}

var (
	imageExternalCacheMap  = map[string]*ImageS{}
	imageExternalCacheLock = &sync.Mutex{}
)

func (image *ImageS) Save() {
	extIRurl := ""
	extIRtoken := ""
	if extIRurl == "" && extIRtoken == "" {

		err := driver.Write("image", image.ID, image)
		log.Error(err)
		return

	}

	buf, err := json.Marshal(image)
	if err != nil {
		log.Error(err)
		return
	}

	out := ""
	res, err := http.Post(extIRurl+"/api/image-external-save?token="+extIRtoken, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		log.Error(err)
		return
	}
	buf2, _ := io.ReadAll(res.Body)
	out = string(buf2)

	fmt.Println("$$----------------------------------------- save:")
	log.PrintJSON(image)
	fmt.Println(out)
	fmt.Println("$$-----------------------------------------")
}

func (image *ImageS) Delete() {
	err := driver.Delete("image", image.ID)
	if log.Error(err) != nil {
		return
	}
}

func (image *ImageS) UpdateSize() {

	manifest := getImageManifest(image)
	if manifest == nil {
		image.SizeMBWithParents = 0
		image.SizeMBWithoutParent = 0
		return
	}

	parent := &ImageManifestS{}
	if image.ParentImageID != "" {
		parent = getImageManifest(ImageGetByID(image.ParentImageID))
		if manifest == nil {
			image.SizeMBWithParents = 0
			image.SizeMBWithoutParent = 0
			return
		}
	}

	// Sum layers
	size := 0
	parentSize := 0
	for i, layer := range manifest.Layers {
		size += layer.Size

		// Count parent layers
		if parent != nil && i < len(parent.Layers) && parent.Layers[i].Digest == layer.Digest {
			parentSize += layer.Size
		}
	}

	// Size = bytes to MB
	image.SizeMBWithParents = size / (1000 * 1000)
	image.SizeMBWithoutParent = (size - parentSize) / (1000 * 1000)

	// For small sizes set to 1
	if image.SizeMBWithParents == 0 && size > 0 {
		image.SizeMBWithParents = 1
	}

	if image.SizeMBWithoutParent == 0 && (size-parentSize) > 0 {
		image.SizeMBWithoutParent = 1
	}

	// Save digest
	image.Digest = manifest.Config.Digest
}

// Exists ...
func (image *ImageS) Exists() bool {
	return getImageManifest(image) != nil
}

func ImageCacheReset(imageID string) {
	imageExternalCacheLock.Lock()
	defer imageExternalCacheLock.Unlock()

	delete(imageExternalCacheMap, imageID)
}

func ImageGetByID(id string) *ImageS {

	image := new(ImageS)

	extIRurl := ""
	extIRtoken := ""
	if extIRurl == "" && extIRtoken == "" {

		if err := driver.Read("image", id, image); err != nil {
			return nil
		}

	} else {

		imageExternalCacheLock.Lock()
		defer imageExternalCacheLock.Unlock()

		if imageExternalCacheMap[id] != nil {
			return imageExternalCacheMap[id]
		}


		res, err := http.Get(extIRurl + "/api/image-external-load?token=" + extIRtoken + "&imageID=" + id)
		if log.Error(err) == nil {
			buf, err := io.ReadAll(res.Body)

			if log.Error(err) == nil && len(buf) > 0 {
				if buf[0] == 1 {
					err = json.Unmarshal(buf[1:], image)
					if log.Error(err) != nil {
						fmt.Println(string(buf))
					}

				} else {
					s := string(buf)
					if strings.Contains(s, "cant find image ID:") {
						return nil

					} else {
						log.Error(s)
					}
				}
			}

		}

		if image.BuildStatus == "ready" || image.BuildStatus == "fail" {
			imageExternalCacheMap[id] = image
		}

	}

	return image
}

func ImageGetLatest(limit int) []*ImageS {
	dir, err := os.Open(filepath.Join(config.DataPath(), "image"))
	if err != nil {
		log.Error(err)
		return nil
	}

	files, err := dir.Readdir(0)
	if err != nil {
		log.Error(err)
		return nil
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	returnFiles := []*ImageS{}
	for i, file := range files {
		file := ImageGetByID(file.Name()[0 : len(file.Name())-5])
		returnFiles = append(returnFiles, file)
		if i >= limit {
			break
		}
	}

	return returnFiles
}

func ImageGetList() []*ImageS {
	imageIDs, err := driver.List("image")
	if err != nil {
		return nil
	}

	images := make([]*ImageS, len(imageIDs))
	for i, imageID := range imageIDs {
		if err := driver.Read("image", imageID, &images[i]); err != nil {
			log.Error(err)
			return nil
		}
	}

	return images
}

func ImageGetListOrderByTime() []*ImageS {
	dir, err := os.Open(filepath.Join(config.DataPath(), "image"))
	if err != nil {
		log.Error(err)
		return nil
	}

	files, err := dir.Readdir(0)
	if err != nil {
		log.Error(err)
		return nil
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	// the IDs of collection
	var imageIDs []string

	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, ".json") && !strings.HasPrefix(name, ".#") {
			imageIDs = append(imageIDs, name[:len(name)-5])
		}
	}

	images := make([]*ImageS, len(imageIDs))
	for i, imageID := range imageIDs {
		if err := driver.Read("image", imageID, &images[i]); err != nil {
			log.Error(err)
			return nil
		}
	}

	return images
}

// ImageManifestS - V2
type ImageManifestS struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"layers"`
}

func getImageManifest(image *ImageS) *ImageManifestS {

	var url string
	if image.ID != "" {
		imageTag := strings.Split(image.ID, ":")
		url = fmt.Sprintf("%s/v2/%s/manifests/%s", db2.TheDomain.URL(""), imageTag[0], imageTag[1])
	}

	req, err := http.NewRequest("GET", url, nil)
	if log.Error(err, url) != nil {
		return nil
	}
	req.Header.Add("User-Agent", "timoni")
	// req.SetBasicAuth("ImageBuilder", global.Config.ImageBuilder.Default_Config().Token)

	// Accept v2 manifests
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	// Do request
	r, err := http.DefaultClient.Do(req)
	if log.Error(err, url) != nil {
		return nil
	}
	defer r.Body.Close()

	buf, err := io.ReadAll(r.Body)
	if log.Error(err, url) != nil {
		return nil
	}

	manifest := &ImageManifestS{}
	err = json.Unmarshal(buf, manifest)
	if log.Error(err, log.Vars{
		"url":  url,
		"body": string(buf),
	}) != nil {
		return nil
	}

	// No digest -> image do not exists in cntreg
	if manifest.Config.Digest == "" {
		return nil
	}

	return manifest
}

func (bp *ImageBuildBlueprintS) RunCMD(args ...string) bool {

	logVars := log.Vars{
		"imageID": bp.ImageID,
		"buildID": bp.BuildID,
		"gitRepo": bp.SourceGit.RepoName,
		"cmd":     strings.Join(args, " "),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	sErr, err := cmd.StderrPipe()
	if log.Error(err, logVars) != nil {
		return false
	}
	defer sErr.Close()

	sOut, err := cmd.StdoutPipe()
	if log.Error(err, logVars) != nil {
		return false
	}
	defer sOut.Close()

	go pipeReader(sErr, "INFO", logVars)
	go pipeReader(sOut, "INFO", logVars)

	cmd.Env = []string{
		"HOME=/tmp",
	}

	// For docker enable buildkit
	if args[0] == "docker" {
		cmd.Env = append(cmd.Env, "DOCKER_BUILDKIT=1")
	}

	err = cmd.Start()
	if log.Error(err, logVars) != nil {
		return false
	}

	cmd.Wait()

	if ctx.Err() == context.DeadlineExceeded {
		log.Error("Command timed out", logVars)
		return false
	}

	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		logVars["exitCode"] = exitCode
		log.Error("exitCode != 0", logVars)
		return false
	}

	return true
}

var dockerBuildKitLineRegex = regexp.MustCompile(`^#([0-9]{1,4}) (?:\[([a-zA-Z0-9._-]+))?`)

func pipeReader(ioIn io.ReadCloser, lvl string, lvars log.Vars) {
	var logLine func(in ...interface{}) *log.RecordS

	switch lvl {
	case "ERROR":
		logLine = log.Error
	default:
		logLine = log.Info
	}

	reader := bufio.NewReader(ioIn)
	stages := map[string]string{}
	for {
		line, _, err := reader.ReadLine()

		if len(line) > 0 {
			line := strings.TrimSpace(string(line))

			// parse docker buildkit step
			matches := dockerBuildKitLineRegex.FindStringSubmatch(line)

			if len(matches) == 0 {
				// line is not a docker buildkit step
				logLine(line, lvars)
				continue
			}

			step := matches[1]

			logVars := *utils.DeepCopy(lvars)
			logVars["step"] = step

			if stageName := matches[2]; stageName != "" {
				stages[step] = stageName
				logVars["stage"] = stageName

			} else if stageName, ok := stages[step]; ok {
				logVars["stage"] = stageName
			}

			logLine(line, logVars)
		}

		if err != nil {
			break
		}
	}

	data := make([]byte, 4096)
	n, _ := reader.Read(data)
	if n > 0 {
		logLine(string(data[:n]), lvars)
	}
}

// ------------------------------------------------------------

func ImageGetNextToBuild() *ImageS {
	for _, image := range ImageGetListOrderByTime() {
		if image.BuildStatus != "pending" {
			continue
		}

		if image.ParentImageID == "" {
			return image
		}

		imageParent := ImageGetByID(image.ParentImageID)
		if imageParent == nil {
			log.Error("imageParent not found", log.Vars{
				"imageParentID": image.ParentImageID,
				"imageID":       image.ID,
				"git-repo":      image.SourceGit.RepoName,
			})
			continue
		}
		if imageParent.BuildStatus == "ready" {
			return image
		}
		if imageParent.BuildStatus == "fail" {
			image.BuildStatus = "fail"
			image.TimeEnd = time.Now()
			image.UpdateSize()
			image.Save()
			continue
		}
	}
	return nil
}
