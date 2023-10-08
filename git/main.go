package main

import (
	gitserver "git/server"
	log "lib/tlog"
	"net/http"

	"github.com/gorilla/mux"
)

// Standalone git server
func main() {
	r := mux.NewRouter()
	// gitserver.StartAuth(r, "./data/git", basicAuthTest)

	gitserver.Start(r, "./data/git") // Without auth

	log.Fatal(http.ListenAndServe(":7101", r))
}

// func basicAuthTest(w http.ResponseWriter, r *http.Request, mode gitserver.Mode) bool {
// 	// enable basic auth
// 	if _, exist := r.Header["Authorization"]; !exist {
// 		w.Header().Add("WWW-Authenticate", `Basic realm="Timoni Git Server Authorization"`)
// 		return false
// 	}

// 	user, pass, ok := r.BasicAuth()
// 	if !ok {
// 		log.Error("Non basic auth")
// 		return false
// 	}

// 	log.Debug(mode, mode.Action())

// 	if user == "rw" {
// 		switch mode.Action() {
// 		case gitserver.Pull:
// 			return pass == "read"
// 		case gitserver.Push:
// 			return pass == "write"
// 		}
// 	}
// 	return user == "test" && pass == "test"
// }
