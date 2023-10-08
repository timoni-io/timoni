package api

import (
	"core/db"
	log "lib/tlog"
	"net/http"
	"time"
)

func apiActionStatus(w http.ResponseWriter, r *http.Request) {
	envID := r.FormValue("envID")
	if envID == "" {
		w.Write([]byte("envID is required"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	env := db.EnvironmentMap.Get(envID)
	if env == nil {
		w.Write([]byte("envID is invalid"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	elementName := r.FormValue("elementName")
	if elementName == "" {
		w.Write([]byte("elementName is required"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	element := env.GetElement(elementName)
	if element == nil {
		w.Write([]byte("elementName is invalid"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if element.GetType() != db.ElementSourceTypeAction {
		w.Write([]byte("element is not an action"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	elementAction, ok := db.GetElementAction(element)
	if !ok {
		w.Write([]byte("element is not an action"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	actionToken := r.FormValue("actionToken")
	if actionToken == "" {
		w.Write([]byte("actionToken is required"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if elementAction.ActionToken != actionToken {
		w.Write([]byte("actionToken is invalid"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status := r.FormValue("status")
	if status == "" {
		w.Write([]byte("status is required"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}


	switch s := db.ElementActionStatusT(status); s {
	case db.ElementActionStatusRunning:
		elementAction.Status = s
		element.Save(nil)
	case db.ElementActionStatusSucceeded, db.ElementActionStatusFailed:
		elementAction.Status = s
		if elementAction.TimeEnd == 0 {
			elementAction.TimeEnd = time.Now().UnixMilli()
		}
		element.Save(nil)

		
	default:
		w.Write([]byte("status is invalid"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func apiEnvironmentElementActionsRun(r *http.Request, user *db.UserS) interface{} {
	environmentID := r.FormValue("env")
	if environmentID == "" {
		return log.Error("`env` is required")
	}

	env := db.EnvironmentMap.Get(environmentID)
	if env == nil {
		return log.Error("environment not found", log.Vars{
			"env": environmentID,
		})
	}

	elementName := r.FormValue("element")
	if elementName == "" {
		return log.Error("`element` is required")
	}

	element := env.GetElement(elementName)
	if element == nil {
		return log.Error("element not found", log.Vars{
			"env":  env.ID,
			"name": elementName,
			"user": user.Email,
		})
	}

	if element.GetType() != db.ElementSourceTypePod {
		return log.Error("element is not container type", log.Vars{
			"env":  env.ID,
			"name": elementName,
			"user": user.Email,
		})
	}

	elementPod, ok := db.GetElementPod(element)
	if !ok {
		return log.Error("element is not container type", log.Vars{
			"env":  env.ID,
			"name": elementName,
			"user": user.Email,
		})
	}

	action := r.FormValue("action")
	if action == "" {
		return log.Error("`action` is required")
	}

	if err := elementPod.RunAction(action, user); err != nil {
		return log.Error("running action failed", log.Vars{
			"env":    env.ID,
			"name":   elementName,
			"user":   user.Email,
			"action": action,
			"error":  err,
		})
	}
	log.Info("running action "+action, log.Vars{
		"env":   env.ID,
		"event": true,
		"user":  user.Email,
	})
	return nil
}
