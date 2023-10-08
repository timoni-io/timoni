package imagebuilder

import (
	"core/db2"
	"core/kube"
	"core/modulestate"
	"fmt"
	"lib/tlog"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var imageBuilderDepoyment *kube.DeploymentS
var imageBuilderStatefulSet *kube.StatefulSetS

func Setup() {
	modulestate.StatusByModulesAdd("image-builder", CheckImageBuilder)
	modulestate.StatusByModulesAdd("image-builder-queue", CheckImageBuilderQueue)

	if db2.TheSettings.APIInternalIP() == "" {
		tlog.Fatal("TheSettings.APIInternalIP is empty")
	}

	deployment()
	go Loop()
}

func deployment() {
	kClient := kube.GetKube()

	imageBuilderDepoyment = &kube.DeploymentS{
		KubeClient:             kClient,
		Namespace:              "timoni",
		Name:                   "image-builder",
		Image:                  "timoni/image-builder:" + db2.TheSettings.ReleaseGitTag(),
		ImagePullAlways:        true,
		Privileged:             true,
		WritableRootFilesystem: true,
		// Replicas:               global.Config.ImageBuilderDefault().MaxCount,
		Replicas:    1,
		ExposePorts: []int32{6666},
		Probe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   "/status",
					Port:   intstr.FromInt(6666),
					Scheme: v1.URISchemeHTTP,
				},
			},
		},

		// CPUGuaranteed:          100,
		// RAMGuaranteed:          1000,
		// CPUMax:                 500,
		// RAMMax:                 6000,

		Envs: map[string]string{
			"BUILDKIT_STEP_LOG_MAX_SIZE":  "5242880",
			"BUILDKIT_STEP_LOG_MAX_SPEED": "1000000",
			"DOCKER_HOST":                 "unix:///var/run/docker.sock",
			"HOME":                        "/tmp",
			"TIMONI_IP":                   db2.TheSettings.APIInternalIP(),
			"TIMONI_DOMAIN":               db2.TheDomain.Name(),
			"TIMONI_PORT":                 fmt.Sprint(db2.TheDomain.Port()),
			"TIMONI_TOKEN":                db2.TheImageBuilder.Timoni_Token(),
			"LOG_MODE":                    "json",
		},
	}

	for {
		_, err := imageBuilderDepoyment.CreateOrUpdate()
		if tlog.Error(err) == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
}