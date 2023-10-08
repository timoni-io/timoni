package term

import (
	"context"
	"core/kube"
	"encoding/json"
	"errors"
	"io"
	log "lib/tlog"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  BufferSize,
	WriteBufferSize: BufferSize,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Socket(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Unable to upgrade connection")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	// IO
	// Use os.Pipe(), not io.Pipe() as it's closing properly
	stdinReader, stdinWriter, _ := os.Pipe()
	stdoutReader, stdoutWriter, _ := os.Pipe()
	resizer := &termResizer{
		Buffer: make(chan *remotecommand.TerminalSize, 10),
	}

	// Cleanup
	go func() {
		<-ctx.Done()

		// Send CTRL+C and CTRL+D 4 times
		for i := 0; i < 4; i++ {
			stdinWriter.Write([]byte{'\x03'})
			stdinWriter.Write([]byte{'\x04'})
		}

		stdinWriter.Close()
		stdoutWriter.Close()
		stdinReader.Close()
		stdoutReader.Close()
		conn.Close()
	}()

	// Shell
	go func() {
		defer cancel()
		err := startShell(
			vars.Get("namespace"), vars.Get("pod"), vars.Get("container"),
			&termIO{
				Stdin:   stdinReader,
				Stdout:  stdoutWriter,
				Resizer: resizer,
			},
			ctx,
		)
		if err != nil {
			conn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
		}
	}()

	// Shell -> WS
	go func() {
		defer cancel()
		buf := make([]byte, BufferSize)
		for {
			read, err := stdoutReader.Read(buf)
			if err != nil {
				if !errors.Is(err, os.ErrClosed) {
					log.Error(err)
				}
				return
			}
			conn.WriteMessage(websocket.BinaryMessage, buf[:read])
		}
	}()

	// WS -> Shell
	go func() {
		defer cancel()

		dataTypeBuf := make([]byte, 1)
		size := &termSize{}

		for {
			_, reader, err := conn.NextReader()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseGoingAway) && !strings.HasSuffix(err.Error(), "use of closed network connection") {
					log.Error(err)
				}
				return
			}

			_, err = reader.Read(dataTypeBuf)
			if err != nil {
				log.Error(err)
				return
			}

			switch dataTypeBuf[0] {
			case 0:
				_, err := io.Copy(stdinWriter, reader)
				if err != nil {
					log.Error(err)
				}
			case 1:
				decoder := json.NewDecoder(reader)
				err = decoder.Decode(size)
				if err != nil {
					conn.WriteMessage(websocket.TextMessage, []byte("Resize failed: "+err.Error()))
					continue
				}
				resizer.Push(size)
			default:
				log.Error("Unknown data type")
			}
		}
	}()
}

func startShell(ns, pod, cnt string, io *termIO, ctx context.Context) error {
	kClient := kube.GetKube()
	req := kClient.API.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(pod).
		Namespace(ns).
		SubResource("exec")
	req.VersionedParams(
		&v1.PodExecOptions{
			Container: cnt,
			Command:   []string{"sh", "-c", "export TERM=xterm; (bash || ash -l || sh -l)"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		},
		scheme.ParameterCodec,
	)

	exec, err := remotecommand.NewSPDYExecutor(kClient.Config, http.MethodPost, req.URL())
	if err != nil {
		return err
	}

	return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Tty:               true,
		Stdin:             io.Stdin,
		Stdout:            io.Stdout,
		TerminalSizeQueue: io.Resizer,
	})
}
