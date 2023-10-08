package server

import (
	"entry-point/global"
	"net/http"
)

func ServeStaticFiles() {
	if global.StaticFilesPath == "" {
		return
	}

	http.Handle("/", http.FileServer(http.Dir(global.StaticFilesPath)))
	http.ListenAndServe(":80", nil)
}
