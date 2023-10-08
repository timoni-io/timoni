package db

import (
	"core/db2"
	"core/db2/fp"
	"core/kube"
	"fmt"
	"lib/tlog"
	"lib/utils/conv"
	"lib/utils/slice"
	"strings"
	"time"

	"github.com/lukx33/lwhelper/out"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (element *elementDomainS) KubeApply() {
	defer PanicHandler()

	kClient := kube.GetKube()
	es := element.GetStatus()

	ingName := conv.KeyString(element.Name + "-" + element.Domain)

	// if global.Config.IngressOldVersion {

	// 	ingress := config.Ingress1S{
	// 		KubeClient:  kClient,
	// 		Namespace:   element.EnvironmentID,
	// 		Name:        conv.KeyString(element.Name + "-" + element.Domain),
	// 		Domain:      element.Domain,
	// 		Annotations: element.Annotations,
	// 		Labels: map[string]string{
	// 			"timoni-env": element.EnvironmentID,
	// 			// "timoni-version": release.Env.CurrentReleaseID,
	// 			// "element": element.Name,
	// 			// "element-port": fmt.Sprint(portNr),
	// 		},
	// 		Paths:         element.Paths,
	// 		MaxUploadSize: element.UploadLimit,
	// 		Timeout:       element.Timeout,
	// 		HTTPS:         element.Scheme == "https",
	// 		Auth:          element.Auth,
	// 	}

	// 	_, err := ingress.CreateOrUpdate()
	// 	if err != nil {
	// 		es.State = ElementStatusFailed
	// 		return

	// 	}

	// } else {

	// cleanup
	{
		ingressCtl := kClient.API.NetworkingV1().Ingresses(element.EnvironmentID)
		ing, err := ingressCtl.List(kClient.CTX, metav1.ListOptions{LabelSelector: "element=" + element.Name})
		if err != nil {
			tlog.Error(err)
		}
		if len(ing.Items) > 0 {
			// Remove old ingresses
			for _, i := range ing.Items {
				if i.Name == ingName {
					continue
				}
				err := ingressCtl.Delete(kClient.CTX, i.Name, metav1.DeleteOptions{})
				if err != nil {
					tlog.Error(err)
				}
			}
		}
	}

	switch element.ExternalProtocol {
	case "tcp", "udp", "lb":
		element.KubeApplyLoadalancer()
		return

	}

	https := false
	if !element.HttpOnly {
		cert := db2.DomainList("Name = '"+element.Domain+"' AND Cert != ''", "", 0, 1).First().Cert()
		https = element.ExternalProtocol == "https" || cert.ID() != ""

		if cert.NotValid() {
			domain := db2.DomainList("Name = '"+element.Domain+"'", "", 0, 1).First()

			if !domain.NotValid() && !domain.HTTPS() {
				element.HttpOnly = true
			}

			if domain.InfoResult() == out.NotFound &&
				(strings.HasSuffix(element.Domain, ".timoni.dev") ||
					strings.HasSuffix(element.Domain, "ps.com")) {

				t := strings.Split(element.Domain, ".")
				tld := strings.Join(t[len(t)-2:], ".")

				res := fp.DomainCreate(
					fp.CertGetByID(""),
					fp.CertProviderList("Name = 'Timoni' AND Organization = ''", "", 0, 1).First(),
					fp.DNSProviderList("Name = 'Timoni' AND Organization = ''", "", 0, 1).First(),
					"",
					element.Domain,
					fp.OrganizationGetByID(db2.TheOrganization.ID()),
					tld,
				)
				if res.NotValid() {
					res.InfoPrint()
				} else {
					db2.SyncWithFP()
				}
			}

			https = false
		}

		if https {
			secret := kube.SecretS{
				KubeClient: kClient,
				Namespace:  element.EnvironmentID,
				Name:       element.Domain + "-tls",
				Type:       corev1.SecretTypeTLS,
				Data: map[string][]byte{
					"tls.crt": []byte(cert.Pem()),
					"tls.key": []byte(cert.Key()),
				},
				Labels: map[string]string{
					"manager": "timoni",
				},
			}
			secret.CreateOrUpdate()
		}
	}

	ingress := kube.Ingress2S{
		KubeClient:  kClient,
		Namespace:   element.EnvironmentID,
		Name:        ingName,
		Domain:      element.Domain,
		Annotations: element.Annotations,
		Labels: map[string]string{
			"timoni-env": element.EnvironmentID,
			// "timoni-version": release.Env.CurrentReleaseID,
			"element": element.Name,
			// "element-port": fmt.Sprint(portNr),
		},
		Paths:         element.Paths,
		MaxUploadSize: element.UploadLimit,
		Timeout:       element.Timeout,
		HTTPS:         https,
		Auth:          element.Auth,
	}

	es.Alerts = []string{}
	_, err := ingress.CreateOrUpdate()
	if err != nil {
		if !slice.Contains(es.Alerts, err.Message) {
			es.Alerts = append(es.Alerts, err.Message)
		}
		es.State = ElementStatusFailed
		return
	}

	// }

	es.State = ElementStatusReady
}

func (element *elementDomainS) KubeApplyLoadalancer() {

	kClient := kube.GetKube()
	es := element.GetStatus()

	switch element.ExternalPort {
	case 22, 80, 443, int(db2.TheDomain.Port()):
		es.Alerts = append(es.Alerts, fmt.Sprintf("Port %d not allowed", element.ExternalPort))
		es.State = ElementStatusFailed
		return
	}

	isvc := kube.ServiceS{
		KubeClient:  kClient,
		Namespace:   element.EnvironmentID,
		Name:        conv.KeyString(element.Name),
		Annotations: element.Annotations,
		Labels: map[string]string{
			"timoni-env": element.EnvironmentID,
			// "timoni-version": release.Env.CurrentReleaseID,
			"element": element.Name,
			// "element-port": fmt.Sprint(portNr),
		},
		TargetSelector: map[string]string{
			"element": element.Paths["/"].ElementName,
		},
		LoadBalancer: true,
		Protocol:     element.ExternalProtocol,
		Internal:     false,
		Ports: map[int32]int32{
			int32(element.ExternalPort): element.Paths["/"].Port,
		},
	}
	_, err := isvc.CreateOrUpdate()
	tlog.Error(err)

	is := isvc.GetObj()
	if is == nil || is.Name != isvc.Name || len(is.Status.LoadBalancer.Ingress) == 0 {
		tlog.Warning("Waiting for service " + conv.KeyString(element.Name))
		time.Sleep(5 * time.Second)
		es.State = ElementStatusDeploying
		return
	}

	es.State = ElementStatusReady
}
