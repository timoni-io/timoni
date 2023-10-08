package kube

import (
	"fmt"
	"os"

	log "lib/tlog"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// PodS ...
type PodS struct {
	KubeClient         *ClientS
	Namespace          string
	Name               string
	ReplicaSetRevision int
	Obj                corev1.Pod
}

type PodStatusS int

var PodStatusLabels = map[int]string{
	0: "New",
	1: "Pending",
	2: "Creating",
	3: "Running",
	4: "Succeeded",
	5: "Failed",
	6: "Terminating",
	7: "Ready",
}

func (ps PodStatusS) String() string {
	return PodStatusLabels[int(ps)]
}

const (
	PodStatusNew PodStatusS = iota
	PodStatusPending
	PodStatusCreating
	PodStatusRunning
	PodStatusSucceeded
	PodStatusFailed
	PodStatusTerminating
	PodStatusReady
)

type WarningS struct {
	Message      string
	Reason       string
	RestartCount int
	ExitCode     int32
}

// Status ...
func (pod *PodS) Status() PodStatusS {

	switch pod.Obj.Status.Phase {
	case corev1.PodPending:
		log.Debug("IsReady: pod is PodPending", pod.Name)
		return PodStatusPending

	case corev1.PodSucceeded:
		log.Debug("IsReady: pod is PodSucceeded", pod.Name)
		return PodStatusSucceeded

	case corev1.PodFailed, corev1.PodUnknown:
		log.Debug("IsReady: pod is PodFailed", pod.Name)
		return PodStatusFailed
	}

	if pod.Obj.DeletionTimestamp != nil {
		log.Debug("IsReady: pod is Terminating", pod.Name)
		return PodStatusTerminating
	}

	for _, cnt := range pod.Obj.Status.ContainerStatuses {
		if cnt.State.Waiting != nil {
			log.Debug("IsReady: pod is Waiting", pod.Name)
			return PodStatusCreating
		}
		if cnt.State.Terminated != nil {
			log.Debug("IsReady: pod is Terminated", pod.Name)
			return PodStatusFailed
		}
		if !cnt.Ready {
			log.Debug("IsReady: pod is Running but not Ready", pod.Name)
			return PodStatusRunning
		}
	}
	return PodStatusReady
}

func (pod *PodS) Alerts() []*WarningS {
	warnings := []*WarningS{}
	for _, status := range pod.Obj.Status.ContainerStatuses {
		if *status.Started {
			continue
		}
		if status.State.Waiting != nil {
			warning := &WarningS{
				RestartCount: int(status.RestartCount),
				Message:      status.State.Waiting.Message,
				Reason:       status.State.Waiting.Reason,
			}
			warnings = append(warnings, warning)
			continue
		}
		if status.State.Terminated != nil {
			warning := &WarningS{
				RestartCount: int(status.RestartCount),
				Message:      status.State.Terminated.Message,
				Reason:       status.State.Terminated.Reason,
				ExitCode:     status.State.Terminated.ExitCode,
			}
			warnings = append(warnings, warning)
		}

	}
	return warnings
}

func (pod *PodS) ExecToStdOut(cmd []string) error {

	req := pod.KubeClient.API.CoreV1().RESTClient().Post().Resource("pods").Name(pod.Name).Namespace(pod.Namespace).SubResource("exec")
	req.VersionedParams(
		&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		},
		scheme.ParameterCodec,
	)

	exec, err := remotecommand.NewSPDYExecutor(pod.KubeClient.Config, "POST", req.URL())
	if err != nil {
		panic(err)
	}
	err = exec.StreamWithContext(pod.KubeClient.CTX, remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	})
	return err
}

func (kube *ClientS) PodListAll(namespace string) map[string]*PodS {
	pods, err := kube.API.CoreV1().Pods(namespace).List(kube.CTX, metav1.ListOptions{})
	if log.Error(err) != nil {
		return nil
	}

	res := map[string]*PodS{}
	for _, pod := range pods.Items {
		res[pod.Name] = &PodS{
			KubeClient: kube,
			Namespace:  pod.Namespace,
			Name:       pod.Name,
			Obj:        pod,
		}
	}

	return res
}

func (kube *ClientS) PodListByNodeIP(namespace, nodeIP string) []*PodS {
	pods := kube.PodListAll(namespace)

	filtered := []*PodS{}
	for _, pod := range pods {
		if pod.Obj.Status.HostIP == nodeIP {
			filtered = append(filtered, pod)
		}
	}

	return filtered
}

func (pod *PodS) IsStorageReady() bool {

	allOK := true

	for _, vol := range pod.Obj.Spec.Volumes {
		if vol.VolumeSource.PersistentVolumeClaim != nil {
			pvcName := vol.VolumeSource.PersistentVolumeClaim.ClaimName
			pvc, err := pod.KubeClient.API.CoreV1().PersistentVolumeClaims(pod.Namespace).Get(pod.KubeClient.CTX, pvcName, metav1.GetOptions{})
			if err != nil {
				fmt.Println("ERROR: pod.IsStorageReady()", pod.Name, err)
				return false
			}

			if pvc.Status.Phase != corev1.ClaimBound {
				allOK = false
			}
		}
	}

	return allOK
}

func (pod *PodS) ReplicaSet() *appsv1.ReplicaSet {

	if len(pod.Obj.OwnerReferences) == 0 {
		return nil
	}

	firstOwner := pod.Obj.OwnerReferences[0]
	if firstOwner.Kind != "ReplicaSet" {
		return nil
	}

	rs, err := pod.KubeClient.API.AppsV1().ReplicaSets(pod.Namespace).Get(pod.KubeClient.CTX, firstOwner.Name, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	return rs
}

func (pod *PodS) RestartCount() int32 {
	if len(pod.Obj.Status.ContainerStatuses) == 0 {
		return 0
	}
	return pod.Obj.Status.ContainerStatuses[0].RestartCount
}

func (pod *PodS) StartTime() int64 {
	if len(pod.Obj.Status.ContainerStatuses) == 0 {
		return 0
	}
	ContainerStateRunning := pod.Obj.Status.ContainerStatuses[0].State.Running
	if ContainerStateRunning == nil {
		return 0
	}
	return ContainerStateRunning.StartedAt.Unix()
}

func (pod *PodS) CreationTime() int64 {
	if len(pod.Obj.Status.ContainerStatuses) == 0 {
		return 0
	}
	return pod.Obj.Status.StartTime.Unix()
}

func (pod *PodS) Delete() {
	pod.KubeClient.API.CoreV1().Pods(pod.Namespace).Delete(pod.KubeClient.CTX, pod.Name, metav1.DeleteOptions{})
}
