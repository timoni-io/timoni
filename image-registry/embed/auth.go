package embed

import (
	log "lib/tlog"
	"net/http"
)

func checkAuth(r *http.Request) bool {
	userID, _, ok := r.BasicAuth()
	if !ok {
		return false
	}

	// TODO: check user
	_ = userID

	// If user not found
	// if !db.Users().Exists(userID) {
	// 	log.Infof("User %s not found", userID)
	// 	return false
	// }

	return true
}

// Basic Auth middleware for registry
func authHandler(registry http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, exist := r.Header["Authorization"]; !exist {
			// Unauthorized
			w.Header()["WWW-Authenticate"] = []string{`Basic realm="Registry Authorization"`}

			// Headers
			w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
			w.WriteHeader(http.StatusUnauthorized)

			w.Write([]byte("Unauthorized"))
			log.Error("Unauthorized", r)
			return
		}

		// Auth with BasicAuth and return user
		if !checkAuth(r) {
			// Headers
			w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
			w.WriteHeader(http.StatusUnauthorized)

			w.Write([]byte("Unauthorized"))
			log.Error("Unauthorized", r)
			return
		}

		// [ ]: Check permissions

		// Pull:
		// r.RequestURI -> /v2/projectName/manifests && Method:HEAD/GET
		// /v2/projectName/blobs && Method: GET

		// Push:
		// r.RequestURI -> /v2/projectName/blobs/ && Method:HEAD
		// r.RequestURI -> /v2/projectName/blobs/uploads && Method:PUT/POST/PATCH
		// r.RequestURI -> /v2/projectName/manifests && Method:PUT

		// Serve Registry
		registry.ServeHTTP(w, r)
	})
}
