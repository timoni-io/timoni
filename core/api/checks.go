package api

import (
	"core/db2"
	"core/kube"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func Check() (db2.StateT, string) {

	kClient := kube.GetKube()
	obj := &kube.StatefulSetS{
		KubeClient: kClient,
		Namespace:  "timoni",
		Name:       "core",
	}
	if !obj.IsReady() {
		return db2.State_deploying, "StatefulSet `core` not ready"
	}

	_, imageTag := kube.GetImageInfo(obj.Obj.Spec.Template.Spec.Containers[0].Image)
	if imageTag != db2.TheSettings.ReleaseGitTag() {
		return db2.State_error, fmt.Sprintf("StatefulSet `core` ImageTag is wrong `%s` != `%s`", imageTag, db2.TheSettings.ReleaseGitTag())
	}

	// ---

	return CheckDomain()
}


func CheckDomain() (db2.StateT, string) {

	url := db2.TheDomain.URL("/api/check-http")

	client := http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, time.Duration(5*time.Second))
			},
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return db2.State_error, url + " get error: " + err.Error()
	}

	if resp.StatusCode != 200 {
		return db2.State_error, fmt.Sprintf("%s http code: %d", url, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return db2.State_error, url + " resp.Body read fail: " + err.Error()
	}
	bodyString := string(bodyBytes)
	if bodyString != "ok" {
		return db2.State_error, url + " wrong resp: " + bodyString
	}

	return db2.State_ready, ""
}
