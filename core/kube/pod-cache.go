package kube

import (
	"lib/utils/maps"
	"strings"

	log "lib/tlog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	PodCache = maps.NewSafe[string, podCacheDataS](nil)
)

type podCacheDataS struct {
	EnvID       string
	ElementName string
}

func (ctl *ClientS) PodCacheUpdate() {

	if ctl == nil {
		log.Warning("kubeClient is empty")
		return
	}

	pods, err := ctl.API.CoreV1().Pods("").List(ctl.CTX, metav1.ListOptions{})
	if log.Error(err) != nil {
		return
	}
	tmp := map[string]podCacheDataS{}
	for _, pod := range pods.Items {

		podID := strings.ReplaceAll(string(pod.ObjectMeta.UID), "-", "_")
		elementName := pod.ObjectMeta.Labels["element"]

		if elementName == "" {
			if len(pod.GetOwnerReferences()) > 0 {
				owner := pod.GetOwnerReferences()[0]
				elementName = owner.Name

				if owner.Kind == "ReplicaSet" {
					tmp := strings.Split(owner.Name, "-")
					elementName = strings.Join(tmp[:len(tmp)-1], "-")
				}

			}
		}
		tmp[podID] = podCacheDataS{
			EnvID:       pod.Namespace,
			ElementName: elementName,
		}
	}
	PodCache = maps.New(tmp).Safe()
}

func PodCacheIter() map[string]podCacheDataS {

	res := map[string]podCacheDataS{}
	for v := range PodCache.Iter() {
		res[v.Key] = v.Value
	}

	return res
}
