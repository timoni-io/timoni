package api

import (
	"core/config"
	"core/db2"
	"core/imageregistry"
	"core/term"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	log "lib/tlog"
	"lib/utils"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lukx33/lwhelper/out"
	"github.com/prometheus/client_golang/prometheus"
)

const timoniLogoSvgBase64 = "PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIGlkPSJXYXJzdHdhXzIiIHZpZXdCb3g9IjAgMCAxMTUuNDUgNDYuODMiIHN0eWxlPSImIzEwOyAgICBmaWxsOiBibGFjazsmIzEwOyAgICBjb2xvcjogYmxhY2s7JiMxMDsiPjxkZWZzPjxzdHlsZT4uY2xzLTF7ZmlsbDojMDAwO308L3N0eWxlPjwvZGVmcz48ZyBpZD0iV2Fyc3R3YV8xLTIiIHN0eWxlPSJjb2xvcjogYmxhY2s7Ij48Zz48Zz48cGF0aCBjbGFzcz0iY2xzLTEiIGQ9Ik03LjUsMjkuMDdjLTIuNTMsMC00LjQxLS41NC01LjYzLTEuNjNDLjY1LDI2LjM1LC4wNCwyNC42NiwuMDQsMjIuMzZsLS4wNC0xNi45SDguMDR2NC4yNGg0LjA5djUuOTJoLTQuMDl2NS4wM2MwLC45MSwuMiwxLjUyLC41OSwxLjg1LC4zOSwuMzIsLjk0LC40OCwxLjYzLC40OCwuMjYsMCwuNTQtLjAyLC44My0uMDcsLjI5LS4wNSwuNTEtLjEsLjY4LS4xNHY1LjY3Yy0uNDMsLjE5LTEuMDQsLjM1LTEuODEsLjQ3LS43OCwuMTItMS42LC4xOC0yLjQ2LC4xOCIvPjxwYXRoIGNsYXNzPSJjbHMtMSIgZD0iTTE0LjksOS42OWg4LjI5VjI4LjU3SDE0LjlWOS42OVptLS4xNC01LjQ2YzAtLjU4LC4xMS0xLjEyLC4zNC0xLjYzLC4yMy0uNTEsLjU0LS45NiwuOTMtMS4zNSwuNC0uMzgsLjg2LS42OSwxLjM4LS45MSwuNTMtLjIzLDEuMDktLjM0LDEuNjktLjM0czEuMTEsLjExLDEuNjEsLjM0Yy41LC4yMywuOTUsLjUzLDEuMzUsLjkxLC4zOSwuMzgsLjcxLC44MywuOTMsMS4zNSwuMjMsLjUxLC4zNCwxLjA2LC4zNCwxLjYzcy0uMTEsMS4xNS0uMzQsMS42N2MtLjIzLC41Mi0uNTQsLjk2LS45MywxLjMzLS4zOSwuMzctLjg0LC42Ni0xLjM1LC44OC0uNSwuMjItMS4wNCwuMzItMS42MSwuMzJzLTEuMTYtLjExLTEuNjktLjMyYy0uNTMtLjIyLS45OS0uNTEtMS4zOC0uODgtLjM5LS4zNy0uNzEtLjgxLS45My0xLjMzLS4yMy0uNTEtLjM0LTEuMDctLjM0LTEuNjciLz48cGF0aCBjbGFzcz0iY2xzLTEiIGQ9Ik0zOC41OCwyOC41N3YtOS45MWMwLS45Ni0uMTItMS42NS0uMzYtMi4wOC0uMjQtLjQzLS42My0uNjUtMS4xOC0uNjUtLjQxLDAtLjgsLjE5LTEuMTcsLjU3LS4zNywuMzktLjU2LDEuMDctLjU2LDIuMDV2MTAuMDFoLTguMjlWOS42OWg4LjA3djIuMDhoLjA3Yy4yNC0uMzEsLjUzLS42MSwuODgtLjkxLC4zNS0uMywuNzQtLjU3LDEuMTgtLjgzLC40NC0uMjUsLjk0LS40NSwxLjQ5LS42MSwuNTUtLjE2LDEuMTUtLjIzLDEuNzktLjIzLDEuMjQsMCwyLjMyLC4yMiwzLjIzLC42NiwuOTEsLjQ0LDEuNjMsMS4xMywyLjE1LDIuMDcsLjYyLS43NywxLjM5LTEuNDEsMi4zLTEuOTQsLjkxLS41MywyLjA2LS43OSwzLjQ1LS43OXMyLjQ1LC4yNiwzLjMyLC43OWMuODcsLjUzLDEuNTcsMS4xOSwyLjA4LDEuOTcsLjUxLC43OSwuODcsMS42NCwxLjA4LDIuNTYsLjIsLjkyLC4zLDEuNzgsLjMsMi41N3YxMS40OGgtOC4yOXYtMTAuMDVjMC0xLjAzLS4xMy0xLjcyLS4zOC0yLjA3LS4yNS0uMzUtLjY0LS41Mi0xLjE3LS41Mi0uNDgsMC0uODksLjIyLTEuMjIsLjY2LS4zNCwuNDQtLjUsMS4xLS41LDEuOTZ2MTAuMDFoLTguMjlaIi8+PHBhdGggY2xhc3M9ImNscy0xIiBkPSJNNzEuMSwxNS45N2MtLjkxLDAtMS42NSwuMzEtMi4yMSwuOTQtLjU2LC42Mi0uODQsMS4zNS0uODQsMi4xOSwwLC44OSwuMjksMS42NCwuODYsMi4yNiwuNTcsLjYyLDEuMzEsLjkzLDIuMjIsLjkzczEuNjUtLjMxLDIuMjMtLjkzYy41Ny0uNjIsLjg2LTEuMzgsLjg2LTIuMjYsMC0uODMtLjI5LTEuNTctLjg4LTIuMTktLjU5LS42Mi0xLjM0LS45NC0yLjI0LS45NG0xMC44LDMuMDljMCwxLjYzLS4yOSwzLjA4LS44OCw0LjM0LS41OSwxLjI3LTEuMzgsMi4zNC0yLjM3LDMuMjEtLjk5LC44Ny0yLjE0LDEuNTQtMy40NSwxLjk5LTEuMzEsLjQ1LTIuNjcsLjY4LTQuMTEsLjY4cy0yLjc3LS4yMy00LjA3LS42OGMtMS4zLS40Ni0yLjQ1LTEuMTItMy40My0xLjk5LS45OC0uODctMS43Ni0xLjk1LTIuMzUtMy4yMS0uNTktMS4yNy0uODgtMi43Mi0uODgtNC4zNHMuMjktMy4wMywuODgtNC4yOWMuNTktMS4yNiwxLjM3LTIuMzIsMi4zNS0zLjE4LC45OC0uODYsMi4xMi0xLjUxLDMuNDMtMS45NiwxLjMtLjQ0LDIuNjYtLjY2LDQuMDctLjY2czIuOCwuMjIsNC4xMSwuNjZjMS4zLC40NCwyLjQ1LDEuMDksMy40NSwxLjk2LC45OSwuODYsMS43OCwxLjkyLDIuMzcsMy4xOCwuNTksMS4yNiwuODgsMi42OCwuODgsNC4yOSIvPjxwYXRoIGNsYXNzPSJjbHMtMSIgZD0iTTg0LjAyLDkuNjloOC4wN3YyLjA4aC4wN2MuNDgtLjY5LDEuMTktMS4zLDIuMTMtMS44MSwuOTQtLjUxLDIuMDItLjc3LDMuMjQtLjc3LDEuMzEsMCwyLjQsLjI0LDMuMjcsLjcyLC44NywuNDgsMS41NiwxLjEsMi4wNywxLjg1LC41MSwuNzUsLjg3LDEuNiwxLjA3LDIuNTMsLjIsLjkzLC4zLDEuODcsLjMsMi44djExLjQ5aC04LjI5di0xMC4wNWMwLS45Ni0uMTUtMS42My0uNDUtMi4wMS0uMy0uMzgtLjc1LS41Ny0xLjM0LS41Ny0uNTUsMC0xLC4yMi0xLjM1LC42Ni0uMzUsLjQ0LS41MiwxLjA4LS41MiwxLjkydjEwLjA1aC04LjI5VjkuNjlaIi8+PHBhdGggY2xhc3M9ImNscy0xIiBkPSJNMTA3LjAyLDkuNjloOC4yOVYyOC41N2gtOC4yOVY5LjY5Wm0tLjE0LTUuNDZjMC0uNTgsLjExLTEuMTIsLjM0LTEuNjMsLjIzLS41MSwuNTQtLjk2LC45My0xLjM1LC40LS4zOCwuODYtLjY5LDEuMzgtLjkxLC41My0uMjMsMS4wOS0uMzQsMS42OS0uMzRzMS4xMSwuMTEsMS42MSwuMzRjLjUsLjIzLC45NSwuNTMsMS4zNSwuOTEsLjM5LC4zOCwuNzEsLjgzLC45MywxLjM1LC4yMywuNTEsLjM0LDEuMDYsLjM0LDEuNjNzLS4xMSwxLjE1LS4zNCwxLjY3Yy0uMjMsLjUyLS41NCwuOTYtLjkzLDEuMzMtLjM5LC4zNy0uODQsLjY2LTEuMzUsLjg4LS41LC4yMi0xLjA0LC4zMi0xLjYxLC4zMnMtMS4xNi0uMTEtMS42OS0uMzJjLS41My0uMjItLjk5LS41MS0xLjM4LS44OC0uMzktLjM3LS43MS0uODEtLjkzLTEuMzMtLjIzLS41MS0uMzQtMS4wNy0uMzQtMS42NyIvPjwvZz48Zz48cGF0aCBjbGFzcz0iY2xzLTEiIGQ9Ik03LjY2LDQyLjljLS4wNywxLjAxLS40NSwxLjgxLTEuMTIsMi4zOXMtMS41NiwuODctMi42NiwuODdjLTEuMiwwLTIuMTUtLjQxLTIuODQtMS4yMi0uNjktLjgxLTEuMDQtMS45Mi0xLjA0LTMuMzR2LS41N2MwLS45LC4xNi0xLjcsLjQ4LTIuMzksLjMyLS42OSwuNzctMS4yMiwxLjM2LTEuNTgsLjU5LS4zNywxLjI4LS41NSwyLjA2LS41NSwxLjA4LDAsMS45NiwuMjksMi42MiwuODcsLjY2LC41OCwxLjA1LDEuNCwxLjE1LDIuNDVoLTEuOTRjLS4wNS0uNjEtLjIyLTEuMDUtLjUxLTEuMzItLjI5LS4yNy0uNzMtLjQxLTEuMzMtLjQxLS42NSwwLTEuMTMsLjIzLTEuNDUsLjY5cy0uNDksMS4xOC0uNDksMi4xNXYuNzFjMCwxLjAxLC4xNSwxLjc2LC40NiwyLjIzcy43OSwuNywxLjQ1LC43Yy42LDAsMS4wNC0uMTQsMS4zNC0uNDFzLjQ2LS43LC41MS0xLjI3aDEuOTRaIi8+PHBhdGggY2xhc3M9ImNscy0xIiBkPSJNMTAuOTksNDYuMDJoLTEuOTR2LTkuNGgxLjk0djkuNFoiLz48cGF0aCBjbGFzcz0iY2xzLTEiIGQ9Ik0xMy4yNSw0Ni44M2gtMS4zOWwzLjQ2LTEwLjJoMS4zOWwtMy40NiwxMC4yWiIvPjxwYXRoIGNsYXNzPSJjbHMtMSIgZD0iTTI1LjEsNDIuOWMtLjA3LDEuMDEtLjQ1LDEuODEtMS4xMiwyLjM5cy0xLjU2LC44Ny0yLjY2LC44N2MtMS4yLDAtMi4xNS0uNDEtMi44NC0xLjIyLS42OS0uODEtMS4wNC0xLjkyLTEuMDQtMy4zNHYtLjU3YzAtLjksLjE2LTEuNywuNDgtMi4zOSwuMzItLjY5LC43Ny0xLjIyLDEuMzYtMS41OCwuNTktLjM3LDEuMjgtLjU1LDIuMDYtLjU1LDEuMDgsMCwxLjk2LC4yOSwyLjYyLC44NywuNjYsLjU4LDEuMDUsMS40LDEuMTUsMi40NWgtMS45NGMtLjA1LS42MS0uMjItMS4wNS0uNTEtMS4zMi0uMjktLjI3LS43My0uNDEtMS4zMy0uNDEtLjY1LDAtMS4xMywuMjMtMS40NSwuNjlzLS40OSwxLjE4LS40OSwyLjE1di43MWMwLDEuMDEsLjE1LDEuNzYsLjQ2LDIuMjNzLjc5LC43LDEuNDUsLjdjLjYsMCwxLjA0LS4xNCwxLjM0LS40MXMuNDYtLjcsLjUxLTEuMjdoMS45NFoiLz48cGF0aCBjbGFzcz0iY2xzLTEiIGQ9Ik0yNi4zNyw0Ni4wMnYtOS40aDIuODljLjgzLDAsMS41NiwuMTksMi4yMiwuNTYsLjY1LC4zNywxLjE2LC45LDEuNTMsMS41OSwuMzcsLjY5LC41NSwxLjQ3LC41NSwyLjM0di40M2MwLC44Ny0uMTgsMS42NS0uNTQsMi4zMy0uMzYsLjY4LS44NywxLjIxLTEuNTIsMS41OC0uNjUsLjM3LTEuMzksLjU2LTIuMjEsLjU3aC0yLjkxWm0xLjk0LTcuODN2Ni4yN2guOTRjLjc2LDAsMS4zNC0uMjUsMS43NC0uNzQsLjQtLjUsLjYtMS4yLC42MS0yLjEydi0uNWMwLS45Ni0uMi0xLjY4LS41OS0yLjE3LS40LS40OS0uOTctLjc0LTEuNzQtLjc0aC0uOTZaIi8+PHBhdGggY2xhc3M9ImNscy0xIiBkPSJNNDEuNzIsNDIuNTloLTEuNTR2My40NGgtMS45NHYtOS40aDMuNDljMS4xMSwwLDEuOTcsLjI1LDIuNTcsLjc0LC42LC41LC45LDEuMTksLjksMi4xLDAsLjY0LS4xNCwxLjE4LS40MiwxLjYtLjI4LC40My0uNywuNzctMS4yNiwxLjAybDIuMDMsMy44NHYuMDloLTIuMDhsLTEuNzYtMy40NFptLTEuNTQtMS41N2gxLjU2Yy40OSwwLC44Ni0uMTIsMS4xMy0uMzcsLjI3LS4yNSwuNC0uNTksLjQtMS4wMnMtLjEzLS43OS0uMzgtMS4wNWMtLjI1LS4yNS0uNjQtLjM4LTEuMTYtLjM4aC0xLjU2djIuODJaIi8+PHBhdGggY2xhc3M9ImNscy0xIiBkPSJNNTIuMzMsNDEuOTVoLTMuNzJ2Mi41Mmg0LjM2djEuNTZoLTYuM3YtOS40aDYuMjl2MS41N2gtNC4zNXYyLjI0aDMuNzJ2MS41MloiLz48cGF0aCBjbGFzcz0iY2xzLTEiIGQ9Ik01Ni4xNyw0Ni4wMmgtMS45NHYtOS40aDEuOTR2OS40WiIvPjxwYXRoIGNsYXNzPSJjbHMtMSIgZD0iTTY1LjYxLDQ2LjAyaC0xLjk0bC0zLjc3LTYuMTh2Ni4xOGgtMS45NHYtOS40aDEuOTRsMy43Nyw2LjJ2LTYuMmgxLjkzdjkuNFoiLz48cGF0aCBjbGFzcz0iY2xzLTEiIGQ9Ik03MC43Nyw0My43bDIuMTMtNy4wN2gyLjE2bC0zLjI3LDkuNGgtMi4wMmwtMy4yNi05LjRoMi4xNWwyLjEyLDcuMDdaIi8+PHBhdGggY2xhc3M9ImNscy0xIiBkPSJNODEuNTksNDEuOTVoLTMuNzJ2Mi41Mmg0LjM2djEuNTZoLTYuM3YtOS40aDYuMjl2MS41N2gtNC4zNXYyLjI0aDMuNzJ2MS41MloiLz48cGF0aCBjbGFzcz0iY2xzLTEiIGQ9Ik05MS4wMSw0Ni4wMmgtMS45NGwtMy43Ny02LjE4djYuMThoLTEuOTR2LTkuNGgxLjk0bDMuNzcsNi4ydi02LjJoMS45M3Y5LjRaIi8+PHBhdGggY2xhc3M9ImNscy0xIiBkPSJNOTkuNTksMzguMmgtMi44OHY3LjgzaC0xLjk0di03LjgzaC0yLjg0di0xLjU3aDcuNjV2MS41N1oiLz48cGF0aCBjbGFzcz0iY2xzLTEiIGQ9Ik0xMDYuMzQsNDEuOTVoLTMuNzJ2Mi41Mmg0LjM2djEuNTZoLTYuM3YtOS40aDYuMjl2MS41N2gtNC4zNXYyLjI0aDMuNzJ2MS41MloiLz48cGF0aCBjbGFzcz0iY2xzLTEiIGQ9Ik0xMDguMTMsNDYuMDJ2LTkuNGgyLjg5Yy44MywwLDEuNTYsLjE5LDIuMjIsLjU2LC42NSwuMzcsMS4xNiwuOSwxLjUzLDEuNTksLjM3LC42OSwuNTUsMS40NywuNTUsMi4zNHYuNDNjMCwuODctLjE4LDEuNjUtLjU0LDIuMzMtLjM2LC42OC0uODcsMS4yMS0xLjUyLDEuNTgtLjY1LC4zNy0xLjM5LC41Ni0yLjIxLC41N2gtMi45MVptMS45NC03LjgzdjYuMjdoLjk0Yy43NiwwLDEuMzQtLjI1LDEuNzQtLjc0LC40LS41LC42LTEuMiwuNjEtMi4xMnYtLjVjMC0uOTYtLjItMS42OC0uNTktMi4xNy0uNC0uNDktLjk3LS43NC0xLjc0LS43NGgtLjk2WiIvPjwvZz48L2c+PC9nPjwvc3ZnPg=="

var (
	cloudHttpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "timoni_http_request_count",
		Help: "Number of http requests",
	},
		[]string{"path", "user"},
	)
)

func addSecurityHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !strings.HasPrefix(r.URL.Path, "/grafana") {

			w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline'; img-src * 'self' data:;")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Strict-Transport-Security", "max-age=15552000") // 180 days
			// w.Header().Set("Strict-Transport-Security", "max-age=15552000; includeSubDomains; preload") // 180 days
			// w.Header().Set("Cache-control", "no-store") // to wylaczy calkowicie cache w przegladarce, czego nie chcemy
			// w.Header().Set("Expect-CT", "")
		}

		h.ServeHTTP(w, r)
	})
}

func GetRouter() *mux.Router {
	router := mux.NewRouter()
	router.Use(addSecurityHeaders)

	// Metrics counters
	prometheus.MustRegister(cloudHttpRequests)

	router.HandleFunc("/cli", term.Xterm)
	router.HandleFunc("/term", term.Socket)
	router.PathPrefix("/term/").Handler(http.FileServer(http.Dir(config.WebPublicPath())))

	imageregistry.Start(router)

	router.HandleFunc("/api/git-on-board", apiGitOnBoard)

	router.HandleFunc("/api/check-http", apiCheckHTTP)
	router.HandleFunc("/api/check-all", apiCheckAll)

	router.HandleFunc("/api/system-version", apiSystemVersion)
	router.Handle("/api/system-info", apiMiddleware(apiSystemInfo))
	router.Handle("/api/system-print-stack", apiMiddleware(nil))

	router.Handle("/api/system-resources", apiMiddleware(apiResources))
	router.Handle("/api/system-pod-cache", apiMiddleware(apiPodCache))
	router.Handle("/api/system-element-versions", apiMiddleware(apiAllElementVersionsMatrix))

	router.HandleFunc("/api/user-login", apiUserLogin)
	router.Handle("/api/user-invite", apiMiddleware(apiUserInvite))
	router.Handle("/api/user-invite-qr", apiMiddleware(apiUserInviteQR))
	router.Handle("/api/user-list", apiMiddleware(apiUserList))
	router.Handle("/api/user-info", apiMiddleware(apiUserInfo))
	router.Handle("/api/user-update", apiMiddleware(apiUserUpdate))
	router.Handle("/api/user-theme", apiMiddleware(apiUserTheme))
	router.Handle("/api/user-notification-update", apiMiddleware(userNotificationOnOff))
	router.Handle("/api/user-auto-logout-update", apiMiddleware(apiUserAutoLogoutUpdate))

	router.Handle("/api/team-create", apiMiddleware(apiTeamCreate))
	router.Handle("/api/team-delete", apiMiddleware(apiTeamDelete))
	router.Handle("/api/team-list", apiMiddleware(apiTeamList))
	router.Handle("/api/team-info", apiMiddleware(apiTeamInfo))
	router.Handle("/api/team-update", apiMiddleware(apiTeamUpdate))
	router.Handle("/api/team-perms-set", apiMiddleware(apiTeamSetPerms))
	router.Handle("/api/team-user-remove", apiMiddleware(apiTeamRemoveUser))
	router.Handle("/api/team-user-add", apiMiddleware(apiTeamAddUser))

	router.Handle("/api/perms-list", apiMiddleware(apiPermissionList))
	router.Handle("/api/gitops-repo-map", apiMiddleware(apiGitOpsRepoMap))

	router.Handle("/api/git-repo-create", apiMiddleware(apiGitRepoCreate))
	router.Handle("/api/git-repo-create-remote", apiMiddleware(apiProjectCreateRemote))
	router.Handle("/api/git-repo-get-remote-branches", apiMiddleware(apiProjectGetRemoteBranches))
	router.Handle("/api/git-repo-delete", apiMiddleware(apiGitRepoDelete))
	router.Handle("/api/git-repo-info", apiMiddleware(apiGitRepoInfo))
	router.Handle("/api/git-repo-update", apiMiddleware(apiProjectUpdate))
	router.Handle("/api/git-repo-file-list", apiMiddleware(apiProjectFiles))
	router.Handle("/api/git-repo-file-open", apiMiddleware(apiProjectFileOpen))
	router.Handle("/api/git-repo-branch-list", apiMiddleware(apiProjectBranchList))
	router.Handle("/api/git-repo-tag-list", apiMiddleware(apiProjectTagList))
	router.Handle("/api/git-repo-get-last-commit", apiMiddleware(apiGetLastCommit))
	router.Handle("/api/git-repo-commit-list", apiMiddleware(apiProjectCommitList))
	router.Handle("/api/git-repo-commit-info-file-list", apiMiddleware(apiProjectCommitInfoFileList))
	router.Handle("/api/git-repo-commit-info-file-diff", apiMiddleware(apiProjectCommitInfoFileDiff))
	router.Handle("/api/git-repo-branches-info-file-list", apiMiddleware(apiProjectBranchesInfoFileList))
	router.Handle("/api/git-repo-branches-info-file-diff", apiMiddleware(apiProjectBranchesInfoFileDiff))
	router.Handle("/api/git-repo-branches-merge", apiMiddleware(apiProjectBranchesMerge))
	router.Handle("/api/git-repo-branches-compare", apiMiddleware(apiProjectBranchesCompare))
	router.Handle("/api/git-repo-element-list", apiMiddleware(apiGitRepoElementList))
	router.Handle("/api/git-repo-env-map", apiMiddleware(apiGitRepoEnvMap))
	router.Handle("/api/git-repo-access-info", apiMiddleware(apiProjectAccessControlInfo))
	router.Handle("/api/git-repo-map", apiMiddleware(apiGitRepoMap))
	router.Handle("/api/git-repo-remote-access-update", apiMiddleware(apiProjectRemoteAccessUpdate))
	router.Handle("/api/git-repo-change-limit", apiMiddleware(apiProjectChangeLimit))
	router.Handle("/api/git-repo-notification-update", apiMiddleware(apiProjectNotificationUpdate))
	router.Handle("/api/git-repo-notification-info", apiMiddleware(apiProjectNotificationInfo))

	router.HandleFunc("/git/{repoName}/info/refs", apiGetInfoRefs)
	router.HandleFunc("/git/{repoName}/git-upload-pack", apiGitServiceUploadPack)   // git pull
	router.HandleFunc("/git/{repoName}/git-receive-pack", apiGitServiceReceivePack) // git push
	router.PathPrefix("/git/").HandlerFunc(apiGetGit)

	router.Handle("/api/git-elements-list", apiMiddleware(apiGitElementsList))

	router.Handle("/api/env-element-scale", apiMiddleware(apiEnvironmentElementScale))
	router.Handle("/api/env-element-static-scale", apiMiddleware(apiEnvironmentElementStaticScaling))

	router.Handle("/api/env-create", apiMiddleware(apiEnvironmentCreate2))
	router.Handle("/api/env-map", apiMiddleware(apiEnvironmentShortMap2))
	router.Handle("/api/env-delete", apiMiddleware(apiEnvironmentDelete2))
	router.Handle("/api/env-info", apiMiddleware(apiEnvironmentInfo))
	router.Handle("/api/env-clone", apiMiddleware(apiEnvironmentClone))
	router.Handle("/api/env-tag-create", apiMiddleware(apiEnvironmentCreateTag))
	router.Handle("/api/env-tag-delete", apiMiddleware(apiEnvironmentDeleteTag))
	router.Handle("/api/env-variables", apiMiddleware(apiEnvironmentVariables))
	router.Handle("/api/env-variable-get-secret", apiMiddleware(apiEnvironmentVariableGetSecret))
	router.Handle("/api/env-pods", apiMiddleware(apiEnvironmentPods))
	router.Handle("/api/env-rename", apiMiddleware(apiEnvironmentRename))
	router.Handle("/api/env-schedule-set", apiMiddleware(apiEnvironmentSchedulerSet))
	router.Handle("/api/env-gitops-set", apiMiddleware(apiEnvironmentGitOpsSet))
	router.Handle("/api/env-domain-targets", apiMiddleware(apiEnvironmentDomainTargets))
	// router.Handle("/api/env-team-add", apiMiddleware(apiEnvironmentTeamAdd))
	// router.Handle("/api/env-team-remove", apiMiddleware(apiEnvironmentTeamRemove))
	router.Handle("/api/env-dynamic-sources-map", apiMiddleware(apiEnvironmentDynamicSourcesMap))
	router.Handle("/api/env-dynamic-sources-add", apiMiddleware(apiEnvironmentDynamicSourcesAdd))
	router.Handle("/api/env-dynamic-sources-delete", apiMiddleware(apiEnvironmentDynamicSourcesDelete))

	router.Handle("/api/env-element-create-from-git", apiMiddleware(apiEnvironmentElementCreateFromGitRepo))
	router.Handle("/api/env-element-create-from-toml", apiMiddleware(apiEnvironmentElementCreateFromTOML))
	router.Handle("/api/env-element-update-from-toml", apiMiddleware(apiEnvironmentElementUpdateFromTOML))
	router.Handle("/api/env-element-map", apiMiddleware(apiEnvironmentElementMap))
	router.Handle("/api/env-element-versions", apiMiddleware(apiEnvironmentElementVersionMap))
	router.Handle("/api/env-element-version-change", apiMiddleware(apiEnvironmentElementVersionChange))
	router.Handle("/api/env-element-commit-list", apiMiddleware(apiEnvironmentElementCommitList))
	router.Handle("/api/env-element-docker-file", apiMiddleware(apiEnvironmentElementDockerFile))
	router.Handle("/api/env-element-restart-pods", apiMiddleware(apiEnvironmentElementRestart))
	router.Handle("/api/env-element-delete", apiMiddleware(apiEnvironmentElementDelete))
	router.Handle("/api/env-element-update-mode-set", apiMiddleware(apiEnvironmentElementUpdateModeSet))
	router.Handle("/api/env-element-run-control", apiMiddleware(apiEnvironmentElementRunControl))
	router.Handle("/api/env-export-toml", apiMiddleware(apiEnvironmentExportTOML))

	router.Handle("/api/env-pod-restart", apiMiddleware(apiEnvironmentPodRestart))

	router.Handle("/api/env-element-actions-run", apiMiddleware(apiEnvironmentElementActionsRun))
	router.HandleFunc("/api/entry-point-actions-status", apiActionStatus)

	router.Handle("/api/image-rebuild", apiMiddleware(apiImageRebuild))
	router.Handle("/api/image-list", apiMiddleware(apiImageList))
	router.Handle("/api/image-status-update", apiMiddleware(apiImageUpdateStatus))
	router.Handle("/api/image-external-save", apiMiddleware(apiImageExternalSave))
	router.Handle("/api/image-external-load", apiMiddleware(apiImageExternalLoad))
	router.HandleFunc("/api/image-delete-unused", apiImageDeleteUnused)

	router.PathPrefix("/setup/").HandlerFunc(userSetup)

	router.PathPrefix("/files/").Handler(http.FileServer(http.Dir(config.WebPublicPath())))
	router.PathPrefix("/assets/").Handler(http.FileServer(http.Dir(config.WebPublicPath())))

	// -------------------------

	router.PathPrefix("/logo.svg").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		buf, _ := base64.StdEncoding.DecodeString(timoniLogoSvgBase64)

		w.Header().Add("Content-Type", "image/svg+xml")
		w.Write(buf)
	})
	router.PathPrefix("/vite.svg").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		buf, _ := base64.StdEncoding.DecodeString(timoniLogoSvgBase64)

		w.Header().Add("Content-Type", "image/svg+xml")
		w.Write(buf)
	})

	// -------------------------

	proxyJP := httputil.NewSingleHostReverseProxy(utils.Must(url.Parse("http://journal-proxy-1.timoni:4003")))
	proxyJP.FlushInterval = -1
	router.PathPrefix("/j1/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/j1")
		proxyJP.ServeHTTP(w, r)
	})

	// -------------------------

	proxyGrafana := httputil.NewSingleHostReverseProxy(utils.Must(url.Parse("http://metrics-grafana.timoni:3000")))
	proxyGrafana.FlushInterval = -1
	router.PathPrefix("/grafana/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyGrafana.ServeHTTP(w, r)
	})

	// -------------------------

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-control", "no-store")
		http.ServeFile(w, r, filepath.Join(config.WebPublicPath(), "index.html"))
	})

	return router
}

func Loop() {
	log.Debug("api.loop()")

	domain := db2.TheDomain
	cert := domain.Cert()
	certLeftDays := int(math.Ceil(time.Since(time.Unix(cert.ExpirationTime(), 0)).Hours()/24)) * -1

	if certLeftDays <= 0 {
		fmt.Println("certLeftDays=", certLeftDays, cert.ID(), domain.Name())
		os.Exit(1)
	}

	certKeyFilePath := filepath.Join(config.DataPath(), "cert.key")
	certPemFilePath := filepath.Join(config.DataPath(), "cert.pem")

	os.WriteFile(certKeyFilePath, []byte(cert.Key()), 0644)
	os.WriteFile(certPemFilePath, []byte(cert.Pem()), 0644)

	// ---

	log.Fatal(http.ListenAndServeTLS(
		fmt.Sprintf(":%d", domain.Port()),
		certPemFilePath,
		certKeyFilePath,
		GetRouter(),
	))

}

func pemExpiry(crtData []byte) time.Time {
	block, rest := pem.Decode(crtData)
	if block == nil {
		fmt.Println("ERROR: crtData empty")
		return time.Time{}
	}

	for ; block != nil; block, rest = pem.Decode(rest) {
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if out.New(err).NotValid() {
			continue
		}

		if len(cert.DNSNames) == 0 {
			continue
		}

		return cert.NotAfter
	}
	return time.Time{}
}
