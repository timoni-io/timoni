package kube

import (
	"core/config"
	"core/db2"
	"core/modulestate"
	"lib/tlog"
	"path/filepath"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Setup() {
	modulestate.StatusByModulesAdd("api", Check)

	for {
		state, msg := Check()
		if state == db2.State_ready {
			break
		}
		tlog.Warning(msg)
		time.Sleep(5 * time.Second)
	}

	// longhorn
	if db2.TheSettings.Longhorn() {
		if err := longhornInstall(); err != nil {
			tlog.Fatal(err)
		}
	} else {
		setDefaultStorageClass("local-path")
	}

	// kube-vip
	if db2.TheSettings.VirtualIP() != "" {
		err := InstallVip(db2.TheSettings.VirtualInterface(),db2.TheSettings.VirtualIP())
		if err != nil {
			tlog.Fatal(err)
		}
	}
}

var LastSyncTime time.Time

func Check() (db2.StateT, string) {

	if GetKube() == nil {
		return db2.State_error, "unable to connect to kube cluster"
	}

	if len(GetKube().NamespaceList()) == 0 {
		return db2.State_error, "unable to list namespaces in kube cluster"
	}

	return db2.State_ready, ""
}

func longhornInstall() error {
	tlog.Info("Installing longhorn")

	// INFO: Longhorn requirement
	// iscsiadm/open-iscsi must be installed on host

	kclient := GetKube()
	kclient.ApplyYamlFilesInDir(filepath.Join(config.ModulesPath(), "kube", "longhorn"), nil)

	scClient := kclient.API.StorageV1().StorageClasses()

	// Wait for storage class to be created
	for {
		_, err := scClient.Get(kclient.CTX, "longhorn", v1.GetOptions{})
		if err == nil {
			break
		}

		tlog.Info("Waiting for longhorn storage class to be created")
		time.Sleep(20 * time.Second)
	}

	return setDefaultStorageClass("longhorn")
}

func setDefaultStorageClass(name string) error {

	kclient := GetKube()
	scClient := kclient.API.StorageV1().StorageClasses()

	// get old default storage class
	scs, err := scClient.List(kclient.CTX, v1.ListOptions{})
	if err != nil {
		return err
	}
	for _, sc := range scs.Items {
		if sc.Name == name {
			continue
		}

		if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			sc.Annotations["storageclass.kubernetes.io/is-default-class"] = "false"
			scClient.Update(kclient.CTX, &sc, v1.UpdateOptions{})
		}
	}

	// set class default
	tlog.Info("Setting default storage class: " + name)
	lhsc, _ := scClient.Get(kclient.CTX, name, v1.GetOptions{})
	lhsc.Annotations["storageclass.kubernetes.io/is-default-class"] = "true"
	_, err = scClient.Update(kclient.CTX, lhsc, v1.UpdateOptions{})
	return err
}

// func CheckKubeSync() (db2.StateT, string) {

// 	if GetKube() == nil {
// 		return db2.State_error, "unable to connect to kube cluster"
// 	}

// 	checkTime := time.Since(LastSyncTime)
// 	if checkTime > 2*time.Minute {
// 		return db2.State_error, fmt.Sprintf("Last sync %s ago", checkTime)
// 	}

// 	return db2.State_ready, fmt.Sprintf("Last sync %s ago", checkTime)
// }
