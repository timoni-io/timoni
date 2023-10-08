package imagebuilder

import (
	"core/db"
	"time"
)

func Loop() {
	defer db.PanicHandler()

	// on restart, rebuild all images that were in process of being built

	for _, image := range db.ImageGetList() {
		if image.BuildStatus == "building" {
			image.BuildStatus = "pending"
			image.Save()
		}
	}

	ImageBuilderGetReady()

	for {

		imageToBuild := db.ImageGetNextToBuild()
		if imageToBuild == nil {
			time.Sleep(3 * time.Second)
			continue
		}

		imageBuilder := ImageBuilderGetReady()
		if imageBuilder == nil {
			time.Sleep(3 * time.Second)
			continue
		}

		if imageBuilder.startBuild(imageToBuild) {
			continue
		}

		imageToBuild.BuildStatus = "fail"
		imageToBuild.TimeEnd = time.Now()
		imageToBuild.Save()
	}
}