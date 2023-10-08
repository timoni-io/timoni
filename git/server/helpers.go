package gitserver

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Mode struct {
	Type    string
	Service string
	Branch  string
}

var (
	Invalid     = Mode{}
	ReceivePack = Mode{Type: "ReceivePack"}
	UploadPack  = Mode{Type: "UploadPack"}
	Control     = Mode{Type: "Control"}
)

func Service(s string) Mode {
	return Mode{Type: "Service", Service: s}
}

type Action string

const (
	Undefined Action = ""
	Pull      Action = "Pull"
	Push      Action = "Push"
)

func (m Mode) Action() Action {
	switch {
	case m == UploadPack || m.Service == "upload-pack":
		return Pull
	case m == ReceivePack || m.Service == "receive-pack":
		return Push
	default:
		return Undefined
	}
}

func parsePath(r *http.Request) (repo string, mode Mode) {
	repoName, path, ok := strings.Cut(r.URL.Path, "/")
	if !ok {
		return "", Invalid
	}

	switch path {
	case "control":
		return repoName, Control
	case "info/refs":
		service := getService(r)
		if service == "" {
			return "", Invalid
		}
		return repoName, Service(service)

	case "git-upload-pack":
		return repoName, UploadPack

	case "git-receive-pack":
		return repoName, ReceivePack

	default:
		return "", Invalid
	}
}

// generate packet header
func servicePacket(service string) []byte {
	packet := fmt.Sprintf("# service=git-%s\n", service)
	prefix := strconv.FormatInt(int64(len(packet)+4), 16)
	if len(prefix)%4 != 0 {
		prefix = strings.Repeat("0", 4-len(prefix)%4) + prefix
	}
	magicMarker := "0000"
	return []byte(prefix + packet + magicMarker)
}

// extract service name
func getService(r *http.Request) string {
	s := r.URL.Query().Get("service")
	// ie. git-receive-pack -> receive-pack
	_, service, _ := strings.Cut(s, "-")
	return service
}
