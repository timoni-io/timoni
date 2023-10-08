package gitprovider

import (
	"core/db"
)

func Setup() {
	for _, gitRepoName := range db.GitRepoMap().Keys() {
		gitRepo := db.GitRepoGetByName(gitRepoName)
		gitRepo.Open()
		gitRepo.Unlock()
	}
}
