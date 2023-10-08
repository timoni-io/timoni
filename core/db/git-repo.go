package db

import (
	"bytes"
	"core/config"
	"core/db2"
	"encoding/base64"
	"fmt"
	"io"
	"lib/tlog"
	"lib/utils"
	"lib/utils/archive"
	"lib/utils/conv"
	"lib/utils/maps"
	"lib/utils/set"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

type GitRepoS struct {
	Name               string
	Description        string
	CreatedTime        time.Time
	CreatedByUserID    string
	CreatedByUserEmail string
	RemoteURL          string // mirror repo from other git server
	GitProviderID      string
	RemoteLastSync     int64
	DefaultBranch      string
	Operators          OperatorRoleT
	BranchHashMap      *maps.SafeMap[string, string]
	ResourcesLimit     CodeStats
	ResourcesUsage     CodeStats
	Notifications      GitRepoNotificationsS
	Error              string

	Tags *set.Safe[string]

	mutex     sync.Mutex
	mutexLock bool

	store *git.Repository
}

type CodeStats struct {
	CodeBranches int64
	CodeStorage  int64 // size on disk
}

type CommitS struct {
	SHA            string
	Message        string
	Date           time.Time
	TimeStamp      int64
	AuthorName     string
	AuthorEmail    string
	AuthorInitials string
	Files          []string
}

type GitTagS struct {
	Name                 string
	CommitSHA            string
	CommitAuthorEmail    string
	CommitAuthorInitials string
	TimeStamp            int64
}

var ProjectDefaultResourcesLimit = CodeStats{
	CodeBranches: 20,
	CodeStorage:  1000,
}

const (
	//element value range
	elementMinStorageTemporary  int     = 2      // MB
	elementMaxStorageTemporary  int     = 10000  // MB
	elementMinStoragePersistent int     = 2      // MB
	elementMaxStoragePersistent int     = 200000 // MB
	elementMinRAMGuaranteed     uint    = 5      // MB
	elementMaxRAMGuaranteed     uint    = 50000  // MB
	elementMinCPUGuaranteed     uint    = 10     // % of cores  100 = 1 vcore on host
	elementMaxCPUGuaranteed     uint    = 8000   // % of cores  100 = 1 vcore on host
	elementGuaranteedMaxRatio   float64 = 1.5    // 1.5 means max value will be 150% of guaranteed. Must be >= 1. Ex. 10 CPUGuaranteed will be 15 CPUMax

	//element defaults
	elementDefaultCPUGuaranteed uint = 50  // 50% of one vCore
	elementDefaultRAMGuaranteed uint = 400 // MB
)

type GitRepoNotificationsS struct {
	Member      map[string]bool
	Role        map[string]bool
	YourRole    map[string]bool
	Limit       map[string]bool
	Settings    map[string]bool
	App         map[string]bool
	CodeBranche map[string]bool
	CodeCommit  map[string]map[string]bool
}
type GitRepoOwnerS struct {
	Name  string
	Email string
}

var (
	globalGitRepoMap = maps.NewSafe[string, *GitRepoS](nil) // key=repoName
)

func (project *GitRepoS) Save() {
	err := driver.Write("git-repo", project.Name, project)
	if tlog.Error(err) != nil {
		return
	}
}

func GitRepoMap() *maps.SafeMap[string, *GitRepoS] {

	repoNames, err := driver.List("git-repo")
	if tlog.Error(err) != nil {
		return nil
	}

	tmp := map[string]bool{}
	for _, repoName := range repoNames {
		tmp[repoName] = true
	}

	for _, k := range globalGitRepoMap.Keys() {
		if !tmp[k] {
			globalGitRepoMap.Delete(k)
		}
	}

	for _, repoName := range repoNames {

		if !globalGitRepoMap.Exists(repoName) {
			tmp := new(GitRepoS)
			if err := driver.Read("git-repo", repoName, tmp); err != nil {
				if tlog.Error(err, tlog.Vars{
					"logger":   "git-repo",
					"git-repo": repoName,
				}) != nil {
					return nil
				}
			}
			globalGitRepoMap.Set(repoName, tmp)
		}

	}

	return globalGitRepoMap
}

func GitRepoGetByName(name string) *GitRepoS {
	GitRepoMap()
	return globalGitRepoMap.Get(name)
}

func (gitRepo *GitRepoS) Delete() *tlog.RecordS {

	if !globalGitRepoMap.Exists(gitRepo.Name) {
		return nil
	}

	use := gitRepo.UseOfElements()
	if len(use) > 0 {
		return tlog.Error("this git-repo is used", tlog.Vars{
			"use": use,
		})
	}

	for _, gitRepoAndBranch := range ElementsInGitRepoCache.Keys() {
		if strings.HasPrefix(gitRepoAndBranch, gitRepo.Name+"/") {
			ElementsInGitRepoCache.Delete(gitRepoAndBranch)
		}
	}

	for _, gitElement := range ElementsInGitRepoFlatMap.Values() {
		if gitElement.Source.RepoName == gitRepo.Name {
			ElementsInGitRepoFlatMap.Delete(gitElement.ID)
		}
	}

	err := gitRepo.ArchiveAndDelete()
	globalGitRepoMap.Delete(gitRepo.Name)
	for _, b := range gitRepo.BranchHashMap.Keys() {
		EnvInGitRepoCache.Delete(gitRepo.Name + "/" + b)
	}
	return tlog.Error(err)
}

func (gitRepo *GitRepoS) ArchiveAndDelete() error {
	filename := fmt.Sprintf("%s-%d.tar.lz4.part", gitRepo.Name, time.Now().Unix())
	outputFilePath := filepath.Join(config.DataPath(), "archive", filename)

	// Compress files
	err := archive.Compress(gitRepo.PathGet(), archive.WriteFile(outputFilePath), archive.CompressLZ4)
	if err != nil {
		return err
	}

	// Rename file
	err = os.Rename(outputFilePath, outputFilePath[:len(outputFilePath)-5])
	if err != nil {
		return err
	}

	err = os.RemoveAll(gitRepo.PathGet())
	if err != nil {
		return err
	}

	err = os.RemoveAll(filepath.Join(config.DataPath(), "git-repo", gitRepo.Name+".json"))
	if err != nil {
		return err
	}

	return nil
}

// GitRepoNameCheck ...
func GitRepoNameCheck(name string) *tlog.RecordS {
	if name == "" {
		return tlog.Error("name is required")
	}
	if !reSimpleName1.MatchString(name) {
		return tlog.Error("name contains characters that are not allowed: " + name)
	}
	if strings.Contains(name, "--") {
		return tlog.Error("name contains `--` that are not allowed")
	}
	if strings.HasPrefix(name, "-") {
		return tlog.Error("name starts with `-` which is not allowed")
	}
	if strings.HasSuffix(name, "-") {
		return tlog.Error("name ends with `-` which is not allowed")
	}
	if len(name) > 70 {
		return tlog.Error("name too long, max 70 chars")
	}
	return nil
}

// GetRemoteURL ..
func (project *GitRepoS) GetRemoteURL() string {

	uri := project.RemoteURL

	if project.GitProviderID != "" {
		gp := db2.GitProviderGetByID(project.GitProviderID)
		if gp.Variant() == db2.GitProviderVariant_LoginAndPassword {
			tmp := strings.Split(uri, "://")
			uri = tmp[0] + "://" + gp.Login() + ":" + url.QueryEscape(gp.Password()) + "@" + strings.Join(tmp[1:], "")

		}
	}

	return uri
}

func (project *GitRepoS) NotifyMember(filter map[string]bool, ownerRequired, defaultAction bool) []string {
	if filter == nil {
		filter = map[string]bool{}
	}
	dirty := false
	notifyEmails := []string{}
	for email, role := range project.Operators {
		action, exist := filter[email]
		if !exist {
			filter[email] = defaultAction
			action = defaultAction
			dirty = true
		}
		if action && (!ownerRequired || role == "owner") {
			user := GetUserByEmail(email)
			if user != nil && user.NotificationsSend && !user.Teams.Exists(BlacklistedTeamName) {
				notifyEmails = append(notifyEmails, email)
			}
		}
	}
	if dirty {
		project.Save()
	}

	return notifyEmails
}

func (project *GitRepoS) MembersAsUser() []*UserS {
	res := []*UserS{}
	for email := range project.Operators {
		tmpUser := GetUserByEmail(email)
		if tmpUser == nil {
			tlog.Error("git-repo members user not found", tlog.Vars{
				"email":    email,
				"git-repo": project.Name,
			})
			continue
		}
		res = append(res, tmpUser)
	}
	return res
}

func (project *GitRepoS) GetOwners() []*GitRepoOwnerS {
	owners := []*GitRepoOwnerS{}
	for email, role := range project.Operators {
		if role == "owner" {
			user := GetUserByEmail(email)
			if user != nil {
				owners = append(owners, &GitRepoOwnerS{
					Email: user.Email,
					Name:  user.Name,
				})
			}
		}
	}
	return owners
}

func (project *GitRepoS) SetNotification(userEmail string, Member, Role, YourRole, Limit, Settings, App, CodeBranche bool, CodeCommit map[string]bool) {

	if project.Notifications.Member == nil {
		project.Notifications.Member = map[string]bool{}
	}
	project.Notifications.Member[userEmail] = Member

	if project.Notifications.Role == nil {
		project.Notifications.Role = map[string]bool{}
	}
	project.Notifications.Role[userEmail] = Role

	if project.Notifications.YourRole == nil {
		project.Notifications.YourRole = map[string]bool{}
	}
	project.Notifications.YourRole[userEmail] = YourRole

	if project.Notifications.Limit == nil {
		project.Notifications.Limit = map[string]bool{}
	}
	project.Notifications.Limit[userEmail] = Limit

	if project.Notifications.Settings == nil {
		project.Notifications.Settings = map[string]bool{}
	}
	project.Notifications.Settings[userEmail] = Settings

	if project.Notifications.App == nil {
		project.Notifications.App = map[string]bool{}
	}
	project.Notifications.App[userEmail] = App

	if project.Notifications.CodeBranche == nil {
		project.Notifications.CodeBranche = map[string]bool{}
	}
	project.Notifications.CodeBranche[userEmail] = CodeBranche

	if project.Notifications.CodeCommit == nil {
		project.Notifications.CodeCommit = map[string]map[string]bool{}
	}
	project.Notifications.CodeCommit[userEmail] = CodeCommit

	project.Save()
}

func (gitRepo *GitRepoS) GetLastCommit(branch string) string {

	if gitRepo == nil {
		return ""
	}

	if gitRepo.store == nil {
		return ""
	}

	if gitRepo.BranchHashMap == nil {
		return ""
	}

	return gitRepo.BranchHashMap.Get(branch)
}

func (gitRepo *GitRepoS) Lock() {
	gitRepo.mutex.Lock()
	gitRepo.mutexLock = true
}

func (gitRepo *GitRepoS) Unlock() {
	if gitRepo == nil {
		return
	}

	gitRepo.mutexLock = false
	gitRepo.mutex.Unlock()
}

// ===========================================

func (gitRepo *GitRepoS) Open() (success bool) {

	if gitRepo == nil {
		return false
	}

	gitRepo.Lock()

	if _, err := os.Stat(gitRepo.PathGet()); os.IsNotExist(err) {

		if gitRepo.RemoteURL == "" {
			gitRepo.initLocalGitRepo()
		} else {
			gitRepo.initRemoteGitRepo()
		}
	}

	var err error
	gitRepo.store, err = git.PlainOpen(gitRepo.PathGet())
	if tlog.Error(err) != nil {
		return false
	}

	gitRepo.syncRemote()

	gitRepo.updateBranchHashMap()
	gitRepo.updateGitRepoSize()
	gitRepo.findElementsAndEnvsFiles()

	return true
}

// ===========================================

func (gitRepo *GitRepoS) initLocalGitRepo() (success bool) {

	_, err := git.PlainInit(gitRepo.PathGet(), true)
	if tlog.Error(err) != nil {
		os.RemoveAll(gitRepo.PathGet())
		return false
	}

	// -- Init repo with README.md file --
	filename := "README.md"
	fs := memfs.New()
	gitRepo.store, _ = git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL: gitRepo.PathGet(),
	})

	w, err := gitRepo.store.Worktree()
	if tlog.Error(err) != nil {
		os.RemoveAll(gitRepo.PathGet())
		return false
	}

	// Create README file
	file, err := fs.Create(filename)
	if tlog.Error(err) != nil {
		os.RemoveAll(gitRepo.PathGet())
		return false
	}

	file.Write([]byte("# " + gitRepo.Name))
	file.Close()

	_, err = w.Add(filename)
	if tlog.Error(err) != nil {
		os.RemoveAll(gitRepo.PathGet())
		return false
	}

	_, err = w.Commit("init", &git.CommitOptions{
		Author: &object.Signature{
			Name:  DefaultUser.Name,
			Email: DefaultUser.Email,
			When:  time.Now(),
		},
	})
	if tlog.Error(err) != nil {
		os.RemoveAll(gitRepo.PathGet())
		return false
	}

	// Push to master branch
	err = gitRepo.store.Push(&git.PushOptions{RemoteName: "origin"})
	if tlog.Error(err) != nil {
		os.RemoveAll(gitRepo.PathGet())
		return false
	}

	return true
}

func (gitRepo *GitRepoS) initRemoteGitRepo() (success bool) {

	out, err := utils.ExecCommand("git", "clone", "--mirror", gitRepo.GetRemoteURL(), gitRepo.PathGet())

	if err != nil {
		tlog.Error(out+" "+err.Message, tlog.Vars{
			"logger":     "git",
			"project":    gitRepo.Name,
			"cmd-output": string(out),
			"git-url":    gitRepo.RemoteURL,
		})
		return false
	}
	return true
}

// ===========================================

func (gitRepo *GitRepoS) syncRemote() (success bool) {

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	if gitRepo.RemoteURL == "" {
		return true
	}

	syncDelay := 15
	if syncDelay > 0 && int64(syncDelay)+gitRepo.RemoteLastSync > time.Now().Unix() {
		return true
	}
	if 10+gitRepo.RemoteLastSync > time.Now().Unix() {
		return true
	}

	gitProviderID, _, errx := GetRemoteRepoBranches(gitRepo.GetRemoteURL(), "", "")
	if errx != nil {
		return false
	}

	if gitProviderID != "" && gitRepo.GitProviderID == "" {
		gitRepo.GitProviderID = gitProviderID
		gitRepo.Save()
	}

	utils.ExecCommand(
		"git",
		"--no-pager",
		"-C", gitRepo.PathGet(), // git repo dir
		"remote",
		"set-url",
		"origin",
		gitRepo.GetRemoteURL(),
	)

	remotes, err2 := gitRepo.store.Remotes()
	if tlog.Error(err2, tlog.Vars{
		"project": gitRepo.Name,
	}) != nil {
		return false
	}

	if len(remotes) == 0 {
		tlog.Error("no remotes to fetch", tlog.Vars{
			"project": gitRepo.Name,
		})
		return false
	}

	fetchCMD := []string{
		"git",
		"--no-pager",
		"-C", gitRepo.PathGet(), // git repo dir
		"fetch",
		"--all",
		"-p",
	}

	out, err := utils.ExecCommand(fetchCMD...)

	if err != nil {
		errFull := out + err.Message

		if strings.Contains(errFull, "403") || strings.Contains(errFull, "Authentication failed") {
			EventRemoteGitAccessDenied(gitRepo)
		}

		tlog.Error(errFull, tlog.Vars{
			"logger":     "git",
			"project":    gitRepo.Name,
			"cmd":        strings.Join(fetchCMD, " "),
			"cmd-output": out,
			"git-url":    gitRepo.RemoteURL,
		})

		gitRepo.Error = errFull
		gitRepo.RemoteLastSync = time.Now().Unix() + 120
		gitRepo.Save()
		return false
	}

	gitRepo.Error = ""
	gitRepo.RemoteLastSync = time.Now().Unix()
	gitRepo.Save()

	return true
}

// ===========================================

func (gitRepo *GitRepoS) updateBranchHashMap() (success bool) {

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	cIter, err := gitRepo.store.References()
	if tlog.Error(err) != nil {
		return false
	}

	gitBranchMap := maps.NewSafe[string, string](nil)

	for {
		branch, err := cIter.Next()
		if err == io.EOF || branch == nil {
			break
		}
		if tlog.Error(err) != nil {
			gitRepo.ResourcesUsage.CodeBranches = int64(gitBranchMap.Len())
			return false
		}

		branchName := branch.Name().String()
		if strings.HasPrefix(branchName, "refs/heads/") {
			gitBranchMap.Set(
				strings.TrimPrefix(branchName, "refs/heads/"),
				branch.Hash().String())

		}
		if strings.HasPrefix(branchName, "refs/remotes/origin/") {
			gitBranchMap.Set(
				strings.TrimPrefix(branchName, "refs/remotes/origin/"),
				branch.Hash().String(),
			)
		}
	}

	gitRepo.ResourcesUsage.CodeBranches = int64(gitBranchMap.Len())
	gitRepo.BranchHashMap = gitBranchMap
	return true
}

// ===========================================

func (gitRepo *GitRepoS) updateGitRepoSize() {
	var size int64

	filepath.Walk(gitRepo.PathGet(), func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			tlog.Error(err, tlog.Vars{
				"logger":  "git",
				"project": gitRepo.Name,
			})
			return err
		}

		if info.IsDir() {
			size += 4096

		} else {
			size += info.Size()
		}
		return nil
	})

	gitRepo.ResourcesUsage.CodeStorage = int64(math.Ceil(float64(size) / 1000 / 1000)) // in MB
}

// ===========================================

// GetCommits ...
// limit must be positive
func (gitRepo *GitRepoS) GetCommits(branch string, limit int, from string) []CommitS {

	if gitRepo == nil || gitRepo.store == nil {
		return nil
	}

	// --------------------------------------------------------

	if from == "" {
		branchHead := gitRepo.GetLastCommit(branch)
		if branchHead == "" {
			tlog.Error("branch not found", tlog.Vars{
				"project": gitRepo.Name,
				"branch":  branch,
			})
			return nil
		}
		from = branchHead
	}

	cIter, err := gitRepo.store.Log(&git.LogOptions{
		From:  plumbing.NewHash(from),
		Order: git.LogOrderCommitterTime,
	})
	if tlog.Error(err, tlog.Vars{
		"project": gitRepo.Name,
		"branch":  branch,
	}) != nil {
		return nil
	}

	result := []CommitS{}
	for i := 0; i < limit; i++ {
		commit, err := cIter.Next()
		if err == io.EOF || commit == nil {
			break
		}

		if tlog.Error(err, tlog.Vars{
			"project": gitRepo.Name,
			"branch":  branch,
		}) != nil {
			return nil
		}

		result = append(result, CommitS{
			SHA:            commit.Hash.String(),
			Message:        commit.Message,
			Date:           commit.Author.When,
			TimeStamp:      conv.UnixTimeStamp(commit.Author.When),
			AuthorName:     commit.Author.Name,
			AuthorEmail:    commit.Author.Email,
			AuthorInitials: InitialsFromEmail(commit.Author.Email),
			Files:          []string{},
			// Files:          gitRepo.commitDiffFileList(commitHash, 5),
		})
	}

	return result
}

func (gitRepo *GitRepoS) GetTimoniCommits(branch string, limit int, fromCommit string) []CommitS {

	if gitRepo.store == nil {
		return nil
	}

	// Get latest commit for branch
	if fromCommit == "" {
		branchHead := gitRepo.GetLastCommit(branch)
		if branchHead == "" {
			tlog.Error("branch not found", tlog.Vars{
				"project": gitRepo.Name,
				"branch":  branch,
			})
			return nil
		}
		fromCommit = branchHead
	}

	// Create git iterator
	cIter, err := gitRepo.store.Log(&git.LogOptions{
		From:  plumbing.NewHash(fromCommit),
		Order: git.LogOrderCommitterTime,
	})
	if tlog.Error(err, tlog.Vars{
		"project": gitRepo.Name,
		"branch":  branch,
	}) != nil {
		return nil
	}

	/// Get commits
	result := []CommitS{}
	for i := 0; i < limit; i++ {
		commit, err := cIter.Next()
		if err == io.EOF || commit == nil {
			break
		}

		if tlog.Error(err, tlog.Vars{
			"project": gitRepo.Name,
			"branch":  branch,
		}) != nil {
			return nil
		}

		commitHash := commit.Hash.String()
		result = append(result, CommitS{
			SHA:            commitHash,
			Message:        commit.Message,
			Date:           commit.Author.When,
			TimeStamp:      conv.UnixTimeStamp(commit.Author.When),
			AuthorName:     commit.Author.Name,
			AuthorEmail:    commit.Author.Email,
			AuthorInitials: InitialsFromEmail(commit.Author.Email),
			// Files:          gitRepo.commitDiffFileList(commitHash, 5),
		})
	}

	return result
}

func (gitRepo *GitRepoS) GetCommit(commitHash string) *CommitS {

	if gitRepo == nil {
		return nil
	}

	if gitRepo.store == nil {
		return nil
	}

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	commitObject, err := gitRepo.store.CommitObject(plumbing.NewHash(commitHash))
	if tlog.Error(err, tlog.Vars{
		"project": gitRepo.Name,
		"commit":  commitHash,
	}) != nil {
		return nil
	}

	return &CommitS{
		SHA:            commitObject.Hash.String(),
		Message:        commitObject.Message,
		Date:           commitObject.Author.When,
		TimeStamp:      conv.UnixTimeStamp(commitObject.Author.When),
		AuthorName:     commitObject.Author.Name,
		AuthorEmail:    commitObject.Author.Email,
		AuthorInitials: InitialsFromEmail(commitObject.Author.Email),
		Files:          []string{},
		// Files:          gitRepo.commitDiffFileList(commitHash, 5),
	}
}

func (gitRepo *GitRepoS) commitDiffFileList(commitHash string, limit int) []string {

	if gitRepo.store == nil {
		return nil
	}

	out, err := utils.ExecCommand(
		"git",
		"--no-pager",
		"-C", gitRepo.PathGet(), // git repo dir
		"diff-tree",
		"--name-only", // show only names of changed files
		"--no-commit-id",
		"--root", // include the initial commit as diff against /dev/null
		"-r",     // diff recursively
		"-m",     // show all changes in merge commits
		commitHash,
	)

	if err != nil {
		tlog.Error(out + err.Message)
		return []string{}
	}

	s := strings.TrimSpace(string(out))
	if s == "" {
		return []string{}
	}

	r := strings.Split(s, "\n")
	if limit == 0 || len(r) <= limit {
		return r
	}
	return r[:limit]
}

func (gitRepo *GitRepoS) branchesDiffFileList(branch1 string, branch2 string) []string {

	if gitRepo.store == nil {
		return nil
	}

	out, err := utils.ExecCommand(
		"git",
		"--no-pager",
		"-C", gitRepo.PathGet(), // git repo dir
		"diff-tree",
		"--name-only", // show only names of changed files
		"--no-commit-id",
		"--root", // include the initial commit as diff against /dev/null
		"-r",     // diff recursively
		branch1,
		branch2,
	)

	if err != nil {
		tlog.Error(out + err.Message)
		return []string{}
	}

	s := strings.TrimSpace(string(out))
	if s == "" {
		return []string{}
	}

	r := strings.Split(s, "\n")
	return r
}

// CommitDiffFileList files
func (gitRepo *GitRepoS) CommitDiffFileList(commitHash string, limit int) []string {

	if gitRepo.store == nil {
		return nil
	}

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	return gitRepo.commitDiffFileList(commitHash, limit)
}

// BranchesDiffFileList files
func (gitRepo *GitRepoS) BranchesDiffFileList(branch1 string, branch2 string) []string {

	if gitRepo.store == nil {
		return nil
	}

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	return gitRepo.branchesDiffFileList(branch1, branch2)
}

func (gitRepo *GitRepoS) GetCommitObject(commitHash string) *object.Commit {

	// input string, byHash bool

	if gitRepo.store == nil {
		return nil
	}

	if commitHash == "" {
		tlog.Error("commitHash is empty", tlog.Vars{
			"project": gitRepo.Name,
		})
		return nil
	}

	var commit *object.Commit
	var err error

	commit, err = gitRepo.store.CommitObject(plumbing.NewHash(commitHash)) // get commit directly
	if tlog.Error(err, tlog.Vars{
		"project":    gitRepo.Name,
		"commitHash": commitHash,
	}) != nil {
		return nil
	}

	return commit
}

func (gitRepo *GitRepoS) FileTreeOfCommit(commitHash, branchName, path string) (*object.Tree, *tlog.RecordS) {

	if gitRepo.store == nil {
		return nil, tlog.Error("gitRepo.store == nil")
	}

	path = strings.TrimPrefix(path, "/")

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	if branchName == "" {
		return nil, tlog.Error("branchName is empty")
	}

	if commitHash == "" {
		commitHash = gitRepo.GetLastCommit(branchName)
	}

	commit := gitRepo.GetCommitObject(commitHash)
	if commit == nil {
		return nil, tlog.Error("commit == nil")
	}

	startFileTree, err := commit.Tree()
	if err != nil {
		return nil, tlog.Error(err, tlog.Vars{
			"repo":   gitRepo.Name,
			"branch": branchName,
			"commit": commitHash,
			"path":   path,
		})
	}

	var fileTree *object.Tree
	// if directory is specified get its content
	if path == "" {
		// use top level directory
		fileTree = startFileTree

	} else {
		fileTree, err = startFileTree.Tree(path)
		if err != nil {
			return nil, tlog.Error(err, tlog.Vars{
				"repo":   gitRepo.Name,
				"branch": branchName,
				"commit": commitHash,
				"path":   path,
			})
		}
	}

	return fileTree, nil
}

func (gitRepo *GitRepoS) FileListOfCommit(commitHash, branchName, directory string) (res []*object.File) {

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	fileTree, _ := gitRepo.FileTreeOfCommit(commitHash, branchName, directory)
	if fileTree == nil {
		return
	}

	iter := fileTree.Files()
	defer iter.Close()

	for {
		f, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				return res
			}

			tlog.Error(err)
			return res
		}

		res = append(res, f)
	}
}

func (gitRepo *GitRepoS) GetFile(branch, commit, filePath string) []byte {

	if gitRepo.store == nil {
		return nil
	}

	if filePath == "" {
		return nil
	}
	if filePath[0] == '/' {
		filePath = filePath[1:]
	}

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	if commit == "" {
		commit = gitRepo.GetLastCommit(branch) // get latest commit
	}

	commitObject, err := gitRepo.store.CommitObject(plumbing.NewHash(commit))
	if tlog.Error(err, tlog.Vars{
		"project":  gitRepo.Name,
		"branch":   branch,
		"commit":   commit,
		"filePath": filePath,
	}) != nil {
		return nil
	}

	fileObject, err := commitObject.File(filePath)
	if err != nil {
		tlog.Error("file not found", tlog.Vars{
			"project":  gitRepo.Name,
			"branch":   branch,
			"commit":   commit,
			"filePath": filePath,
		})
		return nil
	}

	reader, err := fileObject.Reader()
	if tlog.Error(err, tlog.Vars{
		"project":  gitRepo.Name,
		"branch":   branch,
		"commit":   commit,
		"filePath": filePath,
	}) != nil {
		return nil
	}
	defer reader.Close()

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(reader)
	if tlog.Error(err, tlog.Vars{
		"project":  gitRepo.Name,
		"branch":   branch,
		"commit":   commit,
		"filePath": filePath,
	}) != nil {
		return nil
	}

	return buf.Bytes()
}

func (gitRepo *GitRepoS) GetAllowBranches(user *UserS, level string) map[string]bool {

	if gitRepo.store == nil {
		return nil
	}

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	res := make(map[string]bool)
	for _, branchName := range gitRepo.BranchHashMap.Keys() {
		res[branchName] = true
	}

	return res
}

func (gitRepo *GitRepoS) GetAllowTags(user *UserS, level string) map[string]GitTagS { // key=tagName, value=tagCommitSHA

	if gitRepo.store == nil {
		return nil
	}

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	tagsIter, _ := gitRepo.store.TagObjects()
	res := make(map[string]GitTagS)

	for {
		tag, err := tagsIter.Next()
		if err == io.EOF || tag == nil {
			break
		}

		commit, _ := tag.Commit()
		if commit == nil {
			continue
		}

		gt := GitTagS{
			Name:                 tag.Name,
			CommitSHA:            tag.Target.String(),
			CommitAuthorEmail:    commit.Author.Email,
			CommitAuthorInitials: InitialsFromEmail(commit.Author.Email),
			TimeStamp:            commit.Author.When.Unix(),
		}
		res[gt.Name] = gt
	}

	return res
}

func (git *GitRepoS) PathGet() string {

	if git.RemoteURL != "" {
		return filepath.Join(config.GitRemotePath, git.Name)
	}
	return filepath.Join(config.GitLocalPath, git.Name)

}

type ElementInUsageTmpS struct {
	EnvID       string
	EnvName     string
	ElementName string
}

func (gitRepo *GitRepoS) UseOfElements() map[string][]ElementInUsageTmpS {

	res := map[string][]ElementInUsageTmpS{}

	for _, env := range EnvironmentMap.Values() {
		for _, elName := range env.Elements.Keys() {
			element := env.GetElement(elName)
			if element.GetSource().RepoName == gitRepo.Name {

				key := element.GetSource().BranchName + "/" + element.GetSource().FilePath
				res[key] = append(res[key], ElementInUsageTmpS{
					EnvID:       env.ID,
					EnvName:     env.Name,
					ElementName: elName,
				})

			}
		}
	}

	return res
}

// ===========================================

func (gitRepo *GitRepoS) GetCommitFileDiff(commitHash, filePath string) (string, *tlog.RecordS) {
	cmd := []string{
		"git",
		"--no-pager",
		"-C",
		gitRepo.PathGet(), // git repo dir
		"log",
		"-p", // Generate patch
		"-1", // only one commit
		commitHash,
		"--",
		filePath,
	}

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	out, err := utils.ExecCommand(cmd...)
	if err != nil {
		return "", tlog.Error("Error while getting commit diff from git server", out+" "+err.Message)
	}
	return out, nil
}

func (gitRepo *GitRepoS) GetBranchesFileDiff(branch1, branch2, filePath string) (string, *tlog.RecordS) {
	cmd := []string{
		"git",
		"--no-pager",
		"-C", gitRepo.PathGet(), // git repo dir
		"diff", "-p", // Generate patch
		branch1,
		branch2,
		"--",
		filePath,
	}

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	out, err := utils.ExecCommand(cmd...)
	if err != nil {
		return "", tlog.Error("Error while getting branch diff from git server", out+" "+err.Message)
	}
	return out, nil
}

func (gitRepo *GitRepoS) CompareBranches(branch1, branch2 string) (string, *tlog.RecordS) {
	cmd := []string{
		"git",
		"-C", gitRepo.PathGet(),
		"rev-list",
		"--left-right",
		"--count",
		branch1 + "..." + branch2,
	}

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}
	out, err := utils.ExecCommand(cmd...)
	if err != nil {
		return "", tlog.Error("Error while comparing branches", out+" "+err.Message)
	}
	return out, nil
}

func (gitRepo *GitRepoS) GetDiffFromCommit(sha string) (string, *tlog.RecordS) {
	cmd := []string{
		"git",
		"--no-pager",
		"-C", gitRepo.PathGet(), // git repo dir
		"diff",
		"--shortstat",
		"HEAD^",
		"HEAD",
		sha,
	}

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	out, err := utils.ExecCommand(cmd...)
	if err != nil {
		return "", tlog.Error("Error while comparing HEADs", out+" "+err.Message)
	}
	return out, nil
}

func (gitRepo *GitRepoS) GetCommitsPerUser(branch string, begin, end int64) map[string]int {
	commitsPerUser := make(map[string]int)
	for _, commit := range gitRepo.GetCommits(branch, 9999, "") {
		t := time.Unix(commit.TimeStamp, 0)
		if t.After(time.Unix(begin, 0)) && t.Before(time.Unix(end, 0)) {
			commitsPerUser[fmt.Sprintf("%s (%s)", commit.AuthorName, commit.AuthorEmail)]++
		}
	}

	return commitsPerUser
}

func (gitRepo *GitRepoS) GetCommitsPerWeekDay(branch string, begin, end int64) map[string]int {
	commitsPerDay := make(map[string]int)
	for _, commit := range gitRepo.GetCommits(branch, 9999, "") {
		t := time.Unix(commit.TimeStamp, 0)
		if t.After(time.Unix(begin, 0)) && t.Before(time.Unix(end, 0)) {
			commitsPerDay[commit.Date.Weekday().String()]++
		}
	}

	return commitsPerDay
}

func (gitRepo *GitRepoS) GetCommitsPerMonth(branch string, begin, end int64) map[string]int {
	commitsPerMonth := make(map[string]int)
	for _, commit := range gitRepo.GetCommits(branch, 9999, "") {
		t := time.Unix(commit.TimeStamp, 0)
		if t.After(time.Unix(begin, 0)) && t.Before(time.Unix(end, 0)) {
			commitsPerMonth[commit.Date.Month().String()]++
		}
	}

	return commitsPerMonth
}

func (gitRepo *GitRepoS) GetCommitsPerYear(branch string, begin, end int64) map[string]int {
	commitsPerYear := make(map[string]int)
	for _, commit := range gitRepo.GetCommits(branch, 9999, "") {
		t := time.Unix(commit.TimeStamp, 0)
		if t.After(time.Unix(begin, 0)) && t.Before(time.Unix(end, 0)) {
			commitsPerYear[fmt.Sprint(commit.Date.Year())]++
		}
	}

	return commitsPerYear
}

func (gitRepo *GitRepoS) GetUserStats(branch string, begin, end int64) map[string]map[string]int {
	usersStats := make(map[string]map[string]int)

	commits := gitRepo.GetCommits(branch, 9999, "")
	for _, commit := range commits {
		t := time.Unix(commit.TimeStamp, 0)
		if t.After(time.Unix(begin, 0)) && t.Before(time.Unix(end, 0)) {

			if usersStats[commit.AuthorName] == nil {
				usersStats[commit.AuthorName] = map[string]int{}
			}

			usersStats[commit.AuthorName]["Commits"]++
			usersStats[commit.AuthorName]["Files"] += len(commit.Files)
		}
	}

	return usersStats
}

// -----------------------------------

func (gitRepo *GitRepoS) findElementsAndEnvsFiles() {

	// tlog.Debug("looking for elements in git repo: " + gitRepo.Name)

	if !gitRepo.mutexLock {
		tlog.Fatal("need gitRepo.Lock()")
	}

	for branch := range gitRepo.BranchHashMap.Iter() {

		branchName := branch.Key
		latestCommitHash := branch.Value

		cacheKey := gitRepo.Name + "/" + branchName
		cache := ElementsInGitRepoCache.Get(cacheKey)
		if cache.LastCommitChecked == latestCommitHash {
			// nic sie nie zmienilo od ostatniego sprawdzenia
			continue
		}

		// tlog.Debug("looking for elements in git repo: " + gitRepoName + " branch: " + branchName)
		elements := []*GitElementS{}
		envs := maps.NewSafe[string, *GitEnvS](nil)

		for _, file := range gitRepo.FileListOfCommit(latestCommitHash, branchName, "timoni") {
			el, env := createGitElement(gitRepo, branchName, latestCommitHash, "timoni", file)
			if el != nil {
				elements = append(elements, el)
			}
			if env != nil {
				envs.Set(env.Source.FilePath, env)
			}
		}

		buf := gitRepo.GetFile(branchName, latestCommitHash, ".timoni")
		if len(buf) > 0 {
			if strings.TrimSpace(string(buf)) == "*" {
				tlog.Info("looking for elements in whole git repo: " + gitRepo.Name)
				for _, file := range gitRepo.FileListOfCommit(latestCommitHash, branchName, "") {
					el, env := createGitElement(gitRepo, branchName, latestCommitHash, "", file)
					if el != nil {
						elements = append(elements, el)
					}
					if env != nil {
						envs.Set(env.Source.FilePath, env)
					}
				}
			}
		}

		ElementsInGitRepoCache.Set(cacheKey, ElementsInGitRepoCacheS{
			LastCommitChecked: latestCommitHash,
			Elements:          elements,
		})
		EnvInGitRepoCache.Set(cacheKey, EnvInGitRepoCacheS{
			LastCommitChecked: latestCommitHash,
			Environments:      envs,
		})

		// Trigger update itself locks gitRepo
		gitRepo.Unlock()
		gitRepo.triggerUpdate(branchName, latestCommitHash)
		gitRepo.Open()
	}
}

func (gitRepo *GitRepoS) triggerUpdate(branch, latestCommit string) {

	for _, env := range EnvironmentMap.Values() {
		for _, elName := range env.Elements.Keys() {
			element := env.GetElement(elName)
			if element.GetSource().RepoName != gitRepo.Name {
				continue
			}
			if element.GetSource().BranchName != branch {
				continue
			}
			if !element.GetAutoUpdate() {
				element.GetStatus().NewerVersion = latestCommit != element.GetSource().CommitHash
				continue
			}

			env.SetElementVersion(elName, branch, latestCommit, nil)
		}
	}
}

func createGitElement(gitRepo *GitRepoS, branchName, commitHash, directory string, file *object.File) (*GitElementS, *GitEnvS) {

	if !strings.HasSuffix(file.Name, ".toml") {
		return nil, nil
	}

	path := filepath.Join("/", directory, file.Name)

	id := base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(
		gitRepo.Name + "|" +
			branchName + "|" +
			commitHash + "|" +
			path,
	)))

	el := &GitElementS{
		ID:   id,
		Name: filepath.Base(file.Name[:len(file.Name)-5]),
		Source: SourceGitS{
			RepoName:   gitRepo.Name,
			BranchName: branchName,
			CommitHash: commitHash,
			CommitTime: gitRepo.GetCommit(commitHash).TimeStamp,
			FilePath:   path,
		},
		Type: ElementSourceTypeUnknown,
		// Description   string
		// Favorite      bool // personalizowane per operator
		// UsageCount    int  // w ilu miejscach aktualnie ten element jest uzyty we wszystkich sr tej instalacji
		// UsageTime     int  // UsageCount * czas jego używania historycznie od początku
	}

	tmp, err := file.Contents()
	el.FileContent = []byte(tmp)
	if err != nil {
		el.Error = "Reading file: " + err.Error()

	} else {
		el.Validate()
	}

	if el.Type == ElementSourceTypeEnv {
		env := &GitEnvS{
			Name:        el.Name,
			Source:      el.Source,
			FileContent: el.FileContent,
		}
		env.Validate()
		return nil, env
	}

	ElementsInGitRepoFlatMap.Set(el.ID, el)
	return el, nil
}
