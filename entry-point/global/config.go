package global

import (
	"fmt"
	"lib/utils"
	"lib/utils/env"
	"os"
	"strings"
	"time"
)

var (
	GitTag string // build version

	ProcessCommand = os.Args[1:]

	JournalProxyURL = func() string {
		tmp := os.Getenv("TIMONI_JOURNAL_PROXY")
		if tmp == "" {
			return "ws://journal-proxy-1.timoni.svc:4003"
		}
		return fmt.Sprintf("ws://journal-proxy-%s.timoni.svc:4003", tmp)
	}()

	PodName            = env.Get[string]("POD_NAME")
	ElementName        = env.Get[string]("ELEMENT_NAME")
	ElementVersion     = env.Get("ELEMENT_VERSION", "")
	ElementGitRepoName = env.Get("ELEMENT_GIT_REPO_NAME", "")

	EnvironmentID = func() string {
		tmp := env.Get[string]("NAMESPACE")

		if tmp == "timoni" {
			if ElementName == "image-builder" || ElementName == "ingress-traefik" {
				tmp = ElementName
			}
		}

		return tmp
	}()

	ApplyVariablesOnFiles = env.Get("EP_FIX_ENV", []string{})
	ShowOutput            = env.Get("EP_SHOW_OUTPUT", true)
	ReadLogFilesFromDir   = env.Get("EP_LOG_DIR", "")
	ActionToken           = env.Get("EP_ACTION_TOKEN", "")
	StaticFilesPath       = env.Get("EP_STATIC_FILES_PATH", "")
	ParserFormat          = env.Get("EP_PARSER_FORMAT", "")

	CronUntilMinutes = env.Get("EP_CRON_UNTIL", 0)
	CronExpression   = func() string {
		tmp := env.Get("EP_CRON_EXPRESSION", "")
		if tmp == "-" {
			tmp = ""
		}
		return tmp
	}()

	TimoniURL = env.Get("TIMONI_URL", "")

	InitialEnvs = GetEnvMap()

	LogWriter = utils.Must(NewLogWriter(ConfigS{
		URL:        JournalProxyURL + "/in",
		ShowOutput: ShowOutput,
		BaseMessage: &Message{
			EnvID:   EnvironmentID,
			Element: ElementName,
			Version: ElementVersion,
			Project: ElementGitRepoName,
			Pod:     PodName,
		},
		BatchDuration: 100 * time.Millisecond,
		BatchCapacity: 200_000,
		ParserFormat:  ParserFormat,
	}))
)

func GetEnvMap() map[string]string {
	res := map[string]string{}
	for _, e := range os.Environ() {
		tmp := strings.SplitN(e, "=", 2)
		res[tmp[0]] = tmp[1]
	}
	return res
}
