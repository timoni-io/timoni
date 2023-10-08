package kube

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	log "lib/tlog"

	jsonpatch "github.com/evanphx/json-patch"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// DeploymentS ...
type DeploymentS struct {
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
	DockerSocket           bool
	ImagePullSecrets       string
	ImagePullAlways        bool
	Probe                  *corev1.Probe
	ProbeLiveness          bool
	Obj                    *appsv1.Deployment
	CPUReservedPC          uint // in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore
	CPULimitPC             uint // in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore
	RAMReservedMB          uint // in MB
	RAMLimitMB             uint // in MB
	CapabilitiesAdd        []corev1.Capability
	CapabilitiesDrop       []corev1.Capability
	ServiceAccountName     string
	ServiceAccountSecret   string
	ServiceAccountMount    bool
	HostAliases            map[string][]string
	JournalProxyPoolNr     uint16
}

// CreateOrUpdate ...
func (d *DeploymentS) CreateOrUpdate() (anyChange bool, err error) {

	if d.KubeClient == nil {
		return false, errors.New("KubeClient cant be empty")
	}
	if d.Name == "" {
		return false, errors.New("name cant be empty")
	}
	if d.Namespace == "" {
		return false, errors.New("namespace cant be empty")
	}

	if d.ServiceAccountName != "" {
		d.ServiceAccountMount = true
	}

	var deployOld []byte
	deploy, err := d.KubeClient.API.AppsV1().Deployments(d.Namespace).Get(d.KubeClient.CTX, d.Name, metav1.GetOptions{})
	if err == nil {
		deployOld, err = json.Marshal(deploy)
		if err != nil {
			panic(err)
		}

	} else {
		deploy = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: d.Name,
			},
			Spec: appsv1.DeploymentSpec{

				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"element": d.Name,
					},
				},
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						ServiceAccountName:           d.ServiceAccountName,
						AutomountServiceAccountToken: &d.ServiceAccountMount,
						EnableServiceLinks:           &d.ServiceAccountMount,
						Containers: []corev1.Container{
							{
								Name:            d.Name,
								ImagePullPolicy: corev1.PullIfNotPresent,
							},
						},
					},
				},
			},
		}
	}

	deploy.ObjectMeta.Labels = d.Labels

	if !d.KubeClient.IngressOldVersion {
		// Topology Spread
		deploy.Spec.Template.Spec.TopologySpreadConstraints = []corev1.TopologySpreadConstraint{
			{
				MaxSkew:           1,
				TopologyKey:       "kubernetes.io/hostname",
				WhenUnsatisfiable: corev1.ScheduleAnyway,
				LabelSelector:     &metav1.LabelSelector{MatchLabels: map[string]string{}},
			},
		}
	}

	// ---

	if deploy.ObjectMeta.Annotations == nil {
		deploy.ObjectMeta.Annotations = d.Annotations

	} else {
		for k, v := range d.Annotations {
			deploy.ObjectMeta.Annotations[k] = v
		}
	}

	if deploy.ObjectMeta.Annotations == nil {
		deploy.ObjectMeta.Annotations = map[string]string{}
	}

	// ---

	if d.ImagePullSecrets == "" {
		// deploy.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
		// 	{
		// 		Name: "cntreg",
		// 	},
		// }

	} else if d.ImagePullSecrets != "-" {
		deploy.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: d.ImagePullSecrets,
			},
		}
	}

	// ---

	if d.ImagePullAlways {
		deploy.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullAlways
	}

	// ---

	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

	defMode := int32(420)
	if d.ServiceAccountSecret != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "service-account-token",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  d.ServiceAccountSecret,
					DefaultMode: &defMode,
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
	for k := range d.Storage {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, mountPath := range keys {
		store := d.Storage[mountPath]
		name := getKeyString(d.Name + mountPath)

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

		if store.Type == "shared" {

			// -
			storName := (*string)(nil)
			if store.Class != "" {
				storName = &store.Class
			}

			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      store.Name,
					Namespace: d.Namespace,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
					StorageClassName: storName,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(fmt.Sprintf("%dMi", store.MaxSizeMB)),
						},
					},
				},
			}
			_, err := d.KubeClient.API.CoreV1().PersistentVolumeClaims(d.Namespace).Create(d.KubeClient.CTX, pvc, metav1.CreateOptions{})
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				log.Error(err)
			}

			// -

			volumes = append(volumes, corev1.Volume{
				Name: name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: store.Name,
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
				KubeClient: d.KubeClient,
				Namespace:  d.Namespace,
				Name:       "cifs-" + name,
				Type:       corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"username": []byte(store.Login),
					"password": []byte(store.Password),
				},
				Labels: d.Labels,
			}
			_, err := secret.CreateOrUpdate()
			if err != nil {
				return false, err
			}

			// Create storage class for cifs
			url := "//" + filepath.Join(store.RemoteHost, store.RemotePath)
			storName := fmt.Sprintf("cifs-%s-%s", d.Namespace, name)

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
			_, err = d.KubeClient.API.StorageV1().StorageClasses().Create(d.KubeClient.CTX, stor, metav1.CreateOptions{})
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				log.Error(err)
			}

			// Create pvc using created storage class
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: d.Namespace,
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
			_, err = d.KubeClient.API.CoreV1().PersistentVolumeClaims(d.Namespace).Create(d.KubeClient.CTX, pvc, metav1.CreateOptions{})
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				log.Error(err)
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
	}

	// ---
	deploy.Spec.Template.Spec.Volumes = volumes
	deploy.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts

	// ---

	// deploy.Spec.Template.Spec.Containers[0].LivenessProbe = d.Probe

	if d.Probe != nil {
		if d.Probe.TimeoutSeconds == 0 {
			d.Probe.TimeoutSeconds = 1
		}
		if d.Probe.PeriodSeconds == 0 {
			d.Probe.PeriodSeconds = 10
		}
		if d.Probe.SuccessThreshold == 0 {
			d.Probe.SuccessThreshold = 1
		}
		if d.Probe.FailureThreshold == 0 {
			d.Probe.FailureThreshold = 3
		}
	}
	if d.ProbeLiveness {
		deploy.Spec.Template.Spec.Containers[0].LivenessProbe = d.Probe
		deploy.Spec.Template.Spec.Containers[0].ReadinessProbe = nil

	} else {
		deploy.Spec.Template.Spec.Containers[0].ReadinessProbe = d.Probe
		deploy.Spec.Template.Spec.Containers[0].LivenessProbe = nil
	}

	// ---

	if d.Replicas < 1 {
		d.Replicas = 1
	}

	deploy.Spec.Replicas = Int32Ptr(d.Replicas)

	// ---

	if d.PodLabels == nil {
		d.PodLabels = map[string]string{}
	}

	podLabelsDefault := map[string]string{
		"element": d.Name,
	}

	for k, v := range podLabelsDefault {
		d.PodLabels[k] = v
	}

	deploy.Spec.Template.ObjectMeta.Labels = d.PodLabels

	// ---

	ports := []corev1.ContainerPort{}
	for _, portNr := range d.ExposePorts {
		ports = append(ports, corev1.ContainerPort{
			Name:          fmt.Sprint("p", portNr),
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: portNr,
		})

	}

	deploy.Spec.Template.Spec.Containers[0].Ports = ports

	// ---

	envKeys := []string{}
	for k := range d.Envs {
		envKeys = append(envKeys, k)
	}
	sort.Strings(envKeys)

	envs := []corev1.EnvVar{}
	for _, k := range envKeys {
		envs = append(envs, corev1.EnvVar{
			Name:  k,
			Value: d.Envs[k],
		})
	}

	if d.Envs["ELEMENT_NAME"] == "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "ELEMENT_NAME",
			Value: d.Name,
		})
	}

	envs = append(envs, corev1.EnvVar{
		Name:  "EP_SHOW_OUTPUT",
		Value: "true",
	})

	// NAMESPACE
	if d.Envs["NAMESPACE"] == "" {
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

	if d.JournalProxyPoolNr != 0 {
		envs = append(envs, corev1.EnvVar{
			Name:  "TIMONI_JOURNAL_PROXY",
			Value: fmt.Sprint(d.JournalProxyPoolNr),
		})
	}

	deploy.Spec.Template.Spec.Containers[0].Env = envs

	// ---

	deploy.Spec.Template.Spec.Containers[0].Image = d.Image
	deploy.Spec.Template.Spec.Containers[0].Command = d.Command

	// ---

	runAsUser := Int64Ptr(0)
	if len(d.RunAsUser) > 0 {
		runAsUser = Int64Ptr(d.RunAsUser[0])
	}

	deploy.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser:  runAsUser,
		RunAsGroup: runAsUser,
		FSGroup:    runAsUser,
	}

	deploy.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
		Privileged:             BoolPtr(d.Privileged),
		ReadOnlyRootFilesystem: BoolPtr(!d.WritableRootFilesystem),
	}

	if len(d.CapabilitiesAdd) > 0 || len(d.CapabilitiesDrop) > 0 {
		deploy.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities = &corev1.Capabilities{
			Add:  d.CapabilitiesAdd,
			Drop: d.CapabilitiesDrop,
		}
	}

	// ---
	// DockerSocket

	if d.DockerSocket {
		hostPathSocket := corev1.HostPathSocket
		deploy.Spec.Template.Spec.Volumes = []corev1.Volume{
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
		deploy.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "docker-socket",
				MountPath: "/var/run/docker.sock",
			},
		}
	}

	// ---
	// Resources

	deploy.Spec.Template.Spec.Containers[0].Resources.Limits = corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", d.CPULimitPC*10)),
		corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dM", d.RAMLimitMB)),
	}

	deploy.Spec.Template.Spec.Containers[0].Resources.Requests = corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", d.CPUReservedPC*10)),
		corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dM", d.RAMReservedMB)),
	}

	// ---
	// Host Aliases

	deploy.Spec.Template.Spec.HostAliases = nil
	if len(d.HostAliases) > 0 {
		ips := make([]string, 0, len(d.HostAliases))
		for ip := range d.HostAliases {
			ips = append(ips, ip)
		}
		sort.Strings(ips)
		for _, ip := range ips {
			deploy.Spec.Template.Spec.HostAliases = append(deploy.Spec.Template.Spec.HostAliases, corev1.HostAlias{
				IP:        ip,
				Hostnames: d.HostAliases[ip],
			})
		}
	}

	// ---

	if len(deployOld) == 0 {
		d.Obj, err = d.KubeClient.API.AppsV1().Deployments(d.Namespace).Create(d.KubeClient.CTX, deploy, metav1.CreateOptions{})
		return true, err

	}

	deployNew, err := json.Marshal(deploy)
	if err != nil {
		panic(err)
	}

	patch, err := jsonpatch.CreateMergePatch(deployOld, deployNew)
	if err != nil {
		panic(err)
	}

	if len(patch) == 2 {
		d.GetObj()
		return false, nil
	}

	d.Obj, err = d.KubeClient.API.AppsV1().Deployments(d.Namespace).Patch(d.KubeClient.CTX, d.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	log.Error(err, string(patch))
	if err == nil {
		log.Info("Deployment patched", log.Vars{
			"old":        string(deployOld),
			"new":        string(deployNew),
			"deployment": d.Name,
			"patch":      string(patch),
		})
	}
	return true, err
}

func (d *DeploymentS) GetObj() *appsv1.Deployment {

	var err error
	d.Obj, err = d.KubeClient.API.AppsV1().Deployments(d.Namespace).Get(d.KubeClient.CTX, d.Name, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	return d.Obj
}

func (d *DeploymentS) Exist() bool {
	return d.GetObj() != nil
}

func (d *DeploymentS) IsReady() bool {
	if d == nil {
		log.Error("DeploymentS is nil")
		return false
	}

	d.GetObj()
	if d.Obj == nil {
		return false
	}

	if d.Name != d.Obj.Name {
		log.Debug("Deployment IsReady: not found")
		return false
	}

	if d.Obj.Status.Replicas == 0 {
		log.Debug("Deployment IsReady: "+d.Name+" Status.Replicas == 0", d.Replicas)
		return false
	}

	if d.Obj.Status.Replicas != d.Obj.Status.ReadyReplicas {
		log.Debug("Deployment IsReady: " + d.Name + " Status.Replicas != Status.ReadyReplicas")
		return false
	}

	pods := d.PodList(true)
	readyCount := len(pods)

	if readyCount == 0 {
		log.Debug("Deployment IsReady: " + d.Name + " pod readyCount == 0")
		return false
	}

	return readyCount == int(d.Obj.Status.Replicas)
}

func (d *DeploymentS) PodList(onlyReady bool) []*PodS {

	if d == nil {
		log.Error("DeploymentS is nil")
		return nil
	}
	if d.Obj == nil {
		d.GetObj()
	}
	if d.Obj == nil {
		return nil
	}

	pods, err := d.KubeClient.API.CoreV1().Pods(d.Namespace).List(d.KubeClient.CTX, metav1.ListOptions{
		LabelSelector: "element=" + d.Name,
	})
	if log.Error(err) != nil {
		return nil
	}

	res := []*PodS{}
	for _, podObj := range pods.Items {
		pod := &PodS{
			KubeClient: d.KubeClient,
			Namespace:  d.Namespace,
			Name:       podObj.Name,
			Obj:        podObj,
		}

		rs := pod.ReplicaSet()
		if rs != nil {
			rsRevision, err := strconv.Atoi(rs.Annotations["deployment.kubernetes.io/revision"])
			if log.Error(err) != nil {
				continue
			}
			pod.ReplicaSetRevision = rsRevision
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

func (d *DeploymentS) ExecOnEachPod(cmd []string) error {
	for _, pod := range d.PodList(true) {
		log.Debug("ExecOnPod:", pod.Name)
		err := pod.ExecToStdOut(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (kube *ClientS) DeploymentMap(namespace string) map[string]*DeploymentS {
	deps, err := kube.API.AppsV1().Deployments(namespace).List(kube.CTX, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	res := map[string]*DeploymentS{}
	for _, d := range deps.Items {
		res[d.Name] = &DeploymentS{
			KubeClient: kube,
			Namespace:  namespace,
			Name:       d.Name,
			Obj:        &d,
		}
	}

	return res
}

func (d *DeploymentS) Delete() error {
	return d.KubeClient.API.AppsV1().Deployments(d.Namespace).Delete(d.KubeClient.CTX, d.Name, metav1.DeleteOptions{})
}

func (d *DeploymentS) Cleanup() {
	rsList, err := d.KubeClient.API.AppsV1().ReplicaSets(d.Namespace).List(d.KubeClient.CTX, metav1.ListOptions{
		LabelSelector: "element=" + d.Name,
	})
	if log.Error(err) != nil {
		return
	}

	max := 0
	for _, rs := range rsList.Items {
		i, err := strconv.Atoi(rs.Annotations["deployment.kubernetes.io/revision"])
		if log.Error(err) != nil {
			continue
		}

		if i > max {
			max = i
		}
	}

	for _, rs := range rsList.Items {
		i, err := strconv.Atoi(rs.Annotations["deployment.kubernetes.io/revision"])
		if log.Error(err) != nil {
			continue
		}

		if i == max {
			continue
		}

		log.Error(d.KubeClient.API.AppsV1().ReplicaSets(d.Namespace).Delete(d.KubeClient.CTX, rs.Name, metav1.DeleteOptions{}))
	}
}
