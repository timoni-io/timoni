package kubesync

import (
	"core/db"
	"core/kube"
	"lib/tlog"
	"sync"
	"time"
)

func envCheck(env *db.EnvironmentS) {
	defer db.PanicHandler()

	timeStart := time.Now()
	defer func() {
		t := time.Since(timeStart)
		if t.Seconds() > 10 {
			tlog.Warning("env check {{env}} in {{time}}", tlog.Vars{
				"env":  env.Name,
				"time": t.String(),
			})
		}
	}()

	kClient := kube.GetKube()
	if kClient == nil {
		tlog.Error("kube.Client not ready")
		return
	}

	defer func() {
		defer envLoopCheckingLock.Unlock()
		envLoopCheckingLock.Lock()
		envLoopChecking[env.ID] = false
	}()

	if env.ToDelete {

		// Delete namespace
		ns := kClient.NamespaceGet(env.ID)
		if ns == nil {
			env.Delete(nil)
			return
		}

		if ns.Status.Phase != "Terminating" {
			env.SetStatusForAllElements(db.ElementStatusTerminating, nil)
			kClient.NamespaceDelete(env.ID)
		}

		return
	}


	err := kClient.NamespaceCreate(env.ID)
	if err == nil {
		time.Sleep(2 * time.Second)
	}


	go updateResources(env)

	if env.GitOps.Enabled {
		env.FromGitOps()
	}

	wg := new(sync.WaitGroup)
	for _, elName := range env.Elements.Keys() {
		element := env.GetElement(elName)
		if element.GetToDelete() {
			if element.DeleteFromKube() == nil {
				env.ElementDelete(elName, nil)
			}
			continue
		}

		if element.GetActive() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				element.KubeApply()
			}()
		} else {
			element.DeleteFromKube()
		}
	}
	wg.Wait()
}

func updateResources(env *db.EnvironmentS) {
	requestedCPU := 0
	requestedRAM := 0
	cli := kube.GetKube()
	for _, v := range cli.PodListAll(env.ID) {
		for _, pod := range v.Obj.Spec.Containers {
			requestedCPU += int(pod.Resources.Requests.Cpu().Value())
			requestedRAM += int(pod.Resources.Requests.Memory().Value())
		}
	}

	env.Resources.CPUReservedCores = requestedCPU
	env.Resources.RAMReservedMB = requestedRAM
	env.Resources.CPUMaxCores = int(db.System.ClusterInfo.CPU.Capacity)
	env.Resources.RAMMaxMB = int(db.System.ClusterInfo.RAM.Capacity)
}
