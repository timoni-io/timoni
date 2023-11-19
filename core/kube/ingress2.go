package kube

import (
	"encoding/json"
	"fmt"
	"lib/tlog"
	log "lib/tlog"
	"lib/utils/conv"
	"reflect"
	"sort"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	corev1 "k8s.io/api/core/v1"
	netV1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Ingress2S ...
type Ingress2S struct {
	KubeClient          *ClientS
	Namespace           string
	Name                string
	Domain              string
	Annotations         map[string]string
	Labels              map[string]string
	Paths               map[string]*DomainPathS
	HTTPS               bool
	WWWredirect         bool
	HTTPSSecretName     string
	MaxUploadSize       int // MB
	Timeout             int // default 60 sec
	BuffersNumber       int // default 4, proxy-buffers-number
	BufferSize          int // default 4k, proxy-buffer-size
	HeaderBuffersNumber int // default 4, large-client-header-buffers
	HeaderBufferSize    int // default 8k, large-client-header-buffers
	Obj                 *netV1.Ingress
	Auth                string
}

func (i *Ingress2S) CreateOrUpdate() (diff string, status *log.RecordS) {

	if i.KubeClient == nil {
		return "", log.Error("KubeClient cant be empty")
	}
	if i.Name == "" {
		return "", log.Error("Name cant be empty")
	}
	if i.Namespace == "" {
		return "", log.Error("Namespace cant be empty")
	}

	ingressCtl := i.KubeClient.API.NetworkingV1().Ingresses(i.Namespace)

	// ---

	var ingressOld []byte
	ingress, err := ingressCtl.Get(i.KubeClient.CTX, i.Name, metav1.GetOptions{})
	if err == nil {
		ingressOld, err = json.Marshal(ingress)
		if err != nil {
			panic(err)
		}

	} else {
		ingress = &netV1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name: i.Name,
			},
		}
	}

	// ---

	ingress.ObjectMeta.Annotations = map[string]string{}
	for k, v := range i.Annotations {
		ingress.ObjectMeta.Annotations[k] = v
	}

	// ---

	ingress.ObjectMeta.Labels = map[string]string{}
	for k, v := range i.Labels {
		ingress.ObjectMeta.Labels[k] = v
	}

	// ---
	traefikMiddlewares := []string{}
	pathTypePrefix := netV1.PathTypePrefix
	paths := []netV1.HTTPIngressPath{}

	keys := []string{}
	for k := range i.Paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, path := range keys {
		target := i.Paths[path]

		// Set ingress paths
		paths = append(paths, netV1.HTTPIngressPath{
			Path:     path,
			PathType: &pathTypePrefix,
			Backend: netV1.IngressBackend{
				Service: &netV1.IngressServiceBackend{
					Name: target.ElementName,
					Port: netV1.ServiceBackendPort{
						Number: target.Port,
					},
				},
			},
		})

		// Path map
		if target.Prefix != "" {
			middleware, err := i.traefikPathRewriteAdd(path, target.Prefix)
			if log.Error(err) != nil {
				return "", log.Error(err)
			}
			traefikMiddlewares = append(traefikMiddlewares, middleware)
		}

	}

	ingress.Spec = netV1.IngressSpec{
		Rules: []netV1.IngressRule{{
			Host: i.Domain,
			IngressRuleValue: netV1.IngressRuleValue{
				HTTP: &netV1.HTTPIngressRuleValue{
					Paths: paths,
				},
			},
		}},
	}

	if i.WWWredirect {
		ingress.Spec.Rules = append(ingress.Spec.Rules, netV1.IngressRule{
			Host: "www." + i.Domain,
			IngressRuleValue: netV1.IngressRuleValue{
				HTTP: &netV1.HTTPIngressRuleValue{
					Paths: paths,
				},
			},
		})
	}

	// ---

	if ingress.ObjectMeta.Annotations == nil {
		ingress.ObjectMeta.Annotations = map[string]string{}
	}
	ingress.ObjectMeta.Annotations["hsts"] = "false"
	ingress.ObjectMeta.Annotations["hsts-include-subdomains"] = "false"

	if i.HTTPS && i.HTTPSSecretName == "" {
		i.HTTPSSecretName = i.Domain + "-tls"
	}
	if i.HTTPS {
		ingress.Spec.TLS = []netV1.IngressTLS{{
			Hosts:      []string{i.Domain},
			SecretName: i.HTTPSSecretName,
		}}
		traefikMiddlewares = append(traefikMiddlewares, "timoni-http-to-https@kubernetescrd")

		if i.WWWredirect {
			ingress.Spec.TLS[0].Hosts = append(ingress.Spec.TLS[0].Hosts, "www."+i.Domain)
		}
	}

	if i.Auth != "" {
		// ingress basic auth
		// ---------------------------

		// Secret
		authSecretName := fmt.Sprintf("%s-basic-auth", conv.KeyString(i.Name))

		basicAuthSecret := &SecretS{
			KubeClient: i.KubeClient,
			Namespace:  i.Namespace,
			Name:       authSecretName,
			Type:       corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				"users": []byte(i.Auth),
			},
		}
		_, err := basicAuthSecret.CreateOrUpdate()
		if err != nil {
			return "", log.Error("basicAuthSecret.CreateOrUpdate:" + err.Error())
		}
		// ---------------------------

		// Traefik basic auth middleware
		if e := i.traefikBasicAuthAdd(authSecretName); e != nil {
			return "", log.Error("traefikBasicAuthAdd:" + e.Message)
		}

		// Add middleware
		traefikMiddlewares = append(traefikMiddlewares, fmt.Sprintf("%s-%s@kubernetescrd", i.Namespace, authSecretName))
	}

	// Apply middlewares
	if len(traefikMiddlewares) > 0 {
		ingress.ObjectMeta.Annotations["traefik.ingress.kubernetes.io/router.middlewares"] = strings.Join(traefikMiddlewares, ",")
	} else {
		delete(ingress.ObjectMeta.Annotations, "traefik.ingress.kubernetes.io/router.middlewares")
	}

	// ---

	nginxUpload := "nginx.ingress.kubernetes.io/proxy-body-size"
	if v, ok := i.Annotations[nginxUpload]; ok {
		ingress.ObjectMeta.Annotations[nginxUpload] = v
	} else {
		if i.MaxUploadSize <= 0 {
			i.MaxUploadSize = 1
		}
		ingress.ObjectMeta.Annotations[nginxUpload] = fmt.Sprintf("%dm", i.MaxUploadSize)
	}

	if i.Timeout <= 0 {
		i.Timeout = 60
	}
	writeTimeout := "transport.respondingTimeouts.writeTimeout"
	readTimeout := "transport.respondingTimeouts.readTimeout"
	idleTimeout := "transport.respondingTimeouts.idleTimeout"

	timeoutSeconds := fmt.Sprintf("%d", i.Timeout)

	ingress.ObjectMeta.Annotations[writeTimeout] = timeoutSeconds
	ingress.ObjectMeta.Annotations[readTimeout] = timeoutSeconds
	ingress.ObjectMeta.Annotations[idleTimeout] = timeoutSeconds

	if i.BuffersNumber <= 0 {
		i.BuffersNumber = 4
	}
	ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/proxy-buffers-number"] = fmt.Sprintf("%d", i.BuffersNumber)

	if i.BufferSize <= 0 {
		i.BufferSize = 4
	}
	ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/proxy-buffer-size"] = fmt.Sprintf("%dk", i.BufferSize)

	if i.HeaderBuffersNumber <= 0 {
		i.HeaderBuffersNumber = 4
	}
	if i.HeaderBufferSize <= 0 {
		i.HeaderBufferSize = 8
	}
	ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/large-client-header-buffers"] = fmt.Sprintf("%d %dk", i.HeaderBuffersNumber, i.HeaderBufferSize)

	// ---

	if len(ingressOld) == 0 {
		i.Obj, err = ingressCtl.Create(i.KubeClient.CTX, ingress, metav1.CreateOptions{})
		return "creating new obj", log.Error(err)
	}
	ingressNew, err := json.Marshal(ingress)
	if err != nil {
		panic(err)
	}

	patch, err := jsonpatch.CreateMergePatch(ingressOld, ingressNew)
	if err != nil {
		panic(err)
	}

	if len(patch) == 2 {
		return "", nil
	}

	i.Obj, err = ingressCtl.Patch(i.KubeClient.CTX, i.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	return string(patch), log.Error(err)
}

func (i *Ingress2S) GetObj() *netV1.Ingress {

	var err error
	i.Obj, err = i.KubeClient.API.NetworkingV1().Ingresses(i.Namespace).Get(i.KubeClient.CTX, i.Name, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	return i.Obj
}

func (i *Ingress2S) Exist() bool {
	return i.GetObj() != nil
}

func (i Ingress2S) traefikMiddlewareCRD(name string) (obj *unstructured.Unstructured, err error) {
	obj = &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "traefik.containo.us",
		Version: "v1alpha1",
		Kind:    "Middleware",
	})
	obj.SetNamespace(i.Namespace)
	obj.SetName(name) // Middleware name
	err = i.KubeClient.CRD.Get(i.KubeClient.CTX, client.ObjectKeyFromObject(obj), obj)
	return
}

func (i Ingress2S) traefikBasicAuthAdd(name string) *log.RecordS {
	// Get Traefik Basic Auth middleware
	obj, err := i.traefikMiddlewareCRD(name)

	if err == nil {
		// Basic Auth exists

		new := map[string]interface{}{
			"basicAuth": map[string]string{
				"secret": name,
			},
		}

		// Check for changes
		if reflect.DeepEqual(obj.Object["spec"], new) {
			// brak zmian
			return nil
		}

		// Update pod scrape
		obj.Object["spec"] = new
		return log.Error(i.KubeClient.CRD.Update(i.KubeClient.CTX, obj))
	}

	// Create Basic Auth middleware
	obj.Object["spec"] = map[string]interface{}{
		"basicAuth": map[string]string{
			"secret": name,
		},
	}

	return log.Error(i.KubeClient.CRD.Create(i.KubeClient.CTX, obj))
}

func (i Ingress2S) traefikPathRewriteAdd(from, to string) (name string, err error) {
	// Set middleware name
	name = conv.KeyString(fmt.Sprintf("%s-%s-path", from, to))

	// Get Traefik Path Rewrite middleware
	obj, err := i.traefikMiddlewareCRD(name)

	// Set real middleware name
	name = fmt.Sprintf("%s-%s@kubernetescrd", i.Namespace, name)

	// Create middleware spec
	spec := map[string]interface{}{
		"replacePathRegex": map[string]string{
			"regex":       fmt.Sprintf("^%s(.*)", from),
			"replacement": fmt.Sprintf("%s$1", to),
		},
	}

	if err == nil {
		// Middleware exists

		// Check for changes
		if reflect.DeepEqual(obj.Object["spec"], spec) {
			return name, nil
		}

		// Update pod scrape
		obj.Object["spec"] = spec
		return name, i.KubeClient.CRD.Update(i.KubeClient.CTX, obj)
	}

	// Create new middleware
	obj.Object["spec"] = spec

	return name, i.KubeClient.CRD.Create(i.KubeClient.CTX, obj)
}

func (i *Ingress2S) Delete() error {
	tlog.Info("Ingress Delete")
	return i.KubeClient.API.NetworkingV1().Ingresses(i.Namespace).Delete(i.KubeClient.CTX, i.Name, metav1.DeleteOptions{})
}
