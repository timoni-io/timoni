package api

import (
	"core/db"
	"core/db/permissions"
	perms "core/db/permissions"
	"core/db2"
	"core/db2/fp"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"lib/tlog"
	"lib/utils/random"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

type publicUserS struct {
	ID                  string
	Email               string
	Name                string
	Theme               string
	NotificationsSend   bool
	AutoLogout          int8
	CreatedTimeStamp    int64
	Logout              int8
	LastActionTimeStamp int64

	CanCreate         bool
	HideGitRepoLocal  bool
	PermissionsGlobal map[string]permissions.PermExplained
	Teams             []string
}

func apiUserList(r *http.Request, user *db.UserS) interface{} {

	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	usrs := map[string]publicUserS{}

	for u := range db.UserMapID.Iter() {
		usrs[u.Key] = publicUserS{
			ID:                  u.Value.ID,
			Email:               u.Value.Email,
			Name:                u.Value.Name,
			Theme:               u.Value.Theme,
			NotificationsSend:   u.Value.NotificationsSend,
			AutoLogout:          u.Value.Logout,
			CreatedTimeStamp:    u.Value.CreatedTimeStamp,
			LastActionTimeStamp: u.Value.LastActionTimeStamp,
			Teams:               u.Value.Teams.List(),
		}

	}

	return usrs
}

func apiUserLogin(w http.ResponseWriter, r *http.Request) {
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	if email == "" {
		fmt.Fprint(w, `{"error": "email cant be empty"}`)
		return
	}

	re, _ := regexp.Compile(`^(?:[A-Z0-9a-z._%+-]{2,64}@[A-Za-z0-9.-]{2,64}\.[A-Za-z]{2,16})$`)
	if !re.MatchString(email) {
		fmt.Fprint(w, `{"error": "invalid email address"}`)
		return
	}

	if name := strings.Split(email, "@")[0]; name == "default" {
		fmt.Fprint(w, `{"error": "email can not be 'default'"}`)
		return
	}

	token := r.FormValue("token")
	user := db.GetUserByEmail(email)
	isAdmin := false
	if user == nil {
		// registration

		user = &db.UserS{
			Email:               email,
			Name:                email,
			CreatedTimeStamp:    time.Now().UTC().Unix(),
			CreatedGitRepoLimit: 0,
			Theme:               "light",
			GitToken:            random.ID(),
		}
		user.Save() // save generates ID

		org := fp.OrganizationGetByID(db2.TheOrganization.ID())
		if org.NotValid() {
			fmt.Printf("fp.OrganizationGetByID not valid `%s`\n", db2.TheOrganization.ID())
			fmt.Fprint(w, `{"error": "organization is not valid"}`)
			return
		}
		for _, admin := range org.Admins().Iter() {
			if email == admin.Email() {
				user.TotpIssuer = "Timoni"
				user.TotpSecret = admin.TotpSecret()
				db.GetTeamByName(db.AdminTeamName).AddUser(user)
				isAdmin = true
				break
			}
		}

		fmt.Printf("new user `%s` (admin: %t)\n", email, isAdmin)

		if !isAdmin {
			user.TotpIssuer = db2.TheSettings.Name()
			db.GenerateTOTP(user)
		}

		// db.GetTeamByName(db.EveryoneTeamName).AddUser(user)
		user.Save()
	}

	if token == "" {
		msg := fmt.Sprintf("Enter the token from your authenticator app. <br>%s (%s)",
			user.TotpIssuer,
			email,
		)

		if !user.Activated && !isAdmin {
			msg = fmt.Sprintf("Information how to obtain the token has been sent to <b>your email</b> (%s).", email)
		}

		fmt.Fprint(w, `{"error": "empty token", "msg": "`+msg+`"}`)
		return
	}

	if user.Teams.Exists(db.BlacklistedTeamName) {
		fmt.Fprint(w, `{"error": "user is blacklisted"}`)
		return
	}

	if !totp.Validate(token, user.TotpSecret) {
		fmt.Fprint(w, `{"error": "invalid token"}`)
		return
	}
	user.Activated = true
	user.Save()
	sess := db.SessionCreate(user.ID, r)
	fmt.Fprint(w, `{"error": "","session": "`+sess.ID+`"}`)
}

func apiUserInvite(r *http.Request, user *db.UserS) interface{} {

	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	email := strings.ToLower(strings.TrimSpace(r.FormValue("Email")))
	if email == "" {
		return tlog.Error("email cant be empty")
	}

	re, _ := regexp.Compile(`^(?:[A-Z0-9a-z._%+-]{2,64}@[A-Za-z0-9.-]{2,64}\.[A-Za-z]{2,16})$`)
	if !re.MatchString(email) {
		return tlog.Error("invalid email address")
	}

	if name := strings.Split(email, "@")[0]; name == "default" {
		return tlog.Error("email can not be 'default'")
	}

	// teamID := r.FormValue("TeamID")
	// if teamID == "" {
	// 	return tlog.Error("teamID is empty")
	// }
	// team := db.TeamMap.Get(teamID)
	// if team == nil {
	// 	return tlog.Error("team not found")
	// }

	newUser := db.GetUserByEmail(email)
	if newUser == nil {
		newUser = &db.UserS{
			Email:               email,
			Name:                email,
			CreatedTimeStamp:    time.Now().UTC().Unix(),
			CreatedGitRepoLimit: 0,
			Theme:               "light",
			GitToken:            random.ID(),
		}
		newUser.Save() // save generates ID

		newUser.TotpIssuer = db2.TheSettings.Name()
		db.GenerateTOTP(newUser)

		// db.GetTeamByName(db.EveryoneTeamName).AddUser(newUser)
		newUser.Save()
	}

	// if !team.Members.Exists(newUser.ID) {
	// 	team.AddUser(newUser)
	// 	newUser.Save()
	// }

	return "ok"
}

func apiUserInviteQR(r *http.Request, user *db.UserS) interface{} {

	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	email := strings.ToLower(strings.TrimSpace(r.FormValue("Email")))
	if email == "" {
		return tlog.Error("email cant be empty")
	}

	newUser := db.GetUserByEmail(email)
	if newUser == nil {
		return tlog.Error("user not found")
	}

	if newUser.Activated {
		return tlog.Error("user already activated")
	}

	return struct {
		QRcode string
	}{
		QRcode: base64.StdEncoding.EncodeToString(db.GenerateTOTP(newUser)),
	}
}

func apiUserInfo(r *http.Request, user *db.UserS) interface{} {

	if user.Theme == "" {
		user.Theme = "light"
		user.Save()
	}

	publicUser := &publicUserS{
		ID:                user.ID,
		Email:             user.Email,
		Name:              user.Name,
		Theme:             user.Theme,
		NotificationsSend: user.NotificationsSend,
		AutoLogout:        user.Logout,
		HideGitRepoLocal:  db2.TheSettings.GitServerLocalHide(),
		PermissionsGlobal: user.GetPerms().ToFrontPerm().Global,
		Teams:             user.Teams.List(),
	}

	// for _, t := range user.Teams.List() {
	// 	te := db.TeamMap.Get(t)
	// 	if te == nil {
	// 		continue
	// 	}

	// 	publicUser.Teams = append(publicUser.Teams, frontTeam{
	// 		ID:   te.ID,
	// 		Name: te.Name,
	// 		Perm: te.Permissions.ToFrontPerm(),
	// 	})
	// }

	return publicUser
}

func apiUserUpdate(r *http.Request, user *db.UserS) interface{} {

	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	data := struct {
		Email string
		Name  string
		Teams map[string]bool // key=team.ID
	}{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return tlog.Error("Invalid JSON")
	}

	if data.Email == "" {
		return tlog.Error("`Email` is required")
	}

	// get user
	editedUser := db.GetUserByEmail(data.Email)
	if editedUser == nil {
		return tlog.Error("User not found: " + data.Email)
	}

	data.Name = strings.TrimSpace(data.Name)
	if data.Name == "" {
		return tlog.Error("`Name` is required")
	}
	editedUser.Name = data.Name

	// ---
	// removing user from Teams

	for _, teamID := range db.TeamMap.Keys() {
		if data.Teams[teamID] || !editedUser.Teams.Exists(teamID) {
			continue
		}

		team, ok := db.TeamMap.GetFull(teamID)
		if !ok {
			return tlog.Error("Team not found: " + teamID)
		}
		team.Members.Delete(editedUser.ID)
		editedUser.Teams.Delete(teamID)
		team.Save()
	}

	// ---
	// adding user to selected Teams

	for teamID := range data.Teams {
		team, ok := db.TeamMap.GetFull(teamID)
		if !ok {
			return tlog.Error("Team not found: " + teamID)
		}

		team.Members.Add(editedUser.ID)
		editedUser.Teams.Add(teamID)
		team.Save()
	}

	// ---

	editedUser.Save()

	tlog.Info("User updated", tlog.Vars{
		"user":       user.Email,
		"editedUser": editedUser.Email,
		"event":      true,
	})
	return "ok"
}

// func apiUserAliasAdd(r *http.Request, user *db.UserS) interface{} {
// 	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
// 		return tlog.Error("permission denied")
// 	}

// 	email := r.FormValue("email")
// 	if email == "" {
// 		return tlog.Error("`email` is required")
// 	}

// 	// get user
// 	tmpUser := db.GetUserByEmail(email)
// 	if tmpUser == nil {
// 		return tlog.Error("User not found: " + email)
// 	}

// 	newAlias := strings.TrimSpace(strings.ToLower(r.FormValue("alias")))
// 	if newAlias == "" {
// 		return tlog.Error("`alias` is required")
// 	}

// 	// add alias if it's not already there
// 	for _, alias := range tmpUser.EmailAliases {
// 		if alias == newAlias {
// 			return tlog.Error("Alias " + newAlias + " is already added for " + tmpUser.Email)
// 		}
// 	}

// 	tmp := db.GetUserByEmail(newAlias)
// 	if tmp != nil {
// 		return tlog.Error("`alias` already use by other user")
// 	}

// 	tmpUser.EmailAliases = append(tmpUser.EmailAliases, newAlias)
// 	tmpUser.Save()

// 	tlog.Info("User alias added", tlog.Vars{
// 		"user":  tmpUser.Email,
// 		"alias": newAlias,
// 	})

// 	return "ok"
// }

// func apiUserAliasRemove(r *http.Request, user *db.UserS) interface{} {
// 	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
// 		return tlog.Error("permission denied")
// 	}

// 	email := r.FormValue("email")
// 	if email == "" {
// 		return tlog.Error("`email` is required")
// 	}

// 	rmAlias := r.FormValue("alias")
// 	if rmAlias == "" {
// 		return tlog.Error("`alias` is required")
// 	}

// 	//get user
// 	tmpUser := db.GetUserByEmail(email)
// 	if user == nil {
// 		return tlog.Error("User with email `" + tmpUser.Email + "` does not exists.")
// 	}

// 	// remove alias
// 	for i, alias := range tmpUser.EmailAliases {
// 		if alias == rmAlias {
// 			tmpUser.EmailAliases[i] = tmpUser.EmailAliases[len(tmpUser.EmailAliases)-1]
// 			tmpUser.EmailAliases = tmpUser.EmailAliases[:len(tmpUser.EmailAliases)-1]
// 			break
// 		}
// 	}
// 	tmpUser.Save()

// 	tlog.Info("User alias removed", tlog.Vars{
// 		"user":    user.Email,
// 		"tmpUser": tmpUser,
// 		"alias":   rmAlias,
// 	})

// 	return "ok"
// }

func apiUserTheme(r *http.Request, user *db.UserS) interface{} {
	// if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
	// 	return tlog.Error("permission denied")
	// }

	theme := r.FormValue("theme")

	if theme == "" {
		return tlog.Error("`theme` is required")
	}
	if theme != "dark" && theme != "light" {
		return tlog.Error("`theme` unsupported", tlog.Vars{
			"user":  user.Email,
			"theme": theme,
		})
	}

	user.Theme = theme
	user.Save()
	return "ok"
}

func userNotificationOnOff(r *http.Request, user *db.UserS) interface{} {

	user.NotificationsSend = r.FormValue("value") == "true"
	user.Save()
	return "ok"
}

func apiUserAutoLogoutUpdate(r *http.Request, user *db.UserS) interface{} {
	autoLogoutTime := r.FormValue("autoLogin")
	logout, err := strconv.Atoi(autoLogoutTime)
	if err != nil {
		return tlog.Error(err, tlog.Vars{
			"user": user.Email,
			"time": autoLogoutTime,
		})
	}
	user.Logout = int8(logout)
	user.Save()
	return "ok"
}

func apiDocsVOD(r *http.Request, user *db.UserS) interface{} {

	docID := r.FormValue("id")
	docName := r.FormValue("name")

	tlog.Info("VOD", tlog.Vars{
		"docVOD":   true,
		"docsID":   docID,
		"docsName": docName,
		"user":     user.Email,
	})

	return "ok"
}

// not related to users
func apiPermissionList(r *http.Request, user *db.UserS) interface{} {
	if !user.HasGlobPerm(perms.Glob_ManageGlobalMemebers) {
		return tlog.Error("permission denied")
	}

	return perms.DefaultGroup.ToFrontPerm()
}
