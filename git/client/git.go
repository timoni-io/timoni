package gitc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	log "lib/tlog"
	"git/structs"
	"lib/utils"
	"lib/utils/cmd"
	"lib/utils/iter"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrPathAlreadyExists = errors.New("path already exists")
	ErrPathNotExists     = errors.New("path not exists")
	ErrIncorrectArgs     = errors.New("incorrect args")
)

type Git struct {
	ServerUrl string
	Dir       string // Base path
	Repo      string // Repo name

	GlobalOptions cmd.RunOptions
}

func (g *Git) url() string {
	return fmt.Sprintf(
		"%s/%s",
		strings.TrimSuffix(g.ServerUrl, "/"),
		g.Repo,
	)
}

func (g *Git) Path() string {
	return filepath.Join(g.Dir, g.Repo)
}

func (g *Git) Clone() error {
	repoPath := g.Path()
	if utils.PathExists(repoPath) {
		return ErrPathAlreadyExists
	}
	return cmd.
		NewCommand(
			"git",
			"clone",
			g.url(),
			repoPath,
		).Run(&g.GlobalOptions)
}

func (g *Git) CloneCommit(commit string) error {
	repoPath := g.Path()
	if utils.PathExists(repoPath) {
		return ErrPathAlreadyExists
	}

	// https://stackoverflow.com/a/43136160
	// git init
	// git remote add origin <url>
	// git fetch --depth 1 origin <sha>
	// git checkout FETCH_HEAD

	// when `fetch --depth 1` not avaliable (server side) then use standard clone
	fallback := func() error {
		if err := os.RemoveAll(repoPath); err != nil {
			return err
		}
		if err := g.Clone(); err != nil {
			return err
		}
		if err := g.ResetHard(commit); err != nil {
			return err
		}
		return nil
	}

	opts := g.GlobalOptions
	opts.Dir = repoPath

	var err error
	if err = os.MkdirAll(repoPath, 0755); err != nil {
		return err
	}
	if err = cmd.NewCommand("git", "init", "-b", "main").Run(&opts); err != nil {
		log.Debug("CloneCommit: Fallback after init", err)
		return fallback()
	}
	if err = cmd.NewCommand("git", "remote", "add", "origin", g.url()).Run(&opts); err != nil {
		log.Debug("CloneCommit: Fallback after remote add", err)
		return fallback()
	}
	if err = cmd.NewCommand("git", "fetch", "--depth", "1", "origin", commit).Run(&opts); err != nil {
		log.Debug("CloneCommit: Fallback after fetch --depth 1", err)
		return fallback()
	}
	if err = cmd.NewCommand("git", "checkout", "FETCH_HEAD").Run(&opts); err != nil {
		log.Debug("CloneCommit: Fallback after checkout FETCH_HEAD", err)
		return fallback()
	}

	return nil
}

func (g *Git) Fetch() error {
	opts := g.GlobalOptions
	opts.Dir = g.Path()
	return cmd.
		NewCommand("git", "fetch", "--all").
		Run(&opts)
}

func (g *Git) Checkout(commit string) error {
	opts := g.GlobalOptions
	opts.Dir = g.Path()
	return cmd.
		NewCommand("git", "checkout", commit, "--").
		Run(&opts)
}

func (g *Git) ResetHard(commit string) error {
	opts := g.GlobalOptions
	opts.Dir = g.Path()
	return cmd.
		NewCommand("git", "reset", "--hard", commit, "--").
		Run(&opts)
}

func (g *Git) CommitCount(since string) (string, error) {
	stdout, stderr, err := cmd.
		NewCommand("git", "rev-list", "--count", since, "--").
		RunStdString(&cmd.RunOptions{
			Dir: g.Path(),
		})
	if stderr != "" {
		err = errors.New(stderr)
	}
	return stdout, err
}

func (g *Git) InitBare() error {
	path := g.Path()
	if utils.PathExists(path) {
		return ErrPathAlreadyExists
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	return cmd.NewCommand("git", "init", "--bare").
		Run(&cmd.RunOptions{
			Dir: path,
		})
}

func (g *Git) Service(service string) ([]byte, error) {
	if !utils.PathExists(g.Path()) {
		return nil, ErrPathNotExists
	}

	// git service requires path as last argument
	stdout, stderr, err := cmd.NewCommand("git", service, "--stateless-rpc", "--advertise-refs", ".").
		RunStdBytes(&cmd.RunOptions{
			Dir: g.Path(),
		})
	return append(stdout, stderr...), err
}

func (g *Git) UploadPack(stdin io.Reader) ([]byte, error) {
	// git upload-pack requires path as last argument
	stdout, stderr, err := cmd.NewCommand("git", "upload-pack", "--stateless-rpc", ".").
		RunStdBytes(&cmd.RunOptions{
			Dir:   g.Path(),
			Stdin: stdin,
		})

	return append(stdout, stderr...), err
}

func (g *Git) ReceivePack(stdin io.Reader) ([]byte, error) {
	// git receive-pack requires path as last argument
	stdout, stderr, err := cmd.NewCommand("git", "receive-pack", "--stateless-rpc", ".").
		RunStdBytes(&cmd.RunOptions{
			Dir:   g.Path(),
			Stdin: stdin,
		})

	return append(stdout, stderr...), err
}

func (g *Git) Branches() ([]string, error) {
	out, _, err := cmd.NewCommand("git", "branch").
		RunStdString(&cmd.RunOptions{
			Dir:          g.Path(),
			DisablePager: true,
		})

	return iter.MapSlice(strings.Split(out, "\n"), func(s string) string {
		return strings.TrimSpace(strings.TrimLeft(s, "* "))
	}), err
}

// branch can be empty
func (g *Git) Log(branch string) ([]*git.GitLog, error) {
	args := []string{"git", "log", "--format=" + PrettyJson[git.GitLog]()}
	if branch != "" {
		args = append(args, "-b", branch)
	}

	data, _, err := cmd.
		NewCommand(args[0], args[1:]...).
		RunStdBytes(&cmd.RunOptions{
			Dir:          g.Path(),
			DisablePager: true,
		})

	if err != nil {
		return nil, err
	}

	// Format output as json array
	data = []byte(fmt.Sprintf("[%s]", string(data[:len(data)-1])))

	gitLogs := []*git.GitLog{}
	err = json.Unmarshal(data, &gitLogs)
	if err != nil {
		return nil, err
	}
	return gitLogs, nil
}

func (g *Git) Diff(diff git.CommitDiff) ([]*git.GitDiff, error) {
	if diff.A == "" || diff.B == "" {
		return nil, ErrIncorrectArgs
	}
	data, _, err := cmd.
		NewCommand("git", "show", "--name-status", "--format=", diff.A, diff.B).
		RunStdString(&cmd.RunOptions{
			Dir:          g.Path(),
			DisablePager: true,
		})

	if err != nil {
		return nil, err
	}

	return iter.MapSlice(strings.Split(data, "\n"), func(line string) *git.GitDiff {
		status, file, _ := strings.Cut(line, "\t")
		return &git.GitDiff{
			Filename: file,
			Status:   git.GitStatus(status),
		}
	}), nil
}

func (g *Git) Files(commit string) ([]string, error) {
	if commit == "" {
		return nil, ErrIncorrectArgs
	}
	data, _, err := cmd.
		NewCommand("git", "ls-tree", "--name-only", "-r", commit).
		RunStdString(&cmd.RunOptions{
			Dir:          g.Path(),
			DisablePager: true,
		})

	if err != nil {
		return nil, err
	}

	return strings.Split(data, "\n"), nil
}

func (g *Git) OpenFile(commit, filename string) (string, error) {
	if commit == "" || filename == "" {
		return "", ErrIncorrectArgs
	}
	data, _, err := cmd.
		NewCommand("git", "show", fmt.Sprintf("%s:%s", commit, filename)).
		RunStdString(&cmd.RunOptions{
			Dir:          g.Path(),
			DisablePager: true,
		})
	return data, err
}
