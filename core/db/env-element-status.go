package db

import "core/kube"

type ElementStatusS struct {
	State        ElementState
	Alerts       []string
	NewerVersion bool
	Next         *ElementNextS

	PodCount int
	pods     map[string]*ElementKubePodS // key=podName

	restart bool
}

type ElementNextS struct {
	SourceGit   SourceGitS `toml:"-"`
	StepCurrent int
	StepCount   int
	Message     string
	State       ElementState
}

type ElementKubePodS struct {
	PodName      string
	ElementName  string
	Type         string // pod, action, debug
	NodeName     string
	Status       kube.PodStatusS
	CreationTime int64
	ReadyTime    int64
	RestartCount int32 // since creation
	Debug        bool  // true - means, its a special ExtraDebugPod
	Alerts       []string

	CPUUsagePC   uint // % of CPU, 100% = 1 cpu core
	CPUUsageProc uint
	RAMUsedMB    uint
	RAMUsedProc  uint

	restart bool
}

func (es *ElementStatusS) Save() {
	if es.Next != nil {
		es.Next.State = es.State
	}
}

func (es *ElementStatusS) PodsGet() map[string]*ElementKubePodS {
	return es.pods
}

func (pod *ElementKubePodS) Restart() {
	pod.Status = kube.PodStatusPending
	pod.RestartCount++
	pod.restart = true
}
