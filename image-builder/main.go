package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"lib/tlog"
	"lib/utils/cmd"
	"image-builder/global"
)

// Helpful links:
// https://docs.docker.com/registry/spec/api/
// https://github.com/google/go-containerregistry

var (
	building   = false
	buildQueue = make(chan *ImageBuildBlueprintS)
)

func main() {

	setupGitAccess()
	setupDockerAccess()
	startDocker()
	buildEntryPointImage()

	go builderLoop()

	http.HandleFunc("/build", httpBuild)
	http.HandleFunc("/status", httpStatus)

	tlog.Info("ImageBuilder started on port :6666")
	tlog.Error(http.ListenAndServe(":6666", nil))
}

func setupDockerAccess() {
	// Login to image registry
	os.MkdirAll("/tmp/.docker/", 0644)

	config := map[string]map[string]map[string]string{
		"auths": {
			global.TimoniDomainAndPort(): {
				"auth": base64.StdEncoding.EncodeToString(
					[]byte(fmt.Sprintf("%s:%s", "ImageBuilder", global.Token)),
				),
			},
		},
	}

	d, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile("/tmp/.docker/global.json", d, 0644)
}

func setupGitAccess() {
	cookie := fmt.Sprintf("Set-Cookie:token=%s; Path=/; Domain=%s\n", global.Token, global.TimoniDomain)
	tlog.Fatal(os.WriteFile("/tmp/cookie", []byte(cookie), 0644))
	tlog.Fatal(os.WriteFile("/tmp/.gitconfig", []byte("[http]\n\tcookieFile = /tmp/cookie"), 0644))

	if global.TimoniIP != "" {
		_, err := exec.Command("sh", "-c", "echo '"+global.TimoniIP+"  "+global.TimoniDomain+"' >> /etc/hosts").CombinedOutput()
		tlog.Fatal(err)
		time.Sleep(2 * time.Second)
	}
}

func startDocker() {

	tlog.Info("Waiting for Docker...")

	go func() {
		cmd.NewCommand("dockerd", "--iptables=false").Run(&cmd.RunOptions{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		})
	}()

	for {
		err := cmd.NewCommand("docker", "-D", "ps").Run(&cmd.RunOptions{
			// Stdout: os.Stdout,
			// Stderr: os.Stderr,
		})
		tlog.Error(err)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}

	tlog.Info("Docker is ready")
}

func buildEntryPointImage() {

	// # RUN echo "until docker pull {{domain}}/entry-point:{{commitSHA}}; do echo 'Retrying pull'; sleep 2; done" >> /init.sh
	// # RUN echo "docker tag {{domain}}/entry-point:{{commitSHA}} ep" >> /init.sh
	// # COPY --from={{domain}}/entry-point:{{commitSHA}} /bin/ep /bin/ep

	// ---

	tlog.Error(os.Chdir("/bin"))
	os.WriteFile("/bin/.dockerignore", []byte("*\n!ep\n"), 0644)
	os.WriteFile("/bin/entry-point-dockerfile", []byte("FROM scratch\nCOPY ep /bin/ep\n"), 0644)

	err := cmd.NewCommand(
		"docker", "build",
		"--network", "host",
		"-t", "ep",
		"-f", "entry-point-dockerfile",
		"/bin",
	).Run(&cmd.RunOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	tlog.Error(err)
}
