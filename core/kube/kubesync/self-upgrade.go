package kubesync

import (
	"core/kube"
	"encoding/json"
	"lib/tlog"

	jsonpatch "github.com/evanphx/json-patch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func selfUpgrade(imageTag string) {
	kClient := kube.GetKube()

	statefulSet, err := kClient.API.AppsV1().StatefulSets("timoni").Get(kClient.CTX, "core", metav1.GetOptions{})
	if err != nil {
		tlog.Fatal("StatefulSet timoni.core not found: " + err.Error())
	}

	statefulSetOld, err := json.Marshal(statefulSet)
	if err != nil {
		tlog.Fatal(err)
	}

	statefulSet.Spec.Template.Spec.Containers[0].Image = "timoni/core:" + imageTag

	statefulSetNew, err := json.Marshal(statefulSet)
	if err != nil {
		tlog.Fatal(err)
	}

	patch, err := jsonpatch.CreateMergePatch(statefulSetOld, statefulSetNew)
	if err != nil {
		tlog.Fatal(err)
	}

	if len(patch) == 2 {
		// nie ma zadnych zmian
		return
	}

	_, err = kClient.API.AppsV1().StatefulSets("timoni").Patch(kClient.CTX, "core", types.MergePatchType, patch, metav1.PatchOptions{})
	tlog.Fatal(err)
}
