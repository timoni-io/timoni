package db

import (
	"core/config"
	"core/db/scribble"
	"encoding/json"
	"lib/tlog"
	"lib/utils/bitmap"
	"lib/utils/maps"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	driver *scribble.Driver

	reSimpleName1 = regexp.MustCompile(`^[a-z0-9\-]+$`)
	reSimpleName2 = regexp.MustCompile(`^[a-z0-9\-\.]+$`)
	reSimpleName3 = regexp.MustCompile(`^[a-z]+$`)

	// JournalReader *journal.ReaderS
	// JournalWriter *journal.WriterS

	DefaultUser = &UserS{Email: "Timoni", Name: "Timoni"}

	loadFromGitCacheMap = *maps.NewSafe[string, []byte](nil)
)

const (
	AdminTeamName       = "Administrators"
	BlacklistedTeamName = "Blacklisted"
)

type errorMessageT int

// go install golang.org/x/tools/cmd/stringer@latest
//
//go:generate stringer -type=errorMessageT -output error_message_string.go
const (
	error_ElementNotFound      errorMessageT = 10
	error_VariableNotFound     errorMessageT = 11
	error_InvalidReference     errorMessageT = 12 // self or cycle reference
	error_EmptyValue           errorMessageT = 20
	error_InvalidValidator     errorMessageT = 30
	error_InvalidValidatorArgs errorMessageT = 31
	error_InvalidName          errorMessageT = 40
	error_ValidationFailed     errorMessageT = 41
)

func Open() {

	for {
		var err error
		driver, err = scribble.New(config.DataPath(), nil)
		if err == nil {
			break
		}
		tlog.Error(err)
		time.Sleep(10 * time.Second)
	}

	// ----------------------------------------------------------

	os.Mkdir(filepath.Join(config.DataPath(), "env"), 0755)
	os.Mkdir(filepath.Join(config.DataPath(), "image"), 0755)
	os.Mkdir(filepath.Join(config.DataPath(), "user"), 0755)
	os.Mkdir(filepath.Join(config.DataPath(), "query-hash"), 0755)
	os.Mkdir(filepath.Join(config.DataPath(), "git-repo"), 0755)
	os.Mkdir(filepath.Join(config.DataPath(), "git-stats"), 0755)
	os.Mkdir(filepath.Join(config.DataPath(), "team"), 0755)

	// ----------------------------------------------------------
	// applyFixtures

	_, err := os.Stat(filepath.Join(config.DataPath(), "git-remote"))
	if os.IsNotExist(err) {
		applyFixtures()
	}

	// ----------------------------------------------------------
	LoadTeams()

	// Create Admin team
	if GetTeamByName(AdminTeamName) == nil {
		t := NewTeam(AdminTeamName)
		// all permissions
		bitmap.SetAll(&t.Permissions.Global)
		t.Permissions.Envs["id:*"] = t.Permissions.Global
		t.Permissions.GitRepos["id:*"] = t.Permissions.Global
		t.Save()
	} else {
		t := GetTeamByName(AdminTeamName)
		// all permissions
		bitmap.SetAll(&t.Permissions.Global)
		t.Permissions.Envs["id:*"] = t.Permissions.Global
		t.Permissions.GitRepos["id:*"] = t.Permissions.Global
		t.Save()
	}
	if GetTeamByName(BlacklistedTeamName) == nil {
		NewTeam(BlacklistedTeamName)
		// t := NewTeam(BlacklistedTeamName)
		// bitmap.SetBit(&t.Permissions.Global, perms.Glob_AccessToWebUI)
		// for user := range UserMapEmail.Iter() {
		// 	if user.Value.Name == "ImageBuilder" {
		// 		continue
		// 	}
		// 	t.AddUser(user.Value)
		// }
		// t.Save()
	}

	// ----------------------------------------------------------

	buf, err := os.ReadFile(config.DataPath() + "/git-cache.json")
	if err == nil {
		tlog.Error(json.Unmarshal(buf, &loadFromGitCacheMap))
	}

	// ----------------------------------------------------------

	go SyncWithDiskLoop()

	// ----------------------------------------------------------

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		Close()
		tlog.Fatal("db.Close()")
	}()
}

// --------------------------------------------------

// Close ...
func Close() {
	// System.Save()
}

func IsOpen() bool {
	return driver != nil
}

// --------------------------------------------------

var src = rand.NewSource(time.Now().UnixNano())

const (
	letterBytes   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // 52 possibilities
	letterIdxBits = 6                                                                // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1                                             // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 70 / letterIdxBits                                               // # of letter indices fitting in 63 bits
)

// RandString ...
func RandString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

// --------------------------------------------------

func applyFixtures() {

	// 	tlog.Debug("db > applyFixtures...")

	// admin := global.Config.Admin.Email
	gitRepo := &GitRepoS{
		Name:        "elements",
		CreatedTime: time.Now(),
		// CreatedByUserID:    user.ID,
		// CreatedByUserEmail: user.Email,
		RemoteURL:     "https://github.com/timoni-io/elements.git",
		DefaultBranch: "main",
		// Operators: map[string]string{
		// 	user.Email: "owner",
		// },
		// ResourcesUsage: db.CodeStats{},
		// ResourcesLimit: db.ProjectDefaultResourcesLimit,
	}
	gitRepo.Save()

	// gitRepo.Open()
	// defer gitRepo.Unlock()
	// gitRepo.Save()

	// gitserver.FindElementsInGitRepo(gitRepo.Name)

	// (&UserS{
	// 	ID:           "",
	// 	Email:        "",
	// 	Name:         "",
	// 	YivoAPIAllow: []string{"*"},
	// 	CreatedTime:  time.Now().UTC(),
	// }).Save()

	// ----------------------------------------------------

	// (&clusterS{
	// 	Name:            "lan",
	// 	AllowUserGroups: []string{"devs"},
	// }).save()

	// (&clusterS{
	// 	Name: "dmz",
	// }).save()

	// (&clusterS{
	// 	Name: "prod",
	// }).save()

	// ----------------------------------------------------

	// (&namespacesS{
	// 	ID: "lan.lw",
	// 	DeployRules: []deployRuleS{
	// 		{
	// 			Type:  "allow",
	// 			Users: []string{""},
	// 		},
	// 	},
	// }).save()

	// ----------------------------------------------------
}
