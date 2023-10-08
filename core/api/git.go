package api

import (
	"bytes"
	"compress/gzip"
	"core/db"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "lib/tlog"

	"github.com/gorilla/mux"
)

func apiGetGit(w http.ResponseWriter, r *http.Request) {
	if !gitRefsAuthorized(r) {
		w.Header().Add("x-git", "apiGetInfoRefs")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	projectName := mux.Vars(r)["repoName"]
	if projectName == "" {
		projectName = strings.Split(
			strings.Split(r.URL.Path, "/")[2],
			"@",
		)[0]

	}

	out := `<html><head><meta name="go-import" content="cs.syslabit.com/git/` + projectName + ` git https://cs.syslabit.com/git/` + projectName + `" />`
	w.Write([]byte(out))

}

// git repo metadata invoked by all git actions
func apiGetInfoRefs(w http.ResponseWriter, r *http.Request) {
	if !gitRefsAuthorized(r) {
		w.Header().Add("x-git", "apiGetInfoRefs")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	gitRepoName := mux.Vars(r)["repoName"]

	gitRepo := db.GitRepoGetByName(gitRepoName)
	if gitRepo == nil {
		log.Error("git repo not found: " + gitRepoName)
		return
	}
	repoPath := gitRepo.PathGet()

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.Error("dir not exist:", repoPath)
		return
	}

	service := r.FormValue("service")

	if !(service == "git-upload-pack" || service == "git-receive-pack") {
		log.Error("wrong service")
		return
	}

	env := []string{}

	protocol := r.Header.Get("Git-Protocol")
	if protocol != "" {
		env = append(env, "GIT_PROTOCOL="+protocol)
	}

	serviceType := strings.Replace(service, "git-", "", 1)
	cmd := exec.Command("git", serviceType, "--stateless-rpc", "--advertise-refs", repoPath)
	cmd.Env = env
	out, err := cmd.Output()
	if err != nil {
		log.Error(err)
		return
	}

	buf := bytes.Buffer{}

	w.Header().Set("Content-Type", "application/x-"+service+"-advertisement")
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
	w.WriteHeader(http.StatusOK)

	buf.Write(packetWrite("# service=" + service + "\n"))
	buf.Write([]byte("0000"))
	buf.Write(out)

	w.Write(buf.Bytes())
}

// ------------------------------------------------------------------
// pull
func apiGitServiceUploadPack(w http.ResponseWriter, r *http.Request) {
	if gitPullAuthorized(r) != nil {
		w.Header().Add("x-git", "apiGitServiceUploadPack")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	servicePack(w, r, "upload-pack")
}

// push branch creation deletion
func apiGitServiceReceivePack(w http.ResponseWriter, r *http.Request) {
	if gitPushAuthorized(r) != nil {
		w.Header().Add("x-git", "apiGitServiceReceivePack")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	servicePack(w, r, "receive-pack")
}

func servicePack(w http.ResponseWriter, r *http.Request, service string) {

	w.Header().Set("Content-Type", "application/x-git-"+service+"-result")
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
	w.WriteHeader(http.StatusOK)

	var err error
	var reqBody = r.Body

	if r.Header.Get("Content-Encoding") == "gzip" {
		reqBody, err = gzip.NewReader(reqBody)
		if err != nil {
			log.Error("Fail to create gzip reader:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	repoName := mux.Vars(r)["repoName"]
	gitRepo := db.GitRepoGetByName(repoName)
	repoPath := gitRepo.PathGet()

	env := []string{}

	protocol := r.Header.Get("Git-Protocol")
	if protocol != "" {
		env = append(env, "GIT_PROTOCOL="+protocol)
	}

	cmd := exec.Command("git", service, "--stateless-rpc", repoPath)
	cmd.Stdin = reqBody
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	cmd.Env = env
	err = cmd.Run()
	if err != nil {
		log.Error(err)
		return
	}

	// Update branch list
	// "receive-pack" push branch creation deletion
	if service == "receive-pack" {
		gitRepo.Open()
		defer gitRepo.Unlock()
		gitRepo.Save()
	}
}

func gitPushAuthorized(r *http.Request) *log.RecordS {
	return nil
	// token, err := r.Cookie("token")
	// if err != nil {
	// 	return log.Error("git push blocked - token is empty")
	// }

	// user := db.GetUserByGitToken(token.Value)
	// if user == nil {
	// 	return log.Error("git push blocked - user not found")
	// }

	// projectName := mux.Vars(r)["repoName"]
	// gitRepo := db.GitRepoGetByName(projectName)
	// if gitRepo == nil {
	// 	return log.Error("git push blocked - project not found", log.Vars{
	// 		"project": projectName,
	// 		"user":    user.Email,
	// 	})
	// }
	// if gitRepo.RemoteURL != "" {
	// 	return log.Error("git push blocked - project is remote", log.Vars{
	// 		"project": projectName,
	// 		"user":    user.Email,
	// 	})
	// }

	// // --------------

	// if gitRepo.ResourcesUsage.CodeStorage >= gitRepo.ResourcesLimit.CodeStorage {
	// 	return log.Error("git push blocked - max repo size reached", log.Vars{
	// 		"project": projectName,
	// 		"user":    user.Email,
	// 	})
	// }

	// // --------------

	// // PARSING GIT REQUEST for branch name
	// rawRequestSTR := net.ReadBodyFromRequest(r)
	// if rawRequestSTR == "0000" {
	// 	return nil
	// }

	// // log.Debug(rawRequestSTR)
	// requestSplit := strings.Split(rawRequestSTR, " ")

	// if len(requestSplit) < 2 {
	// 	return log.Error("git push blocked - something went wrong", log.Vars{
	// 		"project":      projectName,
	// 		"user":         user.Email,
	// 		"requestSplit": requestSplit,
	// 	})
	// }

	// ref := strings.Split(requestSplit[2][:len(requestSplit[2])-1], "/")
	// var gitBranch string
	// gitBranchSplit := []string{}
	// for i := 2; i < len(ref); i++ {
	// 	gitBranchSplit = append(gitBranchSplit, ref[i])
	// }
	// gitBranch = strings.Join(gitBranchSplit[:], "/")

	// if gitBranch == "" {
	// 	return log.Error("git push blocked - branch not found", log.Vars{
	// 		"project": projectName,
	// 		"user":    user.Email,
	// 		"branch":  gitBranch,
	// 	})
	// }

	// //checking limits of branches count only if branch is new (as deleting and pushing on existing branch can not cause breaching of this limit)

	// lastCommit := gitRepo.GetLastCommit(gitBranch)
	// if lastCommit == "" {
	// 	if gitRepo.ResourcesUsage.CodeBranches >= gitRepo.ResourcesLimit.CodeBranches {
	// 		return log.Error("git push blocked - max branch reached", log.Vars{
	// 			"project":            gitRepo.Name,
	// 			"user":               user.Email,
	// 			"Usage.CodeBranches": gitRepo.ResourcesUsage.CodeBranches,
	// 			"Limit.CodeBranches": gitRepo.ResourcesLimit.CodeBranches,
	// 		})
	// 	}

	// 	db.EventCodeBranchCreated(gitRepo, gitBranch, *user)
	// }

	// db.EventCodeCommit(gitRepo, gitBranch, *user)
	// return nil
}

func gitPullAuthorized(r *http.Request) *log.RecordS {
	return nil

	// projectName := mux.Vars(r)["repoName"]
	// if projectName == "" {
	// 	projectName = strings.Split(
	// 		strings.Split(r.URL.Path, "/")[2],
	// 		"@",
	// 	)[0]
	// }

	// token, err := r.Cookie("token")
	// if err != nil {
	// 	return log.Error("git pull blocked - token is empty", log.Vars{
	// 		"url": r.URL.Path,
	// 	})
	// }
	// user := db.GetUserByGitToken(token.Value)
	// if user == nil {
	// 	return log.Error("git pull blocked - user not found")
	// }
	// project := db.GitRepoGetByName(projectName)
	// if project == nil {
	// 	return log.Error("git pull blocked - project not found", log.Vars{
	// 		"project": projectName,
	// 		"user":    user.Email,
	// 	})

	// }

	// if user.Email == "ImageBuilder" {
	// 	return nil
	// }

	// if project.RemoteURL != "" {
	// 	return log.Error("git pull blocked - project is remote", log.Vars{
	// 		"project": projectName,
	// 		"user":    user.Email,
	// 	})
	// }

	// userRole := project.Operators[user.Email]
	// if userRole == "owner" {
	// 	return nil
	// }
	// if userRole == "no-access" {
	// 	return log.Error("git pull blocked - you have no-access", log.Vars{
	// 		"project": projectName,
	// 		"user":    user.Email,
	// 	})
	// }
	// return nil
}

func packetWrite(str string) []byte {
	s := strconv.FormatInt(int64(len(str)+4), 16)
	if len(s)%4 != 0 {
		s = strings.Repeat("0", 4-len(s)%4) + s
	}
	return []byte(s + str)
}

func gitRefsAuthorized(r *http.Request) bool {
	return true

	// projectName := mux.Vars(r)["repoName"]
	// if projectName == "" {
	// 	projectName = strings.Split(
	// 		strings.Split(r.URL.Path, "/")[2],
	// 		"@",
	// 	)[0]
	// }

	// token, err := r.Cookie("token")
	// if err != nil {
	// 	log.Error("token error", log.Vars{
	// 		"token": token,
	// 	})
	// 	return false
	// }
	// user := db.GetUserByGitToken(token.Value)
	// if user == nil {
	// 	return false
	// }
	// gitRepo := db.GitRepoGetByName(projectName)
	// if gitRepo == nil {
	// 	return false
	// }

	// if user.Email == "ImageBuilder" {
	// 	return true
	// }

	// gitRepo.Open()
	// defer gitRepo.Unlock()

	// branches := gitRepo.GetAllowBranches(user, "db.ProjectRoleSelectorRead")

	// return len(branches) > 0
}

func apiGitOpsRepoMap(r *http.Request, user *db.UserS) interface{} {

	res := map[string]map[string][]string{}
	for item := range db.EnvInGitRepoCache.Iter() {
		if item.Value.Environments.Len() == 0 {
			continue
		}

		tmp := strings.SplitN(item.Key, "/", 2)
		gitRepoName := tmp[0]
		gitBranchName := tmp[1]

		if res[gitRepoName] == nil {
			res[gitRepoName] = map[string][]string{}
		}
		res[gitRepoName][gitBranchName] = item.Value.Environments.Keys()

	}

	return res
}
