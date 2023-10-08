package api

import (
	"core/db"
	"net/http"
	"strings"
)

func apiGitElementsList(r *http.Request, user *db.UserS) interface{} {

	filter := strings.ToLower(r.FormValue("filter"))
	typ := r.FormValue("type")

	res := []*db.GitElementS{}
	for _, cache := range db.ElementsInGitRepoCache.Values() {
		for _, element := range cache.Elements {
			if element.Error != "" {
				continue
			}

			if filterMatch(element, filter, typ) {
				res = append(res, element)
			}
		}
	}

	return res
}

func filterMatch(el *db.GitElementS, filterText, filterType string) bool {
	if filterType != "" && string(el.Type) != filterType {
		return false
	}

	return strings.Contains(el.Name, filterText) ||
		strings.Contains(el.Source.RepoName, filterText) ||
		strings.Contains(el.Source.BranchName, filterText) ||
		strings.Contains(el.Description, filterText)
}
