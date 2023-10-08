package main

import "net/http"

func httpStatus(w http.ResponseWriter, r *http.Request) {

	if building {
		w.Write([]byte("building"))
		return
	}

	w.Write([]byte("ready"))
}
