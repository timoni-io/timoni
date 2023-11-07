package main

import (
	"core/api"
	"core/db"
	"core/db2"
	"core/gitprovider"
	"core/imagebuilder"
	"core/imageregistry"
	"core/ingress"
	"core/journal"
	"core/kube"
	"core/kube/kubesync"
	"core/metrics"
	"core/modulestate"
	"os"
	"syscall"
)

func main() {
	os.Setenv("LOG_MODE", "mjson")
	os.Setenv("TZ", "UTC")
	os.Setenv("GIT_TERMINAL_PROMPT", "0")
	os.Setenv("GIT_SSH_COMMAND", "ssh -i /data/ssh_key_global -o IdentitiesOnly=yes")

	defer db.PanicHandler()

	// -----------------------------------------------------------------------------

	setResolvConf()

	db.Open()  // old local db
	db2.Open() // new local db
	go modulestate.Loop()

	// main modules
	kube.Setup()    // kube client
	api.Setup()     // api and webui
	journal.Setup() // journal-proxy

	// other modules
	gitprovider.Setup()
	go ingress.Setup()
	go imageregistry.Setup()
	go imagebuilder.Setup()

	// certprovider.Setup()
	go metrics.Setup()

	kubesync.Loop()
}

func setResolvConf() {
	if syscall.Geteuid() != 0 {
		return
	}

	os.WriteFile("/etc/resolv.conf", []byte("nameserver 8.8.8.8\n"), 0644)
}
