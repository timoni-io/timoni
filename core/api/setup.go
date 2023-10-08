package api

import (
	"context"
	"core/config"
	"core/db2"
	"core/kube"
	"core/modulestate"
	"crypto/tls"
	"fmt"
	"lib/tlog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	LoadBalancerIPv4   string
	LoadBalancerDomain string
)

func Setup() {

	modulestate.StatusByModulesAdd("api", Check)

	go Loop() // Start api

	createLoadBalancer()

	commitSHA := config.CommitSHA
	if len(commitSHA) > 8 {
		commitSHA = commitSHA[:8]
	}

	tlog.Info("Timoni {{gitTag}} ({{commitSHA}}) is starting {{url}} ...", tlog.Vars{
		"gitTag":    config.GitTag,
		"commitSHA": commitSHA,
		"url":       db2.TheDomain.URL(""),
	})

}

func httpGetFromIP(url, targetAddr string, ignoreCertVericication bool) (success bool) {
	client := http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCertVericication},
		},
	}

	if targetAddr != "" {
		dialer := &net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 3 * time.Second,
		}
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCertVericication},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, targetAddr)
			},
		}
	}

	res, err := client.Get(url)
	if err == nil && res != nil && res.StatusCode == 200 {
		return true
	}

	statusCode := 0
	if res != nil {
		statusCode = res.StatusCode
	}

	tlog.Info("check URL {{url}} failed with {{addr}}", tlog.Vars{
		"url":        url,
		"addr":       targetAddr,
		"err":        err,
		"statusCode": statusCode,
	})
	return false
}

func getNodeIP() string {
	kClient := kube.GetKube()

	nodes, err := kClient.API.CoreV1().Nodes().List(kClient.CTX, metav1.ListOptions{})
	if tlog.Error(err) != nil {
		return ""
	}

	for _, node := range nodes.Items {

		for _, addr := range node.Status.Addresses {
			if addr.Type == "InternalIP" {
				return addr.Address
			}
		}

	}

	return ""
}

type webPanelDomainIPorCNAME struct {
	Source string
	IPv4   string
	CNAME  string
}

func getDomainIP() webPanelDomainIPorCNAME {

	localURL := fmt.Sprintf("https://127.0.0.1:%d/api/check-http", db2.TheDomain.Port())
	for {
		if httpGetFromIP(localURL, "", true) {
			break
		}

		tlog.Warning("Waiting for API to work locally...")
		time.Sleep(5 * time.Second)
	}

	result := webPanelDomainIPorCNAME{}

	if db2.TheDomain.IP() != "" {
		result.IPv4 = db2.TheDomain.IP()
		result.Source = "StaticIPv4"
		return result
	}

	// `Config.LastIP` check if its accessable, if not use other sources
	// check access
	if httpGetFromIP(
		db2.TheDomain.URL("/api/check-http"),
		fmt.Sprintf("%s:%d", db2.TheSettings.LastIP(), db2.TheDomain.Port()),
		false,
	) {
		// OK
		result.IPv4 = db2.TheSettings.LastIP()
		result.Source = "FocalPointLastIP"
		return result
	}

	if LoadBalancerIPv4 != "" {
		result.IPv4 = LoadBalancerIPv4
		result.Source = "LoadBalancerIPv4"
		return result
	}

	if LoadBalancerDomain != "" {

		result.IPv4 = resolveIP(LoadBalancerDomain)
		result.CNAME = LoadBalancerDomain
		result.Source = "LoadBalancerDomain"
		return result
	}

	nodeIP := getNodeIP()
	if nodeIP != "" {
		result.IPv4 = nodeIP
		result.Source = "NodeIP"
		return result
	}

	tlog.Fatal("NodeIP and LoadBalancerIPv4 are empty")
	return result
}

func resolveIP(domain string) string {
	addr, err := net.LookupIP(domain)
	tlog.Error(err)

	if len(addr) > 0 {
		return addr[0].String()
	}
	return ""
}

func createLoadBalancer() {

	kClient := kube.GetKube()

	domain := db2.TheDomain
	domainPort := int32(domain.Port())

	svc := kube.ServiceS{
		KubeClient: kClient,
		Namespace:  "timoni",
		Name:       "core",
		Ports:      map[int32]int32{domainPort: domainPort},
		TargetSelector: map[string]string{
			"element": "core",
		},
		Labels: map[string]string{
			"element": "core",
		},
		LoadBalancer: true,
		Internal:     true,
	}
	for {
		_, err := svc.CreateOrUpdate()
		if tlog.Error(err) == nil {
			break
		}
		tlog.Warning("Waiting for core LB ...")
		time.Sleep(5 * time.Second)
	}

	// ---

	for {
		is := svc.GetObj()
		if is == nil || is.Name != svc.Name || len(is.Status.LoadBalancer.Ingress) == 0 {
			tlog.Warning("Waiting for core service IP...")
			time.Sleep(5 * time.Second)
			continue
		}

		db2.TheSettings.SetAPIInternalIP(is.Spec.ClusterIP)

		LoadBalancerIPv4 = is.Status.LoadBalancer.Ingress[0].IP
		if LoadBalancerIPv4 != "" {
			break
		}

		if len(is.Status.LoadBalancer.Ingress) == 1 &&
			is.Status.LoadBalancer.Ingress[0].IP == "" &&
			is.Status.LoadBalancer.Ingress[0].Hostname == "localhost" {
			if domain.IP() != "" {
				LoadBalancerIPv4 = domain.IP()
			} else {
				LoadBalancerIPv4 = resolveIP("gateway.docker.internal")
			}
			break
		}

		if len(is.Status.LoadBalancer.Ingress) == 1 &&
			is.Status.LoadBalancer.Ingress[0].IP == "" &&
			is.Status.LoadBalancer.Ingress[0].Hostname != "" {
			LoadBalancerDomain = is.Status.LoadBalancer.Ingress[0].Hostname
			break
		}

		tlog.Warning("Waiting for core service IP ...")
		time.Sleep(5 * time.Second)
	}
}

func WaitForAPI() {

	domain := db2.TheDomain
	data := getDomainIP()

	if data.IPv4 == "" {
		tlog.PrintJSON(data)
		tlog.Fatal("cant find IP for WebPanel.Domain")
	}

	if LoadBalancerIPv4 == "" {
		LoadBalancerIPv4 = data.IPv4
	}

	if data.Source == "LoadBalancerDomain" && data.CNAME == "" {
		tlog.Fatal("LoadBalancerDomain/CNAME is empty")
	}

	_, err := exec.Command("sh", "-c", "echo '"+data.IPv4+"  "+domain.Name()+"' >> /etc/hosts").CombinedOutput()
	if os.Getenv("DEV_MODE") != "true" {
		tlog.Fatal(err)
	}
}
