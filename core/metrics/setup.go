package metrics

import (
	"core/config"
	"core/db2"
	"core/kube"
	"core/modulestate"
	"fmt"
	"lib/tlog"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Setup() {

	if !db2.TheMetrics.Enabled() {
		tlog.Debug("SKIP_METRICS")
		return
	}

	modulestate.StatusByModulesAdd("metrics-victoria", CheckVictoria)
	modulestate.StatusByModulesAdd("metrics-grafana", CheckGrafana)
	modulestate.StatusByModulesAdd("metrics-node-agent", CheckNodeAgent)

	if db2.TheSettings.APIInternalIP() == "" {
		tlog.Fatal("db2.TheSettings.APIInternalIP() is empty")
	}

	kClient := kube.GetKube()
	kClient.NamespaceCreate("timoni-metrics")

	// -------------------------------------------------

	for {
		if kClient.ApplyYamlFilesInDir(filepath.Join(config.ModulesPath(), "metrics"), map[string]string{
			"VictoriaMetricsClusterSize": fmt.Sprint(db2.TheMetrics.VictoriaMetricsClusterSize()),
		}) {
			break
		}
		time.Sleep(5 * time.Second)
	}

	// -------------------------------------------------
	// node-agent

	nodeAgent := &kube.DaemonSetS{
		KubeClient:             kClient,
		Namespace:              "timoni",
		Name:                   "node-agent",
		Image:                  "timoni/node-agent:" + db2.TheSettings.ReleaseGitTag(),
		ImagePullAlways:        true,
		WritableRootFilesystem: true,
		Privileged:             true,
		HostPID:                true,
		// Probe: &v1.Probe{
		// 	ProbeHandler: v1.ProbeHandler{
		// 		HTTPGet: &v1.HTTPGetAction{
		// 			Path:   "/",
		// 			Port:   intstr.FromInt(3000),
		// 			Scheme: v1.URISchemeHTTP,
		// 		},
		// 	},
		// },
		HostAliases: map[string][]string{
			db2.TheSettings.APIInternalIP(): {db2.TheDomain.Name()},
		},
		Envs: map[string]string{
			"TIMONI_URL":    db2.TheDomain.URL(""),
			"VM_AGENT_ADDR": "vmagent-vmagent2.timoni-metrics:4242",
		},
	}
	for {
		_, err := nodeAgent.CreateOrUpdate()
		if tlog.Error(err) == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	if db2.TheMetrics.Grafana() {
		setupGrafana()
	}
}

func setupGrafana() {
	// -------------------------------------------------
	// grafana

	kClient := kube.GetKube()

	grafana := &kube.DeploymentS{
		KubeClient:             kClient,
		Namespace:              "timoni",
		Name:                   "metrics-grafana",
		Image:                  "timoni/metrics-grafana:" + db2.TheSettings.ReleaseGitTag(),
		ImagePullAlways:        true,
		ExposePorts:            []int32{3000},
		WritableRootFilesystem: true,
		Probe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   "/",
					Port:   intstr.FromInt(3000),
					Scheme: v1.URISchemeHTTP,
				},
			},
		},

		Envs: map[string]string{
			"GF_SECURITY_ADMIN_PASSWORD": db2.TheMetrics.GrafanaAdminPassword(),
		},
	}
	for {
		_, err := grafana.CreateOrUpdate()
		if tlog.Error(err) == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	// ---

	svc := kube.ServiceS{
		KubeClient: kClient,
		Namespace:  "timoni",
		Name:       "metrics-grafana",
		Ports:      map[int32]int32{3000: 3000},
		TargetSelector: map[string]string{
			"element": "metrics-grafana",
		},
		Labels: map[string]string{
			"element": "metrics-grafana",
		},
	}
	for {
		_, err := svc.CreateOrUpdate()
		if tlog.Error(err) == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	// -------------------------------------------------
}
