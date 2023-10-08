package db

import (
	"core/config"
	"core/kube"
	"encoding/json"
	"fmt"
	"lib/tlog"
	"lib/utils"
	"lib/utils/conv"
	"lib/utils/maps"
	"lib/utils/random"
	"lib/utils/set"
	"lib/utils/slice"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)

var (
	EnvironmentMap               = maps.NewSafe[string, *EnvironmentS](nil) // key=env-id
	EnvironmentSchedulersMap     = maps.NewSafe[string, *gocron.Scheduler](nil)
	ElementMap                   = maps.NewSafe[string, EnvElementS](nil)           // key='env_id/el_name'
	EnvironmentDynamicGitSources = maps.NewSafe[string, *EnvDynamicGitSourceS](nil) // key=id
)

type EnvironmentShortS struct {
	ID            string
	Name          string
	Status        map[ElementState]int
	ToDelete      bool 
	ElementsCount int
}

type EnvironmentS struct {
	ID          string // random, unique
	Name        string 
	ClusterName string 

	Schedule EnvironmentScheduleS
	GitOps   EnvironmentGitOpsConfigS
	ToDelete bool 

	CreationTime   int64
	LastChangeTime int64

	Elements            *maps.SafeMap[string, ElementType] // element-name => element-type
	Tags                *set.Safe[string]
	GlobalVariableCache *maps.SafeMap[string, ElementVariableS]

	Status    map[ElementState]int // env.CalculateElementsStatuses()
	Readiness ReadinessS           // env.CalculateElementsReadiness()
	Resources MetricsInfo
	// Operators maps.SafeMap[string, string] // user-id => role-in-env

	statuses *maps.SafeMap[string, *ElementStatusS] // key=element-name
}

type MetricsInfo struct {
	CPUUsageAvgCores int
	CPUReservedCores int // in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore
	CPUMaxCores      int // in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore

	RAMUsageAvgMB int
	RAMReservedMB int // MB
	RAMMaxMB      int // MB
}

type ReadinessS struct {
	ElementReady    int
	ElementNotReady int
	PodReady        int
	PodNotReady     int
}

type EnvironmentGitOpsConfigS struct {
	Enabled            bool
	DynamicEnvironment bool
	GitRepoName        string
	BranchName         string
	FilePath           string
}

type MostChangedElementsS struct {
	Name         string
	TotalChanges int
	Successes    int
	Failures     int
	InProgress   int
}

type OperatorRoleT map[string]string // user-id => role-in-env

type EnvironmentScheduleS struct {
	Active   bool
	Timezone *time.Location
	OnCrons  []string
	OffCrons []string
}

func (s EnvironmentScheduleS) Configured() bool {
	return len(s.OnCrons)+len(s.OffCrons) > 0
}

// --------------------------------------------

func SyncWithDiskLoop() {
	for {

		environmentIDs, _ := os.ReadDir(filepath.Join(config.DataPath(), "env"))
		tmp := map[string]bool{}
		for _, environmentID := range environmentIDs {
			tmp[environmentID.Name()] = true
		}

		for _, k := range EnvironmentMap.Keys() {
			if !tmp[k] {
				EnvironmentMap.Delete(k)
			}
		}

		for _, environmentID := range environmentIDs {

			envID := environmentID.Name()

			if EnvironmentMap.Get(envID) == nil {

				env := new(EnvironmentS)
				if err := driver.Read(filepath.Join("env", envID), "env", env); err != nil {
					tlog.Error(err)
					continue
				}

				EnvironmentMap.Set(envID, env)

			}
		}

		time.Sleep(10 * time.Second)
	}
}

// --------------------------------------------

func (env *EnvironmentS) Save(user *UserS) *tlog.RecordS {

	// ---
	// GenerateID:

	if env.ID == "" {
		env.ID = "env-" + conv.KeyString(RandString(12))
	}

	err := env.CheckID()
	if err != nil {
		return err
	}

	err = env.CheckName()
	if err != nil {
		return err
	}

	// ---

	env.LastChangeTime = time.Now().Unix()

	err2 := driver.Write(filepath.Join("env", env.ID), "env", env)
	if err2 != nil {
		return tlog.Error(err)
	}


	EnvironmentMap.Set(env.ID, env)

	return nil
}

func (env *EnvironmentS) Create(user *UserS) *tlog.RecordS {

	if env.Name == "" {
		return tlog.Error("`name` is required")
	}

	if env.ID == "" {
		env.ID = random.EnvID()
	}

	if EnvironmentMap.Get(env.ID) != nil {
		return tlog.Error("Environment with this ID already exists")
	}

	env.CreationTime = time.Now().Unix()

	err := env.Save(user)
	if err != nil {
		return err
	}

	return nil
}

func (env *EnvironmentS) Delete(user *UserS) *tlog.RecordS {

	if !env.ToDelete {
		return nil
	}

	EnvironmentMap.Delete(env.ID)
	os.RemoveAll(filepath.Join(config.DataPath(), "env", env.ID))

	var email string
	if user != nil {
		email = user.Email
	}

	tlog.Info("env deleted", tlog.Vars{
		"env":   env.ID,
		"user":  email,
		"event": true,
	})

	return nil
}

func (env *EnvironmentS) ElementAdd(elementName string, element EnvElementS, user *UserS, start bool) *tlog.RecordS {

	if env.Elements.Exists(elementName) {
		return tlog.Error("Element with this name already exists", tlog.Vars{
			"envID":       env.ID,
			"elementName": elementName,
		})
	}

	if !start {
		element.SetStopped(true)
	}

	if element == nil {
		return tlog.Error("element not found", tlog.Vars{
			"envID":       env.ID,
			"elementName": elementName,
		})
	}

	if errx := element.Save(user); errx != nil {
		return errx
	}

	if env.Elements == nil {
		env.Elements = maps.NewSafe[string, ElementType](nil)
	}

	env.statuses.Delete(elementName)
	env.Elements.Set(element.GetName(), element.GetType())
	env.Save(user)

	env.ReRender(elementName)

	return nil
}

func (env *EnvironmentS) ReRender(elementName string) {
	// reRender variables
	for i := range ElementMap.Iter() {
		if !strings.HasPrefix(i.Key, env.ID) {
			continue
		}

		_, elName, _ := strings.Cut(i.Key, "/")
		if elementName == elName {
			continue
		}
		i.Value.RenderVariables()
		i.Value.Save(nil)
	}
}

// ElementLoad loads element from disk and returns unpatched element
func (env *EnvironmentS) ElementLoad(elementName string) EnvElementS {

	elementType := env.Elements.Get(elementName)
	if elementType == "" {
		tlog.Error("element not found in env", tlog.Vars{
			"element": elementName,
			"env":     env.ID,
		})
		return nil
	}

	path := filepath.Join("env", env.ID, "element", elementName)

	var el EnvElementS
	switch elementType {
	case ElementSourceTypeConfig:
		el = &elementConfigS{}
	case ElementSourceTypeDomain:
		el = &elementDomainS{}
	case ElementSourceTypeElasticsearch:
		el = &elementElasticsearchS{}
	case ElementSourceTypeMongodb:
		el = &elementDomainS{}
	case ElementSourceTypePod:
		el = &elementPodS{}
	case ElementSourceTypeAction:
		el = &elementActionS{}
	default:
		tlog.Error("unknown element type")
		return nil
	}

	tlog.Error(driver.Read(path, "0-current", el))
	return el
}

func (env *EnvironmentS) GetElement(elementName string) EnvElementS {
	element, exists := ElementMap.GetFull(fmt.Sprintf("%s/%s", env.ID, elementName))
	if exists {
		return element
	}

	element = env.ElementLoad(elementName)
	if element != nil {
		element, _ = ApplyPatch(element)
		ElementMap.Set(fmt.Sprintf("%s/%s", env.ID, elementName), element)
	}
	return element
}

func (env *EnvironmentS) cloneElement(elementName string) EnvElementS {
	in := env.GetElement(elementName)
	if in == nil {
		return nil
	}

	buf, err := json.Marshal(in)
	tlog.Error(err)

	var out EnvElementS
	switch in.GetType() {
	case ElementSourceTypeConfig:
		out = &elementConfigS{}
	case ElementSourceTypeDomain:
		out = &elementDomainS{}
	case ElementSourceTypeElasticsearch:
		out = &elementElasticsearchS{}
	case ElementSourceTypeMongodb:
		out = &elementDomainS{}
	case ElementSourceTypePod:
		out = &elementPodS{}
	case ElementSourceTypeAction:
		out = &elementActionS{}
	default:
		tlog.Error("unknown element type")
		return nil
	}

	err = json.Unmarshal(buf, out)
	if err != nil {
		tlog.Error(err)
	}

	return out
}

func (env *EnvironmentS) ElementCloneWithoutSecrets(elementName string) EnvElementS {
	out := env.cloneElement(elementName)
	out.hideSecrets()
	return out
}

func (env *EnvironmentS) Clone(user *UserS, targetEnvID, targetEnvName string) (*EnvironmentS, *tlog.RecordS) {
	clonedEnv := utils.DeepCopy(*env)
	clonedEnv.ID = targetEnvID
	clonedEnv.Name = targetEnvName
	clonedEnv.CreationTime = time.Now().Unix()
	clonedEnv.Elements = maps.New[string, ElementType](nil).Safe()
	clonedEnv.GitOps.DynamicEnvironment = false

	errx := clonedEnv.Create(user)
	if errx != nil {
		return nil, errx
	}

	for _, elementName := range env.Elements.Keys() {
		element := env.cloneElement(elementName)
		element.generateSecrets(true)
		element.setEnvID(targetEnvID)
		clonedEnv.Elements.Set(element.GetName(), element.GetType())
		element.Save(user)
	}

	return clonedEnv, nil
}

func (env *EnvironmentS) SetToDelete(t bool, user *UserS) *tlog.RecordS {
	env.ToDelete = true
	return env.Save(user)
}

func (env *EnvironmentS) ElementToDelete(elementName string, user *UserS) *tlog.RecordS {

	element := env.GetElement(elementName)
	if element == nil {
		return tlog.Error("element does not exist")
	}

	element.setToDelete(true)

	if err := element.Save(user); err != nil {
		return tlog.Error(err)
	}

	return nil
}

func (env *EnvironmentS) SetElementVersion(elementName, branchName, commitSHA string, user *UserS) *tlog.RecordS {

	elementOld := env.ElementLoad(elementName)
	if elementOld == nil {
		return tlog.Error("element does not exist")
	}

	elementNew := ElementFromGit(elementName, env.ID, SourceGitS{
		RepoName:   elementOld.GetSource().RepoName,
		BranchName: branchName,
		CommitHash: commitSHA,
		FilePath:   elementOld.GetSource().FilePath,
	})
	if elementNew == nil {
		return tlog.Error("error while changing version")
	}

	return ElementUpdate(
		elementOld,
		elementNew,
		user,
		false,
	)
}

type ElementVersionPatchS struct {
	Data ElementVersionS
}

func (env *EnvironmentS) ElementVersionMap(elementName string) map[string]*ElementVersionS { // key = commit SHA

	versions := map[string]*ElementVersionS{}

	path := filepath.Join("env", env.ID, "element", elementName)
	versionTimes, err := driver.List(path)
	if err != nil {
		return versions
	}

	if len(versionTimes) == 0 {
		return versions
	}

	sort.Strings(versionTimes)
	currentVestionTime := versionTimes[0]

	previousVestionTime := ""
	if len(versionTimes) > 1 {
		previousVestionTime = versionTimes[len(versionTimes)-1] 
	}

	for _, versionTime := range versionTimes {

		tmp := new(ElementVersionPatchS)
		if err := driver.Read(path, versionTime, tmp); err != nil {
			tlog.Error(err)
			continue
		}

		if tmp.Data.SourceGit.CommitHash == "" {
			// tlog.Error("element version file `" + versionTime + ".json` is empty")
			continue
		}

		if currentVestionTime == versionTime {
			tmp.Data.Current = true

		} else if previousVestionTime == versionTime {
			tmp.Data.Previous = true
		}
		versions[tmp.Data.SourceGit.CommitHash] = &tmp.Data
	}

	// ---

	tmp := new(ElementVersionS)
	if err := driver.Read(path, "0-current", tmp); err != nil {
		tlog.Error(err)
	}
	tmp.Current = true
	versions[tmp.SourceGit.CommitHash] = tmp

	// ---

	// fmt.Println("000000000000000000000000000000000")
	// tlog.PrintJSON(versions)
	// os.Exit(1)

	return versions
}

func (env *EnvironmentS) CalculateElementsStatuses() map[ElementState]int {
	states := map[ElementState]int{}
	for _, elementName := range env.Elements.Keys() {
		element := env.GetElement(elementName)
		status := element.GetStatus()
		states[status.State]++
	}

	env.Status = states

	return states
}

func (env *EnvironmentS) CalculateElementsReadiness() ReadinessS {
	readiness := ReadinessS{}
	for _, elementName := range env.Elements.Keys() {
		element := env.GetElement(elementName)
		status := element.GetStatus()

		switch status.State {
		case ElementStatusReady:
			readiness.ElementReady++

		// case ElementStatusTerminating, ElementStatusStopped:
		// 	// skip counting this element and its pods
		// 	continue

		default:
			readiness.ElementNotReady++
		}

		for _, pod := range status.PodsGet() {
			switch pod.Status {
			case kube.PodStatusReady, kube.PodStatusSucceeded:
				readiness.PodReady++
			default:
				readiness.PodNotReady++
			}
		}
	}

	env.Readiness = readiness

	return readiness
}

func (env *EnvironmentS) SetSchedule(timezone string, onCron, offCron []string, user *UserS) *tlog.RecordS {

	if timezone == "" || len(onCron)+len(offCron) == 0 {
		if s, ok := EnvironmentSchedulersMap.GetFull(env.ID); ok {
			s.Stop()
			EnvironmentSchedulersMap.Delete(env.ID)
		}

		tlog.Info("Disable schedule", tlog.Vars{
			"event": true,
			"env":   env.ID,
			"user":  user.Email,
		})

		env.Schedule = EnvironmentScheduleS{}
		return env.Save(user)
	}

	if slice.Equal(onCron, env.Schedule.OnCrons) &&
		slice.Equal(offCron, env.Schedule.OffCrons) {
		// no changes
		return nil
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return tlog.Error("Invalid location", tlog.Vars{
			"Location": timezone,
			"error":    err,
		})
	}

	env.Schedule = EnvironmentScheduleS{
		Timezone: location,
		OnCrons:  onCron,
		OffCrons: offCron,
	}

	tlog.Info("Enable schedule", tlog.Vars{
		"event": true,
		"env":   env.ID,
		"user":  user.Email,
	})

	errx := env.Save(user)
	if errx != nil {
		return errx
	}

	return env.StartSchedule(user)
}

func (env *EnvironmentS) StartSchedule(user *UserS) *tlog.RecordS {
	scheduler := gocron.NewScheduler(env.Schedule.Timezone)
	scheduler.SingletonModeAll()

	for _, cron := range env.Schedule.OnCrons {
		if _, err := scheduler.Cron(cron).Do(scheduleEnableEnv, env); err != nil {
			return tlog.Error(err.Error())
		}
	}
	for _, cron := range env.Schedule.OffCrons {
		if _, err := scheduler.Cron(cron).Do(scheduleDisableEnv, env); err != nil {
			return tlog.Error(err.Error())
		}
	}

	if s := EnvironmentSchedulersMap.Get(env.ID); s != nil {
		s.Stop()
	}
	scheduler.StartAsync()
	EnvironmentSchedulersMap.Set(env.ID, scheduler)
	return nil
}

func scheduleEnableEnv(env *EnvironmentS) {
	for _, elementName := range env.Elements.Keys() {
		element := env.GetElement(elementName)
		if element.GetUnschedulable() {
			continue
		}
		element.SetStopped(false)
		element.Save(nil)
	}
}

func scheduleDisableEnv(env *EnvironmentS) {
	for _, elementName := range env.Elements.Keys() {
		element := env.GetElement(elementName)
		if element.GetUnschedulable() {
			continue
		}
		element.SetStopped(true)
		element.Save(nil)
	}
}

func (env *EnvironmentS) MostChangedElementsGet(interavl time.Duration) []MostChangedElementsS {

	res := []MostChangedElementsS{}
	for _, elName := range env.Elements.Keys() {

		mce := MostChangedElementsS{
			Name: elName,
		}

		// tlog.PrintJSON(env.ElementVersionMap(elName))

		for commitSHA, ver := range env.ElementVersionMap(elName) {
			if commitSHA == "" {
				continue
			}

			mce.TotalChanges++

			switch ver.Status {
			case ElementStatusNew, ElementStatusBuilding, ElementStatusDeploying, ElementStatusTerminating:
				mce.InProgress++

			case ElementStatusFailed:
				mce.Failures++

			case ElementStatusReady:
				mce.Successes++
			}
		}

		res = append(res, mce)
	}

	// ---

	sort.Slice(res, func(i, j int) bool { return res[i].TotalChanges < res[j].TotalChanges })
	if len(res) > 10 {
		res = res[:10]
	}
	// tlog.PrintJSON(res)

	return res
}

// SetStatusForAllElements for release and every element type
func (env *EnvironmentS) SetStatusForAllElements(state ElementState, user *UserS) {

	for _, elName := range env.Elements.Keys() {
		el := env.GetElement(elName)
		el.SetState(state)
		el.Save(user)
	}
}

func (env *EnvironmentS) ElementDelete(elementName string, user *UserS) *tlog.RecordS {

	env.Elements.Delete(elementName)
	env.statuses.Delete(elementName)
	ElementMap.Delete(fmt.Sprintf("%s/%s", env.ID, elementName))
	errx := env.Save(user)
	if errx != nil {
		return errx
	}

	path := filepath.Join(config.DataPath(), "env", env.ID, "element", elementName)

	err := os.RemoveAll(path)
	if err != nil {
		return tlog.Error(err)
	}

	env.ReRender(elementName)

	return nil
}

func (env *EnvironmentS) CalculateElementsUsage() {
	var cpu, ram float64
	for _, elName := range env.Elements.Keys() {
		el := env.GetElement(elName)
		if el.GetStopped() {
			continue
		}
		c, r := el.GetResources()
		cpu += c
		ram += r
	}
	env.Resources.CPUUsageAvgCores = int(math.Floor(cpu))
	env.Resources.RAMUsageAvgMB = int(math.Floor(ram))
}

func (env *EnvironmentS) IngressTargetsGet() []string {

	targets := []string{}
	for el := range env.Elements.Iter() {

		if el.Value != ElementSourceTypePod {
			continue
		}

		for portNr, portInfo := range env.GetElement(el.Key).(*elementPodS).ExposePorts {
			if portInfo.Type == "http" || portInfo.Type == "https" {
				targets = append(targets, el.Key+":"+portNr)
			}
		}

	}

	return targets
}

func (env *EnvironmentS) ElementScale(e EnvElementS, scale int, user *UserS) *tlog.RecordS {

	e.GetScale().NrOfPodsMin = int32(scale)
	e.GetScale().NrOfPodsMax = int32(scale)
	e.SetUnschedulable(scale == 0)
	e.SetStopped(scale == 0)
	e.SetState(ElementStatusDeploying)
	return e.Save(user)
}

// func (env *EnvironmentS) AddTeam(team *Team, user *UserS) error {

// 	if env.Teams == nil {
// 		tlog.Debug("env.Teams == nil")
// 		return errors.New("env.Teams is nil")
// 	}

// 	env.Teams.Add(team.ID)
// 	err := env.Save(user)
// 	if err != nil {
// 		tlog.Debug("env.Save(user) error: " + err.Message)
// 		return errors.New("env.Save(user) error: " + err.Message)
// 	}

// 	return nil
// }

// func (env *EnvironmentS) RemoveTeam(team *Team, user *UserS) error {

// 	if env.Teams == nil {
// 		tlog.Debug("env.Teams == nil")
// 		return errors.New("env.Teams is nil")
// 	}

// 	if team.Name == AdminTeamName {
// 		return errors.New("team admin cannot be removed")
// 	}

// 	env.Teams.Delete(team.ID)
// 	err := env.Save(user)
// 	if err != nil {
// 		tlog.Debug("env.Save(user) error: " + err.Message)
// 		return errors.New("env.Save(user) error: " + err.Message)
// 	}

// 	return nil
// }

func (env *EnvironmentS) FromGitOps() *tlog.RecordS {

	if !env.GitOps.Enabled {
		return nil
	}

	envInGit := EnvInGitRepoCache.Get(env.GitOps.GitRepoName + "/" + env.GitOps.BranchName).Environments.Get(env.GitOps.FilePath)
	if envInGit == nil {
		return tlog.Error("env not found in git")
	}


	for elName, elSource := range envInGit.Element {

		if elSource.RepoName == "" {
			elSource.RepoName = env.GitOps.GitRepoName
		}

		if GitRepoGetByName(elSource.RepoName) == nil {
			tlog.Error("git repo `{{gitRepoName}}` not found for element `{{elementName}}`", tlog.Vars{
				"elementName": elName,
				"gitRepoName": elSource.RepoName,
			})
			continue
		}

		if elSource.CommitHash == "" {
			elSource.CommitHash = GitRepoGetByName(elSource.RepoName).GetLastCommit(elSource.BranchName)
			if elSource.CommitHash == "" {
				tlog.Error("unable to get last CommitHash for element `{{elementName}}` in git repo `{{gitRepoName}}`", tlog.Vars{
					"elementName": elName,
					"gitRepoName": elSource.RepoName,
				})
				continue
			}
		}

		element := env.GetElement(elName)
		if element == nil {
			elementNew := ElementFromGit(elName, env.ID, elSource)
			if elementNew == nil {
				tlog.Error("error while changing version", tlog.Vars{
					"envID":       env.ID,
					"elementName": elName,
					"gitRepoName": elSource.RepoName,
				})
				continue
			}

			env.ElementAdd(
				elName,
				elementNew,
				nil,
				true,
			)
			continue
		}

		if !compareSource(element.GetSource(), elSource) {
			env.SetElementVersion(elName, elSource.BranchName, elSource.CommitHash, nil)
			continue
		}
	}

	for _, elName := range env.Elements.Keys() {
		_, exist := envInGit.Element[elName]
		if !exist {
			env.ElementToDelete(elName, nil)
		}
	}

	for elName, elSource := range envInGit.Element {
		element := env.GetElement(elName)
		if element == nil {
			continue
		}

		element.SetAutoUpdate(elSource.CommitHash == "")
		element.Save(nil)
	}

	// -----------------

	return nil
}

func compareSource(a, b SourceGitS) bool {
	if a.RepoName != b.RepoName {
		return false
	}

	if a.BranchName != b.BranchName {
		return false
	}

	if a.FilePath != b.FilePath {
		return false
	}

	if a.CommitHash != b.CommitHash {
		return false
	}

	return true
}
