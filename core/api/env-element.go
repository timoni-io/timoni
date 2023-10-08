package api

import (
	"encoding/json"
	"fmt"
	"html"
	"io"

	"core/db"
	perms "core/db/permissions"
	"lib/tlog"
	"net/http"
	"strconv"
)

type tmpDataS struct {
	Name      string
	EnvID     string
	GitID     string
	DontStart bool
}

func apiEnvironmentElementCreateFromGitRepo(r *http.Request, user *db.UserS) interface{} {

	data := &tmpDataS{}
	err := json.NewDecoder(r.Body).Decode(data)
	if tlog.Error(err) != nil {
		return tlog.Error("cant read POST body JSON")
	}

	if data.EnvID == "" {
		return tlog.Error("`EnvID` is required")
	}
	env := db.EnvironmentMap.Get(data.EnvID)
	if env == nil {
		return tlog.Error("envNotFound")
	}

	if !user.HasEnvPerm(data.EnvID, perms.Env_ElementFullManage) {
		return tlog.Error("permissionDenied")
	}

	if data.Name == "" {
		return tlog.Error("`Name` is required")
	}

	if data.GitID == "" {
		return tlog.Error("`GitID` is required")
	}

	// --------------------

	if env.GitOps.Enabled {
		return tlog.Error("gitOpsEnabled")
	}

	// --------------------

	if err := env.ElementAdd(
		data.Name,
		db.ElementFromGitMap(data.Name, data.EnvID, data.GitID),
		user,
		!data.DontStart,
	); err != nil {
		return tlog.Error(err)
	}

	tlog.Info("element added "+data.Name, tlog.Vars{
		"env":     env.ID,
		"element": data.Name,
		"event":   true,
	})

	return "ok"
}

func apiEnvironmentElementCreateFromTOML(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("Param `env` is required")
	}
	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		tlog.Error("środowisko nie zostało odnalezione", tlog.Vars{
			"env": envID,
		})
		return tlog.Error("środowisko nie zostało odnalezione: " + envID)
	}

	if !user.HasEnvPerm(envID, perms.Env_ElementFullManage) {
		return tlog.Error("permission denied")
	}

	// --------------------

	if env.GitOps.Enabled {
		return tlog.Error("GitOps is enabled")
	}

	// --------------------

	elementName := r.FormValue("element")
	dontStart := r.FormValue("dont-start") == "true"

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return tlog.Error(err)
	}

	el, err2 := db.ElementFromToml(elementName, env.ID, buf, db.SourceGitS{})
	if err2 != nil {
		return err2
	}

	err2 = env.ElementAdd(
		elementName,
		el,
		user,
		!dontStart,
	)
	if err2 != nil {
		return err2
	}

	tlog.Info("element added "+elementName, tlog.Vars{
		"env":     env.ID,
		"element": elementName,
		"event":   true,
	})

	return "ok"
}

func apiEnvironmentElementUpdateFromTOML(r *http.Request, user *db.UserS) interface{} {
	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("Param `env` is required")
	}
	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		tlog.Error("środowisko nie zostało odnalezione", tlog.Vars{
			"env": envID,
		})
		return tlog.Error("środowisko nie zostało odnalezione: " + envID)
	}

	if !user.HasEnvPerm(envID, perms.Env_ElementFullManage) {
		return tlog.Error("permission denied")
	}

	// --------------------

	if env.GitOps.Enabled {
		return tlog.Error("GitOps is enabled")
	}

	// --------------------

	elementName := r.FormValue("element")

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return tlog.Error(err)
	}

	orgElement := env.ElementLoad(elementName)
	if orgElement == nil {
		return tlog.Error("element does not exist")
	}

	newElement, err2 := db.ElementFromToml(elementName, env.ID, buf, orgElement.GetSource())
	if err2 != nil {
		return err2
	}

	currentElement, ok := db.ElementMap.GetFull(fmt.Sprintf("%s/%s", env.ID, elementName))
	if !ok {
		return tlog.Error("element does not exist")
	}

	err2 = currentElement.CopySecrets(newElement)
	if err2 != nil {
		return err2
	}

	err2 = db.ElementUpdate(
		orgElement,
		newElement,
		user,
		true,
	)
	if err2 != nil {
		return err2
	}
	tlog.Info("element update", tlog.Vars{
		"env":   env.ID,
		"event": true,
		"user":  user.Email,
	})

	return "ok"
}

type ElementMapToFrontS struct {
	Info   interface{}
	Status db.ElementStatusS
}

func apiEnvironmentElementMap(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("Param `env` is required")
	}
	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		tlog.Error("środowisko nie zostało odnalezione", tlog.Vars{
			"env": envID,
		})
		return tlog.Error("środowisko nie zostało odnalezione: " + envID)
	}

	if !user.HasEnvPerm(envID, perms.Env_View) {
		return tlog.Error("permission denied")
	}

	elements := map[string]ElementMapToFrontS{}
	for _, elName := range env.Elements.Keys() {
		element := env.ElementCloneWithoutSecrets(elName)
		elements[elName] = ElementMapToFrontS{
			Info:   element,
			Status: *element.GetStatus(),
		}
	}

	return elements
}

func apiEnvironmentElementVersionMap(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("Param `env` is required")
	}
	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		tlog.Error("environment not found", tlog.Vars{
			"env": envID,
		})
		return tlog.Error("environment not found: " + envID)
	}

	elementName := r.FormValue("element")
	if elementName == "" {
		return tlog.Error("Param `element` is required")
	}

	if !env.Elements.Exists(elementName) {
		return tlog.Error("element does not exist")
	}

	return env.ElementVersionMap(elementName)
}

func apiEnvironmentElementCommitList(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("Param `env` is required")
	}

	elementName := r.FormValue("element")
	if elementName == "" {
		return tlog.Error("Param `element` is required")
	}

	branchName := r.FormValue("branch")
	if branchName == "" {
		return tlog.Error("Param `branch` is required")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		tlog.Error("środowisko nie zostało odnalezione", tlog.Vars{
			"env": envID,
		})
		return tlog.Error("środowisko nie zostało odnalezione: " + envID)
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

	// default value
	if limit == 0 {
		limit = 50
	}

	element := env.GetElement(elementName)
	if element == nil {
		tlog.Error("element nie został odnaleziony", tlog.Vars{
			"env":     envID,
			"element": elementName,
		})
		return tlog.Error("element nie został odnaleziony: " + elementName)
	}

	gitRepo := db.GitRepoGetByName(element.GetSource().RepoName)
	gitRepo.Open()
	defer gitRepo.Unlock()
	return struct {
		GitRepo string
		Branch  string
		Commits []db.CommitS
	}{
		GitRepo: gitRepo.Name,
		Branch:  branchName,
		Commits: gitRepo.GetCommits(branchName, limit, from),
	}
}

func apiEnvironmentElementDockerFile(r *http.Request, user *db.UserS) interface{} {

	type tmpDataS struct {
		Env      string
		Elements []string
	}

	buf, err := io.ReadAll(r.Body)
	if tlog.Error(err) != nil {
		return tlog.Error("cant read POST body")
	}

	data := &tmpDataS{}
	err = json.Unmarshal(buf, data)
	if tlog.Error(err) != nil {
		return tlog.Error("cant read POST body JSON")
	}

	env := db.EnvironmentMap.Get(data.Env)
	if env == nil {
		return tlog.Error("env not found")
	}

	if !user.HasEnvPerm(env.ID, perms.Env_ElementFullManage) {
		return tlog.Error("permission denied")
	}

	res := map[string]string{}

	for _, elementName := range data.Elements {
		if elementName == "zmiany środowiska" || elementName == "wszystkie zdarzenia" {
			continue
		}

		element := env.GetElement(elementName)
		image := element.GetImage()
		if image == nil {
			tlog.Error("image not found")
			continue
		}

		elementPod, ok := db.GetElementPod(element)
		if ok && elementPod.Build.Script != "" {
			res[elementName] = html.EscapeString(elementPod.Build.Script)
			continue
		}

		var dockerfile string
		if image.DockerFileContent != "" {
			dockerfile = image.DockerFileContent
		} else {
			repo := db.GitRepoGetByName(image.SourceGit.RepoName)
			repo.Lock()
			dockerfileContent := repo.GetFile(
				image.SourceGit.BranchName,
				image.SourceGit.CommitHash,
				image.DockerFilePath,
			)
			repo.Unlock()

			dockerfile = string(dockerfileContent)
		}

		res[elementName] = html.EscapeString(dockerfile)
	}

	return res
}

func apiEnvironmentElementRestart(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("`env` is required")
	}

	elementName := r.FormValue("element")
	if elementName == "" {
		return tlog.Error("`element` is required")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		return tlog.Error("env not found")
	}

	// XXX: ImageBuilder = temporary workaround
	if user.ID != "ImageBuilder" && !user.HasEnvPerm(envID, perms.Env_ElementStartStopRestart) {
		return tlog.Error("permission denied")
	}

	element := env.GetElement(elementName)
	if element.GetStatus().State == db.ElementStatusStopped {
		return tlog.Error("element is disabled and can not be restarted")
	}

	element.RestartAllPods(user)

	tlog.Info("restarting pods", tlog.Vars{
		"env":   envID,
		"event": true,
		"user":  user.Name,
	})
	return "ok"
}

func apiEnvironmentPodRestart(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("`env` is required")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		return tlog.Error("env not found")
	}

	if !user.HasEnvPerm(envID, perms.Env_ElementStartStopRestart) {
		return tlog.Error("permission denied")
	}

	elementName := r.FormValue("element")
	if elementName == "" {
		return tlog.Error("`element` is required")
	}

	podName := r.FormValue("pod")
	if podName == "" {
		return tlog.Error("`pod` is required")
	}

	element := env.GetElement(elementName)
	if element.GetStatus().State == db.ElementStatusStopped {
		return tlog.Error("element is disabled and can not be restarted")
	}

	es := element.GetStatus()
	if es.PodsGet()[podName] == nil {
		return tlog.Error("pod not found")
	}
	tlog.Info("pod restarted"+element.GetName(), tlog.Vars{
		"env":   env.ID,
		"event": true,
		"user":  user.Email,
	})
	es.PodsGet()[podName].Restart()
	return "ok"
}

func apiEnvironmentElementDelete(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("Param `env` is required")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		tlog.Error("środowisko nie zostało odnalezione", tlog.Vars{
			"env": envID,
		})
		return tlog.Error("środowisko nie zostało odnalezione: " + envID)
	}

	if !user.HasEnvPerm(envID, perms.Env_ElementFullManage) {
		return tlog.Error("permission denied")
	}

	if env.GitOps.Enabled {
		return tlog.Error("GitOps is enabled")
	}

	elementName := r.FormValue("element")
	if elementName == "" {
		return tlog.Error("Param `element` is required")
	}

	env.ElementToDelete(elementName, user)

	tlog.Info("deleting element "+elementName, tlog.Vars{
		"env":   envID,
		"user":  user.Name,
		"event": true,
	})
	return "ok"
}

func apiEnvironmentElementUpdateModeSet(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("Param `env` is required")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		return tlog.Error("env not found")
	}

	if !user.HasEnvPerm(envID, perms.Env_ElementFullManage) {
		return tlog.Error("permission denied")
	}

	if env.GitOps.Enabled {
		return tlog.Error("GitOps is enabled")
	}

	elementName := r.FormValue("element")
	if elementName == "" {
		return tlog.Error("Param `element` is required")
	}

	element := env.GetElement(elementName)

	switch r.FormValue("mode") {
	case "auto":
		element.SetAutoUpdate(true)
		element.Save(user)
		env := element.GetEnvironment()
		newestCommit := db.GitRepoGetByName(element.GetSource().RepoName).GetLastCommit(element.GetSource().BranchName)
		if err := env.SetElementVersion(element.GetName(), element.GetSource().BranchName, newestCommit, user); err != nil {
			tlog.Error(err)
		}

	case "manual":
		element.SetAutoUpdate(false)
		err := element.Save(user)
		if err != nil {
			return err
		}

	default:
		return tlog.Error("uknown update mode: " + r.FormValue("mode"))
	}

	tlog.Info("element mode set", tlog.Vars{
		"mode":  r.FormValue("mode"),
		"env":   env.ID,
		"event": true,
		"user":  user.Email,
	})

	return "ok"
}

func apiEnvironmentElementVersionChange(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("Param `env` is required")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		tlog.Error("środowisko nie zostało odnalezione", tlog.Vars{
			"env": envID,
		})
		return tlog.Error("środowisko nie zostało odnalezione: " + envID)
	}

	if !user.HasEnvPerm(envID, perms.Env_ElementFullManage) && !user.HasEnvPerm(envID, perms.Env_ElementVersionChangeOnly) {
		return tlog.Error("permission denied")
	}

	elementName := r.FormValue("element")
	if elementName == "" {
		return tlog.Error("Param `element` is required")
	}

	branchName := r.FormValue("branch")
	commitSHA := r.FormValue("commit")

	// --------------------

	element := env.GetElement(elementName)
	if element.GetAutoUpdate() {
		return tlog.Error("element is in auto update mode and can not be changed manually")
	}

	// --------------------

	if env.GitOps.Enabled {
		return tlog.Error("GitOps is enabled")
	}

	// --------------------

	// element.SetState(db.ElementStatusDeploying)

	err := env.SetElementVersion(elementName, branchName, commitSHA, user)
	if err != nil {
		return err
	}

	tlog.Info("element version change", tlog.Vars{
		"env":   env.ID,
		"event": true,
		"user":  user.Email,
	})

	return "ok"
}

func apiEnvironmentElementRunControl(r *http.Request, user *db.UserS) interface{} {
	req := &struct {
		EnvID   string
		Element string
		Control int
	}{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return tlog.Error(err)
	}

	env, ok := db.EnvironmentMap.GetFull(req.EnvID)
	if !ok {
		return tlog.Error("Env not found", tlog.Vars{"EnvID": req.EnvID})
	}

	elements := []string{req.Element}
	if req.Element == "" {
		elements = env.Elements.Keys()
	}

	for _, elementName := range elements {
		element := env.GetElement(elementName)

		switch req.Control {
		case 1:
			// Start
			element.SetStopped(false)
		case 2:
			// Stop
			element.SetStopped(true)

		case 3:
			// Enable scheduler
			element.SetUnschedulable(false)
		case 4:
			// Disable scheduler
			element.SetUnschedulable(true)
		default:
			return tlog.Error("unknown control code")
		}

		element.SetState(db.ElementStatusDeploying)
		element.Save(user)
	}

	tlog.Info("elements {{elements}} {{action}}", tlog.Vars{
		"elements": elements,
		"action":   req.Control,
		"env":      env.ID,
		"event":    true,
		"user":     user.Email,
	})

	return "ok"
}
