package api

import (
	"core/db"
	"core/db/permissions"
	perms "core/db/permissions"
	"core/db2"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"lib/tlog"
	"lib/utils/random"
	"lib/utils/set"
	"net/http"
	"strings"
	"time"
)

func apiEnvironmentShortMap2(r *http.Request, user *db.UserS) interface{} {

	res := map[string]db.EnvironmentShortS{}
	for env := range db.EnvironmentMap.Iter() {

		if !user.HasEnvPerm(env.Value.ID, perms.Env_View) {
			continue
		}

		res[env.Key] = db.EnvironmentShortS{
			ID:            env.Value.ID,
			Name:          env.Value.Name,
			Status:        env.Value.CalculateElementsStatuses(),
			ToDelete:      env.Value.ToDelete,
			ElementsCount: env.Value.Elements.Len(),
		}
	}

	return res
}

type envPost struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Teams []string `json:"teams"`
}

func apiEnvironmentCreate2(r *http.Request, user *db.UserS) interface{} {
	if !user.HasGlobPerm(perms.Glob_CreateAndDeleteEnvs) {
		return tlog.Error("permission denied")
	}

	var envP envPost
	{
		err := json.NewDecoder(r.Body).Decode(&envP)
		if err != nil {
			return tlog.Error("invalid json", tlog.Error(err))
		}
	}

	if envP.Name == "" {
		return tlog.Error("`name` is required")
	}

	env := db.EnvironmentS{
		ID:   envP.ID,
		Name: envP.Name,
	}

	err := env.Create(user)
	if err != nil {
		return err
	}

	tlog.Info("env created", tlog.Vars{
		"env":   env.ID,
		"event": true,
		"user":  user.Email,
	})

	return env.ID
}

func apiEnvironmentDelete2(r *http.Request, user *db.UserS) interface{} {
	if !user.HasGlobPerm(perms.Glob_CreateAndDeleteEnvs) {
		return tlog.Error("permission denied")
	}

	envID := r.FormValue("id")
	if envID == "" {
		return tlog.Error("`id` is required")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		return tlog.Error("env not found")
	}

	if env.GitOps.DynamicEnvironment {
		return tlog.Error("this is dynamic environment")
	}

	err := env.SetToDelete(true, user)
	if err != nil {
		return err
	}

	tlog.Info("env deleted", tlog.Vars{
		"env":   env.ID,
		"event": true,
		"user":  user.Email,
	})

	return "ok"
}

// type tmpTeam struct {
// 	ID     string
// 	Name   string
// 	Global map[string]perms.PermExplained
// 	Userki map[string]publicUserS
// }

func apiEnvironmentInfo(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("`env` is required")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		return tlog.Error("environment not found", tlog.Vars{
			"env": envID,
		})
	}

	env.CalculateElementsStatuses()
	env.CalculateElementsReadiness()
	env.CalculateElementsUsage()

	if env.Resources.RAMUsageAvgMB > 0 && env.Resources.RAMMaxMB < env.Resources.RAMUsageAvgMB {
		env.Resources.RAMMaxMB = env.Resources.RAMUsageAvgMB
	}

	if env.Resources.CPUUsageAvgCores > 0 && env.Resources.CPUMaxCores < env.Resources.CPUUsageAvgCores {
		env.Resources.CPUMaxCores = env.Resources.CPUUsageAvgCores
	}

	// --------------------
	// alerts:

	type tmpAlerts struct {
		Messages  []string
		Variables map[string]db.ErrorsT
	}
	alerts := map[string]tmpAlerts{}

	for _, elName := range env.Elements.Keys() {
		element := env.GetElement(elName)
		es := element.GetStatus()

		alerts[elName] = tmpAlerts{
			Messages:  es.Alerts,
			Variables: element.GetVariableError(),
		}
	}

	// --------------------

	type tmpEnv struct {
		db.EnvironmentS
	}

	tmp := tmpEnv{
		EnvironmentS: *env,
	}

	type tmpMetricsS struct {
		Enabled bool
	}

	tmpMetrics := tmpMetricsS{
		Enabled: db2.TheMetrics.Enabled(),
	}

	// --------------------

	envTeams := map[string]bool{}
	envMembers := map[string]map[string]bool{}
	blacklistedTeamID := ""
	for _, t := range db.TeamMap.Values() {
		if t.Name == db.BlacklistedTeamName {
			blacklistedTeamID = t.ID
		}
	}

	collect := func(t *db.Team) {
		envTeams[t.Name] = true
		for _, userID := range t.Members.List() {
			user := db.GetUserByID(userID)
			if user.Teams.Exists(blacklistedTeamID) {
				continue
			}
			if envMembers[user.Email] == nil {
				envMembers[user.Email] = map[string]bool{}
			}
			envMembers[user.Email][t.Name] = true
		}
	}

	for _, t := range db.TeamMap.Values() {
		if _, exist := t.Permissions.Envs["id:*"]; exist {
			collect(t)
		}
		if _, exist := t.Permissions.Envs["id:"+envID]; exist {
			collect(t)
		}
		if t.Name == db.AdminTeamName {
			collect(t)
		}
	}

	// --------------------

	urls := map[string]string{}
	for _, elName := range env.Elements.Keys() {
		element := env.GetElement(elName)

		if element == nil || element.GetType() != db.ElementSourceTypeDomain {
			continue
		}

		domain, ok := db.GetElementDomain(element)
		if !ok {
			continue
		}

		for path, pathData := range domain.Paths {
			if pathData.Label != "" {
				urls[pathData.Label] = "https://" + domain.Domain + path
			}
		}
	}

	// --------------------

	return struct {
		Env                 tmpEnv
		Alerts              map[string]tmpAlerts
		MostChangedElements []db.MostChangedElementsS
		Metrics             tmpMetricsS
		Permissions         map[string]permissions.PermExplained
		Teams               map[string]bool
		Members             map[string]map[string]bool
		URLs                map[string]string
	}{
		Env:                 tmp,
		Alerts:              alerts,
		MostChangedElements: env.MostChangedElementsGet(24 * time.Hour),
		Metrics:             tmpMetrics,
		Permissions:         user.GetPerms().ToFrontPerm().Env(envID, env.Tags.List()),
		Teams:               envTeams,
		Members:             envMembers,
		URLs:                urls,
	}
}

func apiEnvironmentClone(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("Param `env` is required")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		return tlog.Error("environment not found")
	}

	if !user.HasEnvPerm(envID, perms.Env_View) && !user.HasGlobPerm(perms.Glob_CreateAndDeleteEnvs) {
		return tlog.Error("permission denied")
	}

	id := r.FormValue("targetID")
	if id == "" {
		id = random.EnvID()
	}
	name := r.FormValue("targetName")
	if name == "" {
		return tlog.Error("Param `targetName` is required")
	}

	clonedEnv, err := env.Clone(user, id, name)
	if err != nil {
		return err
	}

	err = clonedEnv.Save(user)
	if err != nil {
		return err
	}

	tlog.Info("env cloned", tlog.Vars{
		"env":   env.ID,
		"event": true,
		"user":  user.Email,
	})

	return clonedEnv.ID
}

func apiEnvironmentCreateTag(r *http.Request, user *db.UserS) interface{} {

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
	if !user.HasEnvPerm(envID, perms.Env_ManageTags) {
		return tlog.Error("permission denied")
	}

	name := r.FormValue("name")
	if name == "" {
		return tlog.Error("Param `name` is required")
	}

	if env.Tags == nil {
		env.Tags = set.NewSafe[string](nil)
	}

	env.Tags.Add(name)

	err := env.Save(user)
	if err != nil {
		return err
	}

	tlog.Info("env tag created"+name, tlog.Vars{
		"env":   env.ID,
		"event": true,
		"user":  user.Email,
	})

	return "ok"
}

func apiEnvironmentDeleteTag(r *http.Request, user *db.UserS) interface{} {
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

	if !user.HasEnvPerm(envID, perms.Env_ManageTags) {
		return tlog.Error("permission denied")
	}

	name := r.FormValue("name")
	if name == "" {
		return tlog.Error("Param `name` is required")
	}

	env.Tags.Delete(name)

	err := env.Save(user)
	if err != nil {
		return err
	}

	tlog.Info("env tag deleted"+name, tlog.Vars{
		"env":   env.ID,
		"event": true,
		"user":  user.Email,
	})

	return "ok"
}

func apiEnvironmentVariables(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		tlog.Error("environment not found", tlog.Vars{
			"env": envID,
		})
		return tlog.Error("environment not found: " + envID)
	}

	variables := map[string]db.ElementVariableS{}
	for _, elName := range env.Elements.Keys() {
		element := env.GetElement(elName)
		for varName, varData := range element.GetVariablesMap(false) {
			variables[elName+"."+varName] = varData
		}
	}

	return variables
}

func apiEnvironmentVariableGetSecret(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	elementName := r.FormValue("element")
	variableName := r.FormValue("variable")

	if !user.HasEnvPerm(envID, perms.Env_CopyAndViewSecrets) {
		return tlog.Error("permission denied")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		tlog.Error("environment not found", tlog.Vars{
			"env": envID,
		})
		return tlog.Error("environment not found: " + envID)
	}

	element := env.GetElement(elementName)
	if element == nil {
		tlog.Error("element not found", tlog.Vars{
			"env":     envID,
			"element": elementName,
		})
		return tlog.Error("element not found: " + elementName)
	}

	return element.GetVariablesMap(true)[variableName].ResolvedValue
}

func apiEnvironmentPods(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		tlog.Error("environment not found", tlog.Vars{
			"env": envID,
		})
		return tlog.Error("environment not found: " + envID)
	}

	if !user.HasEnvPerm(envID, perms.Env_View) {
		return tlog.Error("permission denied")
	}

	pods := map[string]*db.ElementKubePodS{}
	for _, elName := range env.Elements.Keys() {
		for _, pod := range env.GetElement(elName).GetStatus().PodsGet() {
			pods[pod.PodName] = pod
		}
	}

	return pods
}

func apiEnvironmentRename(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	newName := r.FormValue("newName")
	if newName == "" {
		return tlog.Error("new name cant be empty")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		return tlog.Error("env not found")
	}

	if !user.HasEnvPerm(envID, perms.Env_Rename) {
		return tlog.Error("permission denied")
	}

	if newName == env.Name {
		return "ok"
	}

	tlog.Info(fmt.Sprintf("env renamed %s -> %s", env.Name, newName), tlog.Vars{
		"event": true,
		"env":   envID,
		"user":  user.Name,
	})
	env.Name = newName

	err := env.Save(user)
	if err != nil {
		return err
	}

	return "ok"
}

func apiEnvironmentSchedulerSet(r *http.Request, user *db.UserS) interface{} {

	type requestS struct {
		EnvID    string
		Location string
		Active   bool
		OnCrons  []string
		OffCrons []string
	}

	var request requestS
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return tlog.Error("Bad request", tlog.Vars{
			"error": err.Error(),
		})
	}

	if request.EnvID == "" {
		return tlog.Error("Param `EnvID` is required")
	}
	env := db.EnvironmentMap.Get(request.EnvID)
	if env == nil {
		return tlog.Error("environment not found")
	}

	if !user.HasEnvPerm(request.EnvID, perms.Env_ManageSchedule) {
		return tlog.Error("permission denied")
	}

	if err := env.SetSchedule(request.Location, request.OnCrons, request.OffCrons, user); err != nil {
		return err
	}

	env.Schedule.Active = request.Active
	env.Save(user)
	tlog.Info("env schedule set", tlog.Vars{
		"env":   env.ID,
		"event": true,
		"user":  user.Email,
	})
	return "ok"
}

func apiEnvironmentGitOpsSet(r *http.Request, user *db.UserS) interface{} {

	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("Param `envID` is required")
	}

	var request db.EnvironmentGitOpsConfigS
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return tlog.Error("Bad request", tlog.Vars{
			"error": err.Error(),
		})
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		return tlog.Error("environment not found")
	}

	if !user.HasEnvPerm(envID, perms.Env_ManageGitOPS) {
		return tlog.Error("permission denied")
	}

	if env.GitOps.DynamicEnvironment {
		return tlog.Error("this is dynamic environment")
	}

	env.GitOps = request

	if errx := env.FromGitOps(); errx != nil {
		return errx
	}

	if errx := env.Save(user); errx != nil {
		return errx
	}

	tlog.Info("env gitops changed", tlog.Vars{
		"event":  true,
		"env":    env.ID,
		"user":   user.Name,
		"gitops": env.GitOps.Enabled,
	})

	return "ok"
}

func apiEnvironmentDomainTargets(r *http.Request, user *db.UserS) interface{} {
	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("Param `envID` is required")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		return tlog.Error("environment not found")
	}

	return env.IngressTargetsGet()
}

func apiEnvironmentElementScale(r *http.Request, user *db.UserS) interface{} {

	type dataS struct {
		EnvID         string
		Element       string
		NrOfPodsMin   int32
		NrOfPodsMax   int32
		CPUTargetProc uint
	}

	data := new(dataS)
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		return tlog.Error("Invalid request", err)
	}

	if data.EnvID == "" {
		return tlog.Error("Param `EnvID` is required")
	}
	env := db.EnvironmentMap.Get(data.EnvID)
	if env == nil {
		return tlog.Error("Invalid app ID")
	}

	if data.Element == "" {
		return tlog.Error("Param `Element` is required")
	}
	if strings.HasSuffix(data.Element, "-debug") {
		return tlog.Error("element is a debug element")
	}

	if !user.HasEnvPerm(data.EnvID, perms.Env_ElementFullManage) {
		return tlog.Error("permission denied")
	}

	element := env.GetElement(data.Element)
	if element == nil {
		return tlog.Error("Element not found")
	}

	if data.NrOfPodsMin <= 0 {
		return tlog.Error("Param `Min` is required")
	}
	if data.NrOfPodsMax <= 0 {
		return tlog.Error("Param `Max` is required")
	}
	if data.CPUTargetProc <= 0 {
		return tlog.Error("Param `TargetCPU` is required")
	}

	if data.NrOfPodsMin > 200 {
		return tlog.Error("Min value cannot be higher than 200")
	}
	if data.NrOfPodsMax > 200 {
		return tlog.Error("Max value cannot be higher than 200")
	}
	if data.NrOfPodsMin > data.NrOfPodsMax {
		return tlog.Error("Min value cannot be higher than max value")
	}

	if element.GetType() == db.ElementSourceTypePod && element.GetScale().MaxOnePod {
		return tlog.Error("Ten element ma zablokowane skalowanie, dlatego może być maksymalnie jedna instancja.")
	}

	if data.CPUTargetProc < 20 || data.CPUTargetProc > 80 {
		return tlog.Error("Target CPU should be between 20 and 80")
	}

	if data.NrOfPodsMin != element.GetScale().NrOfPodsMin {
		element.GetScale().NrOfPodsMin = data.NrOfPodsMin
		element.GetScale().NrOfPodsCurrent = data.NrOfPodsMin

	}
	if data.NrOfPodsMax != element.GetScale().NrOfPodsMax {
		element.GetScale().NrOfPodsMax = data.NrOfPodsMax

	}
	if data.CPUTargetProc != element.GetScale().CPUTargetProc {
		element.GetScale().CPUTargetProc = data.CPUTargetProc

	}
	element.Save(user)

	return "ok"
}

func apiEnvironmentElementStaticScaling(r *http.Request, user *db.UserS) interface{} {
	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("`env` jest wymagane")
	}
	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		return tlog.Error("env not found")
	}

	if !user.HasEnvPerm(envID, perms.Env_ElementFullManage) {
		return tlog.Error("permission denied")
	}

	elements := map[string]int{}
	err := json.NewDecoder(r.Body).Decode(&elements)
	if err != nil {
		return tlog.Error("Bad request")
	}

	for elementName, scale := range elements {
		element := env.GetElement(elementName)
		if element == nil {
			return tlog.Error("Nie odnaleziono elementu")
		}

		if element.GetType() != db.ElementSourceTypePod {
			return tlog.Error("elementu nie mozna skalowac")
		}

		if scale < 0 {
			return tlog.Error("nieprawiadlowa ilosc instancji")
		}

		if element.GetScale().MaxOnePod {
			return tlog.Error("Ten element ma zablokowane skalowanie, dlatego może być maksymalnie jedna instancja.")
		}

		err2 := env.ElementScale(element, scale, user)
		if err2 != nil {
			return err2
		}

		tlog.Info("element "+element.GetName()+" change scale", tlog.Vars{
			"env":   env.ID,
			"event": true,
			"user":  user.Email,
		})
	}
	return "ok"
}

func apiEnvironmentDynamicSourcesMap(r *http.Request, user *db.UserS) interface{} {
	db.EnvironmentDynamicGitSourcesRefresh()
	return db.EnvironmentDynamicGitSources
}

func apiEnvironmentDynamicSourcesAdd(r *http.Request, user *db.UserS) interface{} {

	envSource := new(db.EnvDynamicGitSourceS)
	err := json.NewDecoder(r.Body).Decode(envSource)
	if err != nil {
		return tlog.Error("invalid json", tlog.Error(err))
	}

	envSource.ID = ""
	envSource.Save(user)

	return "ok"
}

func apiEnvironmentDynamicSourcesDelete(r *http.Request, user *db.UserS) interface{} {

	envSourceID := r.FormValue("envSourceID")
	if envSourceID == "" {
		return tlog.Error("Param `envSourceID` is required")
	}

	envSource := db.EnvironmentDynamicGitSources.Get(envSourceID)
	envSource.Delete(user)

	return "ok"
}

func apiEnvironmentExportTOML(r *http.Request, user *db.UserS) interface{} {
	envID := r.FormValue("env")
	if envID == "" {
		return tlog.Error("Param `env` is required")
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		return tlog.Error("environment not found")
	}

	if !user.HasEnvPerm(envID, perms.Env_ManageMembers) {
		return tlog.Error("permission denied")
	}

	out := []string{`type = "env"`, ""}

	for _, elName := range env.Elements.Keys() {
		element := env.GetElement(elName)
		es := element.GetSource()

		out = append(out, "[element."+elName+"]")
		out = append(out, `git-repo-name = "`+es.RepoName+`"`)
		out = append(out, `branch = "`+es.BranchName+`"`)
		out = append(out, `file-path = "`+es.FilePath+`"`)

		if !element.GetAutoUpdate() {
			out = append(out, `commit = "`+es.CommitHash+`"`)
		}

		out = append(out, "")
	}

	return base64.StdEncoding.EncodeToString([]byte(strings.Join(out, "\n")))
}
