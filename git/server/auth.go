package gitserver

import (
	gitc "git/client"
	log "lib/tlog"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

type GitAuthFunc func(http.ResponseWriter, *http.Request, Mode) bool

type gitServerAuth struct {
	gitServer
	auth GitAuthFunc
}

func StartAuth(r *mux.Router, dataPath string, auth GitAuthFunc) {
	log.Info("Starting git server @ /git")
	dataPath, err := filepath.Abs(dataPath)
	if err != nil {
		log.Error(err)
	}

	server := &gitServerAuth{
		gitServer: gitServer{
			Git: gitc.Git{
				Dir: dataPath,
			},
		},
		auth: auth,
	}

	// Register server
	r.PathPrefix("/git").Handler(http.StripPrefix("/git/", server))
}

func (gs gitServerAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	repo, mode := parsePath(r)

	if !gs.auth(w, r, mode) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	gs.Git.Repo = repo
	gs.Serve(w, r, mode)
}
