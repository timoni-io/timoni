package api

import (
	"core/db"
	perms "core/db/permissions"
	"core/imagebuilder"
	"core/kube"
	"encoding/json"
	"lib/tlog"
	"lib/utils/conv"
	"lib/utils/maps"
	"math"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func apiCheckHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func apiCheckAll(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func apiSystemVersion(w http.ResponseWriter, r *http.Request) {
	buf, err := json.MarshalIndent(db.System.Version(), "", "  ")
	if tlog.Error(err) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}

	w.Write(buf)
}

func apiSystemInfo(r *http.Request, user *db.UserS) interface{} {

	if !user.HasGlobPerm(perms.Glob_AccessToAdminZone) {
		return tlog.Error("Permission denied", tlog.Vars{
			"user": user.Email,
		})
	}

	// ---------------------------------------------------

	type tmpImageS struct {
		ID           string
		Status       string
		TimeCreation int64
		TimeBegin    int64
		// BuilderName  string
	}

	res := []*tmpImageS{}
	for _, img := range db.ImageGetListOrderByTime() {
		if img.BuildStatus == "pending" || img.BuildStatus == "building" {
			res = append(res, &tmpImageS{
				ID:           img.ID,
				Status:       img.BuildStatus,
				TimeCreation: conv.UnixTimeStamp(img.TimeCreation),
				TimeBegin:    conv.UnixTimeStamp(img.TimeBegin),
			})
		}
	}

	// ---------------------------------------------------

	return struct {
		Versions           map[string]interface{}
		Nodes              map[string]*db.NodeS
		ImageBuilderQueue  []*tmpImageS
		ImageBuilderStatus *maps.SafeMap[string, *imagebuilder.ImageBuilderS]
		NotificationsSend  bool
		DiskUsage          map[string]int64
	}{
		Versions:           db.System.Version(),
		Nodes:              db.System.Nodes,
		ImageBuilderQueue:  res,
		ImageBuilderStatus: &imagebuilder.ImageBuilderMap,
		DiskUsage:          db.System.DiskUsage,
	}
}

type sResourceS struct {
	Guaranteed int64
	Usage      int64
	Max        int64
}

type gResourceS struct {
	CPU sResourceS
	RAM sResourceS
}

type appResourcesS struct {
	Elements map[string]*gResourceS
	Total    gResourceS
}

type apiResourcesS struct {
	CPUCapacity int64
	RAMCapacity int64
	Resources   map[string]*appResourcesS
}

func apiResources(r *http.Request, user *db.UserS) interface{} {
	const Mi = int64(1024 * 1024)

	if !user.HasGlobPerm(perms.Glob_AccessToAdminZone) {
		return tlog.Error("must be super user", tlog.Vars{
			"user": user.Email,
		})
	}

	kCtl := kube.GetKube()
	if kCtl == nil {
		return tlog.Error("kube client is nil")
	}

	// Get namespaces
	apps, err := kCtl.API.CoreV1().Namespaces().List(kCtl.CTX, metav1.ListOptions{})
	if err != nil {
		return tlog.Error(err)
	}

	out := apiResourcesS{
		Resources:   map[string]*appResourcesS{},
		CPUCapacity: db.System.ClusterInfo.CPU.Capacity,
		RAMCapacity: db.System.ClusterInfo.RAM.Capacity,
	}
	for _, app := range apps.Items {

		// Get pods
		pods, err := kCtl.API.CoreV1().Pods(app.Name).List(kCtl.CTX, metav1.ListOptions{})
		if err != nil {
			tlog.Error(err)
			continue
		}

		// Skip namespaces without pods
		if len(pods.Items) == 0 {
			continue
		}

		// Init
		resources := &appResourcesS{
			Elements: map[string]*gResourceS{},
		}
		out.Resources[app.Name] = resources

		// Get requests and limits
		for _, pod := range pods.Items {
			res := &gResourceS{}
			resources.Elements[pod.Name] = res

			for _, cnt := range pod.Spec.Containers {
				lim := cnt.Resources.Limits
				req := cnt.Resources.Requests

				// Element resources
				res.CPU.Max += lim.Cpu().ScaledValue(-2)
				res.CPU.Guaranteed += req.Cpu().ScaledValue(-2)

				res.RAM.Max += lim.Memory().Value() / Mi        //MiB
				res.RAM.Guaranteed += req.Memory().Value() / Mi //MiB

				// Total
				resources.Total.CPU.Max += res.CPU.Max
				resources.Total.CPU.Guaranteed += res.CPU.Guaranteed

				resources.Total.RAM.Max += res.RAM.Max
				resources.Total.RAM.Guaranteed += res.RAM.Guaranteed
			}

		}

		// Get usage metrics
		podM, err := kCtl.Metrics.MetricsV1beta1().PodMetricses(app.Name).List(kCtl.CTX, metav1.ListOptions{})
		if err != nil {
			tlog.Error(err)
			continue
		}

		for _, pod := range podM.Items {
			res, exists := resources.Elements[pod.Name]
			if !exists {
				res = &gResourceS{}
				resources.Elements[pod.Name] = res
			}

			for _, cnt := range pod.Containers {
				res.CPU.Usage += cnt.Usage.Cpu().ScaledValue(-2)
				res.RAM.Usage += cnt.Usage.Memory().Value() / Mi //MiB
			}
		}

	}

	return out
}

type metrics struct {
	CPU float64
	RSS float64
}

func apiPodCache(r *http.Request, user *db.UserS) interface{} {
	metrics := map[string]map[string]metrics{}
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		return tlog.Error(err)
	}

	defer r.Body.Close()
	go func() {
		for envID, elements := range metrics {
			// var cpu, mem float64
			env := db.EnvironmentMap.Get(envID)
			if env == nil {
				continue
			}
			for elementName, elementMetrics := range elements {
				_, ok := env.Elements.GetFull(elementName)
				if !ok {
					continue
				}
				env.GetElement(elementName).SetMetrics(elementMetrics.CPU, math.Floor(elementMetrics.RSS)/1024)
			}
		}
	}()
	return kube.PodCacheIter()
}

type tmpElementInMatrixS struct {
	GitRepoName   string
	GitBranchName string
	GitCommitSHA  string
	GitTag        string
	VersionTime   int64
}

type tmpColumnS struct {
	Title string `json:"title"`
	Key   string `json:"key"`
}

func apiAllElementVersionsMatrix(r *http.Request, user *db.UserS) interface{} {

	envMap := map[string]bool{}
	elMap := map[string]map[string]tmpElementInMatrixS{} // key= elName => envName => elInfo
	latestVersionMap := map[string]int64{}               // key= {gitRepo}-{commitTime}

	for _, env := range db.EnvironmentMap.Values() {
		for _, elName := range env.Elements.Keys() {
			element := env.GetElement(elName)
			es := element.GetSource()
			if es.CommitHash == "" {
				continue
			}

			if elMap[elName] == nil {
				elMap[elName] = map[string]tmpElementInMatrixS{}
			}
			envName := element.GetEnvironment().Name
			elMap[elName][envName] = tmpElementInMatrixS{
				GitRepoName:   es.RepoName,
				GitBranchName: es.BranchName,
				GitCommitSHA:  es.CommitHash,
				VersionTime:   es.CommitTime,
			}
			envMap[envName] = true

			key := es.RepoName + elName
			if latestVersionMap[key] == 0 {
				latestVersionMap[key] = es.CommitTime
			} else if es.CommitTime > latestVersionMap[key] {
				latestVersionMap[key] = es.CommitTime
			}

		}
	}
	tlog.PrintJSON(elMap)

	columns := []tmpColumnS{{
		Title: "Element Name",
		Key:   "name",
	}}
	data := []map[string]string{}

	for envName := range envMap {
		columns = append(columns, tmpColumnS{
			Title: envName,
			Key:   envName,
		})
	}
	for elName, env := range elMap {
		row := map[string]string{
			"name": elName,
		}
		for envName, elData := range env {
			txt := elData.GitCommitSHA[:8]
			if elData.GitTag != "" {
				txt = elData.GitTag
			}
			if elData.VersionTime != latestVersionMap[elData.GitRepoName+elName] {
				txt += " ^"
			}

			row[envName] = txt
		}
		data = append(data, row)
	}

	return struct {
		Columns []tmpColumnS
		Data    []map[string]string
	}{
		Columns: columns,
		Data:    data,
	}
}
