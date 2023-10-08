package db2

import (
	"core/config"
	"core/db2/fp"
	"fmt"
	"os"
	"time"
)

var (
	TheOrganization Organization
	TheSettings     Settings
	TheDomain       Domain
	TheKube         Kube
	TheGitProvider  GitProvider
	TheImageBuilder ImageBuilder
	TheIngress      Ingress
	TheMetrics      Metrics
)

func Open() {
	// otwieram lokalna baze danych
	// ktora nawiazuje polaczenie z FP, aby pobierac z niego ustawienia
	// FP powinien w petli modyfikowac lokalna baze danych i zaden inny modul (poza lokalna db) nie powinien uzywac FP

	if !SyncWithFP() {
		fmt.Println("ERROR: conection to FP failed")
		os.Exit(1)
	}

	go Loop()
}

func Loop() {
	for {
		time.Sleep(3 * time.Minute)
		SyncWithFP()
	}
}

func SyncWithFP() bool {

	FPInstallation := fp.InstallationGetByID(config.InstallationID())
	if FPInstallation.NotValid() {
		lastSyncMinutes := int64(0)
		if TheSettings != nil {
			lastSyncMinutes = int64(time.Since(time.Unix(TheSettings.LastSync(), 0)).Minutes())
		}
		fmt.Println(time.Now(), "syncWithFP failed ", lastSyncMinutes)
		if lastSyncMinutes > 50 {
			fmt.Println("ERROR: conection to FP failed")
			os.Exit(1)
		}
		return false
	}

	if !FPInstallation.Active() {
		fmt.Println("Installation is not active")
		os.Exit(1)
		return false
	}

	// FPInstallation.SetLastIP() TODO
	FPInstallation.SetLastSync(time.Now().Unix())

	FPOrganization := FPInstallation.Organization()
	whereOrgID := "Organization = '" + FPOrganization.ID() + "' OR Organization = ''"

	// --------------------

	dbLock.Lock()
	var exists bool
	dbConnection.Model(&organizationS{}).Select("count(*) > 0").Where("id = ?", FPOrganization.ID()).Find(&exists)
	if !exists {
		obj := &organizationS{
			IDC: FPOrganization.ID(),
		}
		dbConnection.Create(obj)
	}
	dbLock.Unlock()
	TheOrganization = OrganizationGetByID(FPOrganization.ID())

	// --------------------

	for _, fpItem := range fp.CertList(whereOrgID, "", 0, 1000).Iter() {
		certCreateOrUpdate(fpItem.InfoJSON())
		// fmt.Println(fpItem.DomainName())
	}

	for _, fpItem := range fp.CertProviderList(whereOrgID, "", 0, 1000).Iter() {
		certProviderCreateOrUpdate(fpItem.InfoJSON())
	}

	for _, fpItem := range fp.DNSProviderList(whereOrgID, "", 0, 1000).Iter() {
		dNSProviderCreateOrUpdate(fpItem.InfoJSON())
	}

	for _, fpItem := range fp.DomainList(whereOrgID, "", 0, 1000).Iter() {
		domainCreateOrUpdate(fpItem.InfoJSON())
	}

	for _, fpItem := range fp.ImageBuilderList(whereOrgID, "", 0, 1000).Iter() {
		imageBuilderCreateOrUpdate(fpItem.InfoJSON())
	}

	for _, fpItem := range fp.KubeList(whereOrgID, "", 0, 1000).Iter() {
		kubeCreateOrUpdate(fpItem.InfoJSON())
	}

	for _, fpItem := range fp.GitProviderList(whereOrgID, "", 0, 1000).Iter() {
		gitProviderCreateOrUpdate(fpItem.InfoJSON())
	}

	for _, fpItem := range fp.IngressList(whereOrgID, "", 0, 1000).Iter() {
		ingressCreateOrUpdate(fpItem.InfoJSON())
	}

	for _, fpItem := range fp.MetricsList(whereOrgID, "", 0, 1000).Iter() {
		metricsCreateOrUpdate(fpItem.InfoJSON())
	}

	for _, fpItem := range fp.NotificationProviderList(whereOrgID, "", 0, 1000).Iter() {
		notificationProviderCreateOrUpdate(fpItem.InfoJSON())
	}

	// --------------------

	settingsCreateOrUpdate(FPInstallation.InfoJSON())
	TheSettings = SettingsList("", "", 0, 1).First()
	if TheSettings.NotValid() {
		panic("cos nie tak")
	}

	// --------------------

	TheDomain = TheSettings.WebUIDomain()
	TheKube = TheSettings.Kube()
	TheGitProvider = TheSettings.GitProvider()
	TheImageBuilder = TheSettings.ImageBuilder()
	TheIngress = TheSettings.Ingress()
	TheMetrics = TheSettings.Metrics()

	return true
}
