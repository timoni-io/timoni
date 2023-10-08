package imagebuilder

import (
	"bytes"
	"core/db"
	"core/kube"
	"encoding/json"
	"fmt"
	"io"
	"lib/tlog"
	"lib/utils/maps"
	"lib/utils/net"
	"net/http"
	"time"
)

var ImageBuilderMap = *maps.NewSafe[string, *ImageBuilderS](nil)

type ImageBuilderS struct {
	PodName string
	IP      string

	StatusUpdateTime time.Time
	StatusPodExist   bool
	StatusHTTPAlive  bool
	StatusBuilding   bool
	Blueprint        *db.ImageBuildBlueprintS
}

// ------------------------------------

func (ib *ImageBuilderS) startBuild(image *db.ImageS) (success bool) {
	image.BuildStatus = db.Building
	image.TimeBegin = time.Now()
	image.TimeEnd = time.Time{}
	image.BuildID = fmt.Sprint(time.Now().Unix())
	image.Save()

	ib.Blueprint = &db.ImageBuildBlueprintS{
		ImageID:           image.ID,
		BuildID:           image.BuildID,
		SourceGit:         image.SourceGit,
		DockerFilePath:    image.DockerFilePath,
		BuildRootPath:     image.BuildRootPath,
		DockerFileContent: image.DockerFileContent,
		ImageBuilderID:    ib.PodName,
	}

	blueprintsBuf, err := json.Marshal(ib.Blueprint)
	if err != nil {
		tlog.Error(err, tlog.Vars{
			"imageID":        image.ID,
			"buildID":        image.BuildID,
			"git-repo":       image.SourceGit.RepoName,
			"imageBuilderIP": ib.IP,
		})
		return false
	}

	url := fmt.Sprintf("http://%s:6666/build", ib.IP)
	r, err := http.Post(url, "application/json", bytes.NewBuffer(blueprintsBuf))
	if err != nil {
		tlog.Error(err, tlog.Vars{
			"imageID":        image.ID,
			"buildID":        image.BuildID,
			"git-repo":       image.SourceGit.RepoName,
			"imageBuilderIP": ib.IP,
		})
		return false
	}

	rBody := net.ReadBodyFromResponse(r)
	if rBody != "ok" {
		tlog.Error(rBody, tlog.Vars{
			"imageID":        image.ID,
			"buildID":        image.BuildID,
			"git-repo":       image.SourceGit.RepoName,
			"imageBuilderIP": ib.IP,
		})
		return false
	}

	// Update image builder
	ib.StatusBuilding = true
	// ib.Blueprints = blueprints

	return true
}

// ------------------------------------

func (ib *ImageBuilderS) updateStatus(podExists bool) {

	ib.StatusUpdateTime = time.Now().UTC()

	// ------------------------------------
	// StatusPodExist

	ib.StatusPodExist = podExists
	if !ib.StatusPodExist {
		ib.StatusHTTPAlive = false
		ib.StatusBuilding = false
		return
	}

	// ------------------------------------
	// StatusHTTPAlive

	r, err := http.Get("http://" + ib.IP + ":6666/status")
	if tlog.Error(err) != nil {
		ib.StatusHTTPAlive = false
		ib.StatusBuilding = false
		return
	}
	defer r.Body.Close()

	ib.StatusHTTPAlive = true

	// ------------------------------------
	// StatusBuilding

	status, err := io.ReadAll(r.Body)
	if tlog.Error(err) != nil {
		ib.StatusBuilding = false
		return
	}

	if string(status) == "ready" {
		ib.StatusBuilding = false
		return
	}
	ib.StatusBuilding = true
}

// ------------------------------------

func ImageBuilderGetReady() *ImageBuilderS {

	if imageBuilderDepoyment == nil && imageBuilderStatefulSet == nil {
		tlog.Error("imagebuilder.Setup() nie zostal wykonany")
		return nil
	}

	var podList []*kube.PodS
	if imageBuilderDepoyment != nil {
		podList = imageBuilderDepoyment.PodList(true)
	}
	if imageBuilderStatefulSet != nil {
		podList = imageBuilderStatefulSet.PodList(true)
	}

	// list to map
	podMap := map[string]*kube.PodS{}
	for _, pod := range podList {
		podMap[pod.Name] = pod
	}

	ImageBuilderMap.Commit(func(data map[string]*ImageBuilderS) {
		for _, pod := range podList {
			ib, exists := data[pod.Name]

			// Add new image builder
			if !exists {
				ib = &ImageBuilderS{
					PodName:          pod.Name,
					IP:               pod.Obj.Status.PodIP,
					StatusUpdateTime: time.Now().UTC(),
					StatusPodExist:   true,
				}
				data[pod.Name] = ib
			}
			// Update status
			ib.updateStatus(true)
		}

		// delete keys that are not in the pod map
		if len(podList) != len(data) {
			for name := range data {
				if _, exists := podMap[name]; !exists {
					delete(data, name)
				}
			}
		}
	})

	if imageBuilderDepoyment != nil {
		imageBuilderDepoyment.Cleanup()
	}
	if imageBuilderStatefulSet != nil {
		imageBuilderStatefulSet.Cleanup()
	}

	for _, ib := range ImageBuilderMap.Values() {
		if ib.StatusPodExist && ib.StatusHTTPAlive && !ib.StatusBuilding {
			return ib
		}
	}

	return nil
}

// ------------------------------------

func ImageBuilderSetReady(imageBuilderPodName string) {
	ib, exist := ImageBuilderMap.GetFull(imageBuilderPodName)
	if exist && ib != nil {
		ib.StatusBuilding = false
		// ib.Blueprints = nil
	}
}
