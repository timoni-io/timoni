package kube

import (
	"encoding/json"
	"errors"
	"fmt"
	"lib/tlog"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type StatefulSetS struct {
	KubeClient             *ClientS
	Namespace              string
	Name                   string
	Image                  string
	Command                []string
	Replicas               int32
	Envs                   map[string]string
	Labels                 map[string]string
	Annotations            map[string]string
	PodLabels              map[string]string
	ExposePorts            []int32
	Storage                map[string]*StorageS
	RunAsUser              []int64
	Privileged             bool
	WritableRootFilesystem bool
	ServiceName            string
	ImagePullSecrets       string
	ImagePullAlways        bool
	Obj                    *appsv1.StatefulSet
	Probe                  *corev1.Probe
	ProbeLiveness          bool
	CapabilitiesAdd        []corev1.Capability
	CapabilitiesDrop       []corev1.Capability
	CPUReservedPC          uint // in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore
	CPULimitPC             uint // in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore
	RAMReservedMB          uint // in MB
	RAMLimitMB             uint // in MB
	ServiceAccountName     string
	ServiceAccountSecret   string
	HostAliases            map[string][]string
	JournalProxyPoolNr     uint16
}

func (s *StatefulSetS) CreateOrUpdate() (anyChange bool, err error) {

	if s.KubeClient == nil {
		return false, errors.New("KubeClient cant be empty")
	}
	if s.Name == "" {
		return false, errors.New("name cant be empty")
	}
	if s.Namespace == "" {
		return false, errors.New("namespace cant be empty")
	}

	var statefulSetOld []byte
	statefulSet, err := s.KubeClient.API.AppsV1().StatefulSets(s.Namespace).Get(s.KubeClient.CTX, s.Name, metav1.GetOptions{})
	if err == nil {
		statefulSetOld, err = json.Marshal(statefulSet)
		if err != nil {
			panic(err)
		}

	} else {
		statefulSet = &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: s.Name,
			},
			Spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"element": s.Name,
					},
				},
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						ServiceAccountName:           s.ServiceAccountName,
						AutomountServiceAccountToken: BoolPtr(false),
						EnableServiceLinks:           BoolPtr(false),
						Containers: []corev1.Container{
							{
								Name: s.Name,
							},
						},
					},
				},
			},
		}
	}

	if !s.KubeClient.IngressOldVersion {
		// Topology Spread
		statefulSet.Spec.Template.Spec.TopologySpreadConstraints = []corev1.TopologySpreadConstraint{
			{
				MaxSkew:           1,
				TopologyKey:       "kubernetes.io/hostname",
				WhenUnsatisfiable: corev1.ScheduleAnyway,
				LabelSelector:     &metav1.LabelSelector{MatchLabels: map[string]string{}},
			},
		}
	}

	if len(statefulSetOld) > 0 {
		for i, pvct := range statefulSet.Spec.VolumeClaimTemplates {
			for mountPath, store := range s.Storage {
				if store.Type != "block" {
					continue
				}
				volName := getKeyString(s.Name + mountPath)
				if volName == pvct.Name {
					currentVolSize := pvct.Spec.Resources.Requests.Storage().String()
					newVolSize := fmt.Sprintf("%dMi", store.MaxSizeMB)
					if currentVolSize != newVolSize {
						tlog.Debug(fmt.Sprint(i, pvct.Name, currentVolSize, "=>", newVolSize))

						// TODO: Resize disk
						// https://kubernetes.io/blog/2018/07/12/resizing-persistent-volumes-using-kubernetes/

						// kubectl edit pvc <name> for each PVC in the StatefulSet, to increase its capacity.
						// kubectl delete sts --cascade=false <name> to delete the StatefulSet and leave its pods.
						// kubectl apply -f <name> to recreate the StatefulSet.
						// kubectl rollout restart sts <name> to restart the StatefulSet/pods
					}
					break
				}
			}
		}
	}

	// -------------------------------------------------------------

	statefulSet.ObjectMeta.Labels = s.Labels
	statefulSet.Spec.ServiceName = s.ServiceName

	// ---
	// Resources

	statefulSet.Spec.Template.Spec.Containers[0].Resources.Limits = corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", s.CPULimitPC*10)),
		corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dM", s.RAMLimitMB)),
	}

	statefulSet.Spec.Template.Spec.Containers[0].Resources.Requests = corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", s.CPUReservedPC*10)),
		corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dM", s.RAMReservedMB)),
	}

	// ---

	if statefulSet.ObjectMeta.Annotations == nil {
		statefulSet.ObjectMeta.Annotations = s.Annotations

	} else {
		for k, v := range s.Annotations {
			statefulSet.ObjectMeta.Annotations[k] = v
		}
	}

	if statefulSet.ObjectMeta.Annotations == nil {
		statefulSet.ObjectMeta.Annotations = map[string]string{}
	}

	// ---

	if s.ImagePullSecrets == "" {
		// statefulSet.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
		// 	{
		// 		Name: "cntreg",
		// 	},
		// }

	} else if s.ImagePullSecrets != "-" {
		statefulSet.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: s.ImagePullSecrets,
			},
		}
	}

	// ---

	if s.ImagePullAlways {
		statefulSet.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullAlways
	}

	// ---

	pvcMap := map[string]corev1.PersistentVolumeClaim{}
	for _, pvc := range statefulSet.Spec.VolumeClaimTemplates {
		pvcMap[pvc.Name] = pvc
	}

	// ---

	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}
	pvcList := []corev1.PersistentVolumeClaim{}

	if s.ServiceAccountSecret != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "service-account-token",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: s.ServiceAccountSecret,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "service-account-token",
			MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
			ReadOnly:  true,
		})
	}

	keys := []string{}
	for k := range s.Storage {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, mountPath := range keys {
		store := s.Storage[mountPath]
		name := getKeyString(s.Name + mountPath)

		// ---

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      name,
			MountPath: mountPath,
		})

		// ---

		if store.Type == "host" {
			volumes = append(volumes, corev1.Volume{
				Name: name,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: store.RemotePath,
					},
				},
			})
		}

		// ---

		if store.Type == "config-map" {
			volumes = append(volumes, corev1.Volume{
				Name: name,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: store.RemotePath,
						},
					},
				},
			})
		}

		// ---

		if store.Type == "temp" {
			sizeLimit := resource.MustParse(fmt.Sprintf("%dMi", store.MaxSizeMB))
			volumes = append(volumes, corev1.Volume{
				Name: name,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						// Medium:    corev1.StorageMediumMemory,
						Medium:    corev1.StorageMediumDefault,
						SizeLimit: &sizeLimit,
					},
				},
			})
		}

		// ---

		if store.Type == "nfs" {
			volumes = append(volumes, corev1.Volume{
				Name: name,
				VolumeSource: corev1.VolumeSource{
					NFS: &corev1.NFSVolumeSource{
						Server:   store.RemoteHost,
						Path:     store.RemotePath,
						ReadOnly: store.ReadOnly,
					},
				},
			})
		}

		// ---

		if store.Type == "cifs" {
			// Required driver: "https://github.com/kubernetes-csi/csi-driver-smb"
			if store.Options == "" {
				store.Options = "dir_mode=0755,file_mode=0644,noperm"
			}

			// Create login secret
			secret := SecretS{
				KubeClient: s.KubeClient,
				Namespace:  s.Namespace,
				Name:       "cifs-" + name,
				Type:       corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"username": []byte(store.Login),
					"password": []byte(store.Password),
				},
				Labels: s.Labels,
			}
			_, err := secret.CreateOrUpdate()
			if err != nil {
				return false, err
			}

			// Create storage class for cifs
			url := "//" + filepath.Join(store.RemoteHost, store.RemotePath)
			storName := fmt.Sprintf("cifs-%s-%s", s.Namespace, name)

			stor := &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: storName,
				},
				Provisioner: "smb.csi.k8s.io",
				Parameters: map[string]string{
					"source": url,
					"csi.storage.k8s.io/node-stage-secret-name":      secret.Name,
					"csi.storage.k8s.io/node-stage-secret-namespace": secret.Namespace,
				},
				MountOptions: strings.Split(store.Options, ","),
			}
			_, err = s.KubeClient.API.StorageV1().StorageClasses().Create(s.KubeClient.CTX, stor, metav1.CreateOptions{})
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				tlog.Error(err)
			}

			// Create pvc using created storage class
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: s.Namespace,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
					StorageClassName: &storName,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("100Gi"),
						},
					},
				},
			}
			_, err = s.KubeClient.API.CoreV1().PersistentVolumeClaims(s.Namespace).Create(s.KubeClient.CTX, pvc, metav1.CreateOptions{})
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				tlog.Error(err)
			}

			// Add pvc to volume list
			volumes = append(volumes, corev1.Volume{
				Name: name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: name,
					},
				},
			})

		}

		// ---

		if store.Type == "block" {

			pvc, exists := pvcMap[name]
			if !exists {
				pvc = corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
				}
			}

			storName := (*string)(nil)
			if store.Class != "" {
				storName = &store.Class
			}

			fs := corev1.PersistentVolumeFilesystem
			pvc.Spec = corev1.PersistentVolumeClaimSpec{
				VolumeMode:       &fs,
				AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				StorageClassName: storName,
				Resources: corev1.ResourceRequirements{
					// Limits: corev1.ResourceList{
					// corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(fmt.Sprintf("%dMi", volSize)),
					// },
					Requests: corev1.ResourceList{
						corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(fmt.Sprintf("%dMi", store.MaxSizeMB)),
					},
				},
			}

			pvcList = append(pvcList, pvc)
		}

		if store.Type == "shared" {
			pvc, exists := pvcMap[name]
			if !exists {
				pvc = corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
				}
			}

			storName := (*string)(nil)
			if store.Class != "" {
				storName = &store.Class
			}

			pvc.Spec = corev1.PersistentVolumeClaimSpec{
				AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
				StorageClassName: storName,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(fmt.Sprintf("%dMi", store.MaxSizeMB)),
					},
				},
			}

			pvcList = append(pvcList, pvc)
		}

		// ---
	}

	// ---

	statefulSet.Spec.Template.Spec.Volumes = volumes
	statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts
	statefulSet.Spec.VolumeClaimTemplates = pvcList

	// ---

	if s.Replicas < 1 {
		s.Replicas = 1
	}

	statefulSet.Spec.Replicas = Int32Ptr(s.Replicas)

	// ---

	podLabels := map[string]string{
		"element": s.Name,
	}
	if s.PodLabels != nil {
		for k, v := range s.PodLabels {
			podLabels[k] = v
		}
		podLabels["element"] = s.Name
	}

	statefulSet.Spec.Template.ObjectMeta.Labels = podLabels

	// ---

	// TODO: s.ExposePorts - posortowac

	ports := []corev1.ContainerPort{}
	for _, portNr := range s.ExposePorts {
		ports = append(ports, corev1.ContainerPort{
			Name:          fmt.Sprint("p", portNr),
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: portNr,
		})
	}
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Name < ports[j].Name
	})

	statefulSet.Spec.Template.Spec.Containers[0].Ports = ports

	// Probe

	if s.Probe != nil {
		if s.Probe.TimeoutSeconds == 0 {
			s.Probe.TimeoutSeconds = 1
		}
		if s.Probe.PeriodSeconds == 0 {
			s.Probe.PeriodSeconds = 10
		}
		if s.Probe.SuccessThreshold == 0 {
			s.Probe.SuccessThreshold = 1
		}
		if s.Probe.FailureThreshold == 0 {
			s.Probe.FailureThreshold = 3
		}
	}
	if s.ProbeLiveness {
		statefulSet.Spec.Template.Spec.Containers[0].LivenessProbe = s.Probe
		statefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe = nil

	} else {
		statefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe = s.Probe
		statefulSet.Spec.Template.Spec.Containers[0].LivenessProbe = nil
	}

	// ---

	envKeys := []string{}
	for k := range s.Envs {
		envKeys = append(envKeys, k)
	}
	sort.Strings(envKeys)

	envs := []corev1.EnvVar{}
	for _, k := range envKeys {
		envs = append(envs, corev1.EnvVar{
			Name:  k,
			Value: s.Envs[k],
		})
	}

	if s.Envs["ELEMENT_NAME"] == "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "ELEMENT_NAME",
			Value: s.Name,
		})
	}

	// EP_SHOW_OUTPUT
	envs = append(envs, corev1.EnvVar{
		Name:  "EP_SHOW_OUTPUT",
		Value: "true",
	})

	// NAMESPACE
	if s.Envs["NAMESPACE"] == "" {
		envs = append(envs, corev1.EnvVar{
			Name: "NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		})
	}

	// POD_IP
	envs = append(envs, corev1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	})

	// POD_NAME
	envs = append(envs, corev1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	})

	// NODE_NAME
	envs = append(envs, corev1.EnvVar{
		Name: "NODE_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "spec.nodeName",
			},
		},
	})

	// NODE_IP
	envs = append(envs, corev1.EnvVar{
		Name: "NODE_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.hostIP",
			},
		},
	})

	// DD_AGENT_HOST - datadog host agent ip
	envs = append(envs, corev1.EnvVar{
		Name: "DD_AGENT_HOST",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.hostIP",
			},
		},
	})

	if s.JournalProxyPoolNr != 0 {
		envs = append(envs, corev1.EnvVar{
			Name:  "TIMONI_JOURNAL_PROXY",
			Value: fmt.Sprint(s.JournalProxyPoolNr),
		})
	}

	statefulSet.Spec.Template.Spec.Containers[0].Env = envs

	// ---

	statefulSet.Spec.Template.Spec.Containers[0].Image = s.Image
	statefulSet.Spec.Template.Spec.Containers[0].Command = s.Command

	// ---

	runAsUser := Int64Ptr(0)
	if len(s.RunAsUser) > 0 {
		runAsUser = Int64Ptr(s.RunAsUser[0])
	}

	statefulSet.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser:  runAsUser,
		RunAsGroup: runAsUser,
		FSGroup:    runAsUser,
	}

	statefulSet.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
		Privileged:             BoolPtr(s.Privileged),
		ReadOnlyRootFilesystem: BoolPtr(!s.WritableRootFilesystem),
	}

	// ---

	if s.Image == "docker:19.03.11" {
		hostPathSocket := corev1.HostPathSocket
		statefulSet.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "docker-socket",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/run/docker.sock",
						Type: &hostPathSocket,
					},
				},
			},
		}
		statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "docker-socket",
				MountPath: "/var/run/docker.sock",
			},
		}
	}

	// ---
	// Host Aliases

	statefulSet.Spec.Template.Spec.HostAliases = nil
	if len(s.HostAliases) > 0 {
		ips := make([]string, 0, len(s.HostAliases))
		for ip := range s.HostAliases {
			ips = append(ips, ip)
		}
		sort.Strings(ips)
		for _, ip := range ips {
			statefulSet.Spec.Template.Spec.HostAliases = append(statefulSet.Spec.Template.Spec.HostAliases, corev1.HostAlias{
				IP:        ip,
				Hostnames: s.HostAliases[ip],
			})
		}
	}

	// ---

	if len(statefulSetOld) == 0 {
		s.Obj, err = s.KubeClient.API.AppsV1().StatefulSets(s.Namespace).Create(s.KubeClient.CTX, statefulSet, metav1.CreateOptions{})
		return true, err

	}

	statefulSetNew, err := json.Marshal(statefulSet)
	if err != nil {
		panic(err)
	}

	patch, err := jsonpatch.CreateMergePatch(statefulSetOld, statefulSetNew)
	if err != nil {
		panic(err)
	}

	if len(patch) == 2 {
		return false, nil
	}

	s.Obj, err = s.KubeClient.API.AppsV1().StatefulSets(s.Namespace).Patch(s.KubeClient.CTX, s.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	tlog.Error(err, string(patch))
	if err == nil {
		tlog.Info("StatefulSet patched", tlog.Vars{
			"old":        string(statefulSetOld),
			"new":        string(statefulSetNew),
			"deployment": s.Name,
			"patch":      string(patch),
		})
	}
	return true, err
}

func (s *StatefulSetS) GetObj() *appsv1.StatefulSet {

	var err error
	s.Obj, err = s.KubeClient.API.AppsV1().StatefulSets(s.Namespace).Get(s.KubeClient.CTX, s.Name, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	if s.Name != s.Obj.Name {
		return nil
	}

	return s.Obj
}

func (s *StatefulSetS) Exist() bool {
	return s.GetObj() != nil
}

func (s *StatefulSetS) Delete() error {
	return s.KubeClient.API.AppsV1().StatefulSets(s.Namespace).Delete(s.KubeClient.CTX, s.Name, metav1.DeleteOptions{})
}

func (s *StatefulSetS) PodList(onlyReady bool) []*PodS {

	pods, err := s.KubeClient.API.CoreV1().Pods(s.Namespace).List(s.KubeClient.CTX, metav1.ListOptions{
		LabelSelector: "element=" + s.Name,
	})
	if tlog.Error(err) != nil {
		return nil
	}

	res := []*PodS{}
	for _, podObj := range pods.Items {

		pod := &PodS{
			KubeClient: s.KubeClient,
			Namespace:  s.Namespace,
			Name:       podObj.Name,
			Obj:        podObj,
		}

		if onlyReady {
			if pod.Status() == PodStatusReady {
				res = append(res, pod)
			}
		} else {
			res = append(res, pod)
		}

	}

	return res
}

func (s *StatefulSetS) IsReady() bool {

	if s == nil {
		tlog.Error("StatefulSetS is nil")
		return false
	}

	s.GetObj()
	if s.Obj == nil {
		return false
	}

	if s.Name != s.Obj.Name {
		tlog.Debug("StatefulSetS IsReady: not found")
		return false
	}

	if s.Obj.Status.Replicas == 0 {
		tlog.Debug("StatefulSetS IsReady: " + s.Name + " Status.Replicas == 0")
		return false
	}

	if s.Obj.Status.Replicas != s.Obj.Status.ReadyReplicas {
		tlog.Debug("StatefulSetS IsReady: " + s.Name + " Status.Replicas != Status.ReadyReplicas")
		return false
	}

	pods := s.PodList(true)
	readyCount := len(pods)

	if readyCount == 0 {
		tlog.Debug("StatefulSetS IsReady: " + s.Name + " pod readyCount == 0")
		return false
	}

	return readyCount >= int(s.Obj.Status.Replicas)
}

func (s *StatefulSetS) Cleanup() {
	rsList, err := s.KubeClient.API.AppsV1().ReplicaSets(s.Namespace).List(s.KubeClient.CTX, metav1.ListOptions{
		LabelSelector: "element=" + s.Name,
	})
	if tlog.Error(err) != nil {
		return
	}

	max := 0
	for _, rs := range rsList.Items {
		i, err := strconv.Atoi(rs.Annotations["deployment.kubernetes.io/revision"])
		if tlog.Error(err) != nil {
			continue
		}

		if i > max {
			max = i
		}
	}

	for _, rs := range rsList.Items {
		i, err := strconv.Atoi(rs.Annotations["deployment.kubernetes.io/revision"])
		if tlog.Error(err) != nil {
			continue
		}

		if i == max {
			continue
		}

		tlog.Error(s.KubeClient.API.AppsV1().ReplicaSets(s.Namespace).Delete(s.KubeClient.CTX, rs.Name, metav1.DeleteOptions{}))
	}
}
