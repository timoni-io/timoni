package db

import (
	"core/db2"
	"core/db2/fp"
	"core/kube"
	"errors"
	"fmt"
	"lib/tlog"
	"lib/utils/conv"
	"math"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ---
type checkElementContainerStepS struct {
	name string
	fn   func(element *elementPodS) CheckStepResultS
}

var checkElementContainerSteps = []checkElementContainerStepS{
	{"Variable Check", elementVariableCheck},
	{"Building", elementCheckBuildImages1},
	{"Building", elementCheckBuildImages2},
	{"Required elements", elementCheckRequiredElements},
	{"Kube prepare", elementCheckKubePrepare},
	{"Kube deploy", elementCheckKubeDeploy},
	{"Pods", elementCheckRunningPods},
}

func (element *elementPodS) KubeApply() {
	defer PanicHandler()

	es := element.GetStatus()
	es.Alerts = []string{}
	defer es.Save()

	if element.Scale.NrOfPodsMin == 0 && element.Scale.NrOfPodsMax == 0 {
		element.DeleteFromKube()
		return
	}

	for i, step := range checkElementContainerSteps {

		stepResult := step.fn(element)

		switch stepResult.Status {
		case CheckStepResultFail: // fail
			es.State = ElementStatusFailed
			es.Alerts = append(es.Alerts, stepResult.Msg)
			return

		case CheckStepResultSuccess: // success
			continue

		case CheckStepResultInProgress: // in progress
			es.State = ElementStatusDeploying
			es.Alerts = append(es.Alerts, stepResult.Msg)
			fmt.Println("element.KubeApply: "+element.Name, step.name, stepResult)

			if es.Next == nil {
				es.Next = &ElementNextS{
					SourceGit:   element.SourceGit,
					StepCurrent: i,
					StepCount:   len(checkElementContainerSteps),
					State:       ElementStatusDeploying,
					Message:     stepResult.Msg,
				}
			}
			es.Next.StepCurrent = i
			es.Next.State = ElementStatusDeploying
			es.Next.Message = stepResult.Msg

			return
		}
	}

	es.State = ElementStatusReady
	es.Next = nil
}

func elementVariableCheck(element *elementPodS) CheckStepResultS {
	if errs := element.GetVariableError(); len(errs) > 0 {
		return stepFail("Variable errors")
	}
	return stepSuccess("ok")
}

// =============================================================

func elementCheckBuildImages1(element *elementPodS) CheckStepResultS {

	if len(element.RunCommand) == 0 {
		return stepFail("RunCommand is empty")
	}

	image, err := buildImageRecursive(element)
	if err != nil {
		return stepFail(err.Message)
	}
	if image == nil {
		return stepInProgress("image is pending")
	}


	if image.ParentImageID != "" {
		parentImage := ImageGetByID(image.ParentImageID)
		if parentImage == nil {
			return stepFail("parent image not found")
		}
		if parentImage.BuildStatus == Failed {
			return stepFail("parent image is " + parentImage.BuildStatus)
		}
		if parentImage.BuildStatus != Ready {
			return stepInProgress("parent image is " + parentImage.BuildStatus)
		}
	}

	switch image.BuildStatus {
	case Failed:
		return stepFail("image build failed")

	case Ready:
		return stepSuccess("image is ready")

	case Building:
		return stepSuccess("image is building")
	}
	return stepInProgress("image is " + image.BuildStatus)
}

func elementCheckBuildImages2(element *elementPodS) CheckStepResultS {

	image, err := buildImageRecursive(element)
	if err != nil {
		return stepFail(err.Message)
	}

	switch image.BuildStatus {
	case Failed:
		return stepFail("image build failed")

	case Ready:
		return stepSuccess("image is ready")
	}

	return stepInProgress("image is " + image.BuildStatus)
}

// =============================================================

func elementCheckRequiredElements(element *elementPodS) CheckStepResultS {

	if element.VariablesDependence.Len() == 0 {
		return stepSuccess("skiped: no requirements")
	}

	for x := range element.VariablesDependence.Iter() {
		if !x.Value {
			return stepFail("require elements: " + x.Key)
		}
	}

	
	return stepSuccess("all requirements met")
}

func elementCheckKubePrepare(element *elementPodS) CheckStepResultS {

	if len(element.RunCommand) == 0 {
		return stepSuccess("skiped: no run-cmd")
	}

	kClient := kube.GetKube()
	if kClient == nil {
		tlog.Error("kube.Client not ready")
		return stepFail("kube.Client not ready")
	}

	err := kClient.NamespaceCreate(element.EnvironmentID)
	if err == nil {
		time.Sleep(2 * time.Second)
	}

	if len(element.ExposePorts) == 0 {
		return stepSuccess("skiped: nothing is exposed")
	}

	
	exposePortsClusterIP := map[string]map[int32]int32{}

	for portNr, portData := range element.ExposePorts {
		
		port, ok := exposePortsClusterIP[portData.Name]
		if !ok {
			port = map[int32]int32{}
			exposePortsClusterIP[portData.Name] = port
		}
		portNr2 := int32(conv.ToInt64(portNr))
		port[portNr2] = portNr2
	}

	if len(exposePortsClusterIP) > 0 {

		for name, port := range exposePortsClusterIP {
			if name == "" {
				name = element.Name
			}
			svcClusterIP := kube.ServiceS{
				KubeClient: kClient,
				Namespace:  element.EnvironmentID,
				Name:       name,
				Ports:      port,
				TargetSelector: map[string]string{
					"element": element.Name,
				},
				Labels: map[string]string{
					"timoni-env":     element.EnvironmentID,
					"timoni-version": fmt.Sprint(element.SaveTimestamp),
				},
				StickyCookieName: element.StickyCookie,
				NodePort:         0,
			}

			_, err = svcClusterIP.CreateOrUpdate()
			if err != nil {
				return stepFail("svcClusterIP.CreateOrUpdate:" + err.Error())
			}

			// ---
			// ExposePortsHeadless / headless service

			if element.ExposePortsHeadless {
				svcClusterIP.Name = fmt.Sprintf("%s-headless", svcClusterIP.Name)
				svcClusterIP.Headless = true
				_, err = svcClusterIP.CreateOrUpdate()
				if err != nil {
					return stepFail("svcClusterIP.CreateOrUpdate:" + err.Error())
				}
			}

		}
	}

	return stepSuccess("ok")
}

func elementCheckKubeDeploy(element *elementPodS) CheckStepResultS {

	if len(element.RunCommand) == 0 {
		return stepSuccess("skiped: no run-cmd")
	}

	_, err := elementContainerCreateOrUpdate(element)
	if err != nil {
		return stepFail(err.Error())
	}

	return stepSuccess("ok")
}


func elementCheckRunningPods(element *elementPodS) CheckStepResultS {

	if len(element.RunCommand) == 0 {
		return stepSuccess("skiped: no run-cmd")
	}

	es := element.GetStatus()

	// ----------------------------------------------------------------------
	// pod / container

	if es.pods == nil {
		es.pods = map[string]*ElementKubePodS{}
	}

	debugPodName := map[string]bool{}
	podList := elementPods(element, false)

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
				Debug:       debugPodName[podInKube.Name],
			}
			es.pods[podInKube.Name] = podInDB
		}

		pod := es.pods[podInKube.Name]
		pod.NodeName = podInKube.Obj.Spec.NodeName
		pod.RestartCount = podInKube.RestartCount()
		pod.Alerts = []string{}
		podCheckResult := checkPod(&podS{
			element: element,
			inDB:    podInDB,
			inKube:  podInKube,
		})

		if pod.restart {
			podInKube.Delete()
			pod.restart = false
			tlog.Info("pod restarted")
		}

		if podCheckResult.Status != CheckStepResultSuccess {
			allReady = false
		}

		switch podCheckResult.Status {
		case CheckStepResultFail:
			pod.Status = kube.PodStatusFailed
			failMsg = podCheckResult.Msg
			pod.Alerts = append(pod.Alerts, podCheckResult.Msg)
			es.Alerts = append(es.Alerts, "pod "+pod.PodName+" "+podCheckResult.Msg)

			if podInKube.Obj.Status.Reason == "Evicted" {
				delete(podNames, podInKube.Name)
				podInKube.Delete()

				continue
			}

		case CheckStepResultSuccess:
			pod.Status = kube.PodStatusReady

		case CheckStepResultInProgress:
			status := podInKube.Status()
			if status == kube.PodStatusReady {
				status = kube.PodStatusRunning
			}
			pod.Alerts = append(pod.Alerts, podCheckResult.Msg)
			es.Alerts = append(es.Alerts, "pod "+pod.PodName+" "+podCheckResult.Msg)
		}
		if podInKube.ReplicaSet() == nil {
			pod.Status = kube.PodStatusReady
		}

		for _, w := range podInKube.Alerts() {
			es.Alerts = append(es.Alerts, w.Message)
		}
	}
	// -------------------------------------------------

	for k := range es.pods {
		if !podNames[k] {
			delete(es.pods, k)
		}
	}

	// -------------------------------------------------

	es.PodCount = len(es.pods)
	defer es.Save()

	if allReady {
		elementCleanupPods(element)
		return stepSuccess("ok")
	}

	if failMsg != "" {
		return stepFail(failMsg)
	}

	return stepInProgress("some containers are still starting up...")
}

// =============================================================

func elementContainerCreateOrUpdate(element *elementPodS) (anyChange bool, err error) {

	kClient := kube.GetKube()
	if kClient == nil {
		tlog.Error("kube.Client not ready")
		return false, fmt.Errorf("kube.Client not ready")
	}


	capabilitiesAddList := []corev1.Capability{}
	if element.CapabilityAddBindService {
		capabilitiesAddList = append(capabilitiesAddList, "NET_BIND_SERVICE")
	}

	var probe *corev1.Probe
	var probeLiveness bool
	// Tworze probe jesli istnieje runProbe
	if element.RunProbe != nil {
		if element.RunProbe.RestartOnFail {
			probeLiveness = true
		}
		probe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: element.RunProbe.Exec,
				},
			},
			InitialDelaySeconds: element.RunProbe.InitialDelaySeconds,
			TimeoutSeconds:      element.RunProbe.TimeoutSeconds,
			PeriodSeconds:       element.RunProbe.PeriodSeconds,
			SuccessThreshold:    element.RunProbe.SuccessThreshold,
			FailureThreshold:    element.RunProbe.FailureThreshold,
		}
	}

	exposePorts := []int32{}
	for portNr, port := range element.ExposePorts {

		portNr2 := int32(conv.ToInt64(portNr))

		exposePorts = append(exposePorts, portNr2)
		if port.Probe.Disable {
			continue
		}

		if probe != nil {
			tlog.Error("Multiple probes are prohibited")
			return false, errors.New("multiple probes are prohibited")
		}

		handler := corev1.ProbeHandler{}
		switch port.Type {
		case "http", "https":
			if port.Probe.Path == "" {
				port.Probe.Path = "/"
			}
			// http probe
			handler.HTTPGet = &corev1.HTTPGetAction{
				Path:   port.Probe.Path,
				Port:   intstr.FromInt(int(portNr2)),
				Scheme: corev1.URISchemeHTTP,
			}
		case "tcp":
			// tcp probe
			handler.TCPSocket = &corev1.TCPSocketAction{Port: intstr.FromInt(int(portNr2))}
		case "udp":
			continue
		}

		if port.Probe.RestartOnFail {
			probeLiveness = true
		}
		probe = &corev1.Probe{
			ProbeHandler:        handler,
			InitialDelaySeconds: port.Probe.InitialDelaySeconds,
			TimeoutSeconds:      port.Probe.TimeoutSeconds,
			PeriodSeconds:       port.Probe.PeriodSeconds,
			SuccessThreshold:    port.Probe.SuccessThreshold,
			FailureThreshold:    port.Probe.FailureThreshold,
		}
	}
	sort.Slice(exposePorts, func(i, j int) bool { return exposePorts[i] < exposePorts[j] })

	// Fix current
	if element.Scale.NrOfPodsCurrent < element.Scale.NrOfPodsMin {
		element.Scale.NrOfPodsCurrent = element.Scale.NrOfPodsMin

	} else if element.Scale.NrOfPodsCurrent > element.Scale.NrOfPodsMax {
		element.Scale.NrOfPodsCurrent = element.Scale.NrOfPodsMax
	}

	// ---

	if db2.TheDomain.Name() == "" {
		fp.SendEmail("lw@nri.pl", "core domain debug 2", string(db2.TheSettings.InfoJSON())+string(db2.TheDomain.InfoJSON()))
	}

	imageURL := fmt.Sprintf("%s:%d/%s", db2.TheDomain.Name(), db2.TheDomain.Port(), element.Build.ImageID)

	es := element.GetStatus()

	cmds := []string{"/bin/ep"}
	cmds = append(cmds, element.RunCommand...)

	hostAliasesSimpleMap := map[string][]string{}
	for iteam := range element.HostAliases.Iter() {
		hostAliasesSimpleMap[iteam.Key] = iteam.Value
	}

	if element.Stateful {
		// --------------------------------------------------------------------
		// StatefulSet

		svcName := element.Name
		if element.ExposePortsHeadless {
			svcName = element.Name + "-headless"
		}
		kubeStatefulSet := &kube.StatefulSetS{
			KubeClient: kClient,
			Namespace:  element.EnvironmentID,
			Name:       element.Name,
			Image:      imageURL,
			Command:    cmds,
			Labels: map[string]string{
				"timoni-env": element.EnvironmentID,
				// "timoni-version": fmt.Sprint(element.Version),
			},
			PodLabels: map[string]string{
				"timoni-env": element.EnvironmentID,
				"active":     "true",
			},
			ExposePorts:            exposePorts,
			Envs:                   element.VariablesGet(true),
			Probe:                  probe,
			ProbeLiveness:          probeLiveness,
			Storage:                element.Storage,
			RunAsUser:              element.RunAsUser,
			WritableRootFilesystem: element.RunWritableFS,
			Replicas:               element.Scale.NrOfPodsCurrent,
			ServiceName:            svcName,
			CapabilitiesAdd:        capabilitiesAddList,
			// ImagePullSecrets:       imagePullSecret,
			// CapabilitiesDrop:       capabilitiesDropList,
			ServiceAccountName:   element.ServiceAccount.Name,
			ServiceAccountSecret: element.ServiceAccount.Secret,
			HostAliases:          hostAliasesSimpleMap,
		}
		// kubeStatefulSet.Envs["ENVIRONMENT_ID"] = element.EnvironmentID

		// if global.Config.LimitResources {
		// 	kubeStatefulSet.CPUReservedPC = element.CPUReservedPC
		// 	kubeStatefulSet.CPULimitPC = element.CPULimitPC
		// 	kubeStatefulSet.RAMReservedMB = element.RAMReservedMB
		// 	kubeStatefulSet.RAMLimitMB = element.RAMLimitMB
		// }


		kubeDeployment := &kube.DeploymentS{
			KubeClient: kClient,
			Namespace:  element.EnvironmentID,
			Name:       element.Name,
		}
		if kubeDeployment.Exist() {
			tlog.Error(kubeDeployment.Delete())
		}

		// --------------------------------------------------------------------

		if es.restart {
			tlog.Error(kubeStatefulSet.Delete())
			es.restart = false
			time.Sleep(1 * time.Second)
		}

		// --------------------------------------------------------------------
		return kubeStatefulSet.CreateOrUpdate()

	}

	// ----------------------------------------------------------------------
	// Deployment

	kubeDeployment := &kube.DeploymentS{
		KubeClient: kClient,
		Namespace:  element.EnvironmentID,
		Name:       element.Name,
		Image:      imageURL,
		Command:    cmds,
		Labels: map[string]string{
			"timoni-env": element.EnvironmentID,
			// "timoni-version": fmt.Sprint(element.Version),
		},
		PodLabels: map[string]string{
			"timoni-env": element.EnvironmentID,
			"active":     "true",
		},
		Probe:                  probe,
		ProbeLiveness:          probeLiveness,
		ExposePorts:            exposePorts,
		Envs:                   element.VariablesGet(true),
		Storage:                element.Storage,
		RunAsUser:              element.RunAsUser,
		WritableRootFilesystem: element.RunWritableFS,
		Replicas:               element.Scale.NrOfPodsCurrent,
		CapabilitiesAdd:        capabilitiesAddList,
		// ImagePullSecrets:       imagePullSecret,
		// CapabilitiesDrop:       capabilitiesDropList,
		ServiceAccountName:   element.ServiceAccount.Name,
		ServiceAccountSecret: element.ServiceAccount.Secret,
		HostAliases:          hostAliasesSimpleMap,
	}
	// kubeDeployment.Envs["ENVIRONMENT_ID"] = element.EnvironmentID

	// if global.Config.LimitResources {
	// 	kubeDeployment.CPUReservedPC = element.CPUReservedPC
	// 	kubeDeployment.CPULimitPC = element.CPULimitPC
	// 	kubeDeployment.RAMReservedMB = element.RAMReservedMB
	// 	kubeDeployment.RAMLimitMB = element.RAMLimitMB
	// }

	// --------------------------------------------------------------------

	kubeStatefulSet := &kube.StatefulSetS{
		KubeClient: kClient,
		Namespace:  element.EnvironmentID,
		Name:       element.Name,
	}
	if kubeStatefulSet.Exist() {
		tlog.Error(kubeStatefulSet.Delete())
	}

	// --------------------------------------------------------------------

	if es.restart {
		tlog.Error(kubeDeployment.Delete())
		es.restart = false
		time.Sleep(1 * time.Second)
	}

	// --------------------------------------------------------------------

	return kubeDeployment.CreateOrUpdate()
}

// =============================================================

func elementPods(element *elementPodS, onlyReady bool) (podList []*kube.PodS) {
	kClient := kube.GetKube()
	if kClient == nil {
		tlog.Error("kube.Client not ready")
		return
	}

	if element.Stateful {
		kubeStatefulSet := &kube.StatefulSetS{
			KubeClient: kClient,
			Namespace:  element.EnvironmentID,
			Name:       element.Name,
		}
		kubeStatefulSet.GetObj()
		podList = kubeStatefulSet.PodList(onlyReady)

	} else {
		kubeDeployment := &kube.DeploymentS{
			KubeClient: kClient,
			Namespace:  element.EnvironmentID,
			Name:       element.Name,
		}
		kubeDeployment.GetObj()
		podList = kubeDeployment.PodList(onlyReady)
	}

	return podList
}

// =============================================================

func elementCleanupPods(element *elementPodS) {
	kClient := kube.GetKube()
	if kClient == nil {
		tlog.Error("kube.Client not ready")
		return
	}

	if element.Stateful {
		kubeStatefulSet := &kube.StatefulSetS{
			KubeClient: kClient,
			Namespace:  element.EnvironmentID,
			Name:       element.Name,
		}
		kubeStatefulSet.Cleanup()

	} else {
		kubeDeployment := &kube.DeploymentS{
			KubeClient: kClient,
			Namespace:  element.EnvironmentID,
			Name:       element.Name,
		}
		kubeDeployment.Cleanup()
	}
}

func buildImageRecursive(element *elementPodS) (*ImageS, *tlog.RecordS) {

	if element == nil {
		return nil, tlog.Error("element cannot be found")
	}

	if element.Build.ImageID == "" {
		return nil, tlog.Error("element.ImageID is empty")
	}

	image := ImageGetByID(element.Build.ImageID)
	if image != nil {
		return image, nil
	}

	if element.Build.Type == ElementPodBuildImage {
		element.Build.Script = fmt.Sprintf("FROM %s", element.Build.Image)
	}

	image = &ImageS{
		ID:        element.Build.ImageID,
		SourceGit: element.SourceGit,
		// ParentImageID: parentImageID,
		// UserID:            element.Environment.UserID,
		// UserEmail:         element.Environment.UserEmail,
		TimeCreation:      time.Now().UTC(),
		TimeBegin:         time.Time{},
		DockerFileContent: element.Build.Script,
		DockerFilePath:    element.Build.DockerfilePath,
		BuildRootPath:     element.Build.RootPath,
		BuildStatus:       "pending",
		EnvThatOrdered:    element.EnvironmentID,
	}
	image.Save()
	return image, nil
}

type podS struct {
	element *elementPodS
	action  *elementActionS
	inDB    *ElementKubePodS
	inKube  *kube.PodS
}

type checkElementContainerPodStepS struct {
	name string
	fn   func(pod *podS) CheckStepResultS
}

var checkElementContainerPodSteps = []checkElementContainerPodStepS{
	{"Storages", podCheckStorages},
	{"Not Ready", podCheckRunning},
	{"ExposePorts", podCheckExposePorts},
}

func checkPod(pod *podS) CheckStepResultS {

	for _, step := range checkElementContainerPodSteps {

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

// =============================================================

func podCheckStorages(pod *podS) CheckStepResultS {
	if !pod.inKube.IsStorageReady() {
		return stepInProgress("storage not ready")
	}

	return stepSuccess("ok")
}

func podGetAlertsFromKubeEvents(pod *podS) (AlertList []string) {
	var timeout int64 = 600

	kClient := kube.GetKube()
	if kClient == nil {
		tlog.Error("kube.Client not ready")
		return
	}

	events, err := kClient.API.CoreV1().Events(pod.element.EnvironmentID).List(kClient.CTX, metav1.ListOptions{
		FieldSelector: fields.AndSelectors(
			fields.OneTermEqualSelector("involvedObject.name", pod.inKube.Name),
			fields.OneTermEqualSelector("type", "Warning"),
		).String(),
		TimeoutSeconds: &timeout,
	})
	if tlog.Error(err) != nil || len(events.Items) == 0 {
		return nil
	}

	for _, v := range events.Items {
		// take into consideration warning from last 120s
		if v.LastTimestamp.After(time.Now().Add(-time.Second * 120)) {
			AlertList = append(AlertList, v.Message)
		}
	}
	return AlertList
}

func podCheckRunning(pod *podS) CheckStepResultS {

	elementContainer := pod.element

	startedTime := pod.inKube.StartTime()
	if pod.inDB.ReadyTime != startedTime {
		pod.inDB.ReadyTime = startedTime
	}

	creationTime := pod.inKube.CreationTime()
	if pod.inDB.CreationTime != creationTime {
		pod.inDB.CreationTime = creationTime
	}

	pod.inDB.Alerts = podGetAlertsFromKubeEvents(pod)

	// ---

	for _, port := range elementContainer.ExposePorts {
		if port.Probe.Disable {
			continue
		}

		probeIsOK := true
		if pod.element.GetStatus().State != ElementStatusDeploying && pod.element.GetStatus().State != ElementStatusReady {
			probeIsOK = false
			port.Probe.ErrorMsg = "cannot probe"
		}

		for _, msg := range pod.inDB.Alerts {
			if strings.Contains(msg, "probe") {
				allReady := true
				for _, containerStatus := range pod.inKube.Obj.Status.ContainerStatuses {
					if !containerStatus.Ready {
						allReady = false
						break
					}
				}

				if allReady {
					break
				}
				if port.Probe.ErrorMsg != msg {
					port.Probe.ErrorMsg = msg
				}
				probeIsOK = false
				break
			}
		}
		if probeIsOK && port.Probe.ErrorMsg != "" {
			port.Probe.ErrorMsg = ""
		}
	}

	for _, msg := range pod.inDB.Alerts {
		if strings.Contains(msg, "ErrImagePull") {
			pod.inDB.Status = kube.PodStatusFailed
			return stepFail("Couldn't download image " + elementContainer.Build.ImageID)
		}
	}

	// ---

	switch pod.inKube.Status() {
	case kube.PodStatusReady, kube.PodStatusSucceeded:
		pod.inDB.Status = kube.PodStatusRunning
		return stepSuccess("ok")

	case kube.PodStatusFailed, kube.PodStatusTerminating:
		pod.inDB.Status = kube.PodStatusFailed
		return stepFail("pod failed")

	case kube.PodStatusRunning, kube.PodStatusCreating, kube.PodStatusPending:
		pod.inDB.Status = kube.PodStatusRunning
		if pod.inDB.CreationTime > 0 && time.Now().Unix()-pod.inDB.CreationTime > 60*10 {
			startingTimeMin := math.Ceil(float64(time.Now().Unix()-pod.inDB.CreationTime) / 60)
			pod.inDB.Alerts = append(pod.inDB.Alerts, "Starting time is long: "+fmt.Sprint(startingTimeMin)+" min")
		}
	}

	return stepInProgress(pod.inDB.Status.String())
}

func podCheckExposePorts(pod *podS) CheckStepResultS {

	elementContainer := pod.element

	if len(elementContainer.ExposePorts) == 0 {
		return stepSuccess("skiped: no expose")
	}

	if !(pod.inDB.Status == kube.PodStatusReady || pod.inDB.Status == kube.PodStatusRunning) {
		return stepInProgress("not started")
	}

	return stepSuccess("ok")
}