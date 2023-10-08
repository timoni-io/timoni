package gitserver

import (
	"encoding/json"
	gitc "git/client"
	git "git/structs"
	log "lib/tlog"
	"net/http"
	"os"
)

type ControlA struct {
	Git gitc.Git
	Req git.Control
}

func (a ControlA) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch a.Req.Action {
	case git.New:
		a.New(w, r)
	case git.Delete:
		a.Delete(w, r)
	case git.Branches:
		a.Branches(w, r)
	case git.Log:
		a.Log(w, r)
	case git.Diff:
		a.Diff(w, r)
	case git.Files:
		a.Files(w, r)
	case git.Open:
		a.OpenFile(w, r)
	}
}

func (a ControlA) New(w http.ResponseWriter, r *http.Request) {
	err := a.Git.InitBare()
	if err != nil {
		log.Error(err)
		RespErr(w, err)
		return
	}

	RespOk(w, "Repo created")
}

func (a ControlA) Delete(w http.ResponseWriter, r *http.Request) {
	err := os.RemoveAll(a.Git.Path())
	if err != nil {
		log.Error(err)
		RespErr(w, err)
		return
	}

	RespOk(w, "Deleted repo")
}

func (a ControlA) Branches(w http.ResponseWriter, r *http.Request) {
	branches, err := a.Git.Branches()
	if err != nil {
		log.Error(err)
		RespErr(w, err)
		return
	}
	RespOk(w, branches)
}

func (a ControlA) Log(w http.ResponseWriter, r *http.Request) {
	logs, err := a.Git.Log(a.Req.Branch)
	if err != nil {
		log.Error(err)
		RespErr(w, err)
		return
	}
	RespOk(w, logs)
}

func (a ControlA) Diff(w http.ResponseWriter, r *http.Request) {
	diff, err := a.Git.Diff(a.Req.Diff)
	if err != nil {
		log.Error(err)
		RespErr(w, err)
		return
	}
	RespOk(w, diff)
}

func (a ControlA) Files(w http.ResponseWriter, r *http.Request) {
	files, err := a.Git.Files(a.Req.Commit)
	if err != nil {
		log.Error(err)
		RespErr(w, err)
		return
	}
	RespOk(w, files)
}

func (a ControlA) OpenFile(w http.ResponseWriter, r *http.Request) {
	file, err := a.Git.OpenFile(a.Req.Commit, a.Req.Filename)
	if err != nil {
		log.Error(err)
		RespErr(w, err)
		return
	}
	RespOk(w, file)
}

func RespErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func RespOk(w http.ResponseWriter, resp any) {
	data, err := json.Marshal(resp)
	if err != nil {
		RespErr(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
