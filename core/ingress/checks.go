package ingress

import (
	"core/db2"
	"core/kube"
)

func Check() (db2.StateT, string) {
	kClient := kube.GetKube()
	obj := &kube.DaemonSetS{
		KubeClient: kClient,
		Namespace:  "timoni",
		Name:       "ingress-traefik",
	}
	if !obj.IsReady() {
		return db2.State_deploying, "DaemonSet `ingress-traefik` not ready"
	}

	return db2.State_ready, ""
}
