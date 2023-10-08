package db

import (
	"core/config"
	"encoding/base64"
	"encoding/json"
	"lib/tlog"
	"os"
	"path/filepath"
)

type EnvDynamicGitSourceS struct {
	ID           string
	RepoName     string
	BranchName   string
	DirPath      string
	Teams        []string
	Environments map[string]*EnvDynamicTargetS // key=file-path in git-repo
}

type EnvDynamicTargetS struct {
	ID   string
	Name string
}

var environmentDynamicGitSourcesLoad = false

func EnvironmentDynamicGitSourcesRefresh() {

	if !environmentDynamicGitSourcesLoad {
		buf, _ := os.ReadFile(filepath.Join(config.DataPath(), "dynamic-envs.json"))
		json.Unmarshal(buf, EnvironmentDynamicGitSources)
	}

	for _, des := range EnvironmentDynamicGitSources.Values() {
		key := des.RepoName + "/" + des.BranchName

		for _, filePath := range EnvInGitRepoCache.Get(key).Environments.Keys() {
			if des.Environments[filePath] != nil {
				continue
			}

			name := filepath.Base(filePath)
			name = name[:len(name)-5]
			env := EnvironmentS{
				Name: name,
			}

			env.Create(nil)
			env.GitOps = EnvironmentGitOpsConfigS{
				Enabled:            true,
				DynamicEnvironment: true,
				GitRepoName:        des.RepoName,
				BranchName:         des.BranchName,
				FilePath:           filePath,
			}
			env.Save(nil)

			if des.Environments == nil {
				des.Environments = map[string]*EnvDynamicTargetS{}
			}
			des.Environments[filePath] = &EnvDynamicTargetS{
				ID:   env.ID,
				Name: env.Name,
			}
			EnvironmentDynamicGitSourcesSave()

		}
	}
}

func EnvironmentDynamicGitSourcesSave() {
	buf, _ := json.MarshalIndent(EnvironmentDynamicGitSources, "", "  ")
	tlog.Error(os.WriteFile(
		filepath.Join(config.DataPath(), "dynamic-envs.json"),
		buf,
		0644,
	))
}

func (env *EnvDynamicGitSourceS) Save(user *UserS) *tlog.RecordS {

	// ---
	// GenerateID:

	if env.ID == "" {
		env.ID = base64.StdEncoding.EncodeToString([]byte(env.RepoName + "|" + env.BranchName + "|" + env.DirPath))
	}

	// ---

	EnvironmentDynamicGitSources.Set(env.ID, env)
	EnvironmentDynamicGitSourcesSave()
	return nil
}

func (des *EnvDynamicGitSourceS) Delete(user *UserS) *tlog.RecordS {

	if des == nil {
		return nil
	}

	for _, envData := range des.Environments {
		env := EnvironmentMap.Get(envData.ID)
		env.SetToDelete(true, user)
	}

	EnvironmentDynamicGitSources.Delete(des.ID)
	EnvironmentDynamicGitSourcesSave()
	return nil
}
