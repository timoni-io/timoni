package kubesync

import (
	"core/config"
	"core/db"
	"core/db2"
	"core/kube"
	"lib/tlog"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	envLoopCheckingLock = &sync.Mutex{}
	envLoopChecking     = map[string]bool{}
)

func Loop() {
	defer db.PanicHandler()
	kClient := kube.GetKube()

	for _, env := range db.EnvironmentMap.Values() {

		// Render Variables
		for _, ele := range env.Elements.Keys() {
			env.GetElement(ele).RenderVariables()
		}

		// start scheduler
		if env.Schedule.Configured() {
			env.StartSchedule(nil)
		}
	}

	go GarbagePodCollectorLoop()

	targetGitTag := db2.TheSettings.ReleaseGitTag()
	if config.GitTag != "???" && config.GitTag != targetGitTag {
		selfUpgrade(targetGitTag)
		tlog.Fatal("Timoni version change {{old}} => {{new}}", tlog.Vars{
			"old": config.GitTag,
			"new": targetGitTag,
		})
	}

	// ------------------------------------------------------------------

	for {
		go kClient.PodCacheUpdate()
		kube.LastSyncTime = time.Now().UTC()
		kClient.UpdateNodesInfo()
		cpu, ram, _ := kube.GetKube().GetClusterUsage(false)
		db.System.ClusterInfo.Resources.CPURequested = cpu
		db.System.ClusterInfo.Resources.RAMRequested = ram
		db.System.ClusterInfo.Resources.CPUCapacity = kube.NodesInfo.TotalCpus
		db.System.ClusterInfo.Resources.RAMCapacity = kube.NodesInfo.TotalMem

		envLoopCheckingLock.Lock()
		for _, env := range db.EnvironmentMap.Values() {

			if !envLoopChecking[env.ID] {
				envLoopChecking[env.ID] = true
				go envCheck(env)
			}

		}
		envLoopCheckingLock.Unlock()

		time.Sleep(4 * time.Second)
	}
}

func GarbagePodCollectorLoop() {

	timeout := 30 * time.Minute

	for {
		kCtl := kube.GetKube()
		if kCtl == nil {
			return
		}

		pods, err := kCtl.API.CoreV1().Pods("").List(kCtl.CTX, metav1.ListOptions{})
		if tlog.Error(err) != nil {
			time.Sleep(time.Minute)
			continue
		}

		for _, pod := range pods.Items {
			if pod.DeletionTimestamp == nil {
				continue
			}

			if time.Since(pod.DeletionTimestamp.Time) < timeout {
				continue
			}

			// Force delete terminating pod
			tlog.Warning("Force deleting {{pod}} from {{app}}", tlog.Vars{
				"pod": pod.Name,
				"env": pod.Namespace,
			})
			err := kCtl.API.CoreV1().Pods(pod.Namespace).Delete(kCtl.CTX, pod.Name, *metav1.NewDeleteOptions(0))
			tlog.Error(err)
		}

		time.Sleep(timeout)
	}

}
