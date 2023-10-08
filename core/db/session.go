package db

import (
	"core/config"
	"encoding/json"
	"lib/tlog"
	"lib/utils/maps"
	"lib/utils/net"
	"lib/utils/random"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type SessionObjectS struct {
	ID           string
	CreationTime time.Time
	UpdateTime   time.Time
	UserID       string
	Used         bool
	Expires      time.Time
	UserAgent    string
	IP           string
}

const sessionsFileName = "user-sessions.json"

var SessionMap = SessionsAllLoad()

func SessionCreate(userID string, req *http.Request) SessionObjectS {
	id := random.String(32)
	sess := SessionObjectS{
		ID:           id,
		UserID:       userID,
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
		Expires:      time.Now().Add(4 * 24 * time.Hour),
		IP:           net.RequestIP(req),
		UserAgent:    req.UserAgent(),
	}
	SessionMap.Set(id, sess)
	SessionsAllSave()
	return sess
}

func (sess *SessionObjectS) Update() {
	sess.UpdateTime = time.Now()
	SessionMap.Set(sess.ID, *sess)
	SessionsAllSave()
}

func (sess *SessionObjectS) Expired() bool {
	return sess.Expires.Before(time.Now())
}

func SessionsAllSave() {
	buf, err := json.MarshalIndent(SessionMap, "", "  ")
	tlog.Error(err)

	tlog.Error(os.WriteFile(filepath.Join(config.DataPath(), sessionsFileName), buf, 0644))
}

func SessionsAllLoad() *maps.SafeMap[string, SessionObjectS] {
	sess := maps.New[string, SessionObjectS](nil).Safe()
	buf, _ := os.ReadFile(filepath.Join(config.DataPath(), sessionsFileName))
	json.Unmarshal(buf, sess)

	return sess
}

func SessionsAllDelete() {
	SessionMap = maps.New[string, SessionObjectS](nil).Safe()
	SessionsAllSave()
}
