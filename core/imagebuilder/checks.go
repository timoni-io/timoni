package imagebuilder

import (
	"core/db"
	"core/db2"
	"fmt"
	"time"
)

func CheckImageBuilder() (db2.StateT, string) {

	var lastestSync time.Time

	readyCount := 0
	for _, ib := range ImageBuilderMap.Values() {

		if !ib.StatusPodExist {
			return db2.State_deploying, "Pod `" + ib.PodName + "` not ready"
		}

		if !ib.StatusHTTPAlive {
			return db2.State_deploying, "HTTP on Pod `" + ib.PodName + "` not ready"
		}

		if lastestSync.IsZero() || ib.StatusUpdateTime.Before(lastestSync) {
			lastestSync = ib.StatusUpdateTime
		}

		readyCount++
	}

	if readyCount == 0 {

		return db2.State_ready, "Not one Image builder is ready"
	}

	return db2.State_ready, fmt.Sprintf("Latest sync %s ago", time.Since(lastestSync))
}

func CheckImageBuilderQueue() (db2.StateT, string) {

	i := 0
	for _, image := range db.ImageGetList() {
		if image.BuildStatus == "pending" {
			i++
		}
	}
	if i > 10 {
		return db2.State_error, fmt.Sprint(i) + " images in status 'pending'"
	}

	return db2.State_ready, fmt.Sprint(i) + " images in status 'pending'"
}
