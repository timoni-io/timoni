package api

import (
	"core/db"
	"core/db/permissions"
	perms "core/db/permissions"
	"core/db2"
	"encoding/json"
	"fmt"
	"io"
	"lib/tlog"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type projectCreateS struct {
	Name        string
	Description string
	URL         string
}

func apiProjectGetRemoteBranches(r *http.Request, user *db.UserS) interface{} {

	type responseS struct {
		Branches    []string
		IsTaken     bool
		GitRepoName string
		Owners      []*db.GitRepoOwnerS
	}

	remoteURL := r.FormValue("url")

	// url validation
	if remoteURL == "" {
		return tlog.Error("Param `url` is required")
	}
	if _, err := url.ParseRequestURI(remoteURL); err != nil {
		return tlog.Error("Invalid url address")
	}

	// check if remoteURL is unique
	for _, gitRepo := range db.GitRepoMap().Values() {
		if gitRepo.RemoteURL == remoteURL {
			if !user.HasRepoPerm(gitRepo.Name, perms.Repo_View) {
				return tlog.Error("permission denied")
			}
			return responseS{
				IsTaken:     true,
				GitRepoName: gitRepo.Name,
				Owners:      gitRepo.GetOwners(),
			}
		}
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	// Get remote branches, to check if have access to git repo
	_, branches, errx := db.GetRemoteRepoBranches(remoteURL, login, password)
	if errx != nil {
		return errx
	}

	gitRepo := ""
	urlSplited := strings.Split(remoteURL, "/")
	if len(urlSplited) > 0 {
		gitRepo = strings.TrimSuffix(urlSplited[len(urlSplited)-1], ".git")
		gitRepo = strings.ToLower(gitRepo)
		gitRepo = strings.ReplaceAll(gitRepo, "_", "-")
		gitRepo = strings.ReplaceAll(gitRepo, ".", "-")
	}

	if !user.HasRepoPerm(gitRepo, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	return responseS{
		Branches:    branches,
		GitRepoName: gitRepo,
	}
}

func apiGitRepoCreate(r *http.Request, user *db.UserS) interface{} {
	if !user.HasGlobPerm(perms.Glob_CreateAndDeleteGitRepos) {
		return tlog.Error("permission denied")
	}

	gitRepoCreate := new(projectCreateS)

	err := json.NewDecoder(r.Body).Decode(gitRepoCreate)
	if err != nil {
		return tlog.Error("query is not valid json", err)
	}

	if errx := db.GitRepoNameCheck(gitRepoCreate.Name); errx != nil {
		return errx
	}

	if db.GitRepoGetByName(gitRepoCreate.Name) != nil {
		return tlog.Error("git repo with this name already exist")
	}

	gitRepo := &db.GitRepoS{
		Name:               gitRepoCreate.Name,
		Description:        gitRepoCreate.Description,
		CreatedTime:        time.Now(),
		CreatedByUserID:    user.ID,
		CreatedByUserEmail: user.Email,
		DefaultBranch:      "master",
		Operators: map[string]string{
			user.Email: "owner",
		},
		ResourcesUsage: db.CodeStats{},
		ResourcesLimit: db.ProjectDefaultResourcesLimit,
	}

	gitRepo.Open()
	defer gitRepo.Unlock()
	gitRepo.Save()

	tlog.Info("Git repo created", tlog.Vars{
		"user":    user.Email,
		"gitRepo": gitRepo.Name,
		"remote":  gitRepo.RemoteURL != "",
		"event":   true,
	})
	return gitRepo.Name
}

func apiProjectCreateRemote(r *http.Request, user *db.UserS) interface{} {

	if !user.HasGlobPerm(perms.Glob_CreateAndDeleteGitRepos) {
		return tlog.Error("permission denied")
	}
	projectCreate := new(projectCreateS)

	err := json.NewDecoder(r.Body).Decode(projectCreate)
	if err != nil {
		return tlog.Error("query is not valid json", err)
	}

	if errx := db.GitRepoNameCheck(projectCreate.Name); errx != nil {
		return errx
	}

	if db.GitRepoGetByName(projectCreate.Name) != nil {
		return tlog.Error("git repo with this name already exist")
	}

	// url validation
	remoteURL := strings.TrimSpace(projectCreate.URL)
	if remoteURL == "" {
		return tlog.Error("Param `url` is required")
	}

	if strings.HasPrefix(remoteURL, "http") {
		_, err = url.ParseRequestURI(remoteURL)
		if err != nil {
			return tlog.Error("Invalid url address")
		}
	} else if strings.HasPrefix(remoteURL, "git@") {
	} else {
		return tlog.Error("Invalid url address")
	}

	type responseS struct {
		IsTaken     bool
		GitRepoName string
		Owners      []*db.GitRepoOwnerS
	}

	for _, gitRepo := range db.GitRepoMap().Values() {
		if gitRepo.RemoteURL == remoteURL {
			return responseS{
				IsTaken:     true,
				GitRepoName: gitRepo.Name,
				Owners:      gitRepo.GetOwners(),
			}
		}
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	// branch validation
	gitProviderID, branches, errx := db.GetRemoteRepoBranches(remoteURL, login, password)
	if errx != nil {
		return errx
	}

	if len(branches) == 0 {
		return tlog.Error("empty repo, no branches")
	}

	branchMap := map[string]bool{}
	for _, k := range branches {
		branchMap[k] = true
	}

	sort.Strings(branches)

	defaultBranch := branches[0]
	if branchMap["main"] {
		defaultBranch = "main"

	} else if branchMap["master"] {
		defaultBranch = "master"
	}

	gitRepo := &db.GitRepoS{
		Name:               projectCreate.Name,
		Description:        projectCreate.Description,
		CreatedTime:        time.Now(),
		CreatedByUserID:    user.ID,
		CreatedByUserEmail: user.Email,
		RemoteURL:          remoteURL,
		GitProviderID:      gitProviderID,
		DefaultBranch:      defaultBranch,
		Operators: map[string]string{
			user.Email: "owner",
		},
		ResourcesUsage: db.CodeStats{},
		ResourcesLimit: db.ProjectDefaultResourcesLimit,
	}

	gitRepo.Open()
	defer gitRepo.Unlock()
	gitRepo.Save()

	tlog.Info("Git repo created", tlog.Vars{
		"user":    user.Email,
		"gitRepo": gitRepo.Name,
		"remote":  gitRepo.RemoteURL != "",
		"event":   true,
	})

	return responseS{
		IsTaken:     false,
		GitRepoName: gitRepo.Name,
		Owners:      gitRepo.GetOwners(),
	}
}

func apiGitRepoDelete(r *http.Request, user *db.UserS) interface{} {

	gitRepoName := r.FormValue("name")
	if errx := db.GitRepoNameCheck(gitRepoName); errx != nil {
		return errx
	}

	gitRepo := db.GitRepoGetByName(gitRepoName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", gitRepoName)
	}

	if gitRepo.RemoteLastSync == 0 && !user.HasRepoPerm(gitRepoName, perms.Repo_LocalManage) {
		return tlog.Error("permission denied")
	}

	if !user.HasRepoPerm(gitRepoName, perms.Repo_RemoteManage) {
		return tlog.Error("permission denied")
	}

	err := gitRepo.Delete()
	if err != nil {
		return err
	}

	tlog.Info("Project deleted", tlog.Vars{
		"user":     user.Email,
		"git-repo": gitRepo.Name,
		"remote":   gitRepo.RemoteURL != "",

		"event": true,
	})

	gitRepo.Delete()
	return "ok"
}

func apiGitRepoInfo(r *http.Request, user *db.UserS) interface{} {
	gitRepoName := r.FormValue("name")
	if errx := db.GitRepoNameCheck(gitRepoName); errx != nil {
		return errx
	}

	if !user.HasRepoPerm(gitRepoName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(gitRepoName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", gitRepoName)
	}

	cloneURL := fmt.Sprintf("%s/git/%s", db2.TheDomain.URL(""), gitRepoName)
	remoteURL := gitRepo.RemoteURL

	return struct {
		Name                   string
		Description            string
		CloneURL               string
		RemoteURL              string
		DefaultBranch          string
		AppCreateLimitExceeded bool
		Owners                 []*db.GitRepoOwnerS
		Error                  string
		AccessToCode           string
		Permissions            map[string]permissions.PermExplained
	}{
		Name:          gitRepo.Name,
		Description:   gitRepo.Description,
		CloneURL:      cloneURL,
		AccessToCode:  fmt.Sprintf("curl -sfL %s/api/git-on-board?token=%s", db2.TheDomain.URL(""), user.ID),
		RemoteURL:     remoteURL,
		DefaultBranch: gitRepo.DefaultBranch,
		Owners:        gitRepo.GetOwners(),
		Error:         gitRepo.Error,
	}
}

func apiProjectAccessControlInfo(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("name")
	if errx := db.GitRepoNameCheck(projectName); errx != nil {
		return errx
	}

	if !user.HasRepoPerm(projectName, perms.Repo_SettingsManage) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	// ---------------------------------------------------------

	type resS struct {
		Name           string
		Members        map[string]string
		ResourcesLimit db.CodeStats
		ResourcesUsage db.CodeStats
	}

	if gitRepo.RemoteURL != "" {
		// if repo is remote, don't limit resources
		gitRepo.ResourcesLimit.CodeBranches = gitRepo.ResourcesUsage.CodeBranches
		gitRepo.ResourcesLimit.CodeStorage = gitRepo.ResourcesUsage.CodeStorage
	}
	tmpMembers := map[string]string{}

	for memberEmail, memberRole := range gitRepo.Operators {
		tmpUser := db.GetUserByEmail(memberEmail)
		if tmpUser == nil {
			continue
		}
		if tmpUser.Teams.Exists(db.BlacklistedTeamName) {
			tmpMembers[memberEmail] = "no-access"
		} else {
			tmpMembers[memberEmail] = memberRole
		}
	}

	return resS{
		Name:           gitRepo.Name,
		Members:        tmpMembers,
		ResourcesLimit: gitRepo.ResourcesLimit,
		ResourcesUsage: gitRepo.ResourcesUsage,
	}
}

func apiProjectBranchList(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("name")
	if errx := db.GitRepoNameCheck(projectName); errx != nil {
		return errx
	}

	if !user.HasRepoPerm(projectName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	gitRepo.Open()
	defer gitRepo.Unlock()
	gitRepo.Save()

	// Return list for frontend
	list := []string{}
	for k := range gitRepo.GetAllowBranches(user, "lvl") {
		list = append(list, k)
	}

	if gitRepo.DefaultBranch == "" {
		set := false
		for _, v := range list {
			if v == "master" || v == "main" {
				gitRepo.DefaultBranch = v
				set = true
				break
			}
		}
		if !set {
			// if git repo does not have master/main branch -> set first branch as default
			gitRepo.DefaultBranch = list[0]
		}
	}

	sort.Strings(list)
	return list
}

func apiProjectTagList(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("name")
	if errx := db.GitRepoNameCheck(projectName); errx != nil {
		return errx
	}

	if !user.HasRepoPerm(projectName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	gitRepo.Open()
	defer gitRepo.Unlock()

	// Return list for frontend
	list := []db.GitTagS{}
	for _, tag := range gitRepo.GetAllowTags(user, "lvl") {
		list = append(list, tag)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].TimeStamp > list[j].TimeStamp
	})

	return list
}

func apiGetLastCommit(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("name")
	if projectName == "" {
		return tlog.Error("Param `name` is required")
	}

	if !user.HasRepoPerm(projectName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	branch := r.FormValue("branch")
	if branch == "" {
		return tlog.Error("Param `branch` is required")
	}
	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	commit := gitRepo.GetLastCommit(branch)
	if commit == "" {
		return tlog.Error("Branch " + branch + " nie istnieje")
	}
	return commit
}

func apiProjectCommitList(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("name")
	if projectName == "" {
		return tlog.Error("Param `name` is required")
	}

	if !user.HasRepoPerm(projectName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	if errx := db.GitRepoNameCheck(projectName); errx != nil {
		return errx
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	branchName := r.FormValue("branch")
	if branchName == "" {
		return tlog.Error("Param `branch` is required")
	}

	gitRepo.Open()
	defer gitRepo.Unlock()
	branchList := gitRepo.GetAllowBranches(user, "db.ProjectRoleSelectorRead")

	if _, exist := branchList[branchName]; !exist {
		return tlog.Error("access denied", tlog.Vars{
			"branchName": branchName,
			"user":       user.Email,
		})
	}

	from := r.FormValue("from") // commit hash
	limitStr := r.FormValue("limit")
	var limit int
	var err error

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return tlog.Error("Invalid `limit` param - not an integer")
		}
		if limit > 50 {
			return tlog.Error("Invalid `limit` param value - max is 50")
		}
	}

	page := r.FormValue("page")
	if page != "" {
		pages := 15
		limit, err = strconv.Atoi(page)
		limit *= pages
		if err != nil {
			return tlog.Error("Invalid `page` param - not an integer")
		}
		if limit < pages {
			return tlog.Error("Invalid `page` param - lower than 1")
		}
		commits := gitRepo.GetCommits(branchName, limit, from)
		rLimit := limit
		if len(commits) < limit {
			return commits
		}
		if limit-pages > rLimit {
			return nil
		}
		return commits[limit-pages : rLimit]
	}

	// default value
	if limit == 0 {
		limit = 50
	}

	return gitRepo.GetCommits(branchName, limit, from)
}

func apiProjectCommitInfoFileList(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("name")
	if errx := db.GitRepoNameCheck(projectName); errx != nil {
		return errx
	}

	if !user.HasRepoPerm(projectName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	branch := r.FormValue("branch")
	if branch == "" {
		return tlog.Error("branch is required")
	}

	// Check branch access
	gitRepo.Open()
	defer gitRepo.Unlock()
	branches := gitRepo.GetAllowBranches(user, "db.ProjectRoleSelectorRead")
	if _, exist := branches[branch]; !exist {
		return tlog.Error("access denied", tlog.Vars{
			"branchName": branch,
			"user":       user.Email,
		})
	}

	commitHash := r.FormValue("commit")
	if commitHash == "" {
		return tlog.Error("commit is required")
	}

	return gitRepo.CommitDiffFileList(commitHash, 0)
}

func apiProjectBranchesInfoFileList(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("name")
	if errx := db.GitRepoNameCheck(projectName); errx != nil {
		return errx
	}

	if !user.HasRepoPerm(projectName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	branch1 := r.FormValue("branch1")
	if branch1 == "" {
		return tlog.Error("branch1 is required")
	}

	branch2 := r.FormValue("branch2")
	if branch2 == "" {
		return tlog.Error("branch2 is required")
	}

	// Check branch access
	gitRepo.Open()
	defer gitRepo.Unlock()
	branches := gitRepo.GetAllowBranches(user, "db.ProjectRoleSelectorRead")
	if _, exist := branches[branch1]; !exist {
		return tlog.Error("access denied", tlog.Vars{
			"branchName": branch1,
			"user":       user.Email,
		})
	}

	if _, exist := branches[branch2]; !exist {
		return tlog.Error("access denied", tlog.Vars{
			"branchName": branch2,
			"user":       user.Email,
		})
	}

	return gitRepo.BranchesDiffFileList(branch1, branch2)
}

func apiProjectCommitInfoFileDiff(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("name")
	if errx := db.GitRepoNameCheck(projectName); errx != nil {
		return errx
	}

	if !user.HasRepoPerm(projectName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	branch := r.FormValue("branch")
	if branch == "" {
		return tlog.Error("branch is required")
	}

	// Check branch access
	gitRepo.Open()
	defer gitRepo.Unlock()
	branches := gitRepo.GetAllowBranches(user, "db.ProjectRoleSelectorRead")
	if _, exist := branches[branch]; !exist {
		return tlog.Error("access denied", tlog.Vars{
			"branchName": branch,
			"user":       user.Email,
		})
	}

	commitHash := r.FormValue("commit")
	if commitHash == "" {
		return tlog.Error("commit is required")
	}

	filePath := r.FormValue("path")
	if filePath == "" {
		return tlog.Error("path is required")
	}

	out, errx := gitRepo.GetCommitFileDiff(commitHash, filePath)
	if errx != nil {
		return errx
	}
	return out
}

func apiProjectBranchesInfoFileDiff(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("name")
	if errx := db.GitRepoNameCheck(projectName); errx != nil {
		return errx
	}

	if !user.HasRepoPerm(projectName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	branch1 := r.FormValue("branch1")
	if branch1 == "" {
		return tlog.Error("branch1 is required")
	}
	branch2 := r.FormValue("branch2")
	if branch2 == "" {
		return tlog.Error("branch2 is required")
	}
	filePath := r.FormValue("filePath")
	if filePath == "" {
		return tlog.Error("filePath is required")
	}

	// Check branch access
	gitRepo.Open()
	defer gitRepo.Unlock()
	branches := gitRepo.GetAllowBranches(user, "db.ProjectRoleSelectorRead")
	if _, exist := branches[branch1]; !exist {
		return tlog.Error("access denied", tlog.Vars{
			"branchName": branch1,
			"user":       user.Email,
		})
	}
	if _, exist := branches[branch2]; !exist {
		return tlog.Error("access denied", tlog.Vars{
			"branchName": branch2,
			"user":       user.Email,
		})
	}

	out, errx := gitRepo.GetBranchesFileDiff(branch1, branch2, filePath)
	if errx != nil {
		return errx
	}

	return out
}

func apiProjectBranchesMerge(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("name")
	if errx := db.GitRepoNameCheck(projectName); errx != nil {
		return errx
	}

	if !user.HasRepoPerm(projectName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	if gitRepo.RemoteURL != "" {
		return tlog.Error("can't merge remote git repo", tlog.Vars{
			"user":     user.Email,
			"git-repo": gitRepo.Name,
			"remote":   gitRepo.RemoteURL,
		})
	}

	branch1 := r.FormValue("branch1")
	if branch1 == "" {
		return tlog.Error("branch1 is required")
	}
	branch2 := r.FormValue("branch2")
	if branch2 == "" {
		return tlog.Error("branch2 is required")
	}
	squash := false
	if r.FormValue("squash") == "true" {
		squash = true
	}

	removeBranch := false
	if r.FormValue("removeBranch") == "true" {
		removeBranch = true
		if branch2 == "master" {
			return tlog.Error("Can't remove master branch")
		}
	}

	commitMessage := r.FormValue("commitMessage")
	if commitMessage == "" {
		commitMessage = fmt.Sprint("merge ", branch2)
	}

	// Check branch access
	gitRepo.Open()
	defer gitRepo.Unlock()
	branches := gitRepo.GetAllowBranches(user, "db.ProjectRoleSelectorReadWrite")
	if _, exist := branches[branch1]; !exist {
		return tlog.Error("access denied", tlog.Vars{
			"branchName": branch1,
			"user":       user.Email,
		})
	}
	branches = gitRepo.GetAllowBranches(user, "db.ProjectRoleSelectorRead")
	if _, exist := branches[branch2]; !exist {
		return tlog.Error("access denied", tlog.Vars{
			"branchName": branch2,
			"user":       user.Email,
		})
	}

	removeWorktree := func(tmpIndex string, name string) error {
		cmd := []string{
			"git", "-C", gitRepo.PathGet() + tmpIndex, "worktree", "remove", "--force", name + tmpIndex,
		}
		_, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if tlog.Error(err) != nil {
			return err
		}
		return nil
	}

	tmpIndex := fmt.Sprint("_tmp_", time.Now().UnixNano(), "_")

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// create worktree
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	cmd := []string{
		"git", "-C", gitRepo.PathGet(), "worktree", "add", "../" + gitRepo.Name + tmpIndex, branch2,
	}
	// fmt.Println(cmd)
	out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if tlog.Error(err) != nil {
		return tlog.Error("git worktree error", tlog.Vars{
			"cmd":    cmd,
			"err":    err,
			"output": string(out),
		})
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	//checkout
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	cmd = []string{
		"git", "checkout", branch1,
	}
	cmdToRun := exec.Command(cmd[0], cmd[1:]...)
	cmdToRun.Dir = gitRepo.PathGet() + tmpIndex
	out, err = cmdToRun.CombinedOutput()
	if err != nil {
		err1 := removeWorktree(tmpIndex, gitRepo.Name)
		if err1 != nil {
			tlog.Error(err1)
		}
		return tlog.Error("git checkout error", tlog.Vars{
			"cmd":    cmd,
			"err":    err,
			"output": string(out),
			"dir":    cmdToRun.Dir,
		})
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	//merge
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	cmd = []string{"git", "merge"}
	if squash {
		cmd = append(cmd, "--squash")

		cmdTMP := []string{"git", "merge-base", branch1, branch2}
		cmdToRun = exec.Command(cmdTMP[0], cmdTMP[1:]...)
		cmdToRun.Dir = gitRepo.PathGet() + tmpIndex
		commitStopByte, err := cmdToRun.CombinedOutput()
		if tlog.Error(err) != nil {
			return tlog.Error("git merge-base error", tlog.Vars{
				"cmd":    cmd,
				"err":    err,
				"output": string(out),
			})
		}

		commitStopSHA := string(commitStopByte)
		branch1CommitsSHA := map[string]bool{}
		for _, commit := range gitRepo.GetCommits(branch1, 500, "") {
			branch1CommitsSHA[commit.SHA] = true
		}

		commitMessage += "\n\nauthors: "
		autors := map[string]bool{}
		for _, commit := range gitRepo.GetCommits(branch2, 500, "") {
			if commit.SHA == commitStopSHA {
				break
			}
			if commit.AuthorEmail == "Timoni" {
				continue
			}
			if !branch1CommitsSHA[commit.SHA] && !autors[commit.AuthorEmail] {
				autors[commit.AuthorEmail] = true
				commitMessage += "\n" + commit.AuthorEmail
			}
		}
	}
	cmd = append(cmd, branch2)
	cmdToRun = exec.Command(cmd[0], cmd[1:]...)
	cmdToRun.Dir = gitRepo.PathGet() + tmpIndex
	out, err = cmdToRun.CombinedOutput()
	if tlog.Error(err) != nil {
		err1 := removeWorktree(tmpIndex, gitRepo.Name)
		if err1 != nil {
			tlog.Error(err1)
		}
		return tlog.Error("commit #2 error", tlog.Vars{
			"cmd":    cmd,
			"err":    err,
			"output": string(out),
		})
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	//commit
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	cmd = []string{"git", "commit", "-am", commitMessage}
	cmdToRun = exec.Command(cmd[0], cmd[1:]...)
	cmdToRun.Dir = gitRepo.PathGet() + tmpIndex
	out, err = cmdToRun.CombinedOutput()
	if tlog.Error(err) != nil && err.Error() != "exit status 1" {
		err1 := removeWorktree(tmpIndex, gitRepo.Name)
		if err1 != nil {
			tlog.Error(err1)
		}
		return tlog.Error("commit error", tlog.Vars{
			"cmd":    cmd,
			"err":    err,
			"output": string(out),
		})
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	//remove branch
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	if removeBranch {
		cmd = []string{
			"git", "-C", gitRepo.PathGet() + tmpIndex, "branch", "-D", branch2,
		}
		out, err = exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if tlog.Error(err) != nil {
			err1 := removeWorktree(tmpIndex, gitRepo.Name)
			if err1 != nil {
				tlog.Error(err1)
			}
			return tlog.Error(err, tlog.Vars{
				"cmd":    cmd,
				"err":    err,
				"output": string(out),
			})
		}
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	//remove worktree
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	err = removeWorktree(tmpIndex, gitRepo.Name)
	if tlog.Error(err) != nil {
		return tlog.Error("remove worktree error")
	}
	tlog.Info("branches merged "+branch1+" to "+branch2, tlog.Vars{
		"event": true,
		"user":  user.Email,
	})
	return "ok"
}

func apiProjectBranchesCompare(r *http.Request, user *db.UserS) interface{} {
	type tmpS struct {
		Branch string
		Behind int
		Ahead  int
	}

	tmpData := []tmpS{}

	projectName := r.FormValue("git-repo")
	branchName := r.FormValue("branch")

	if !user.HasRepoPerm(projectName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(projectName)

	if gitRepo == nil {
		return tlog.Error("Project not found", tlog.Vars{
			"git-repo": projectName,
			"user":     user.Email,
		})
	}

	gitRepo.Open()
	defer gitRepo.Unlock()
	branches := gitRepo.GetAllowBranches(user, "db.ProjectRoleSelectorRead")
	if _, exist := branches[branchName]; !exist {
		return tlog.Error("access denied", tlog.Vars{
			"branchName": branchName,
			"user":       user.Email,
		})
	}

	for _, branch2 := range gitRepo.BranchHashMap.Keys() {
		if _, exist := branches[branch2]; !exist {
			continue
		}
		if branch2 != branchName {
			out, errx := gitRepo.CompareBranches(branchName, branch2)
			if errx != nil {
				return errx
			}

			re := regexp.MustCompile("[0-9]+")
			tab := re.FindAllString(out, -1)
			b, _ := strconv.Atoi(tab[0])
			a, _ := strconv.Atoi(tab[1])
			tmpData = append(tmpData, tmpS{
				Branch: branch2,
				Behind: b,
				Ahead:  a,
			})
		}
	}

	return tmpData
}

func apiProjectFiles(r *http.Request, user *db.UserS) interface{} {
	type fileObjectS struct {
		Name   string `json:"name"`
		IsFile string `json:"isFile"`
	}

	projectName := r.FormValue("name")
	if !user.HasRepoPerm(projectName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	if errx := db.GitRepoNameCheck(projectName); errx != nil {
		return errx
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	branchName := r.FormValue("branch")
	commitHash := r.FormValue("commit")

	// validate input
	if commitHash == "" && branchName == "" {
		return tlog.Error("must specify either commithash or branch name")
	}
	if commitHash != "" && branchName != "" {
		return tlog.Error("branchName and Commit hash cant both be specified")
	}

	gitRepo.Open()
	defer gitRepo.Unlock()
	branches := gitRepo.GetAllowBranches(user, "db.ProjectRoleSelectorRead")
	if _, exist := branches[branchName]; !exist {
		return tlog.Error("access denied", tlog.Vars{
			"branchName": branchName,
			"user":       user.Email,
		})
	}

	currentDirectory := r.FormValue("directory")
	fileTree, err := gitRepo.FileTreeOfCommit(commitHash, branchName, currentDirectory)
	if err != nil {
		tlog.Error(err)
		return tlog.Error("server error", tlog.Vars{
			"commitHash":       commitHash,
			"branchName":       branchName,
			"currentDirectory": currentDirectory,
		})
	}
	if fileTree == nil {
		return tlog.Error("server error", tlog.Vars{
			"commitHash":       commitHash,
			"branchName":       branchName,
			"currentDirectory": currentDirectory,
		})
	}

	directory := []fileObjectS{}
	file := []fileObjectS{}
	// send its contents
	for _, entry := range fileTree.Entries {
		sth := fileObjectS{
			Name:   entry.Name,
			IsFile: "dir",
		}
		if entry.Mode.IsFile() {
			sth.IsFile = "file"
			file = append(file, sth)
			continue
		}
		directory = append(directory, sth)
	}
	return append(directory, file...)
}

func apiProjectFileOpen(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("name")
	if errx := db.GitRepoNameCheck(projectName); errx != nil {
		return errx
	}

	if !user.HasRepoPerm(projectName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	branchName := r.FormValue("branch")
	commitHash := r.FormValue("commit")

	// validate input
	if commitHash == "" && branchName == "" {
		return tlog.Error("must specify either commithash or branch name")
	}
	if commitHash != "" && branchName != "" {
		return tlog.Error("branchName and Commit hash cant both be specified")
	}

	gitRepo.Open()
	defer gitRepo.Unlock()
	branchList := gitRepo.GetAllowBranches(user, "db.ProjectRoleSelectorRead")
	if _, exist := branchList[branchName]; !exist {
		return tlog.Error("access denied", tlog.Vars{
			"branchName": branchName,
			"user":       user.Email,
		})
	}

	path := r.FormValue("path")
	if path == "" {
		return "select file"
	}

	splitPath := strings.Split(path, "/")
	fileName := splitPath[len(splitPath)-1]
	currentDirectory := strings.Join(splitPath[:len(splitPath)-1], "/")
	if len(currentDirectory) != 0 && currentDirectory[0] == '/' { // got to remove the first slash with it it wont find the dir
		currentDirectory = currentDirectory[1:]
	}

	if fileName == "" {
		return tlog.Error("select file", tlog.Vars{
			"path": path,
		})
	}
	fileTree, err2 := gitRepo.FileTreeOfCommit(commitHash, branchName, currentDirectory)
	if err2 != nil {
		tlog.Error(err2)
		return tlog.Error("server error")
	}
	if fileTree == nil {
		return tlog.Error("server error")
	}

	file, err := fileTree.FindEntry(fileName)
	if tlog.Error(err) != nil || file == nil {
		return tlog.Error("could not find the file")
	}

	descriptor, err := fileTree.TreeEntryFile(file)
	if tlog.Error(err) != nil || file == nil {
		return tlog.Error(err)
	}
	isBinFile, err := descriptor.IsBinary()
	if err != nil {
		return tlog.Error(err)
	}
	if isBinFile {
		return tlog.Error("file is not human readable")
	}

	fileContent, err := descriptor.Contents()
	if tlog.Error(err) != nil || file == nil {
		return tlog.Error(err)
	}
	return fileContent
}

type gitRepoTmpS struct {
	Name          string
	DefaultBranch string
	Local         bool
	Status        db.ElementState
	StatusMessage string
}

func apiGitRepoMap(r *http.Request, user *db.UserS) interface{} {
	repos := map[string]gitRepoTmpS{}

	for _, repo := range db.GitRepoMap().Values() {

		if !user.HasRepoPerm(repo.Name, perms.Repo_View) {
			continue
		}

		repos[repo.Name] = gitRepoTmpS{
			Name:          repo.Name,
			DefaultBranch: repo.DefaultBranch,
			Local:         repo.RemoteURL == "",
			Status:        db.ElementStatusReady,
			StatusMessage: "",
		}

	}

	return repos
}

func apiProjectRemoteAccessUpdate(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("git-repo")
	remoteURL := r.FormValue("url")

	if errx := db.GitRepoNameCheck(projectName); errx != nil {
		return errx
	}

	if !user.HasRepoPerm(projectName, perms.Repo_SettingsManage) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", projectName)
	}

	if remoteURL == "" && gitRepo.RemoteURL != "" {
		remoteURL = gitRepo.RemoteURL
	}

	// url validation
	_, err := url.ParseRequestURI(remoteURL)
	if err != nil {
		return tlog.Error("Invalid url address")
	}

	gitProviderID, _, errx := db.GetRemoteRepoBranches(remoteURL, "", "")
	if errx != nil {
		return errx
	}

	gitRepo.RemoteURL = remoteURL
	gitRepo.GitProviderID = gitProviderID
	gitRepo.Error = ""
	gitRepo.Save()

	tlog.Info("Project credentials updated", tlog.Vars{
		"git-repo": gitRepo.Name,
		"user":     user.Email,
		"event":    true,
	})

	db.EventProjectSetingsChanged(gitRepo, *user)
	return "ok"
}

func apiProjectChangeLimit(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("git-repo")

	if !user.HasRepoPerm(projectName, perms.Repo_SettingsManage) {
		return tlog.Error("permission denied")
	}

	limitName := r.FormValue("limit")
	value, err := strconv.Atoi(r.FormValue("value"))

	if err != nil {
		return tlog.Error("Value must be int", tlog.Vars{
			"limit": value,
		})
	}

	if value < 0 {
		return tlog.Error("Value must be greater than or equal to 0 ", tlog.Vars{
			"limit": value,
		})
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("Project not found", tlog.Vars{
			"user":     user.Email,
			"git-repo": projectName,
		})
	}

	var oldLimit int64 = 0
	val := int64(value)
	switch limitName {

	case "CodeBranches":
		oldLimit = gitRepo.ResourcesLimit.CodeBranches
		gitRepo.ResourcesLimit.CodeBranches = val
	case "CodeStorage":
		oldLimit = gitRepo.ResourcesLimit.CodeStorage
		gitRepo.ResourcesLimit.CodeStorage = val

	case "AppCount", "PodsMin", "PodsMax",
		"CPUGuaranteed", "CPUMax", "RAMGuaranteed",
		"RAMMax", "StorageTemporary", "StoragePersistent":

	default:
		return tlog.Error("Wrong limit name", tlog.Vars{
			"user":     user.Email,
			"git-repo": projectName,
			"limit":    limitName,
		})
	}
	tlog.Info("repo limits changed", tlog.Vars{
		"event":    true,
		"git-repo": gitRepo.Name,
		"limit":    limitName,
		"old":      oldLimit,
		"new":      value,
		"user":     user.Email,
	})
	gitRepo.Save()
	db.EventLimitChanged(gitRepo, limitName, int(oldLimit), value, *user)
	return "ok"
}

func apiProjectNotificationUpdate(r *http.Request, user *db.UserS) interface{} {
	type tmpDataS struct {
		GitRepoName string
		Member      bool
		Role        bool
		YourRole    bool
		Limit       bool
		Settings    bool
		App         bool
		CodeBranche bool
		CodeCommit  map[string]bool
	}

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		tlog.Error(err)
		return tlog.Error("cant read POST body")
	}

	data := &tmpDataS{}
	err = json.Unmarshal(buf, data)
	if err != nil {
		tlog.Error(err)
		return tlog.Error("cant read POST body JSON")
	}

	if !user.HasRepoPerm(data.GitRepoName, perms.Repo_SettingsManage) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(data.GitRepoName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", tlog.Vars{
			"user":     user.Email,
			"git-repo": data.GitRepoName,
		})
	}

	gitRepo.SetNotification(
		user.Email,
		data.Member,
		data.Role,
		data.YourRole,
		data.Limit,
		data.Settings,
		data.App,
		data.CodeBranche,
		data.CodeCommit,
	)
	return "ok"
}

func apiProjectNotificationInfo(r *http.Request, user *db.UserS) interface{} {

	projectName := r.FormValue("git-repo")
	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", tlog.Vars{
			"user":     user.Email,
			"git-repo": projectName,
		})
	}

	type tmpDataS struct {
		Member      bool
		Role        bool
		YourRole    bool
		Limit       bool
		Settings    bool
		App         bool
		CodeBranche bool
		CodeCommit  map[string]bool
	}

	return tmpDataS{
		Member:      gitRepo.Notifications.Member[user.Email],
		Role:        gitRepo.Notifications.Role[user.Email],
		YourRole:    gitRepo.Notifications.YourRole[user.Email],
		Limit:       gitRepo.Notifications.Limit[user.Email],
		Settings:    gitRepo.Notifications.Settings[user.Email],
		App:         gitRepo.Notifications.App[user.Email],
		CodeBranche: gitRepo.Notifications.CodeBranche[user.Email],
		CodeCommit:  gitRepo.Notifications.CodeCommit[user.Email],
	}
}

func apiProjectUpdate(r *http.Request, user *db.UserS) interface{} {
	projectName := r.FormValue("git-repo")

	if !user.HasRepoPerm(projectName, perms.Repo_SettingsManage) {
		return tlog.Error("permission denied")
	}

	gitRepo := db.GitRepoGetByName(projectName)
	if gitRepo == nil {
		return tlog.Error("git repo not found", tlog.Vars{
			"user":     user.Email,
			"git-repo": projectName,
		})
	}

	defaultBranch := r.FormValue("default")
	if defaultBranch == "" {
		return tlog.Error("param `default` is required")
	}
	if gitRepo.DefaultBranch != defaultBranch {
		gitRepo.DefaultBranch = defaultBranch
	}

	tlog.Info("repo updated", tlog.Vars{
		"event": true,
		"user":  user.Email,
	})
	return "ok"
}

func apiGitRepoElementList(r *http.Request, user *db.UserS) interface{} {
	gitRepoName := r.FormValue("git-repo-name")

	if !user.HasRepoPerm(gitRepoName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	branch := r.FormValue("branch")

	if branch == "" {
		return tlog.Error("branch is empty")
	}

	type tmpS struct {
		Element *db.GitElementS
		Usage   []db.ElementInUsageTmpS
	}

	gitRepo := db.GitRepoGetByName(gitRepoName)
	if gitRepo == nil {
		return tlog.Error("git-repo not found")
	}

	usageMap := gitRepo.UseOfElements()

	res := []tmpS{}
	for _, cache := range db.ElementsInGitRepoCache.Values() {
		for _, element := range cache.Elements {
			if element.Source.RepoName != gitRepo.Name {
				continue
			}
			if element.Source.BranchName != branch {
				continue
			}
			key := branch + "/" + element.Source.FilePath
			res = append(res, tmpS{
				Element: element,
				Usage:   usageMap[key],
			})
		}
	}

	return res
}

func apiGitRepoEnvMap(r *http.Request, user *db.UserS) interface{} {
	gitRepoName := r.FormValue("git-repo-name")

	if !user.HasRepoPerm(gitRepoName, perms.Repo_View) {
		return tlog.Error("permission denied")
	}

	branch := r.FormValue("branch")

	if branch == "" {
		return tlog.Error("branch is empty")
	}

	gitRepo := db.GitRepoGetByName(gitRepoName)
	if gitRepo == nil {
		return tlog.Error("git-repo not found")
	}

	return db.EnvInGitRepoCache.Get(gitRepoName + "/" + branch).Environments
}
