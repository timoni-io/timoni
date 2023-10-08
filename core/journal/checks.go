package journal

import (
	"core/db2"
	"core/kube"
	"fmt"
)

func CheckDB() (db2.StateT, string) {
	kClient := kube.GetKube()
	obj := &kube.StatefulSetS{
		KubeClient: kClient,
		Namespace:  "timoni",
		Name:       "chi-timoni-cluster-timoni-1-0",
	}
	if !obj.IsReady() {
		return db2.State_deploying, "StatefulSet `chi-timoni-cluster-timoni-1-0` not ready"
	}

	return db2.State_ready, ""
}

func CheckProxy() (db2.StateT, string) {
	kClient := kube.GetKube()
	obj := &kube.DeploymentS{
		KubeClient: kClient,
		Namespace:  "timoni",
		Name:       "journal-proxy-1",
	}
	if !obj.IsReady() {
		return db2.State_deploying, "Deployment `journal-proxy-1` not ready"
	}

	_, imageTag := kube.GetImageInfo(obj.Obj.Spec.Template.Spec.Containers[0].Image)
	if imageTag != db2.TheSettings.ReleaseGitTag() {
		return db2.State_error, fmt.Sprintf("Deployment `journal-proxy-1` ImageTag is wrong `%s` != `%s`", imageTag, db2.TheSettings.ReleaseGitTag())
	}

	return db2.State_ready, ""

}
