package db

import (
	"core/kube"
	"fmt"
	"lib/tlog"
	"lib/utils/conv"
	"lib/utils/maps"
	"path/filepath"
	"strings"
)

type elementPodS struct {
	elementS

	ParentSource *SourceGitS      `toml:"from"`
	Build        elementPodBuildS `toml:"build"`
	RunCommand   []string         `toml:"run-cmd"`

	RunAsUser     []int64 `toml:"run-as-user"` 
	RunWritableFS bool    `toml:"run-writable-file-system"`

	StaticFilesPath string `toml:"serve-static-files-from,omitempty"`

	CapabilityAddBindService bool `toml:"capability-add-net-bind-service"`

	RunProbe            *elementRunProbeS               `toml:"probe"`
	Stateful            bool                            `toml:"-"`       // true = StatefulSet, false = Deployment
	Storage             map[string]*kube.StorageS       `toml:"storage"` // key=mount point in pod
	ExposePorts         map[string]*elementExposePortS  `toml:"expose"`
	ExposePortsHeadless bool                            `toml:"expose-headless"`
	StickyCookie        string                          `toml:"sticky-cookie"`
	ServiceAccount      elementContainerServiceAccountS `toml:"service-account"`

	Schedule string              `toml:"schedule"` // cron schedule format
	Actions  map[string][]string `toml:"actions"`  // key=action name, value=action script

	CPUReservedPC uint `toml:"cpu"` // PC = procent rdzenia, in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore
	CPULimitPC    uint `toml:"-"`   // PC = procent rdzenia, in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore

	RAMReservedMB uint `toml:"ram"` // in MB
	RAMLimitMB    uint `toml:"-"`   // in MB

	// Pods    map[string]*ElementPodS `toml:"-"` // key=podName
	Scale       ElementScaleS                  `toml:"scale"`
	HostAliases maps.SafeMap[string, []string] `toml:"host-aliases"` // key=IP, value=list of domains
}

type elementPodBuildS struct {
	Type           ElementPodBuildTypeS `toml:"type"`
	Script         string               `toml:"script"`
	Image          string               `toml:"image"`
	DockerfilePath string               `toml:"dockerfile"` 

	Variables map[string]string `toml:"dockerfile-variables"`
	RootPath  string            `toml:"root-path"`
	ImageID   string            `toml:"-"`

	Steps []buildStepS `toml:"steps"`
}

type buildStepS struct {
	Name       string
	CMD        map[string][]string 
	TimeoutSec int64               `toml:"timeout-sec"`
}

type ElementPodBuildTypeS string

const (
	ElementPodBuildScript         ElementPodBuildTypeS = "script"
	ElementPodBuildImage          ElementPodBuildTypeS = "image"
	ElementPodBuildDockerfilePath ElementPodBuildTypeS = "dockerfile"
)

type elementExposePortS struct {
	Probe struct {
		Disable             bool   `toml:"disable"`
		Path                string `toml:"path"`                  // default '/', use for HTTP probe
		InitialDelaySeconds int32  `toml:"initial-delay-seconds"` // default 0
		PeriodSeconds       int32  `toml:"period-seconds"`        // default 10
		TimeoutSeconds      int32  `toml:"timeout-seconds"`       // default 1
		SuccessThreshold    int32  `toml:"success-threshold"`     // default 1
		FailureThreshold    int32  `toml:"failure-threshold"`     // default 3
		RestartOnFail       bool   `toml:"restart-on-fail"`
		ErrorMsg            string `toml:"-"`
	} `toml:"probe"`

	Name        string  `toml:"name"`         // created service name
	Type        string  `toml:"type"`         // one of: tcp, udp, http, https
	MetricsPath string  `toml:"metrics-path"` 
	PortAliases []int32 `toml:"port-aliases"`
}

type elementRunProbeS struct {
	Exec                []string `toml:"exec"`
	InitialDelaySeconds int32    `toml:"initial-delay-seconds"`
	PeriodSeconds       int32    `toml:"period-seconds"`
	TimeoutSeconds      int32    `toml:"timeout-seconds"`
	SuccessThreshold    int32    `toml:"success-threshold"`
	FailureThreshold    int32    `toml:"failure-threshold"`
	RestartOnFail       bool     `toml:"restart-on-fail"`
}

type ElementScaleS struct {
	MaxOnePod bool `toml:"max-one"`

	NrOfPodsMin     int32 `toml:"min"`
	NrOfPodsMax     int32 `toml:"max"`
	NrOfPodsCurrent int32 `toml:"-"`

	CPUTargetProc uint `toml:"targetCPU"`
}

type elementContainerServiceAccountS struct {
	Name   string `toml:"name"`
	Secret string `toml:"secret"`
}

type CheckStepResultStatusS int

const (
	CheckStepResultFail       CheckStepResultStatusS = 0 // fail
	CheckStepResultSuccess    CheckStepResultStatusS = 1 // success
	CheckStepResultInProgress CheckStepResultStatusS = 2 // in progress
)

type CheckStepResultS struct {
	Status CheckStepResultStatusS
	Msg    string
}

func (element *elementPodS) RebuildImage(imageID string, user *UserS) *tlog.RecordS {

	if element == nil {
		return nil
	}

	if element.Type != ElementSourceTypePod {
		return nil
	}

	if imageID == element.Build.ImageID {
		es := element.GetStatus()
		es.State = ElementStatusDeploying
		es.Alerts = append(es.Alerts, "BuildImages: rebuilding...")
		es.Save()
	}

	err := element.Save(user)
	if err != nil {
		return err
	}

	return nil
}

func (element *elementPodS) GetImage() *ImageS {
	return ImageGetByID(element.Build.ImageID)
}

func (element *elementPodS) RestartAllPods(user *UserS) *tlog.RecordS {

	es := element.GetStatus()

	if es.restart {
		return nil
	}

	es.restart = true
	element.SetState(ElementStatusDeploying)

	return element.Save(user)
}

func (element *elementPodS) check(user *UserS) *tlog.RecordS {
	if element.ToDelete {
		return nil
	}

	if len(element.RunCommand) == 0 {
		return tlog.Error("run-cmd is empty")
	}

	if element.Scale.CPUTargetProc == 0 {
		element.Scale.CPUTargetProc = 80
	}

	if element.Scale.NrOfPodsMin == 0 {
		element.Scale.NrOfPodsMin = 1
	}

	if element.Scale.NrOfPodsMax == 0 {
		element.Scale.NrOfPodsMax = 1
	}


	element.Stateful = false
	for _, storage := range element.Storage {
		if storage.Type == "block" {
			element.Stateful = true
			if storage.MaxSizeMB <= 1 {
				return tlog.Error("storage MaxSizeMB is invalid")
			}
			break
		}
	}


	if element.SourceGit.RepoName == "" || element.SourceGit.FilePath == "" {
		// element from scratch bez git-repo
		element.Build.ImageID = fmt.Sprintf("%s:%s.%s", element.EnvironmentID, element.Name, "not-implemented")

	} else {
		fileName := strings.TrimSuffix(element.SourceGit.FilePath, ".toml")
		fileName = strings.TrimPrefix(fileName, "/timoni/")
		fileName = strings.TrimPrefix(fileName, "/elements/")
		fileName = conv.KeyString(fileName)

		element.Build.ImageID = fmt.Sprintf("%s:%s.%s", conv.KeyString(element.SourceGit.RepoName), fileName, element.SourceGit.CommitHash)
	}

	// ----------------------------------------------------------
	// Build

	// tlog.PrintJSON(element.Build)

	switch element.Build.Type {
	case ElementPodBuildImage:
		if element.Build.Image == "" {
			return tlog.Error("build.Image is empty")
		}
		element.Build.Script = "FROM " + element.Build.Image

	case ElementPodBuildScript:
		if element.Build.Script == "" {
			return tlog.Error("build.Script is empty")
		}

	case ElementPodBuildDockerfilePath:
		if element.Build.DockerfilePath == "" {
			element.Build.DockerfilePath = "Dockerfile"
		}
		if element.SourceGit.RepoName == "" {
			return tlog.Error("build.Dockerfile not available with element from scratch made in webUI")
		}

		if element.Build.RootPath == "" {
			element.Build.RootPath = filepath.Dir(element.Build.DockerfilePath)
		}
		gitRepo := GitRepoGetByName(element.SourceGit.RepoName)
		gitRepo.Open()
		defer gitRepo.Unlock()

		element.Build.Script = string(
			gitRepo.GetFile(element.SourceGit.BranchName, element.SourceGit.CommitHash, element.Build.DockerfilePath),
		)

	default:
		return tlog.Error("invalid Build.Type")
	}

	for k, v := range element.Build.Variables {
		element.Build.Script = strings.ReplaceAll(element.Build.Script, "{{"+k+"}}", v)
	}

	element.Build.Script = strings.TrimSpace(conv.CleanString(element.Build.Script))
	if !strings.HasPrefix(strings.ToLower(element.Build.Script), "from ") {
		element.Build.Script = "FROM scratch\n\n" + element.Build.Script + "\n"
	}

	// run generic check
	return element.elementS.check(user)
}

func (element *elementPodS) Save(user *UserS) *tlog.RecordS {
	if err := element.check(user); err != nil {
		return err
	}
	return elementSave(element, user)
}

func (element *elementPodS) GetScale() *ElementScaleS {
	return &element.Scale
}

// ---

func (element *elementPodS) Merge(el EnvElementS) *tlog.RecordS {
	e, ok := el.(*elementPodS)
	if !ok {
		return tlog.Error(fmt.Sprintln("wrong type, expected", element.GetType(), "got", el.GetType()))
	}

	e.ParentSource = element.ParentSource
	e.Build.ImageID = element.Build.ImageID
	e.Scale = element.Scale
	e.Stateful = element.Stateful

	for k, v := range e.Variables {
		elementVar, ok := element.Variables[k]
		if !ok {
			continue
		}
		v.FirstValue = elementVar.FirstValue
		v.ResolvedValue = ""
	}

	return element.elementS.Merge(&e.elementS)
}

func (element *elementPodS) CopySecrets(el EnvElementS) *tlog.RecordS {
	e, ok := el.(*elementPodS)
	if !ok {
		return tlog.Error(fmt.Sprintln("wrong type, expected", element.GetType(), "got", el.GetType()))
	}
	return element.elementS.CopySecrets(&e.elementS)
}

func (element *elementPodS) DeleteFromKube() *tlog.RecordS {

	if element.EnvironmentID == "" {
		return tlog.Error("element.EnvironmentID is empty")
	}

	kClient := kube.GetKube()
	if kClient == nil {
		return tlog.Error("kube.Client not ready")
	}

	if element.Stateful {
		kubeStatefulSet := &kube.StatefulSetS{
			KubeClient: kClient,
			Namespace:  element.EnvironmentID,
			Name:       element.Name,
		}

		if kubeStatefulSet.GetObj() == nil {
			element.SetState(ElementStatusStopped)
		} else {
			kubeStatefulSet.Delete()
		}

	} else {
		kubeDeployment := &kube.DeploymentS{
			KubeClient: kClient,
			Namespace:  element.EnvironmentID,
			Name:       element.Name,
		}

		if kubeDeployment.GetObj() == nil && element.GetStatus().State != ElementStatusStopped {
			element.SetState(ElementStatusStopped)
		} else {
			kubeDeployment.Delete()
		}
	}

	if len(element.ExposePorts) > 0 {

		kubeService := &kube.ServiceS{
			KubeClient: kClient,
			Namespace:  element.EnvironmentID,
			Name:       element.Name,
		}

		if kubeService.GetObj() != nil {
			tlog.Info("DeleteFromKube")
			kubeService.Delete()
		}
	}

	es := element.GetStatus()
	es.PodCount = 0
	es.pods = nil
	es.Alerts = nil

	return nil
}

func (element *elementPodS) VariablesGet(returnSecrets bool) map[string]string {
	variables := element.elementS.VariablesGet(returnSecrets)
	
	if element.Schedule != "" {
		variables["EP_CRON_EXPRESSION"] = element.Schedule
	}

	return variables
}
