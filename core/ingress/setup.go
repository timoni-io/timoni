package ingress

import (
	"core/config"
	"core/db2"
	"core/kube"
	"core/modulestate"
	"lib/tlog"
	"path/filepath"
	"time"
)

func Setup() {
	modulestate.StatusByModulesAdd("ingress-traefik", Check)

	kClient := kube.GetKube()

	kClient.ApplyYamlFilesInDir(filepath.Join(config.ModulesPath(), "ingress"), nil)

	// -------------------------------------------------
	// traefik

	traefik := &kube.DaemonSetS{
		KubeClient:          kClient,
		Namespace:           "timoni",
		Name:                "ingress-traefik",
		Image:               "timoni/ingress-traefik:" + db2.TheSettings.ReleaseGitTag(),
		ImagePullAlways:     true,
		ExposePorts:         []int32{80, 443, 9000},
		ServiceAccountName:  "traefik-ingress-controller",
		ServiceAccountMount: true,
		// Annotations:         db2.TheIngress.Traefik_Annotations(), 
		// Privileged:             true,
		// HostPID: true,
		// WritableRootFilesystem: true,

		// 	resources:
		// 	requests:
		// 	  memory: 200Mi
		// 	limits:
		// 	  memory: 1Gi
		// 	  # cpu: '10000m'
		//   ports:
		// 	- name: web
		// 	  containerPort: 80
		// 	- name: websecure
		// 	  containerPort: 443
		// 	- name: traefik
		// 	  containerPort: 9000

	}
	for {
		_, err := traefik.CreateOrUpdate()
		if tlog.Error(err) == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	for {
		isvc := kube.ServiceS{
			KubeClient: kClient,
			Namespace:  "timoni",
			Name:       "ingress-traefik",
			TargetSelector: map[string]string{
				"element": "ingress-traefik",
			},
			LoadBalancer: true,
			// Annotations:  db2.TheIngress.Traefik_Annotations(),
			Internal: db2.TheIngress.Traefik_Internal(),
			Ports: map[int32]int32{
				80:  80,
				443: 443,
			},
		}
		_, err := isvc.CreateOrUpdate()
		tlog.Error(err)

		is := isvc.GetObj()
		if is == nil || is.Name != isvc.Name || len(is.Status.LoadBalancer.Ingress) == 0 {
			tlog.Warning("Waiting for ingress traefik service IP...")
			time.Sleep(5 * time.Second)
			continue
		}
		if len(is.Status.LoadBalancer.Ingress) > 0 {
			break
		}
	}
}
