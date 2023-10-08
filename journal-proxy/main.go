package main

import (
	"journal-proxy/action"
	"journal-proxy/cache"
	"journal-proxy/metrics"
	"journal-proxy/wsb"
	log "lib/tlog"
	um "lib/utils/metrics"
	ws "lib/ws/server"
	"net/http"
	_ "net/http/pprof"
	"os/exec"
	"strings"
)

func pprofSvgGoroutine(w http.ResponseWriter, r *http.Request) {
	cmd := `pprof -svg 'localhost:` + "4003" + `/debug/pprof/goroutine'`
	out, _ := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	sl := strings.Split(string(out), "\n")[2:]
	sl[6] = sl[6][:len(sl[6])-1] + ` style="background-color: #423b38;">`
	s := strings.Join(sl, "\n")
	s = strings.Replace(s, "fill=\"white\"", "fill=\"#423b38\"", 1)
	w.Write([]byte(s))
}

func pprofSvgAllocs(w http.ResponseWriter, r *http.Request) {
	cmd := `pprof -svg 'localhost:` + "4003" + `/debug/pprof/allocs'`
	out, _ := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	sl := strings.Split(string(out), "\n")[2:]
	sl[6] = sl[6][:len(sl[6])-1] + ` style="background-color: #423b38;">`
	s := strings.Join(sl, "\n")
	s = strings.Replace(s, "fill=\"white\"", "fill=\"#423b38\"", 1)
	w.Write([]byte(s))
}

func pprofSvgHeap(w http.ResponseWriter, r *http.Request) {
	cmd := `pprof -svg 'localhost:` + "4003" + `/debug/pprof/heap'`
	out, _ := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	sl := strings.Split(string(out), "\n")[2:]
	sl[6] = sl[6][:len(sl[6])-1] + ` style="background-color: #423b38;">`
	s := strings.Join(sl, "\n")
	s = strings.Replace(s, "fill=\"white\"", "fill=\"#423b38\"", 1)
	w.Write([]byte(s))
}

func pprofSvgCpu(w http.ResponseWriter, r *http.Request) {
	cmd := `pprof -svg 'localhost:` + "4003" + `/debug/pprof/profile?seconds=10&debug=1'`
	out, _ := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	sl := strings.Split(string(out), "\n")[2:]
	sl[6] = sl[6][:len(sl[6])-1] + ` style="background-color: #423b38;">`
	s := strings.Join(sl, "\n")
	s = strings.Replace(s, "fill=\"white\"", "fill=\"#423b38\"", 1)
	w.Write([]byte(s))
}

func main() {
	// debug.SetGCPercent(4000)
	// debug.SetMemoryLimit(1 << 30)
	go func() {
		r := http.NewServeMux()
		r.HandleFunc("/alloc", pprofSvgAllocs)
		r.HandleFunc("/heap", pprofSvgHeap)
		r.HandleFunc("/goro", pprofSvgGoroutine)
		r.HandleFunc("/cpu", pprofSvgCpu)
		http.ListenAndServe(":6060", r)
	}()

	log.Info("Starting journal-proxy...")

	// Initialize Cache with data from DB if connected
	if conn := wsb.ConnPool.GetNoWait(); conn != nil {
		wsb.ConnPool.Add(cache.Init(conn))
	}

	// Register ws actions and http endpoints
	api, err := ws.NewServer(ws.ServerConfig{})
	if err != nil {
		log.Error(err)
	}

	action.Register(api)

	// WS
	http.HandleFunc("/", api.Handler)   // Front websocket api
	http.HandleFunc("/in", wsb.Handler) // Backend websocket
	go wsb.Loop()

	// HTTP
	http.Handle("/metrics", um.HandlerPretty(&metrics.Vars))
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/delete-env", action.DeleteEnvHandler)

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// w.Write([]byte("ok"))
	})

	log.Info("Serving @ :4003")
	log.Fatal(http.ListenAndServe(":4003", nil))
}
