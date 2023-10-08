package api

import (
	"core/db"
	perms "core/db/permissions"
	"core/imagebuilder"
	"core/imageregistry"
	"encoding/json"
	"fmt"
	"io"
	"lib/tlog"
	"net/http"
	"strings"
	"time"
)

func imageRebuild(image *db.ImageS) (string, error) {
	if image.ParentImageID != "" {
		parent := db.ImageGetByID(image.ParentImageID)
		if parent == nil {
			return image.ParentImageID, fmt.Errorf("image not found")
		}
		if parent.BuildStatus == "fail" || !parent.Exists() {
			id, err := imageRebuild(parent)
			if err != nil {
				return id, err
			}
		}
	}

	db.ImageCacheReset(image.ID)

	image.BuildStatus = "pending"
	image.BuildID = ""
	image.TimeCreation = time.Now().UTC()
	image.TimeBegin = time.Time{}
	image.TimeEnd = time.Time{}
	image.SizeMBWithParents = 0
	image.SizeMBWithoutParent = 0
	image.Save()
	return "", nil
}

func apiImageRebuild(r *http.Request, user *db.UserS) interface{} {

	imageID := r.FormValue("imageID")
	if imageID == "" {
		return tlog.Error("`imageID` is required")
	}

	envID := r.FormValue("envID")
	if envID == "" {
		return tlog.Error("`envID` is required")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		return tlog.Error("env not found", tlog.Vars{
			"user":  user.Email,
			"env":   envID,
			"image": imageID,
		})
	}

	logVars := tlog.Vars{
		"user":  user.Email,
		"env":   envID,
		"image": imageID,
		"event": true,
	}

	image := db.ImageGetByID(imageID)
	if image == nil {
		return tlog.Error("cant find image ID: "+imageID, logVars)
	}

	// if image.BuildStatus != "fail" {
	// 	return tlog.Error("cant rebuild image which not fail: " + imageID)
	// }

	_, err := imageRebuild(image)
	if err != nil {
		return tlog.Error(err, logVars)
	}

	for _, env := range db.EnvironmentMap.Values() {
		for _, elName := range env.Elements.Keys() {
			env.GetElement(elName).RebuildImage(imageID, user)
		}
	}

	return tlog.Info("Image set to rebuild", logVars)
}

func apiImageList(r *http.Request, user *db.UserS) interface{} {

	if !user.HasGlobPerm(perms.Glob_AccessToAdminZone) {
		return tlog.Error("You are not authorized", tlog.Vars{
			"user": user.Email,
		})
	}

	return getImageInfoMap()
}

type tmpImageS struct {
	ID                string
	ParentImageID     string
	BuildStatus       string
	SizeWithParents   int // MB
	SizeWithoutParent int // MB
	Registred         bool
	BuildSHA          string
	Usage             map[string]int // envID => ilosc elementow, która aktualnie uzywa tego obrazu
	UsageCount        int
}

type tmpProjectS struct {
	Name                      string
	ImagesCount               int
	ImagesSizeMBWithoutParent int // MB
	Images                    []tmpImageS
	UsageCount                int
}

func getImageInfoMap() map[string]*tmpProjectS {
	imageUsageMap := getImageUsageMap()
	res := map[string]*tmpProjectS{}
	for _, img := range db.ImageGetList() {
		project, exist := res[img.SourceGit.RepoName]
		if !exist {
			project = &tmpProjectS{
				Name:   img.SourceGit.RepoName,
				Images: []tmpImageS{},
			}
		}

		project.ImagesCount++
		project.ImagesSizeMBWithoutParent += img.SizeMBWithoutParent
		project.Images = append(project.Images, tmpImageS{
			ID:                img.ID,
			ParentImageID:     img.ParentImageID,
			BuildStatus:       img.BuildStatus,
			SizeWithParents:   img.SizeMBWithParents,
			SizeWithoutParent: img.SizeMBWithoutParent,
			Usage:             imageUsageMap[img.ID],
			UsageCount:        sumCount(imageUsageMap[img.ID]),
		})
		res[img.SourceGit.RepoName] = project

		for _, i := range imageUsageMap[img.ID] {
			res[img.SourceGit.RepoName].UsageCount += i
		}
	}
	return res
}

func apiImageUpdateStatus(r *http.Request, user *db.UserS) interface{} {
	if user.Email != "ImageBuilder" {
		return "Access Denied"
	}

	imageBP := &db.ImageBuildBlueprintS{}
	tlog.Error(json.NewDecoder(r.Body).Decode(imageBP))

	if imageBP.EntryPointGitTag == "" {
		return "EntryPointGitTag required"
	}

	if imageBP.ImageID == "" {
		return "ImageID required"
	}

	if imageBP.ImageBuilderID == "" {
		return "ImageBuilderID required"
	}

	status := imageBP.Status
	if status == "" || (status != "ready" && status != "fail") {
		return "Status must be ready or fail"
	}

	image := db.ImageGetByID(imageBP.ImageID)
	if image == nil {
		return "image with this is ID is missing"
	}
	image.EntryPointVersion = imageBP.EntryPointGitTag
	image.BuildStatus = status
	image.TimeEnd = time.Now()
	image.UpdateSize()
	image.Save()

	// tlog.PrintJSON(imageBP)
	// tlog.PrintJSON(image)

	imagebuilder.ImageBuilderSetReady(imageBP.ImageBuilderID)

	tlog.Info("Image set status", tlog.Vars{
		"user":        user.Email,
		"image":       image.ID,
		"BuildStatus": image.BuildStatus,
		"project":     image.SourceGit.RepoName,
	})

	return "ok"
}

func apiImageExternalSave(r *http.Request, user *db.UserS) interface{} {

	if user.Email != "ImageBuilder" {
		return "Access Denied"
	}

	image := &db.ImageS{}

	buf, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(buf, image)
	if tlog.Error(err, string(buf)) != nil {
		return err.Error()
	}

	image.Save()
	return "ok"
}

func apiImageExternalLoad(r *http.Request, user *db.UserS) interface{} {

	if user.Email != "ImageBuilder" {
		return "Access Denied"
	}

	imageID := r.FormValue("imageID")
	if imageID == "" {
		return tlog.Error("`imageID` is required")
	}

	image := db.ImageGetByID(imageID)
	if image == nil {
		return tlog.Error("cant find image ID: " + imageID)
	}

	return image
}

func getImageUsageMap() map[string]map[string]int { // imageID => envID => count in env

	res := map[string]map[string]int{}

	for _, env := range db.EnvironmentMap.Values() {
		for _, el := range env.Elements.Keys() {
			element := env.GetElement(el)
			if element == nil {
				continue
			}
			image := element.GetImage()
			if image == nil {
				continue
			}
			imageID := image.ID
			if imageID == "" {
				continue
			}

			if res[imageID] == nil {
				res[imageID] = map[string]int{}
			}
			res[imageID][env.ID]++
		}
	}

	return res
}

func sumCount(data map[string]int) int {
	sum := 0
	for _, v := range data {
		sum += v
	}
	return sum
}

func apiImageDeleteUnused(w http.ResponseWriter, r *http.Request) {

	// if config.ExternalImageRegistryURL != "" {
	// 	return log.Error("ExternalImageRegistry")
	// }

	deletedImages := []string{}
	ir := imageregistry.Open()
	for _, repo := range getImageInfoMap() {
		for _, image := range repo.Images {
			if image.UsageCount > 0 {
				continue
			}

			img := db.ImageGetByID(image.ID)
			if img == nil {
				continue
			}

			ir.DeleteImage(image.ID)
			deletedImages = append(deletedImages, " * "+image.ID)
		}
	}

	if len(deletedImages) == 0 {
		w.Write([]byte("There is nothing to delete"))
		return
	}

	ir = imageregistry.Open()
	deletedBlobs := ir.DeleteUnusedBlobs()

	out := "Deleted images:\n"
	out += strings.Join(deletedImages, "\n")

	out += "\nDeleted blobs:\n"
	out += strings.Join(deletedBlobs, "\n")

	out += "\nSuccess"
	w.Write([]byte(out))
}
