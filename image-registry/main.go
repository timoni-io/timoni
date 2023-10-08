package main

import (
	"context"
	"image-registry/embed"
	log "lib/tlog"
	"net/http"
)

func main() {
	embed.Start(context.Background(), http.DefaultServeMux, "./data/registry")
	log.Fatal(http.ListenAndServe(":7101", nil))
}
