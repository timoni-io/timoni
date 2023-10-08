package imageregistry

import (
	"core/db2"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func loadK3SAuthData() bool {
	// Load k3s registries

	type K3SRegistriesS struct {
		Mirrors map[string]map[string][]string          `yaml:"mirrors"`
		Configs map[string]map[string]map[string]string `yaml:"configs"`
	}

	k3sRegs := &K3SRegistriesS{}
	k3sReg, err := os.Open("/etc/rancher/k3s/registries.yaml")
	if err != nil {
		log.Warn("error loading k3s registries: ", err)
		return false
	}
	defer k3sReg.Close()

	err = yaml.NewDecoder(k3sReg).Decode(k3sRegs)
	if err != nil {
		log.Error("error decoding registries file: ", err)
		return false
	}

	// Save Username && password
	if conf, exist := k3sRegs.Configs[db2.TheDomain.Name()]; exist {
		if auth, exist := conf["auth"]; exist {
			k3sUsername = auth["username"]
			k3sPassword = auth["password"]
		}
	}

	return true
}

func checkAuth(r *http.Request) (success bool) {
	username, password, ok := r.BasicAuth()
	if !ok {
		log.Info("Basic Auth failed")
		return false
	}

	if username != "ImageBuilder" {
		log.Error("Invalid username")
		return false
	}

	if password != db2.TheImageBuilder.Timoni_Token() {
		log.Error("Incorrect password")
		return false
	}

	return true
}

// Basic Auth middleware for registry
func authHandler(registry http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if k3sUsername == "" {
			// Serve Registry
			registry.ServeHTTP(w, r)
			return
		}

		if _, exist := r.Header["Authorization"]; !exist {
			// Not Authorized
			w.Header()["WWW-Authenticate"] = []string{`Basic realm="Registry Authorization"`}

			// Headers
			w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
			w.WriteHeader(http.StatusUnauthorized)

			w.Write([]byte("Not Authorized"))
			return
		}

		if !checkAuth(r) {
			w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
			w.WriteHeader(http.StatusUnauthorized)

			w.Write([]byte("Not Authorized"))
			return
		}

		registry.ServeHTTP(w, r)
	})
}
