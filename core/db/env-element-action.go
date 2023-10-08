package db

import (
	"core/kube"
	"errors"
	"fmt"
	"lib/tlog"
	"lib/utils"
	"lib/utils/maps"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ElementActionStatusT string

const (
	ElementActionStatusPending     ElementActionStatusT = "Pending"
	ElementActionStatusRunning     ElementActionStatusT = "Running"
	ElementActionStatusSucceeded   ElementActionStatusT = "Succeeded"
	ElementActionStatusFailed      ElementActionStatusT = "Failed"
	ElementActionStatusTerminating ElementActionStatusT = "Terminating"
)

type elementActionS struct {
	elementS

	ActionName string
	Status     ElementActionStatusT
	TimeBegin  int64 // when action started, timestamp in milliseconds
	TimeEnd    int64 // when action finished, timestamp in milliseconds

	// unique (for every action) token used in communication with entry-point
	ActionToken string

	ParentName string // from which element this action was started
	ImageID    string // from parent

	RunCommand []string
	// RunVariables  map[string]string // from parent
	RunAsUser     []int64 // from parent
	RunWritableFS bool    // from parent

	Storage               map[string]*kube.StorageS      // key=mount point in pod
	HostAliases           maps.SafeMap[string, []string] // key=IP, value=list of domains
	ApplyVariablesOnFiles []string
	// Schedule              string // cron schedule format

	Scale *ElementScaleS
}

func (element *elementPodS) RunAction(actionName string, user *UserS) error {

	if element.Type != ElementSourceTypePod {
		return errors.New("element is not container")
	}

	cmd, ok := element.Actions[actionName]
	if !ok {
		return errors.New("action not found")
	}

	copyElementPod := *utils.DeepCopy(element)

	env := element.GetEnvironment()

	now := time.Now().UnixMilli()
	newElementName := strings.ToLower(fmt.Sprintf("%s-%s-%d", element.Name, actionName, now))

	for key, storage := range copyElementPod.Storage {
		// if container has stateful disk - do not attach it to new element
		if copyElementPod.Stateful && storage.Type == "block" {
			delete(copyElementPod.Storage, key)
		}
	}

	elementAction := elementActionS{
		elementS: elementS{
			Name:                newElementName,
			EnvironmentID:       copyElementPod.EnvironmentID,
			Type:                ElementSourceTypeAction,
			SourceGit:           copyElementPod.SourceGit,
			AutoUpdate:          false,
			ToDelete:            false,
			Variables:           copyElementPod.Variables,
			VariablesDependence: copyElementPod.VariablesDependence,
		},

		ActionName:  actionName,
		Status:      ElementActionStatusPending,
		TimeBegin:   now,
		TimeEnd:     0,
		ActionToken: uuid.NewString(),

		ParentName: copyElementPod.Name,
		ImageID:    copyElementPod.Build.ImageID,

		RunCommand:            cmd,
		RunAsUser:             copyElementPod.RunAsUser,
		RunWritableFS:         copyElementPod.RunWritableFS,
		Storage:               copyElementPod.Storage,
		HostAliases:           *copyElementPod.HostAliases.Copy().Safe(),
		ApplyVariablesOnFiles: copyElementPod.ApplyVariablesOnFiles,

		Scale: &ElementScaleS{
			MaxOnePod:       true,
			NrOfPodsMin:     1,
			NrOfPodsMax:     1,
			NrOfPodsCurrent: 0,
			CPUTargetProc:   0,
		},
	}

	errx := env.ElementAdd(newElementName, &elementAction, user, true)
	if errx != nil {
		return errors.New(errx.Message)
	}

	return nil
}

func (element *elementActionS) getParent() EnvElementS {
	env := EnvironmentMap.Get(element.EnvironmentID)
	if env == nil {
		return nil
	}
	parent := env.GetElement(element.ParentName)
	if parent == nil {
		return nil
	}
	return parent
}

func (element *elementActionS) RebuildImage(imageID string, user *UserS) *tlog.RecordS {
	return element.getParent().RebuildImage(imageID, user)
}
func (element *elementActionS) GetImage() *ImageS {
	return element.getParent().GetImage()
}

func (element *elementActionS) GetScale() *ElementScaleS {
	return element.Scale
}

func (element *elementActionS) Save(user *UserS) *tlog.RecordS {
	return elementSave(element, user)
}

func (element *elementActionS) Merge(el EnvElementS) *tlog.RecordS {
	return nil
}

func (element *elementActionS) CopySecrets(el EnvElementS) *tlog.RecordS {
	return nil
}

func (e *elementActionS) GetResources() (cpu, ram float64) {
	return 0, 0
}

func (e *elementActionS) SetMetrics(cpu, ram float64) {
	e.CPUUsageAvgCores = cpu
	e.RAMUsageAvgMB = ram
}
