package kube

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	jsonpatch "github.com/evanphx/json-patch"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ServiceS struct {
	KubeClient       *ClientS
	Namespace        string
	Name             string
	Ports            map[int32]int32
	Labels           map[string]string
	Annotations      map[string]string
	TargetSelector   map[string]string
	Headless         bool
	Obj              *corev1.Service
	NodePort         int32
	StickyCookieName string
	Internal         bool
	LoadBalancer     bool
	Protocol         string // 'tcp' or 'udp'
}

func (s *ServiceS) CreateOrUpdate() (diff string, err error) {

	if s.KubeClient == nil {
		return "", errors.New("KubeClient cant be empty")
	}
	if s.Name == "" {
		return "", errors.New("name cant be empty")
	}
	if s.Namespace == "" {
		return "", errors.New("namespace cant be empty")
	}

	// ---

	var svcOld []byte
	svc, err := s.KubeClient.API.CoreV1().Services(s.Namespace).Get(s.KubeClient.CTX, s.Name, metav1.GetOptions{})
	if err == nil {
		svcOld, err = json.Marshal(svc)
		if err != nil {
			panic(err)
		}

	} else {
		t := corev1.ServiceTypeClusterIP
		if s.NodePort > 0 {
			t = corev1.ServiceTypeNodePort
		}
		svc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: s.Name,
			},
			Spec: corev1.ServiceSpec{
				Type: t,
			},
		}
	}

	// ---

	svc.ObjectMeta.Labels = s.Labels
	svc.Spec.Selector = s.TargetSelector

	// ---

	if s.Headless {
		svc.Spec.ClusterIP = "None"

	} else if svc.Spec.ClusterIP == "None" {
		svc.Spec.ClusterIP = ""
	}

	// ---

	if s.LoadBalancer {
		svc.Spec.Type = corev1.ServiceTypeLoadBalancer
		svc.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeLocal
	}

	// ---

	if s.NodePort > 0 && len(s.Ports) > 1 {
		return "error", errors.New("node port service must have only one port")
	}

	protocol := corev1.ProtocolTCP
	if s.Protocol == "udp" {
		protocol = corev1.ProtocolUDP
	}

	ports := []corev1.ServicePort{}
	for fromPortNr, toPortNr := range s.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:     fmt.Sprint("p", fromPortNr),
			Protocol: protocol,
			Port:     fromPortNr,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: toPortNr,
			},
			NodePort: s.NodePort,
		})
	}

	sort.Slice(ports, func(i, j int) bool { return ports[i].Port < ports[j].Port })
	svc.Spec.Ports = ports

	// ---

	if svc.ObjectMeta.Annotations == nil {
		svc.ObjectMeta.Annotations = s.Annotations

	} else {
		for k, v := range s.Annotations {
			svc.ObjectMeta.Annotations[k] = v
		}
	}

	if svc.ObjectMeta.Annotations == nil {
		svc.ObjectMeta.Annotations = map[string]string{}
	}

	// Sticky session
	if s.StickyCookieName != "" {
		svc.ObjectMeta.Annotations["traefik.ingress.kubernetes.io/service.sticky.cookie"] = "true"
		svc.ObjectMeta.Annotations["traefik.ingress.kubernetes.io/service.sticky.cookie.name"] = s.StickyCookieName
	} else {
		delete(svc.ObjectMeta.Annotations, "traefik.ingress.kubernetes.io/service.sticky.cookie")
		delete(svc.ObjectMeta.Annotations, "traefik.ingress.kubernetes.io/service.sticky.cookie.name")
	}

	// Internal
	if s.Internal {
		svc.ObjectMeta.Annotations["service.beta.kubernetes.io/azure-load-balancer-internal"] = "true"
		svc.ObjectMeta.Annotations["networking.gke.io/load-balancer-type"] = "Internal"
	} else {
		svc.ObjectMeta.Annotations["service.beta.kubernetes.io/azure-load-balancer-internal"] = "false"
	}

	// ---

	if len(svcOld) == 0 {
		s.Obj, err = s.KubeClient.API.CoreV1().Services(s.Namespace).Create(s.KubeClient.CTX, svc, metav1.CreateOptions{})
		return "creating new obj", err
	}

	svcNew, err := json.Marshal(svc)
	if err != nil {
		panic(err)
	}

	patch, err := jsonpatch.CreateMergePatch(svcOld, svcNew)
	if err != nil {
		panic(err)
	}

	if len(patch) == 2 {
		return "", nil
	}

	s.Obj, err = s.KubeClient.API.CoreV1().Services(s.Namespace).Patch(s.KubeClient.CTX, s.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	return string(patch), err
}

func (s *ServiceS) GetObj() *corev1.Service {

	var err error
	s.Obj, err = s.KubeClient.API.CoreV1().Services(s.Namespace).Get(s.KubeClient.CTX, s.Name, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	return s.Obj
}

func (s *ServiceS) Exist() bool {
	return s.GetObj() != nil
}

func (s *ServiceS) Delete() error {
	return s.KubeClient.API.CoreV1().Services(s.Namespace).Delete(s.KubeClient.CTX, s.Name, metav1.DeleteOptions{})
}
