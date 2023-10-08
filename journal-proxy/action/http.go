package action

import (
	"journal-proxy/wsb"
	"lib/tlog"
	"net/http"
)

func DeleteEnvHandler(w http.ResponseWriter, r *http.Request) {
	conn := wsb.ConnPool.GetNoWait()
	if conn == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	defer wsb.ConnPool.Add(conn)

	envID := r.FormValue("envID")

	tlog.Info("DeleteEnvHandler called", tlog.Vars{
		"envID": envID,
	})

	if envID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conn.DropTables(envID)

	w.WriteHeader(http.StatusNoContent)
}
