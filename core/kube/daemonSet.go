package kube

import (
	"encoding/json"
	"errors"
	"fmt"
	log "lib/tlog"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type DaemonSetS struct {
	KubeClient             *ClientS
	Namespace              string
	Name                   string
	Image                  string
	Command                []string
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
	HostPID                bool
	ImagePullSecrets       string
	ImagePullAlways        bool
	Probe                  *corev1.Probe
	Obj                    *appsv1.DaemonSet
	PodListOnlyReadyCache  []*PodS
	CPUGuaranteed          uint // in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore
	CPUMax                 uint // in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore
	RAMGuaranteed          uint // in MB
	RAMMax                 uint // in MB
	ServiceAccountName     string
	ServiceAccountMount    bool
	HostAliases            map[string][]string
	JournalProxyPoolNr     uint16
}

type StorageS struct {
	Type      string `toml:"type"`        // timoni storage type: block, temp, shared, cifs, nfs, ftp
	Class     string `toml:"class"`       // kube storage class
	MaxSizeMB int    `toml:"max-size-mb"` // max size in MB

	// cifs, nfs, ftp
	Name       string `toml:"name"` // disk name (shared)
	Login      string `toml:"login"`
	Password   string `toml:"password"`
	RemoteHost string `toml:"remote-host"`
	RemotePath string `toml:"remote-path"`
	Options    string `toml:"options"` // eg. dir_mode=0755,file_mode=0644,noperm
	ReadOnly   bool   `toml:"read-only"`
}

func (d *DaemonSetS) CreateOrUpdate() (diff string, err error) {

	if d.KubeClient == nil {
		return "", errors.New("KubeClient cant be empty")
	}
	if d.Name == "" {
		return "", errors.New("name cant be empty")
	}
	if d.Namespace == "" {
		return "", errors.New("namespace cant be empty")
	}

	var daemonOld []byte
	daemon, err := d.KubeClient.API.AppsV1().DaemonSets(d.Namespace).Get(d.KubeClient.CTX, d.Name, metav1.GetOptions{})
	if err == nil {
		daemonOld, err = json.Marshal(daemon)
		if err != nil {
			panic(err)
		}

	} else {
		daemon = &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: d.Name,
			},
			Spec: appsv1.DaemonSetSpec{
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
								Name: d.Name,
							},
						},
					},
				},
			},
		}
	}

	daemon.ObjectMeta.Labels = d.Labels

	if !d.KubeClient.IngressOldVersion {
		// Topology Spread
		daemon.Spec.Template.Spec.TopologySpreadConstraints = []corev1.TopologySpreadConstraint{
			{
				MaxSkew:           1,
				TopologyKey:       "kubernetes.io/hostname",
				WhenUnsatisfiable: corev1.ScheduleAnyway,
				LabelSelector:     &metav1.LabelSelector{MatchLabels: map[string]string{}},
			},
		}
	}

	// ---

	if daemon.ObjectMeta.Annotations == nil {
		daemon.ObjectMeta.Annotations = d.Annotations

	} else {
		for k, v := range d.Annotations {
			daemon.ObjectMeta.Annotations[k] = v
		}
	}

	if daemon.ObjectMeta.Annotations == nil {
		daemon.ObjectMeta.Annotations = map[string]string{}
	}

	// ---

	if d.ImagePullSecrets == "" {
		// deploy.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
		// 	{
		// 		Name: "cntreg",
		// 	},
		// }

	} else if d.ImagePullSecrets != "-" {
		daemon.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: d.ImagePullSecrets,
			},
		}
	}

	// ---

	if d.ImagePullAlways {
		daemon.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullAlways
	}

	// ---

	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

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
				return "", err
			}

			// Create storage class for cifs
			url := "//" + filepath.Join(store.RemoteHost, store.RemotePath)
			storName := fmt.Sprintf("cifs-%s-%s", d.Namespace, name)

			stor := &v1.StorageClass{
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

	daemon.Spec.Template.Spec.Volumes = volumes
	daemon.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts

	// ---

	daemon.Spec.Template.Spec.Containers[0].ReadinessProbe = d.Probe

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

	daemon.Spec.Template.ObjectMeta.Labels = d.PodLabels

	// ---

	ports := []corev1.ContainerPort{}
	for _, portNr := range d.ExposePorts {
		ports = append(ports, corev1.ContainerPort{
			Name:          fmt.Sprint("p", portNr),
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: portNr,
		})
	}

	daemon.Spec.Template.Spec.Containers[0].Ports = ports

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

	envs = append(envs, corev1.EnvVar{
		Name:  "ELEMENT_NAME",
		Value: d.Name,
	})

	// NAMESPACE
	envs = append(envs, corev1.EnvVar{
		Name: "NAMESPACE",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.namespace",
			},
		},
	})

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

	daemon.Spec.Template.Spec.Containers[0].Env = envs

	// ---

	daemon.Spec.Template.Spec.Containers[0].Image = d.Image
	daemon.Spec.Template.Spec.Containers[0].Command = d.Command

	// ---

	runAsUser := Int64Ptr(0)
	if len(d.RunAsUser) > 0 {
		runAsUser = Int64Ptr(d.RunAsUser[0])
	}

	daemon.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser:  runAsUser,
		RunAsGroup: runAsUser,
		FSGroup:    runAsUser,
	}

	daemon.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
		Privileged:             BoolPtr(d.Privileged),
		ReadOnlyRootFilesystem: BoolPtr(!d.WritableRootFilesystem),
	}

	daemon.Spec.Template.Spec.HostPID = d.HostPID

	// ---
	// DockerSocket

	if d.DockerSocket {
		hostPathSocket := corev1.HostPathSocket
		daemon.Spec.Template.Spec.Volumes = []corev1.Volume{
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
		daemon.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "docker-socket",
				MountPath: "/var/run/docker.sock",
			},
		}
	}

	// ---
	// Resources

	daemon.Spec.Template.Spec.Containers[0].Resources.Limits = corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", d.CPUMax*10)),
		corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dM", d.RAMMax)),
	}

	daemon.Spec.Template.Spec.Containers[0].Resources.Requests = corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", d.CPUGuaranteed*10)),
		corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dM", d.RAMGuaranteed)),
	}

	// ---
	// Host Aliases

	daemon.Spec.Template.Spec.HostAliases = nil
	if len(d.HostAliases) > 0 {
		ips := make([]string, 0, len(d.HostAliases))
		for ip := range d.HostAliases {
			ips = append(ips, ip)
		}
		sort.Strings(ips)
		for _, ip := range ips {
			daemon.Spec.Template.Spec.HostAliases = append(daemon.Spec.Template.Spec.HostAliases, corev1.HostAlias{
				IP:        ip,
				Hostnames: d.HostAliases[ip],
			})
		}
	}

	// ---

	if len(daemonOld) == 0 {
		d.Obj, err = d.KubeClient.API.AppsV1().DaemonSets(d.Namespace).Create(d.KubeClient.CTX, daemon, metav1.CreateOptions{})
		return "creating new obj", err

	}
	deployNew, err := json.Marshal(daemon)
	if err != nil {
		panic(err)
	}

	patch, err := jsonpatch.CreateMergePatch(daemonOld, deployNew)
	if err != nil {
		panic(err)
	}

	if len(patch) == 2 {
		d.GetObj()
		return "", nil
	}

	d.Obj, err = d.KubeClient.API.AppsV1().DaemonSets(d.Namespace).Patch(d.KubeClient.CTX, d.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	log.Error(err, string(patch))
	if err == nil {
		log.Debug("DaemonSet patched", log.Vars{
			"old":       string(daemonOld),
			"deamonset": d.Name,
			"patch":     string(patch),
		})
	}
	return string(patch), err
}

// GetObj ...
func (d *DaemonSetS) GetObj() *appsv1.DaemonSet {

	var err error
	d.Obj, err = d.KubeClient.API.AppsV1().DaemonSets(d.Namespace).Get(d.KubeClient.CTX, d.Name, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	return d.Obj
}

func (d *DaemonSetS) Exist() bool {
	return d.GetObj() != nil
}

func (d *DaemonSetS) IsReady() bool {
	d.GetObj()
	if d.Obj == nil {
		return false
	}

	if d.Obj.Status.NumberReady == 0 {
		log.Debug("DaemonSet IsReady: " + d.Name + " Status.NumberReady == 0")
		return false
	}

	// fmt.Println(dObj.Status.Replicas, dObj.Status.ReadyReplicas)
	if d.Obj.Status.NumberReady != d.Obj.Status.DesiredNumberScheduled {
		log.Debug("DaemonSet IsReady: " + d.Name + " Status.NumberReady != Status.ReadyReplicas")
		return false
	}

	pods := d.PodList(true)
	readyCount := len(pods)

	if readyCount == 0 {
		log.Debug("DaemonSet IsReady: " + d.Name + " pod readyCount == 0")
		return false
	}
	// TODO: sprawdzić czy to działa

	return d.Obj.Status.NumberReady == d.Obj.Status.DesiredNumberScheduled
}

func (d *DaemonSetS) PodList(onlyReady bool) []*PodS {

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

		if onlyReady {
			if pod.Status() == PodStatusReady {
				res = append(res, pod)
			}
		} else {
			res = append(res, pod)
		}

	}

	if onlyReady {
		d.PodListOnlyReadyCache = res
	}

	return res
}

// ExecOnEachPod ...
func (d *DaemonSetS) ExecOnEachPod(cmd []string) error {
	for _, pod := range d.PodList(true) {
		log.Debug("ExecOnPod:", pod.Name)
		err := pod.ExecToStdOut(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeploymentMap ...
func (kube *ClientS) DaemonSetMap(namespace string) map[string]*DaemonSetS {
	deps, err := kube.API.AppsV1().DaemonSets(namespace).List(kube.CTX, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	res := map[string]*DaemonSetS{}
	for _, d := range deps.Items {
		res[d.Name] = &DaemonSetS{
			KubeClient: kube,
			Namespace:  namespace,
			Name:       d.Name,
			Obj:        &d,
		}
	}

	return res
}

// Delete ...
func (d *DaemonSetS) Delete() error {
	return d.KubeClient.API.AppsV1().Deployments(d.Namespace).Delete(d.KubeClient.CTX, d.Name, metav1.DeleteOptions{})
}

// Cleanup ...
func (d *DaemonSetS) Cleanup() {
	rsList, err := d.KubeClient.API.AppsV1().ReplicaSets(d.Namespace).List(d.KubeClient.CTX, metav1.ListOptions{
		LabelSelector: "element=" + d.Name,
	})
	if log.Error(err) != nil {
		return
	}

	max := 0
	for _, rs := range rsList.Items {
		i, err := strconv.Atoi(rs.Annotations["deamonset.kubernetes.io/revision"])
		if log.Error(err) != nil {
			continue
		}

		if i > max {
			max = i
		}
	}

	for _, rs := range rsList.Items {
		i, err := strconv.Atoi(rs.Annotations["deamonset.kubernetes.io/revision"])
		if log.Error(err) != nil {
			continue
		}

		if i == max {
			continue
		}

		log.Error(d.KubeClient.API.AppsV1().ReplicaSets(d.Namespace).Delete(d.KubeClient.CTX, rs.Name, metav1.DeleteOptions{}))
	}
}
