package db

import (
	"core/kube"
	"fmt"
	"lib/tlog"
	"lib/utils/conv"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type elementDomainS struct {
	elementS

	Domain           string `toml:"domain"`
	ExternalProtocol string `toml:"external-protocol"` // default 'https', posible: http, tcp, udp, lb
	ExternalPort     int    `toml:"external-port"`

	// nativeLB in https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#kind-ingressrouteudp
	// DontUseLoadBalancer bool `toml:"dont-use-load-balancer"`

	HttpOnly    bool `toml:"http-only"`
	WWWredirect bool `toml:"www-redirect"`

	Auth        string `toml:"auth"`         // basic auth eg. 'user:password_hash' < htpasswd -nb {login} {pass}
	UploadLimit int    `toml:"upload-limit"` // default 1 MB, proxy-body-size
	Timeout     int    `toml:"timeout"`      // default 60 s, proxy-read-timeout, client-header-timeout, client-body-timeout
	// BuffersNumber       int    `toml:"buffers-number"`        // default 4, proxy-buffers-number
	// BufferSize          int    `toml:"buffer-size"`           // default 4 k, proxy-buffer-size
	// HeaderBuffersNumber int    `toml:"header-buffers-number"` // default 4, large-client-header-buffers
	// HeaderBufferSize    int    `toml:"header-buffer-size"`    // default 8 k, large-client-header-buffers

	// StartPath   string                         `toml:"start-path"`
	Paths       map[string]*kube.DomainPathS `toml:"paths"`
	Annotations map[string]string            `toml:"annotations"`
	// URL         string                         `toml:"-"`
}

const (
	defaultDomainTemplate = "{{NAMESPACE}}.{{CLUSTER_DOMAIN}}"
)

func (element *elementDomainS) RebuildImage(imageID string, user *UserS) *tlog.RecordS {
	return nil
}

func (element *elementDomainS) GetImage() *ImageS {
	return nil
}

func (element *elementDomainS) RestartAllPods(user *UserS) *tlog.RecordS {
	return nil
}

func (element *elementDomainS) check(user *UserS) *tlog.RecordS {
	for k, v := range element.Annotations {
		if strings.TrimSpace(k) == "" {
			delete(element.Annotations, k)
		}

		if strings.TrimSpace(v) == "" {
			delete(element.Annotations, k)
		}
	}

	if element.Domain == "" {
		element.Domain = defaultDomainTemplate
	}

	if len(element.Paths) == 0 {
		status := element.GetStatus()
		status.Alerts = append(status.Alerts, "Paths is empty")
		status.Save()
	}

	element.RenderVariables()

	// run generic check
	return element.elementS.check(user)
}

func (element *elementDomainS) Save(user *UserS) *tlog.RecordS {
	if element.ToDelete {
		return nil
	}
	if err := element.check(user); err != nil {
		return err
	}
	return elementSave(element, user)
}

func (element *elementDomainS) RenderVariables() bool {
	element.addSystemVariables()
	element.Variables["DOMAIN_NAME"] = &ElementVariableS{
		Errors:       map[string]errorMessageT{},
		FirstValue:   element.Domain,
		CurrentValue: element.Domain,
		System:       true,
	}
	element.elementS.RenderVariables()
	element.Domain = element.Variables["DOMAIN_NAME"].ResolvedValue
	return true
}

func (element *elementDomainS) Merge(el EnvElementS) *tlog.RecordS {
	e, ok := el.(*elementDomainS)
	if !ok {
		return tlog.Error(fmt.Sprintln("wrong type, expected", element.GetType(), "got", el.GetType()))
	}

	if e.Domain == "" {
		e.Domain = defaultDomainTemplate
	}

	if len(e.Paths) == 0 {
		status := e.GetStatus()
		status.Alerts = append(status.Alerts, "Paths is empty")
		status.Save()
	}

	return element.elementS.Merge(&e.elementS)
}

func (element *elementDomainS) CopySecrets(el EnvElementS) *tlog.RecordS {
	e, ok := el.(*elementDomainS)
	if !ok {
		return tlog.Error(fmt.Sprintln("wrong type, expected", element.GetType(), "got", el.GetType()))
	}

	return element.elementS.CopySecrets(&e.elementS)
}

func (element *elementDomainS) DeleteFromKube() *tlog.RecordS {

	if element.EnvironmentID == "" {
		return tlog.Error("element.EnvironmentID is empty")
	}

	switch element.ExternalProtocol {
	case "tcp", "udp", "lb":
		return element.DeleteLoadBalancer()
	}

	kClient := kube.GetKube()
	if kClient == nil {
		return tlog.Error("kube.Client not ready")
	}

	obj, err := kClient.API.NetworkingV1().Ingresses(element.EnvironmentID).Get(kClient.CTX, conv.KeyString(element.Name+"-"+element.Domain), v1.GetOptions{})
	if err != nil || obj == nil || obj.Name == "" {
		return nil
	}

	tlog.Info("DeleteFromKube", tlog.Vars{
		"envID":       element.EnvironmentID,
		"elementName": element.Name,
	})

	kClient.API.NetworkingV1().Ingresses(element.EnvironmentID).Delete(
		kClient.CTX,
		conv.KeyString(element.Name+"-"+element.Domain),
		v1.DeleteOptions{},
	)
	return nil
}

func (element *elementDomainS) DeleteLoadBalancer() *tlog.RecordS {
	kClient := kube.GetKube()
	if kClient == nil {
		return tlog.Error("kube.Client not ready")
	}

	err := kClient.API.CoreV1().Services(element.EnvironmentID).Delete(kClient.CTX, conv.KeyString(element.Name), v1.DeleteOptions{})
	if err != nil {
		return tlog.Error(err)
	}
	return nil
}

func (element *elementDomainS) GetScale() *ElementScaleS {
	return &ElementScaleS{}
}
