package db

import (
	perms "core/db/permissions"
	"encoding/json"
	"lib/tlog"
	"lib/utils/maps"
	"lib/utils/set"
	"strings"

	"github.com/google/uuid"
)

type Team struct {
	ID          string
	Name        string
	Members     *set.Safe[string]
	Permissions perms.PermGroup
}

var TeamMap = *maps.NewSafe[string, *Team](nil)

func NewTeam(name string) *Team {
	team := &Team{
		ID:      strings.ReplaceAll(uuid.New().String(), "-", ""),
		Name:    name,
		Members: set.NewSafe[string](nil),
		Permissions: perms.PermGroup{
			Envs:     map[string]perms.Mask{},
			GitRepos: map[string]perms.Mask{},
		},
	}
	TeamMap.Set(team.ID, team)
	team.Save()
	return team
}

func GetTeamByName(name string) *Team {
	for _, t := range TeamMap.Values() {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func (t *Team) AddUser(user *UserS) {
	if user == nil {
		tlog.Error("user is nil")
		return
	}
	t.Members.Add(user.ID)
	t.Save()
	user.Teams.Add(t.ID)
	user.Save()
}

func (t *Team) RemoveUser(user *UserS) {
	t.Members.Delete(user.ID)
	t.Save()
	user.Teams.Delete(t.ID)
	user.Save()
}

func (t *Team) Save() {
	err := driver.Write("team", t.ID, t)
	if err != nil {
		tlog.Error(err)
	}
}

func (t *Team) Delete() {

	for u := range UserMapID.Iter() {
		user := u.Value
		if user.Teams.Exists(t.ID) {
			user.Teams.Delete(t.ID)
			user.Save()
		}
	}

	err := driver.Delete("team", t.ID)
	if tlog.Error(err) != nil {
		return
	}

	TeamMap.Delete(t.ID)
}

func LoadTeams() {
	TeamMap = *maps.NewSafe[string, *Team](nil)
	teamsB, err := driver.ReadAll("team")
	if err != nil {
		tlog.Error(err)
		return
	}
	for _, t := range teamsB {
		team := &Team{}
		err = json.Unmarshal(t, team)
		if err != nil {
			tlog.Error(err)
			continue
		}
		TeamMap.Set(team.ID, team)
	}
}

func (t *Team) UsersToMap() map[string]UserS {
	users := map[string]UserS{}

	for _, u := range t.Members.List() {
		if user := GetUserByID(u); user != nil {
			users[user.ID] = *user
		}
	}

	return users
}
