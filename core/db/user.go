package db

import (
	"bytes"
	"core/db/permissions"
	"encoding/base32"
	"image/png"
	log "lib/tlog"
	"lib/utils/bitmap"
	"lib/utils/maps"
	"lib/utils/set"
	"math/rand"
	"strings"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

type UserS struct {
	ID                  string
	Email               string
	EmailAliases        []string
	Name                string
	TotpIssuer          string
	TotpSecret          string
	Activated           bool
	GitToken            string
	CreatedTimeStamp    int64
	CreatedByUserID     string
	CreatedByUserEmail  string
	CreatedGitRepoLimit int32
	Theme               string
	NotificationsSend   bool
	LastActionTimeStamp int64
	Logout              int8
	Teams               *set.Set[string]

	LastVisited *maps.SafeMap[string, int64] // env-id: timestamp
}

var (
	UserMapEmail = maps.New(map[string]*UserS{}).Safe()
	UserMapID    = maps.New(map[string]*UserS{}).Safe()
)

func generateRandomB32Secret(length uint) []byte {
	secretBytes := make([]byte, length)
	rand.Read(secretBytes)

	base32Secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secretBytes)
	return []byte(base32Secret)
}

func GenerateTOTP(user *UserS) (qr []byte) {
	key, err := totp.Generate(totp.GenerateOpts{
		Secret:      generateRandomB32Secret(12),
		Issuer:      user.TotpIssuer,
		AccountName: user.Email,
	})

	if log.Error(err) != nil {
		return nil
	}

	user.TotpSecret = key.Secret()
	user.Save()

	img, _ := key.Image(200, 200)
	buff := new(bytes.Buffer)
	png.Encode(buff, img)
	return buff.Bytes()
}

func (user *UserS) generateID() {
	if user.ID != "" {
		return
	}

	user.ID = strings.ReplaceAll(
		uuid.New().String()+
			uuid.New().String()+
			uuid.New().String()[:16], // cut to < 90 len, because of backup filename length limit
		"-", "")

	tmp := GetUserByID(user.ID)
	if tmp != nil {
		log.Error("user id already taken: " + user.ID)
		return
	}
}

// Save ...
func (user *UserS) Save() {

	user.generateID()
	if user.Teams == nil {
		user.Teams = set.New[string]()
	}
	if user.LastVisited == nil {
		user.LastVisited = maps.NewSafe[string, int64](nil)
	}
	// everyoneTeam := GetTeamByName(EveryoneTeamName)
	// if everyoneTeam != nil && !everyoneTeam.Members.Exists(user.ID) {
	// 	everyoneTeam.Members.Add(user.ID)
	// 	everyoneTeam.Save()
	// }

	if err := driver.Write("user", user.ID, user); err != nil {
		log.Error(err)
		return
	}
}

func userUpdateMap(returnSystemUsers bool) {

	if driver == nil {
		log.Error("db driver not ready")
		return
	}

	usersIDs, err := driver.List("user")
	if err != nil {
		log.Error(err)
	}

	usersOnDisk := set.New(usersIDs...)

	for _, k := range UserMapID.Keys() {
		if !usersOnDisk.Exists(k) {
			UserMapID.Delete(k)
		}
	}

	for _, userID := range usersIDs {
		user, ok := UserMapID.GetFull(userID)
		if !ok {
			user = new(UserS)
			if err := driver.Read("user", userID, user); err != nil {
				log.Error(err)
			}

			if (user.Email == "ImageBuilder" || user.Email == "Timoni") && !returnSystemUsers {
				continue
			}

			UserMapID.Set(userID, user)
		}

		UserMapEmail.Set(strings.ToLower(user.Email), user)

		for _, alias := range user.EmailAliases {
			UserMapEmail.Set(strings.ToLower(alias), user)
		}
	}

}

func GetUserByEmail(email string) *UserS {
	userUpdateMap(true)
	return UserMapEmail.Get(strings.ToLower(email))
}

func GetUserByID(id string) *UserS {
	userUpdateMap(true)
	return UserMapID.Get(id)
}

func GetUserByGitToken(token string) *UserS {
	for u := range UserMapID.Iter() {
		if u.Value.GitToken == token {
			return u.Value
		}
	}
	return nil
}

func (user *UserS) Delete() {
	err := driver.Delete("user", user.ID)
	if log.Error(err) != nil {
		return
	}
}

func InitialsFromEmail(email string) string {
	name := strings.Split(email, "@")[0]
	if name == "" {
		return "??"
	}

	if len(name) == 1 {
		return strings.ToUpper(name)
	}

	if email == "Timoni" {
		return "Ti"
	}

	if email == "ImageBuilder" {
		return "IB"
	}

	if !strings.Contains(name, ".") {
		return strings.ToUpper(name[:2])
	}

	tmp := strings.Split(name, ".")
	name = tmp[0]
	surname := tmp[1]
	return strings.ToUpper(string(name[0]) + string(surname[0]))
}

func (user *UserS) IsOwner(project *GitRepoS) bool {
	return project.Operators[user.Email] == "owner"
}

func (user *UserS) GetKubeName() string {
	if !strings.Contains(user.Email, "@") {
		return ""
	}

	name := strings.Split(user.Email, "@")[0]
	return strings.NewReplacer(
		".", "-",
		"_", "-",
		"+", "-",
		"%", "-",
	).Replace(name)
}

func (user *UserS) HasGlobPerm(perm permissions.GlobPerm) bool {
	mask := permissions.Mask(0)
	for _, tID := range user.Teams.List() {
		team := TeamMap.Get(tID)
		if team == nil {
			continue
		}
		mask.Join(team.Permissions.Global)
	}
	return bitmap.GetBit(mask, perm)
}

func (user *UserS) HasEnvPerm(envID string, perm permissions.EnvPerm) bool {
	if envID == "" {
		return false
	}

	env, ok := EnvironmentMap.GetFull(envID)
	if !ok {
		return false
	}

	mask := permissions.Mask(0)
	for _, tID := range user.Teams.List() {
		team := TeamMap.Get(tID)
		if team == nil {
			continue
		}
		for _, tag := range env.Tags.List() {
			mask.Join(team.Permissions.Envs[tag])
		}
		// join glob env permissions
		mask.Join(team.Permissions.Envs["*"])
	}

	return bitmap.GetBit(mask, perm)
}

func (user *UserS) HasRepoPerm(repoName string, perm permissions.RepoPerm) bool {
	if repoName == "" {
		return false
	}

	repo := GitRepoGetByName(repoName)
	if repo == nil {
		return false
	}

	mask := permissions.Mask(0)
	for _, tID := range user.Teams.List() {
		team := TeamMap.Get(tID)
		if team == nil {
			continue
		}
		for _, tag := range repo.Tags.List() {
			mask.Join(team.Permissions.GitRepos[tag])
		}
		// join glob repo permissions
		mask.Join(team.Permissions.GitRepos["*"])
	}
	return bitmap.GetBit(mask, perm)
}

// returns map[string]PermToFront
func (u *UserS) GetPerms() permissions.PermGroup {

	res := permissions.PermGroup{}

	if u == nil {
		return res
	}

	for _, tID := range u.Teams.List() {
		t, ok := TeamMap.GetFull(tID)
		if !ok {
			continue
		}
		res.Global |= t.Permissions.Global
	}

	return res
}
