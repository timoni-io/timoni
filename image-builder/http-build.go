package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"image-builder/global"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "lib/tlog"
	"lib/utils"

	"github.com/hako/durafmt"
)

type ImageBuildBlueprintS struct {
	ImageID           string
	BuildID           string
	BuildRootPath     string // ścieżka do katalogu w którym wykonuje build
	SourceGit         SourceGitS
	DockerFilePath    string // ścieżka do pliku DockerFile
	DockerFileContent string
	ImageBuilderID    string
	Status            string
	EntryPointGitTag  string
}

type SourceGitS struct {
	RepoName   string `toml:"git-repo-name"`
	BranchName string `toml:"branch"`
	CommitHash string `toml:"commit"`
	FilePath   string `toml:"file-path"`
}

func (bp *ImageBuildBlueprintS) RunCMD(args ...string) bool {

	logVars := log.Vars{
		"imageID": bp.ImageID,
		"buildID": bp.BuildID,
		"gitRepo": bp.SourceGit.RepoName,
		"cmd":     strings.Join(args, " "),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	sErr, err := cmd.StderrPipe()
	if log.Error(err, logVars) != nil {
		return false
	}
	defer sErr.Close()

	sOut, err := cmd.StdoutPipe()
	if log.Error(err, logVars) != nil {
		return false
	}
	defer sOut.Close()

	go pipeReader(sErr, "INFO", logVars)
	go pipeReader(sOut, "INFO", logVars)

	cmd.Env = []string{
		"HOME=/tmp",
	}

	// For docker enable buildkit
	if args[0] == "docker" {
		cmd.Env = append(cmd.Env, "DOCKER_BUILDKIT=1")
	}

	err = cmd.Start()
	if log.Error(err, logVars) != nil {
		return false
	}

	cmd.Wait()

	if ctx.Err() == context.DeadlineExceeded {
		log.Error("Command timed out", logVars)
		return false
	}

	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		logVars["exitCode"] = exitCode
		log.Error("exitCode != 0", logVars)
		return false
	}

	return true
}

func httpBuild(w http.ResponseWriter, r *http.Request) {

	if building {
		w.Write([]byte("Image Builder isn't ready"))
		return
	}

	buf, err := io.ReadAll(r.Body)
	if log.Error(err) != nil {
		w.Write([]byte(err.Error()))
		return
	}

	bp := &ImageBuildBlueprintS{}
	err = json.Unmarshal(buf, bp)
	if log.Error(err) != nil {
		w.Write([]byte(err.Error()))
		return
	}

	if bp.ImageID == "" {
		w.Write([]byte("ImageID is required"))
		return
	}

	if bp.ImageBuilderID == "" {
		w.Write([]byte("ImageBuilderID is required"))
		return
	}

	if bp.BuildID == "" {
		w.Write([]byte("BuildID is required"))
		return
	}

	// Start build
	building = true
	buildQueue <- bp

	w.Write([]byte("ok"))
}

// ------------------------------------------------------------------------

func builderLoop() {
	for bp := range buildQueue {
		log.Info("Image build starting...", bp.ImageID)
		start := time.Now()

		if bp.SourceGit.RepoName == "" {
			buildWithoutGit(bp)
		} else {
			build(bp)
		}

		body, err := json.Marshal(bp)
		if log.Error(err) != nil {
			continue
		}
		for i := 0; ; i++ {

			if i > 9 {
				// failed to post update
				log.Error("send build status to api fail 10 times", log.Vars{
					"imageID": bp.ImageID,
					"buildID": bp.BuildID,
					"gitRepo": bp.SourceGit.RepoName,
				})
				bp.Status = "fail"
				break
			}
			req, _ := http.NewRequest("POST", global.UpdateURL, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Session", global.Token)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				// failed to post update
				log.Error(err, log.Vars{
					"imageID": bp.ImageID,
					"buildID": bp.BuildID,
					"gitRepo": bp.SourceGit.RepoName,
				})
				time.Sleep(5 * time.Second)
				continue
			}

			if res.StatusCode != 200 {
				// failed to post update
				log.Error(global.UpdateURL+" return wrong http status code {{code}}", log.Vars{
					"imageID": bp.ImageID,
					"buildID": bp.BuildID,
					"gitRepo": bp.SourceGit.RepoName,
					"code":    res.StatusCode,
				})
				time.Sleep(5 * time.Second)
				continue
			}

			body, err := io.ReadAll(res.Body)
			log.Error(err, log.Vars{
				"imageID": bp.ImageID,
				"buildID": bp.BuildID,
				"gitRepo": bp.SourceGit.RepoName,
			})
			res.Body.Close()

			if len(body) == 0 {
				// failed to post update
				log.Error(global.UpdateURL+" return wrong empty body", log.Vars{
					"imageID": bp.ImageID,
					"buildID": bp.BuildID,
					"gitRepo": bp.SourceGit.RepoName,
				})
				time.Sleep(5 * time.Second)
				continue
			}

			if body[0] != 1 {
				// failed to post update
				log.Error(global.UpdateURL+" return wrong response body first byte not 1", log.Vars{
					"imageID": bp.ImageID,
					"buildID": bp.BuildID,
					"gitRepo": bp.SourceGit.RepoName,
					"body":    string(body),
				})
				time.Sleep(5 * time.Second)
				continue
			}

			if string(body[1:]) != `"ok"` {
				// failed to post update
				log.Error(global.UpdateURL+" return wrong response body {{body}}", log.Vars{
					"imageID": bp.ImageID,
					"buildID": bp.BuildID,
					"gitRepo": bp.SourceGit.RepoName,
					"body":    string(body[1:]),
				})
				time.Sleep(5 * time.Second)
				continue
			}

			break
		}

		// Set build flag
		building = false
		duration := durafmt.Parse(time.Since(start)).LimitFirstN(2).String()
		if time.Since(start) < time.Second {
			duration = durafmt.Parse(time.Since(start)).LimitFirstN(1).String()
		}

		if bp.Status == "ready" {
			log.Info("image build success, took {{duration}}", log.Vars{
				"imageID":  bp.ImageID,
				"buildID":  bp.BuildID,
				"gitRepo":  bp.SourceGit.RepoName,
				"status":   bp.Status,
				"duration": duration,
			})
		} else {
			log.Error("image build fail, took {{duration}}", log.Vars{
				"imageID":  bp.ImageID,
				"buildID":  bp.BuildID,
				"gitRepo":  bp.SourceGit.RepoName,
				"status":   bp.Status,
				"duration": duration,
			})
		}
	}
}

// ------------------------------------------------------------------------

func build(bp *ImageBuildBlueprintS) {
	gitPath := filepath.Join("/tmp", bp.SourceGit.RepoName)
	bp.EntryPointGitTag = global.GitTag

	// Clone git repo
	if bp.ImageID != "test" {
		if _, err := os.Stat(gitPath); os.IsNotExist(err) {
			if !bp.RunCMD("git", "clone", global.TimoniURL()+"/git/"+bp.SourceGit.RepoName, gitPath) {
				bp.Status = "fail"
				return
			}
		} else {
			if !bp.RunCMD("git", "-C", gitPath, "fetch", "--all") {
				bp.Status = "fail"
				return
			}
		}

		if !bp.RunCMD("git", "-C", gitPath, "checkout", bp.SourceGit.BranchName) {
			bp.Status = "fail"
			return
		}
		if !bp.RunCMD("git", "-C", gitPath, "reset", "--hard", bp.SourceGit.CommitHash) {
			bp.Status = "fail"
			return
		}
	} else {
		// Create dir for test
		os.MkdirAll(gitPath, 0655)
	}

	logVars := log.Vars{
		"imageID": bp.ImageID,
		"buildID": bp.BuildID,
		"gitRepo": bp.SourceGit.RepoName,
	}

	buildRootPath := gitPath
	if bp.BuildRootPath != "" {
		bp.BuildRootPath = strings.TrimLeft(bp.BuildRootPath, "/")
		buildRootPath = filepath.Join(gitPath, bp.BuildRootPath)
	}

	if bp.DockerFilePath != "" {
		bp.DockerFilePath = strings.TrimLeft(bp.DockerFilePath, "/")
		buf, err := os.ReadFile(filepath.Join(gitPath, bp.DockerFilePath))
		if log.Error(err, logVars) != nil {
			bp.Status = "fail"
			return
		}
		bp.DockerFileContent = string(buf)
	}

	bp.DockerFileContent += "\nCOPY --from=ep /bin/ep /bin/ep"

	dockerFilePath := filepath.Join(buildRootPath, ".timoni-docker-file")
	err := os.WriteFile(dockerFilePath, []byte(bp.DockerFileContent), 0644)
	if log.Error(err, logVars) != nil {
		bp.Status = "fail"
		return
	}

	numberOfCommitsBuf, err := exec.Command("git", "-C", gitPath, "rev-list", "--count", bp.SourceGit.BranchName, "--").CombinedOutput()
	if log.Error(err, logVars) != nil {
		bp.Status = "fail"
		return
	}

	// Build image with docker/buildkit
	if !bp.RunCMD(
		"docker", "build",
		"--network", "host",
		"-t", global.TimoniDomainAndPort()+"/"+bp.ImageID,
		"-f", dockerFilePath,
		"--build-arg", "gitcomitsha="+bp.SourceGit.CommitHash,
		"--build-arg", "gitbranchname="+bp.SourceGit.BranchName,
		"--build-arg", "gitcommitscount="+strings.TrimSpace(string(numberOfCommitsBuf)),
		buildRootPath,
	) {
		bp.Status = "fail"
		return
	}

	// Push image to image registry
	if !bp.RunCMD("docker", "push", global.TimoniDomainAndPort()+"/"+bp.ImageID) {
		bp.Status = "fail"
		return
	}

	bp.Status = "ready"
}

func buildWithoutGit(bp *ImageBuildBlueprintS) {
	bp.EntryPointGitTag = global.GitTag
	logVars := log.Vars{
		"imageID": bp.ImageID,
		"buildID": bp.BuildID,
	}

	bp.DockerFileContent += "\nCOPY --from=ep /bin/ep /bin/ep"

	temp, err := os.MkdirTemp("/tmp", "build-")
	if err != nil {
		log.Error(err, logVars)
		bp.Status = "fail"
		return
	}

	dockerfilePath := filepath.Join(temp, "Dockerfile")
	err = os.WriteFile(dockerfilePath, []byte(bp.DockerFileContent), 0644)
	if err != nil {
		log.Error(err, logVars)
		bp.Status = "fail"
		return
	}

	// Build image with docker/buildkit
	if !bp.RunCMD(
		"docker", "build",
		"--network", "host",
		"-t", global.TimoniDomainAndPort()+"/"+bp.ImageID,
		"-f", dockerfilePath,
		temp, // context
	) {
		bp.Status = "fail"
		return
	}

	// Push image to image registry
	if !bp.RunCMD("docker", "push", global.TimoniDomainAndPort()+"/"+bp.ImageID) {
		bp.Status = "fail"
		return
	}

	bp.Status = "ready"
}

var dockerBuildKitLineRegex = regexp.MustCompile(`^#([0-9]{1,4}) (?:\[([a-zA-Z0-9._-]+))?`)

func pipeReader(ioIn io.ReadCloser, lvl string, lvars log.Vars) {
	var logLine func(in ...interface{}) *log.RecordS

	switch lvl {
	case "ERROR":
		logLine = log.Error
	default:
		logLine = log.Info
	}

	reader := bufio.NewReader(ioIn)
	stages := map[string]string{}
	for {
		line, _, err := reader.ReadLine()

		if len(line) > 0 {
			line := strings.TrimSpace(string(line))

			// parse docker buildkit step
			matches := dockerBuildKitLineRegex.FindStringSubmatch(line)

			if len(matches) == 0 {
				// line is not a docker buildkit step
				logLine(line, lvars)
				continue
			}

			step := matches[1]

			logVars := *utils.DeepCopy(lvars)
			logVars["step"] = step

			if stageName := matches[2]; stageName != "" {
				stages[step] = stageName
				logVars["stage"] = stageName

			} else if stageName, ok := stages[step]; ok {
				logVars["stage"] = stageName
			}

			logLine(line, logVars)
		}

		if err != nil {
			break
		}
	}

	data := make([]byte, 4096)
	n, _ := reader.Read(data)
	if n > 0 {
		logLine(string(data[:n]), lvars)
	}
}
