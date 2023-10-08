package term

import (
	"core/db"
	"core/db2"
	"core/kube"
	_ "embed"
	"fmt"
	"net/http"
	"strings"
)

//go:embed index.html
var TerminalHTML string

func Xterm(w http.ResponseWriter, r *http.Request) {
	envID := r.FormValue("env")
	if envID == "" {
		fmt.Fprint(w, "ERROR: `env` is required")
		return
	}

	elementName := r.FormValue("element")
	if elementName == "" {
		fmt.Fprint(w, "ERROR: `element` is required")
		return
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		fmt.Fprint(w, "ERROR: app `"+envID+"` not found")
		return
	}

	podName := r.FormValue("pod")
	debug := r.FormValue("debug")

	element := env.GetElement(elementName)
	if element == nil {
		fmt.Fprint(w, "ERROR: element `"+elementName+"` not found")
		return
	}

	es := element.GetStatus()
	pods := es.PodsGet()
	if len(pods) == 0 {
		fmt.Fprint(w, "ERROR: element `"+elementName+"` does not have pods")
		return
	}

	if podName == "" {
		for k, pod := range pods {
			if debug == "true" && !pod.Debug {
				continue
			}
			if pod.Status == kube.PodStatusReady || pod.Status == kube.PodStatusRunning {
				podName = k
				break
			}
		}

		if podName == "" {
			fmt.Fprint(w, "ERROR: element `"+elementName+"` does not have any ready pods")
			return
		}
	} else {
		if _, ok := pods[podName]; !ok {
			fmt.Fprint(w, "ERROR: element `"+elementName+"` does not have pod "+podName)
			return
		}
	}

	containerName := elementName
	if debug == "true" {
		containerName += "-debug"
	}

	// namespace pod container
	domain := db2.TheDomain
	url := fmt.Sprintf("wss://%s:%d/term?namespace=%s&pod=%s&container=%s", domain.Name(), domain.Port(), envID, podName, containerName)
	strings.NewReplacer(
		"{{URL}}", url,
		"{{TITLE}}", fmt.Sprintf("%s:%s", envID, containerName),
	).WriteString(w, TerminalHTML)
}
