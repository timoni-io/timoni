package global

import (
	"fmt"
	"os"
)

var (
	GitTag = "???"
)

var (
	Token        = os.Getenv("TIMONI_TOKEN")
	TimoniDomain = os.Getenv("TIMONI_DOMAIN")
	TimoniPort   = os.Getenv("TIMONI_PORT")
	TimoniIP     = os.Getenv("TIMONI_IP")
	UpdateURL    = fmt.Sprintf("%s/api/image-status-update", TimoniURL())
)

func TimoniURL() string {
	return fmt.Sprintf("https://%s:%s", TimoniDomain, TimoniPort)
}

func TimoniDomainAndPort() string {
	return fmt.Sprintf("%s:%s", TimoniDomain, TimoniPort)
}
