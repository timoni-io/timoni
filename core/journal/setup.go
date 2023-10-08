package journal

import (
	"core/db2"
	"core/kube"
	"core/modulestate"
	"encoding/base64"
	"encoding/json"
	"lib/tlog"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type JournalProxyS struct {
	Name  string
	Count int32

	DatabaseConnections int `max:"10"` // no more than number of CPU cores
	DatabaseAddress     string

	CacheValuesLimit uint16
	CacheEntryLimit  uint16
	MaxEntriesLimit  int
}

func Setup() {
	modulestate.StatusByModulesAdd("journal-proxy", CheckProxy)

	kClient := kube.GetKube()

	buf, _ := json.Marshal(JournalProxyS{
		Name:  "jp1",
		Count: 1,

		DatabaseConnections: 5,
		DatabaseAddress:     "",

		CacheValuesLimit: 1000,
		CacheEntryLimit:  1000,
		MaxEntriesLimit:  1000,
	})
	conf := base64.StdEncoding.EncodeToString(buf)

	dep := &kube.DeploymentS{
		KubeClient: kClient,
		Namespace:  "timoni",
		Name:       "journal-proxy-1",
		// Command: []string{"/bin/sleep", "infinity"},
		Replicas:               1,
		Image:                  "timoni/journal-proxy:" + db2.TheSettings.ReleaseGitTag(),
		ImagePullAlways:        true,
		ExposePorts:            []int32{4003},
		WritableRootFilesystem: true,
		Probe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   "/status",
					Port:   intstr.FromInt(4003),
					Scheme: v1.URISchemeHTTP,
				},
			},
		},

		// CPUGuaranteed:          100,
		// RAMGuaranteed:          1000,
		// CPUMax:                 500,
		// RAMMax:                 6000,

		Envs: map[string]string{
			"CONF":     conf,
			"LOG_MODE": "mjson",
		},
	}

	for {
		_, err := dep.CreateOrUpdate()
		if tlog.Error(err) == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	// ---

	svc := kube.ServiceS{
		KubeClient: kClient,
		Namespace:  "timoni",
		Name:       "journal-proxy-1",
		Ports:      map[int32]int32{4003: 4003},
		TargetSelector: map[string]string{
			"element": "journal-proxy-1",
		},
		Labels: map[string]string{
			"element": "journal-proxy-1",
		},
	}
	for {
		_, err := svc.CreateOrUpdate()
		if tlog.Error(err) == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	for {
		if dep.IsReady() {
			break
		}
		time.Sleep(3 * time.Second)
	}
}
