package api

import (
	"bytes"
	"core/db"
	"core/db2"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"io"
	"lib/terrors"
	"lib/tlog"
	"lib/utils/net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

func apiEncodeResponse(res interface{}, url string) []byte {

	switch entry := res.(type) {
	case *tlog.RecordS:
		if res.(*tlog.RecordS).Level == tlog.LevelError {

			buf, err := json.MarshalIndent(entry.Message, "", "  ")

			if entry.Message == "wystapił problem" && len(entry.Vars) == 1 {

				tmp := struct {
					Message string
					Details interface{}
				}{
					Message: entry.Message,
					Details: entry.Vars[0],
				}
				buf, err = json.MarshalIndent(tmp, "", "  ")
			}

			if err != nil {
				tlog.Error(err, tlog.Vars{
					"url": url,
				})
				return nil
			}
			buf = append([]byte{0}, buf...) // fail
			return buf
		}
	}

	// ---

	buf, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		tlog.Error(err, tlog.Vars{
			"url": url,
		})
		return nil
	}
	buf = append([]byte{1}, buf...) // success
	return buf
}

func apiMiddleware(fn func(r *http.Request, user *db.UserS) interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Cache-control", "no-store")

		if r.URL.Path == "/api/system-print-stack" {
			tlog.Info("recv /api/system-print-stack")

			buf := make([]byte, 1<<20)
			stacklen := runtime.Stack(buf, true)
			w.Write(buf[:stacklen])
			return
		}

		defer db.PanicHandlerCB(func() {
			buf := append([]byte{0}, []byte("\"internal server error\"")...)
			w.Write(buf)
		})

		if r.URL.Path == "/api/system-pod-cache" {

			fnr := fn(r, nil)
			_, ok := fnr.(io.ReadCloser)
			if ok {
				io.Copy(w, fnr.(io.ReadCloser))

			} else if _, ok = fnr.(io.Reader); ok {
				io.Copy(w, fnr.(io.Reader))
			} else {
				w.Write(apiEncodeResponse(
					fnr,
					r.URL.String(),
				))
			}

			return
		}

		// --------------------------------------------------------

		sessionID := r.Header.Get("Session")
		var user *db.UserS

		if sessionID == db2.TheImageBuilder.Timoni_Token() {
			// Image Builder
			user = &db.UserS{
				ID:       "ImageBuilder",
				Name:     "ImageBuilder",
				Email:    "ImageBuilder",
				GitToken: db2.TheImageBuilder.Timoni_Token(),
			}

		} else {
			user = getUser(sessionID, r, w)
		}

		if user == nil {
			return
		}

		fnr := fn(r, user)
		_, ok := fnr.(io.ReadCloser)
		if ok {
			io.Copy(w, fnr.(io.ReadCloser))

		} else if _, ok = fnr.(io.Reader); ok {
			io.Copy(w, fnr.(io.Reader))
		} else {
			w.Write(apiEncodeResponse(
				fnr,
				r.URL.String(),
			))
		}

	})
}

// --------------------------------------------------

func getUser(sessionID string, r *http.Request, w http.ResponseWriter) *db.UserS {

	userIP := net.RequestIP(r)

	sess, ok := db.SessionMap.GetFull(sessionID)
	if !ok {
		w.Write(apiEncodeResponse(
			tlog.Error("`session` is invalid", tlog.Vars{
				"logger":  "apiMiddleware",
				"url":     r.URL,
				"user-ip": userIP,
			}),
			r.URL.String(),
		))
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	user := db.GetUserByID(sess.UserID)
	if user == nil {
		tlog.Warning("user with this email does not exist", tlog.Vars{
			"logger":     "apiMiddleware",
			"url":        r.URL,
			"user-ip":    userIP,
			"session-id": sessionID,
		})
		w.Write(apiEncodeResponse(
			tlog.Error(terrors.UserIsBlockedOrInvalid),
			r.URL.String(),
		))
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	if user.Teams.Exists(db.BlacklistedTeamName) {
		tlog.Warning("user is blacklisted", tlog.Vars{
			"logger":     "apiMiddleware",
			"url":        r.URL,
			"user-ip":    userIP,
			"user-email": user.Email,
			"session-id": sessionID,
		})
		w.WriteHeader(http.StatusForbidden)
		w.Write(apiEncodeResponse(
			tlog.RecordS{
				Level:   tlog.LevelWarning,
				Message: "user is blacklisted",
			},
			r.URL.String(),
		))
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}

	if sess.Expired() {
		// tlog.Warning("session is expired", tlog.Vars{
		// 	"logger":  "apiMiddleware",
		// 	"url":     r.URL,
		// 	"user-ip": userIP,
		// })
		w.Write(apiEncodeResponse(
			tlog.RecordS{
				Level:   tlog.LevelWarning,
				Message: "`session` is expired",
			},
			r.URL.String(),
		))
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	sess.Update()
	user.LastActionTimeStamp = time.Now().Unix()

	// Increment counter
	cloudHttpRequests.WithLabelValues(r.URL.Path, user.Email).Inc()

	return user
}

func apiGitOnBoard(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	userObj := db.GetUserByID(token)
	if userObj == nil {
		w.Write([]byte("invalid token"))
		return
	}

	if !userObj.Activated || userObj.Teams.Exists(db.BlacklistedTeamName) {
		w.Write([]byte("User not activated"))
		return
	}

	script :=
		`#!/bin/sh

set -e

mkdir -p ~/.timoni
echo 'Set-Cookie:token=` + userObj.GitToken + `; Path=/; Domain=` + db2.TheDomain.Name() + `' >> ~/.timoni/cookie
git config --global http.cookieFile "~/.timoni/cookie"
echo "Git access added"
`
	w.Write([]byte(script))
}

func userSetup(w http.ResponseWriter, r *http.Request) {
	token := strings.Split(r.URL.Path, "/")[2]
	user := db.GetUserByID(token)
	if user == nil {
		w.Write([]byte("invalid token"))
		return
	}

	domain := db2.TheDomain

	url := domain.URL("/?token=" + token)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<html>
	
	Web admin panel:<br>
	<a href="` + url + `">` + url + `</a><br>
	<br><br>
	
	Workstation-agent install script:<br>
	curl ` + domain.URL("/files/workstation-agent") + ` -o workstation-agent && chmod +x workstation-agent && sudo ./workstation-agent setup ` + domain.Name() + ` ` + token + ` 
`

	k, err := totp.Generate(totp.GenerateOpts{
		Secret:      []byte(user.TotpSecret),
		Issuer:      db2.TheSettings.Name(),
		AccountName: user.Email,
	})
	if err != nil {
		tlog.Error(err.Error())
		return
	}
	tlog.PrintJSON(user)
	buff := bytes.NewBuffer([]byte(html))
	qrCode, err := k.Image(200, 200)
	if err != nil {
		tlog.Error(err.Error())
		return
	}
	//"data:image/png;base64," + yourByteArrayAsBase64
	b := bytes.NewBuffer([]byte{})
	if err := png.Encode(b, qrCode); err != nil {
		tlog.Error(err.Error())
		return
	}
	buff.Write([]byte(`<br>Totp QR code:<br><img src="data:image/png;base64, ` + base64.StdEncoding.EncodeToString(b.Bytes()) + `">`))
	buff.WriteTo(w)
}
