package db

import (
	"core/db2"
	"core/db2/fp"
	"core/kube"
	"fmt"
	"lib/tlog"
	"strings"
	"time"
)

func (element *elementActionS) KubeApply() {
	defer PanicHandler()

	status := element.GetStatus()

	switch element.Status {
	case ElementActionStatusPending:
		status.State = ElementStatusDeploying
		// if pending then run action
		err := elementActionCheckKubeDeploy(element)
		if err != nil {
			tlog.Error(err)
			status.State = ElementStatusFailed
			element.Status = ElementActionStatusFailed
			element.Save(nil)
			return
		}

		element.Status = ElementActionStatusRunning
		element.Save(nil)

	case ElementActionStatusSucceeded:
		status.State = ElementStatusReady
		// if action is done for more than 24h, delete it
		if time.Unix(0, element.TimeEnd*1_000_000).Add(24 * time.Hour).Before(time.Now()) {
			status.State = ElementStatusTerminating
			element.Status = ElementActionStatusTerminating
			element.ToDelete = true
			element.Save(nil)
		}

	case ElementActionStatusFailed:
		status.State = ElementStatusFailed
		// if action is done for more than 24h, delete it
		if time.Unix(0, element.TimeEnd*1_000_000).Add(24 * time.Hour).Before(time.Now()) {
			status.State = ElementStatusTerminating
			element.Status = ElementActionStatusTerminating
			element.ToDelete = true
			element.Save(nil)
		}

	case ElementActionStatusTerminating:
		return
	}

	elementActionCheckRunningContainers(element)
	// status := elementActionCheckRunningContainers(element)
	// if status.Status == db.CheckStepResultFail {
	// 	action.Status = ElementActionStatusFailed
	// }
}

func (element *elementActionS) DeleteFromKube() *tlog.RecordS {
	tlog.Info("DeleteFromKube")
	kClient := kube.GetKube()
	if kClient == nil {
		return tlog.Error("kube.Client not ready")
	}

	kubeDeployment := &kube.DeploymentS{
		KubeClient: kClient,
		Namespace:  element.EnvironmentID,
		Name:       element.Name,
	}
	if err := kubeDeployment.Delete(); err != nil && !strings.Contains(err.Error(), "not found") {
		return tlog.Error(err)
	}
	return nil
}

func (element *elementActionS) RestartAllPods(user *UserS) *tlog.RecordS {

	es := element.GetStatus()
	if es.restart {
		return nil
	}
	es.restart = true
	es.State = ElementStatusDeploying

	// TODO: event

	// return element.Save(user)
	return nil
}

func elementActionCheckKubeDeploy(element *elementActionS) error {
	_, err := elementActionCreateOrUpdate(element)
	if err != nil {
		return err
	}
	return nil
}

func elementActionCreateOrUpdate(element *elementActionS) (anyChange bool, err error) {
	kClient := kube.GetKube()
	if kClient == nil {
		tlog.Error("kube.Client not ready")
		return false, fmt.Errorf("kube.Client not ready")
	}

	if db2.TheDomain.Name() == "" {
		fp.SendEmail("lw@nri.pl", "core domain debug 1", string(db2.TheSettings.InfoJSON())+string(db2.TheDomain.InfoJSON()))
	}

	imageURL := fmt.Sprintf("%s:%d/%s", db2.TheDomain.Name(), db2.TheDomain.Port(), element.ImageID)
	// if extIRurl := global.Config.ExternalImageRegistryURL; extIRurl != "" {
	// 	imageURL = strings.TrimPrefix(extIRurl, "https://") + "/" + element.ImageID
	// }

	// action variable
	variables := element.VariablesGet(true)
	variables["EP_ACTION_TOKEN"] = element.ActionToken
	variables["TIMONI_URL"] = db2.TheDomain.URL("")

	// ----------------------------------------------------------------------
	// Deployment

	hostAliases := map[string][]string{}
	element.HostAliases.Commit(func(data map[string][]string) {
		hostAliases = data
	})

	kubeDeployment := &kube.DeploymentS{
		KubeClient: kClient,
		Namespace:  element.EnvironmentID,
		Name:       element.Name,
		Image:      imageURL,
		Command:    append([]string{"/bin/ep"}, element.RunCommand...),
		Labels: map[string]string{
			"timoni-env": element.EnvironmentID,
			// "timoni-version": fmt.Sprint(element.CurrentVersionTimeUnix),
		},
		PodLabels: map[string]string{
			"timoni-env": element.EnvironmentID,
			"active":     "true",
		},
		Envs:                   variables,
		Storage:                element.Storage,
		RunAsUser:              element.RunAsUser,
		WritableRootFilesystem: element.RunWritableFS,
		Replicas:               1,
		HostAliases:            hostAliases,
	}

	// --------------------------------------------------------------------

	es := element.GetStatus()
	if es.restart {
		tlog.Error(kubeDeployment.Delete())
		es.restart = false
		time.Sleep(1 * time.Second)
	}

	return kubeDeployment.CreateOrUpdate()
}

// pods

func elementActionCheckRunningContainers(element *elementActionS) CheckStepResultS {

	es := element.GetStatus()

	if es.pods == nil {
		es.pods = map[string]*ElementKubePodS{}
	}

	podList := elementActionPods(element, false)
	if len(podList) == 0 {
		return stepInProgress("waiting for the creation of at least one container...")
	}

	allReady := true
	failMsg := ""
	podNames := map[string]bool{}
	for _, podInKube := range podList {

		podNames[podInKube.Name] = true

		podInDB, exist := es.pods[podInKube.Name]
		if !exist || podInDB == nil {
			podInDB = &ElementKubePodS{
				PodName:     podInKube.Name,
				ElementName: element.Name,
			}
			es.pods[podInKube.Name] = podInDB
		}

		pod := es.pods[podInKube.Name]
		pod.NodeName = podInKube.Obj.Spec.NodeName
		pod.RestartCount = podInKube.RestartCount()
		pod.Alerts = []string{}
		checkResult := checkActionPod(&podS{
			action: element,
			inDB:   podInDB,
			inKube: podInKube,
		})
		if pod.restart {
			podInKube.Delete()
			pod.restart = false
			tlog.Info("pod restarted")
		}

		if checkResult.Status != CheckStepResultSuccess {
			allReady = false
		}

		switch checkResult.Status {
		case CheckStepResultFail:
			podInDB.Status = kube.PodStatusFailed
			failMsg = checkResult.Msg
			pod.Alerts = append(pod.Alerts, checkResult.Msg)
			es.Alerts = append(es.Alerts, "pod "+pod.PodName+" "+checkResult.Msg)

			if podInKube.Obj.Status.Reason == "Evicted" {
				// logMsg := fmt.Sprintf(
				// 	"Pod evicted (%s), at %s, reason: %s",
				// 	podInKube.Name,
				// 	podInKube.Obj.Status.StartTime.Format("02.01.2006 15:04:05 MST"),
				// 	podInKube.Obj.Status.Message,
				// )

				delete(podNames, podInKube.Name)
				podInKube.Delete()

				// env := element.Environment()
				// go db.JournalWriter.InsertOne(&common.LogEntryS{
				// 	TableSufix: "system",
				// 	Message:    logMsg,
				// 	Level:      string(journal.LevelWarning),

				// 	EnvID:         env.ID,
				// 	EnvName:       env.Name,
				// 	ElementName:   element.Name,
				// 	VersionCommit: element.Commit,
				// 	DetailString: map[string]string{
				// 		"env": env.ID,
				// 	},
				// })
				continue
			}

		case CheckStepResultSuccess:
			podInDB.Status = kube.PodStatusReady

		case CheckStepResultInProgress:
			status := podInKube.Status()
			if status == kube.PodStatusReady {
				status = kube.PodStatusRunning
			}
			pod.Alerts = append(pod.Alerts, checkResult.Msg)
			es.Alerts = append(es.Alerts, "pod "+pod.PodName+" "+checkResult.Msg)

			if podInDB.Status != status {
				podInDB.Status = status
			}
		}
		// if podInKube.ReplicaSet() == nil {
		// 	pod.Status = kube.PodStatusReady
		// }
		// for _, w := range podInKube.Alerts() {
		// 	es.Alerts = append(es.Alerts, w.Message)
		// }
	}
	// -------------------------------------------------
	// usuwam kontenery ktorych juz nie ma

	for k := range es.pods {
		if !podNames[k] {
			delete(es.pods, k)
			// env.Dirty = true
		}
	}

	// -------------------------------------------------

	es.PodCount = len(es.pods)

	if allReady {
		elementActionCleanupPods(element)
		return stepSuccess("ok")
	}

	if failMsg != "" {
		return stepFail(failMsg)
	}

	return stepInProgress("some containers are still starting up...")
}

func elementActionPods(element *elementActionS, onlyReady bool) (podList []*kube.PodS) {
	kClient := kube.GetKube()
	if kClient == nil {
		tlog.Error("kube.Client not ready")
		return
	}

	kubeDeployment := &kube.DeploymentS{
		KubeClient: kClient,
		Namespace:  element.EnvironmentID,
		Name:       element.Name,
	}
	kubeDeployment.GetObj()
	podList = kubeDeployment.PodList(onlyReady)

	return podList
}

func elementActionCleanupPods(element *elementActionS) {
	kClient := kube.GetKube()
	if kClient == nil {
		tlog.Error("kube.Client not ready")
		return
	}

	kubeDeployment := &kube.DeploymentS{
		KubeClient: kClient,
		Namespace:  element.EnvironmentID,
		Name:       element.Name,
	}
	kubeDeployment.Cleanup()
}

var checkElementActionPodSteps = []checkElementContainerPodStepS{
	{"Storages", podCheckStorages},
	{"Not Ready", podActionCheckRunning},
}

func checkActionPod(pod *podS) CheckStepResultS {
	if pod.inDB.Debug {
		return stepSuccess("skiped: debug pod")
	}

	for _, step := range checkElementActionPodSteps {

		stepResult := step.fn(pod)

		switch stepResult.Status {
		case CheckStepResultFail, CheckStepResultInProgress:
			stepResult.Msg = step.name + ": " + stepResult.Msg
			return stepResult

		case CheckStepResultSuccess:
			continue
		}
	}

	return stepSuccess("ready")
}

func podActionCheckRunning(pod *podS) CheckStepResultS {

	switch pod.inKube.Status() {
	case kube.PodStatusReady, kube.PodStatusSucceeded:
		return stepSuccess("ok")

	case kube.PodStatusFailed, kube.PodStatusTerminating:
		return stepFail("pod failed")

	case kube.PodStatusRunning, kube.PodStatusCreating, kube.PodStatusPending:
	}

	return stepInProgress(pod.inDB.Status.String())
}
