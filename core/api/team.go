package api

import (
	"core/db"
	perms "core/db/permissions"
	"encoding/json"
	"lib/tlog"
	"net/http"
	"strings"
)

type frontTeam struct {
	ID               string
	Name             string
	NrOfMembers      int
	NrOfEnvironments int
	NrOfGitRepos     int

	Perm    perms.FrontPerm
	Members map[string]publicUserS
}

func apiTeamCreate(r *http.Request, user *db.UserS) interface{} {
	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	var t struct {
		Name string
	}

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return tlog.Error("Invalid JSON")
	}

	if t.Name == "" {
		return tlog.Error("Team name cannot be empty")
	}

	if db.GetTeamByName(t.Name) != nil {
		return tlog.Error("Team name already exists")
	}

	db.NewTeam(t.Name).Save()
	tlog.Info("Team created: {{team}}", tlog.Vars{
		"team":  t.Name,
		"user":  user.Name,
		"event": true,
	})

	return "ok"
}

func apiTeamList(r *http.Request, user *db.UserS) interface{} {
	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	teams := []frontTeam{}

	for _, t := range db.TeamMap.Values() {

		teams = append(teams, frontTeam{
			ID:               t.ID,
			Name:             t.Name,
			NrOfMembers:      t.Members.Len(),
			NrOfEnvironments: len(t.Permissions.Envs),
			NrOfGitRepos:     len(t.Permissions.GitRepos),
		})
	}

	return teams
}

func apiTeamInfo(r *http.Request, user *db.UserS) interface{} {

	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	teamID := r.FormValue("teamID")
	if teamID == "" {
		return tlog.Error("`teamID` is required")
	}

	team := db.TeamMap.Get(teamID)
	if team == nil {
		return tlog.Error("`teamID` not found")
	}

	members := map[string]publicUserS{}
	for k, v := range team.UsersToMap() {
		members[k] = publicUserS{
			ID:                  v.ID,
			Email:               v.Email,
			Name:                v.Name,
			Theme:               v.Theme,
			NotificationsSend:   v.NotificationsSend,
			AutoLogout:          v.Logout,
			CreatedTimeStamp:    v.CreatedTimeStamp,
			LastActionTimeStamp: v.LastActionTimeStamp,
		}
	}

	return frontTeam{
		ID:               team.ID,
		Name:             team.Name,
		NrOfMembers:      team.Members.Len(),
		NrOfEnvironments: len(team.Permissions.Envs),
		NrOfGitRepos:     len(team.Permissions.GitRepos),
		Members:          members,
		Perm:             team.Permissions.ToFrontPerm(),
	}
}

func apiTeamUpdate(r *http.Request, user *db.UserS) interface{} {
	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	var t struct {
		ID   string
		Name string
	}

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return tlog.Error("Invalid JSON")
	}

	team, ok := db.TeamMap.GetFull(t.ID)
	if !ok {
		return tlog.Error("Team not found")
	}

	if t.Name == "" {
		return tlog.Error("Team name cannot be empty")
	}

	if team.Name == db.AdminTeamName {
		return tlog.Error("Cannot rename admin team")
	}

	if db.GetTeamByName(t.Name) != nil {
		return tlog.Error("Team name already exists")
	}

	team.Name = t.Name
	team.Save()

	tlog.Info("Team created: %s", t.Name, tlog.Vars{
		"team":  t.Name,
		"user":  user.Name,
		"event": true,
	})

	return "ok"
}

func apiTeamDelete(r *http.Request, user *db.UserS) interface{} {
	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	var t struct {
		ID string
	}

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return tlog.Error("Invalid JSON")
	}

	team, ok := db.TeamMap.GetFull(t.ID)
	if !ok {
		return tlog.Error("Team not found")
	}

	if team.Name == db.AdminTeamName {
		return tlog.Error("Cannot delete admin team")
	}

	if team.Members.Len() > 0 {
		return tlog.Error("Cannot delete team with members. Remove members first.")
	}

	team.Delete()
	tlog.Info("Team deleted: %s", team.Name, tlog.Vars{
		"team":  team.Name,
		"user":  user.Name,
		"event": true,
	})

	return "ok"
}

func apiTeamRemoveUser(r *http.Request, user *db.UserS) interface{} {
	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	var t struct {
		TeamID string
		UserID string
	}

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return tlog.Error("Invalid JSON")
	}

	team, ok := db.TeamMap.GetFull(t.TeamID)
	if !ok {
		return tlog.Error("Team not found")
	}

	u := db.GetUserByID(t.UserID)
	if u == nil {
		return tlog.Error("User not found")
	}

	team.Members.Delete(t.UserID)
	u.Teams.Delete(t.TeamID)
	u.Save()
	team.Save()

	return "ok"
}

func apiTeamAddUser(r *http.Request, user *db.UserS) interface{} {
	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	var t struct {
		TeamID string
		UserID string
	}

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return tlog.Error("Invalid JSON")
	}

	team, ok := db.TeamMap.GetFull(t.TeamID)
	if !ok {
		return tlog.Error("Team not found")
	}

	u := db.GetUserByID(t.UserID)
	if u == nil {
		return tlog.Error("User not found")
	}

	team.Members.Add(t.UserID)
	u.Teams.Add(t.TeamID)
	u.Save()
	team.Save()

	return "ok"
}

func apiTeamSetPerms(r *http.Request, user *db.UserS) interface{} {
	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	var t struct {
		TeamID  string
		Global  map[uint8]bool
		Env     map[string]map[uint8]bool
		GitRepo map[string]map[uint8]bool
	}

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return tlog.Error("Invalid JSON")
	}

	team, ok := db.TeamMap.GetFull(t.TeamID)
	if !ok || team == nil {
		return tlog.Error("Team not found")
	}

	if team.Name == db.AdminTeamName {
		return tlog.Error("Cannot change permissions of admin team")
	}

	newPermGroup := perms.PermGroup{
		Global:   perms.FromMap(t.Global),
		Envs:     map[string]perms.Mask{},
		GitRepos: map[string]perms.Mask{},
	}
	for k, v := range t.Env {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		newPermGroup.Envs[k] = perms.FromMap(v)
	}
	for k, v := range t.GitRepo {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		newPermGroup.GitRepos[k] = perms.FromMap(v)
	}

	team.Permissions = newPermGroup
	team.Save()

	tlog.Info("tem permissions updated", tlog.Vars{
		"event": true,
		"user":  user.Email,
	})

	return "ok"
}
