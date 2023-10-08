package db

import (
	"core/kube"
	"fmt"
	"lib/tlog"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type elementElasticsearchS struct {
	elementS

	ExternalIP    bool                                   `toml:"external-ip"`
	XpackSecurity bool                                   `toml:"xpack-security"`
	Backup        map[string]elementElasticsearchBackupS `toml:"backup"`

	// https://www.elastic.co/guide/en/cloud-on-k8s/current/k8s-node-configuration.html
	Version  string                         `toml:"version"`
	NodeSets []elementElasticSearchNodeSetS `toml:"nodeSets"`

	// Limits
	CPUReservedPC uint `toml:"cpu"`     // in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore
	CPUMax        uint `toml:"-"`       // in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore
	RAMReservedMB uint `toml:"ram"`     // in MB
	RAMMax        uint `toml:"-"`       // in MB
	Storage       int  `toml:"storage"` // in MB
}

type elementElasticSearchNodeSetS struct {
	Name  string `toml:"name"`
	Count int    `toml:"count"`
	// Config      map[string]interface{} `toml:"config"` // XXX: z tym jest problem podczas zapisu do bazy json: unsupported type: map[interface {}]interface {}
	PodTemplate struct {
		Spec v1.PodSpec `toml:"spec"`
	} `toml:"podTemplate"`
}

type elementElasticsearchBackupS struct {
	Type     string `toml:"type"` // nfs, azure
	ReadOnly bool   `toml:"read-only"`

	// nfs
	RemoteServer string `toml:"remote-server"`
	RemotePath   string `toml:"remote-path"`

	// azure
	AzureStorageAccountName      string `toml:"azure-storage-account-name"`
	AzureStorageAccountKey       string `toml:"azure-storage-account-key"`
	AzureStorageAccountContainer string `toml:"azure-storage-account-container"`
}

func (element *elementElasticsearchS) RebuildImage(imageID string, user *UserS) *tlog.RecordS {
	return nil
}

func (element *elementElasticsearchS) GetImage() *ImageS {
	return nil
}

func (element *elementElasticsearchS) RestartAllPods(user *UserS) *tlog.RecordS {
	return nil
}

func (element *elementElasticsearchS) check(user *UserS) *tlog.RecordS {
	if element.ToDelete {
		return nil
	}
	return element.elementS.check(user)
}

func (element *elementElasticsearchS) Save(user *UserS) *tlog.RecordS {
	if err := element.check(user); err != nil {
		return err
	}
	return elementSave(element, user)
}

func (element *elementElasticsearchS) Merge(el EnvElementS) *tlog.RecordS {
	e, ok := el.(*elementElasticsearchS)
	if !ok {
		return tlog.Error(fmt.Sprintln("wrong type, expected", element.GetType(), "got", el.GetType()))
	}
	return element.elementS.Merge(&e.elementS)
}

func (element *elementElasticsearchS) CopySecrets(el EnvElementS) *tlog.RecordS {
	e, ok := el.(*elementElasticsearchS)
	if !ok {
		return tlog.Error(fmt.Sprintln("wrong type, expected", element.GetType(), "got", el.GetType()))
	}
	return element.elementS.CopySecrets(&e.elementS)
}

func (element *elementElasticsearchS) DeleteFromKube() *tlog.RecordS {
	tlog.Info("DeleteFromKube")

	if element.GetStatus().State == ElementStatusStopped {
		return nil
	}

	if element.EnvironmentID == "" {
		return tlog.Error("element.EnvironmentID is empty")
	}

	kClient := kube.GetKube()
	if kClient == nil {
		return tlog.Error("kube.Client not ready")
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "elasticsearch.k8s.elastic.co",
		Version: "v1",
		Kind:    "Elasticsearch",
	})
	obj.SetNamespace(element.EnvironmentID)
	obj.SetName(element.Name)

	kClient.CRD.Get(kClient.CTX, client.ObjectKeyFromObject(obj), obj)

	if obj.GetGeneration() == 0 {
		element.SetState(ElementStatusStopped)
	}

	return nil
}

func (element *elementElasticsearchS) GetScale() *ElementScaleS {
	return &ElementScaleS{}
}