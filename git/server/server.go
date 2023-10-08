package gitserver

import (
	"encoding/json"
	"errors"
	"fmt"
	gitc "git/client"
	log "lib/tlog"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

type gitServer struct {
	Git gitc.Git
}

func Start(r *mux.Router, dataPath string) {
	log.Info("Starting unsafe git server @ /git")
	dataPath, err := filepath.Abs(dataPath)
	if err != nil {
		log.Error(err)
	}

	// Register server
	r.PathPrefix("/git").Handler(http.StripPrefix("/git/", gitServer{
		Git: gitc.Git{
			Dir: dataPath,
		},
	}))
}

func (gs gitServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	repo, mode := parsePath(r)
	gs.Git.Repo = repo
	gs.Serve(w, r, mode)
}

func (gs *gitServer) Serve(w http.ResponseWriter, r *http.Request, mode Mode) {
	switch mode {
	case Invalid:
		w.WriteHeader(http.StatusBadRequest)

	// Git server control, like new repo, commit list ...
	case Control:
		gs.control(w, r)

	// Basic git server actions
	case UploadPack:
		gs.uploadPack(w, r)
	case ReceivePack:
		gs.receivePack(w, r)
	default:
		gs.service(w, r, mode.Service)
	}
}

// URL: /{repo_name}/control
func (gs *gitServer) control(w http.ResponseWriter, r *http.Request) {
	ctrl := ControlA{Git: gs.Git}
	err := json.NewDecoder(r.Body).Decode(&ctrl.Req)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctrl.ServeHTTP(w, r)
}

// URL: /{repo_name}/info/refs?service={service}
func (gs *gitServer) service(w http.ResponseWriter, r *http.Request, service string) {
	contentType := fmt.Sprintf("application/x-git-%s-advertisement", service)

	data, err := gs.Git.Service(service)
	if err != nil {
		switch {
		case errors.Is(err, gitc.ErrPathNotExists):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// OK
	w.Header().Add("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(
		append(
			servicePacket(service),
			data...,
		),
	)
}

// URL: /{repo_name}/git-upload-pack
func (gs *gitServer) uploadPack(w http.ResponseWriter, r *http.Request) {
	const contentType = "application/x-git-upload-pack-result"

	data, err := gs.Git.UploadPack(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// OK
	w.Header().Add("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// URL: /{repo_name}/git-receive-pack
func (gs *gitServer) receivePack(w http.ResponseWriter, r *http.Request) {
	const contentType = "application/x-git-receive-pack-result"

	data, err := gs.Git.ReceivePack(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// OK
	w.Header().Add("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
