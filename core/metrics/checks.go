package metrics

import (
	"core/db2"
	"core/kube"
	"fmt"
)

func CheckVictoria() (db2.StateT, string) {
	kClient := kube.GetKube()
	obj := &kube.StatefulSetS{
		KubeClient: kClient,
		Namespace:  "timoni-metrics",
		Name:       "vmstorage-vmcluster-persistent",
	}
	if !obj.IsReady() {
		return db2.State_deploying, "StatefulSet `vmstorage-vmcluster-persistent` not ready"
	}

	return db2.State_ready, ""
}

func CheckGrafana() (db2.StateT, string) {
	kClient := kube.GetKube()
	obj := &kube.DeploymentS{
		KubeClient: kClient,
		Namespace:  "timoni",
		Name:       "metrics-grafana",
	}
	if !obj.IsReady() {
		return db2.State_deploying, "Deployment `metrics-grafana` not ready"
	}

	_, imageTag := kube.GetImageInfo(obj.Obj.Spec.Template.Spec.Containers[0].Image)
	if imageTag != db2.TheSettings.ReleaseGitTag() {
		return db2.State_error, fmt.Sprintf("Deployment `metrics-grafana` ImageTag is wrong `%s` != `%s`", imageTag, db2.TheSettings.ReleaseGitTag())
	}

	return db2.State_ready, ""
}

func CheckNodeAgent() (db2.StateT, string) {
	kClient := kube.GetKube()
	obj := &kube.DaemonSetS{
		KubeClient: kClient,
		Namespace:  "timoni",
		Name:       "node-agent",
	}
	if !obj.IsReady() {
		return db2.State_deploying, "DaemonSet `node-agent` not ready"
	}

	return db2.State_ready, ""
}
