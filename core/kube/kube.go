package kube

import (
	"context"
	"core/db2"
	"os"
	"strings"
	"time"

	log "lib/tlog"
	"lib/utils/maps"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	ctrl "sigs.k8s.io/controller-runtime"
	crd "sigs.k8s.io/controller-runtime/pkg/client"
)

var clientMap = maps.NewSafe[string, *ClientS](nil)

func GetKube() *ClientS {

	// Check if client already connected
	kClient := clientMap.Get("")
	if kClient != nil {
		return kClient
	}

	for {
		kClient = connect()

		// Check if connection failed
		if kClient != nil {
			break
		}
		time.Sleep(3 * time.Second)
		log.Warning("Waiting for kube")
	}

	// Save client
	clientMap.Set("", kClient)

	return kClient
}

func connect() *ClientS {
	// ------------------------------------------------
	// Try Connect

	kClient, err := NewClient(nil)
	if err == nil && kClient != nil && len(kClient.NamespaceList()) > 0 {
		return kClient
	}

	buf, _ := os.ReadFile("/etc/rancher/k3s/k3s.yaml")
	kClient, err = NewClient(buf)
	if err == nil && kClient != nil && len(kClient.NamespaceList()) > 0 {
		return kClient
	}

	configContent := []byte(db2.KubeList("Name = '"+db2.TheSettings.Name()+"'", "", 0, 1).First().Config())
	kClient, err = NewClient(configContent)
	if err == nil && kClient != nil && len(kClient.NamespaceList()) > 0 {
		return kClient
	}

	return nil
}

type ClientS struct {
	Config  *restclient.Config
	API     *kubernetes.Clientset
	CRD     crd.Client
	CTX     context.Context
	Dynamic *dynamic.DynamicClient
	Metrics *metricsv.Clientset

	IngressOldVersion bool
}

func NewClient(config []byte) (*ClientS, *log.RecordS) {

	kube := new(ClientS)
	var err error

	if len(config) == 0 {
		kube.Config, err = ctrl.GetConfig()

	} else {
		kube.Config, err = clientcmd.RESTConfigFromKubeConfig(config)
	}

	if err != nil {
		return nil, log.Error(err)
	}

	kube.Config.RateLimiter = nil
	kube.Config.QPS = 1000
	kube.Config.Burst = 2000

	kube.API, err = kubernetes.NewForConfig(kube.Config)
	if err != nil {
		return nil, log.Error(err)
	}

	kube.CRD, err = crd.New(kube.Config, crd.Options{})
	if err != nil {
		return nil, log.Error(err)
	}

	kube.Metrics, err = metricsv.NewForConfig(kube.Config)
	if err != nil {
		return nil, log.Error(err)
	}

	kube.Dynamic, err = dynamic.NewForConfig(kube.Config)
	if err != nil {
		return nil, log.Error(err)
	}

	kube.CTX = context.TODO()

	return kube, nil
}

func NewClientFromConfigFile(filepath string) (*ClientS, *log.RecordS) {
	buf, err := os.ReadFile(filepath)
	if err != nil {
		return nil, log.Error(err)
	}

	return NewClient(buf)
}

func (kube *ClientS) NamespaceCreate(name string) error {
	_, err := kube.API.CoreV1().Namespaces().Create(kube.CTX, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}, metav1.CreateOptions{})
	return err
}

func (kube *ClientS) NamespaceDelete(name string) error {
	err := kube.API.CoreV1().Namespaces().Delete(kube.CTX, name, metav1.DeleteOptions{})
	return err
}

func (kube *ClientS) NamespaceList() []corev1.Namespace {
	nsList, err := kube.API.CoreV1().Namespaces().List(kube.CTX, metav1.ListOptions{})
	if log.Error(err) != nil {
		return nil
	}

	return nsList.Items
}

func (kube *ClientS) NamespaceGet(name string) *corev1.Namespace {

	for _, ns := range kube.NamespaceList() {
		if ns.Name == name {
			return &ns
		}
	}

	return nil
}

func (kube *ClientS) PodMap(namespace string) map[string]*PodS {
	pods, err := kube.API.CoreV1().Pods(namespace).List(kube.CTX, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	res := map[string]*PodS{}
	for _, podObj := range pods.Items {
		res[podObj.Name] = &PodS{
			KubeClient: kube,
			Namespace:  namespace,
			Name:       podObj.Name,
			Obj:        podObj,
		}
	}

	return res
}

func (kube *ClientS) GetClusterUsage(skipCSApps bool) (cpuRequested, ramRequested, pods int64) {
	podList := kube.PodListAll("")

	for _, pods := range podList {

		if skipCSApps && strings.HasPrefix(pods.Namespace, "app-") {
			continue
		}

		for _, cnt := range pods.Obj.Spec.Containers {
			cpuRequested += cnt.Resources.Requests.Cpu().ScaledValue(-2)
			ramRequested += cnt.Resources.Requests.Memory().Value() / (1024 * 1024) //MiB
		}
	}

	pods = int64(len(podList))
	return
}

type NodesInfoS struct {
	TotalCpus int64 // max recorded cpu count
	TotalMem  int64 // in MiB
	TotalPods int
}

var NodesInfo = NodesInfoS{}

func (kube *ClientS) UpdateNodesInfo() {
	nodes, err := kube.API.CoreV1().Nodes().List(kube.CTX, metav1.ListOptions{})
	if err != nil {
		log.Error(err)
		return
	}

	var nI = NodesInfoS{}
	for _, node := range nodes.Items {
		nI.TotalCpus += node.Status.Capacity.Cpu().Value()
		nI.TotalMem += node.Status.Capacity.Memory().Value() / (1024 * 1024) //MiB
		nI.TotalPods += int(node.Status.Capacity.Pods().Value())
	}

	if NodesInfo.TotalCpus < nI.TotalCpus {
		NodesInfo = nI
	}
}

//---------------------------------

func Int32Ptr(i int32) *int32 { return &i }

func Int64Ptr(i int64) *int64 { return &i }

func BoolPtr(b bool) *bool { return &b }

func getKeyString(s string) string {
	key := strings.ToLower(s)
	key = strings.ReplaceAll(key, "/", "-")
	key = strings.ReplaceAll(key, ".", "-")
	key = strings.ReplaceAll(key, "_", "-")
	key = strings.Trim(key, "-")
	return key
}

func GetImageInfo(imageFull string) (imageName, imageTag string) {
	// eg: timoni/core:230209-04
	// imageName = timoni/core
	// imageTag = 230209-04

	ta := strings.SplitN(imageFull, ":", 2)
	imageName = ta[0]
	imageTag = ta[1]
	return
}
