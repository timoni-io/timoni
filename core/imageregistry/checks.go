package imageregistry

import (
	"core/db2"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func Check() (db2.StateT, string) {
	url := db2.TheDomain.URL("/v2/")

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
	if bodyString != "{}" {
		return db2.State_error, url + " wrong resp: " + bodyString
	}

	return db2.State_ready, ""
}
