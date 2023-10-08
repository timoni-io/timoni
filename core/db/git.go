package db

import (
	"core/db2"
	"fmt"
	"lib/utils"
	"lib/utils/conv"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	log "lib/tlog"
)

func GetRemoteRepoBranches(uri, login, password string) (string, []string, *log.RecordS) {
	// 1. repo is public
	out, err := utils.ExecCommand("git", "ls-remote", "--heads", uri)
	if err == nil {
		// public repo
		return "", getBranches(out), nil
	}

	// 2. repo is private and we have access
	if login != "" && password != "" {

		branches, res := getRemoteRepoBranchesPassword(uri, login, password)
		if res != nil {
			return "", branches, res
		}
		// trzeba stworzyc nowego gitProvider'a
		gp := db2.GitProviderCreate(
			uri,
			db2.TheOrganization,
			db2.GitProviderVariant_LoginAndPassword,
		)
		gp.SetLogin(login)
		gp.SetLogin(password)

		return gp.ID(), branches, res
	}

	// 3. repo is private, and access must be provided by git-provider
	for _, gp := range db2.GitProviderList("Variant = 1", "", 0, 100).Iter() {

		branches, res := getRemoteRepoBranchesPassword(uri, gp.Login(), gp.Password())
		if res == nil {
			return gp.ID(), branches, res
		}
	}

	for _, gp := range db2.GitProviderList("Variant = 2", "", 0, 100).Iter() {

		branches, res := getRemoteRepoBranchesSSHKey(uri, gp.PrivateKey(), gp.PublicKey())
		if res == nil {
			fmt.Println("getRemoteRepoBranchesSSHKey", gp.ID())
			return gp.ID(), branches, res
		}
	}

	return "", nil, log.Error("login and password are required")
}

func getRemoteRepoBranchesPassword(uri, login, password string) ([]string, *log.RecordS) {

	// build url https://user:password@github.com/...
	tmp := strings.Split(uri, "://")
	uri = tmp[0] + "://" + login + ":" + url.QueryEscape(password) + "@" + strings.Join(tmp[1:], "")

	// check access with login and password
	out, err := utils.ExecCommand("git", "ls-remote", "--heads", uri)
	if err != nil {
		log.Error("remote git repo git ls-remote fail", log.Vars{
			"exit":       err,
			"cmd-output": out,
			"url":        uri,
		})
		return nil, log.Error("login and password are invalid")
	}

	return getBranches(out), nil
}

func getRemoteRepoBranchesSSHKey(uri, PrivateKey, PublicKey string) ([]string, *log.RecordS) {

	os.WriteFile("/data/ssh_keys", []byte(PrivateKey), 0600)
	os.WriteFile("/data/ssh_keys.pub", []byte(PublicKey), 0600)

	out, err := utils.ExecCommand("git", "ls-remote", "--heads", uri)
	if err != nil {
		log.Error("remote git repo git ls-remote fail", log.Vars{
			"exit":       err,
			"cmd-output": out,
			"url":        uri,
		})
		return nil, log.Error("login and password are invalid")
	}

	return getBranches(out), nil
}

func getBranches(cmdOut string) []string {
	branches := []string{}

	splitted := strings.Split(cmdOut, "\n")
	length := len(splitted)
	for i, line := range splitted {
		if i == length-1 {
			break
		}
		splitted := strings.Split(line, "refs/heads/")
		if len(splitted) > 1 {
			branches = append(branches, splitted[1])
		}
	}
	return branches
}

func GetCommitCount(gitPath, projectName, commit string) string {
	cmd := []string{
		"git",
		"--no-pager",
		"-C", filepath.Join(gitPath, projectName), // git repo dir
		"rev-list",
		"--count",
		commit,
	}

	out, err := utils.ExecCommand(cmd...)

	if err != nil {
		log.Error(out+" "+err.Message, log.Vars{
			"gitRepo": projectName,
		})
		return ""
	}

	return strings.TrimSpace(out)
}

func GetCommitTime(gitPath, commit string) int64 {
	cmd := []string{
		"git",
		"--no-pager",
		"-C", gitPath, // git repo dir
		"show",
		"-s", "--format=%ct",
		commit,
	}

	out, err := utils.ExecCommand(cmd...)

	if err != nil {
		log.Error(err)
		return 0
	}

	return conv.ToInt64(strings.TrimSpace(out))
}
