package db

import (
	"core/kube"
	"fmt"
	"lib/tlog"
)

type elementMongodbS struct {
	elementS

	BackupAzureStorageAccountName      string `toml:"backup-azure-storage-account-name"`
	BackupAzureStorageAccountKey       string `toml:"backup-azure-storage-account-key"`
	BackupAzureStorageAccountContainer string `toml:"backup-azure-storage-account-container"`

	Version       string                         `toml:"version"`
	MembersCount  int                            `toml:"members-count"`
	StorageSize   int                            `toml:"storage-size-gb"`
	CPUlimit      int                            `toml:"cpu-limit"`
	RAMlimit      int                            `toml:"ram-limit-gb"`
	RAMReservedMB int                            `toml:"ram"`
	CPUReservedPC int                            `toml:"cpu"`
	Users         map[string]elementMongodbUserS `toml:"user"` // key=userLogin

	// Limits
	// CPUGuaranteed uint `toml:"cpu"`     // in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore
	// CPUMax        uint `toml:"-"`       // in % of vCores, eg 100 = 1 vcore, 250 = 2.5 vcore
	// RAMGuaranteed uint `toml:"ram"`     // in MB
	// RAMMax        uint `toml:"-"`       // in MB
	// Storage       int  `toml:"storage"` // in MB
}

type elementMongodbUserS struct {
	Password string   `toml:"password"`
	Database string   `toml:"database"`
	Roles    []string `toml:"roles"`
}

func (element *elementMongodbS) RebuildImage(imageID string, user *UserS) *tlog.RecordS {
	return nil
}

func (element *elementMongodbS) GetImage() *ImageS {
	return nil
}

func (element *elementMongodbS) RestartAllPods(user *UserS) *tlog.RecordS {
	return nil
}

func (element *elementMongodbS) check(user *UserS) *tlog.RecordS {
	if element.ToDelete {
		return nil
	}
	return element.elementS.check(user)
}

func (element *elementMongodbS) Save(user *UserS) *tlog.RecordS {
	if err := element.check(user); err != nil {
		return err
	}
	return elementSave(element, user)
}

func (element *elementMongodbS) Merge(el EnvElementS) *tlog.RecordS {
	e, ok := el.(*elementMongodbS)
	if !ok {
		return tlog.Error(fmt.Sprintln("wrong type, expected", element.GetType(), "got", el.GetType()))
	}
	return element.elementS.Merge(&e.elementS)
}

func (element *elementMongodbS) CopySecrets(el EnvElementS) *tlog.RecordS {
	e, ok := el.(*elementMongodbS)
	if !ok {
		return tlog.Error(fmt.Sprintln("wrong type, expected", element.GetType(), "got", el.GetType()))
	}
	return element.elementS.CopySecrets(&e.elementS)
}

func (element *elementMongodbS) GetScale() *ElementScaleS {
	return &ElementScaleS{}
}

func (element *elementMongodbS) DeleteFromKube() *tlog.RecordS {
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

	return nil
}