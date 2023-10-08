package db

import (
	"core/config"
	"core/db2"
	"fmt"
	"time"

	log "lib/tlog"
)

type SystemS struct {
	Nodes              map[string]*NodeS
	LoadBalancers      map[string]bool
	GuardianChecks     map[int]*GuardianCheckS
	NodePorts          map[string]int32 // key= appID/elementName/portNr
	NotificationsQueue map[string]time.Time
	ClusterInfo        ClusterInfoS
	DiskUsage          map[string]int64
	NodesIPs           []string
}

type NodeS struct {
	ID         string
	Name       string
	IP         string
	Ready      bool
	KubeMaster bool
	Resources  NodeResourcesS
}

type ClusterInfoS struct {
	NodeResourcesS
	Nodes struct {
		Total int64
		Ready int64
	}
	Resources struct {
		CPURequested int64
		CPUCapacity  int64
		RAMRequested int64
		RAMCapacity  int64
	}
}

type ResourceS struct {
	Usage      int64
	Guaranteed int64
	Max        int64
	Capacity   int64
}

type NodeResourcesS struct {
	CPU  ResourceS
	RAM  ResourceS
	Pods struct {
		Capacity int64
		Total    int64
		Ready    int64
	}
	Environments struct {
		Total int64
		Ready int64
	}
}

// GuardianCheckS ...
type GuardianCheckS struct {
	Message            string
	Success            bool
	Traceback          []log.TraceS
	LastCheckTimeStamp int64
	BeginTimeStamp     int64
}

var System = &SystemS{}

func (system *SystemS) Version() map[string]interface{} {
	data := db2.TheSettings

	return map[string]interface{}{
		"name":          data.Name(),
		"clusterDomain": data.WebUIDomain().Name(),
		"gitTag":        data.ReleaseGitTag(),
		"commitSHA":     config.CommitSHA,
		"kernel":        "",
		"k3os":          "",
		"k3s":           "",
	}
}

func (system *SystemS) GetNodePort(envID string, elementName string, targetPort int32) int32 {

	key := fmt.Sprintf("%s/%s/%d", envID, elementName, targetPort)

	nodePortNr, exist := system.NodePorts[key]
	if exist {
		return nodePortNr
	}

	if len(system.NodePorts) == 0 {
		system.NodePorts = map[string]int32{
			key: 30001,
		}
		return 30001
	}

	highest := int32(0)
	for _, npNr := range system.NodePorts {
		if npNr > highest {
			highest = npNr
		}
	}
	highest++

	system.NodePorts[key] = highest
	return highest
}
