package global

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"lib/terrors"
	"lib/utils/slice"
	wsc "lib/ws/client"
	"lib/ws/coder"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// DO NOT USE fmt.Print HERE, ONLY fmt.Fprint IF NEEDED
// fmt.Fprintln(stdout, "logwriter.Send")

// original STDOUT and STDERR
var (
	STDOUT = os.Stdout
	STDERR = os.Stderr
)

type LogWriterS struct {
	cfg    ConfigS
	conn   *wsc.ChatClient
	buffer *slice.Rigid[[]byte, uint]
}

type ConfigS struct {
	URL           string
	ShowOutput    bool
	BaseMessage   *Message
	BatchDuration time.Duration
	// 100_000 = max about 150 MB
	// 1_000_000 = max about 850 MB
	BatchCapacity uint
	ParserFormat  string
}

func (c *ConfigS) validate() error {
	if c == nil {
		return fmt.Errorf("logwriter: %d", terrors.ClusterConfigIsEmpty)
	}
	if c.URL == "" {
		return errors.New("logwriter: URL field empty")
	}
	if c.BaseMessage == nil {
		return errors.New("logwriter: BaseMessage field empty")
	}
	if c.BatchDuration == 0 {
		return errors.New("logwriter: BatchDuration field empty")
	}
	if c.BatchCapacity == 0 {
		return errors.New("logwriter: BatchCapacity field empty")
	}
	return nil
}

func NewLogWriter(cfg ConfigS) (*LogWriterS, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	w := &LogWriterS{
		cfg:    cfg,
		buffer: slice.NewRigid[[]byte](cfg.BatchCapacity).Safe(),
	}

	go func() {
		fmt.Fprintf(STDOUT, "logwriter.New: connecting to %s\n", cfg.URL)
		var err error
		w.conn, err = wsc.NewChatClient(wsc.ClientConfig{
			URL: cfg.URL,
			Dialer: &websocket.Dialer{
				HandshakeTimeout:  3 * time.Second,
				EnableCompression: true,
			},
			Stdout: STDOUT,
			Coder:  coder.GOB{},
		})
		if err != nil {
			fmt.Fprintln(STDOUT, "logwriter: %w", err)
			return
		}
		fmt.Fprintln(STDOUT, "logwriter.New: connected")
	}()

	go w.readwriter()

	return w, nil
}

func (w *LogWriterS) envelope(data []byte) *Message {
	entry := *w.cfg.BaseMessage
	entry.Data = data
	entry.Parser = w.cfg.ParserFormat
	return &entry
}

// Pipe writer
func (w *LogWriterS) Pipe() *os.File {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil
	}

	// start goroutine to read from pipe
	go func() {
		r := bufio.NewReader(reader)
		buf := make([]byte, 0, 1<<20) // 1MiB buffer

		// read lines
		for {
			n, err := r.Read(buf[:cap(buf)])
			if err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Fprintln(STDERR, "logwriter.Pipe:", err)
				}
				return
			}
			if n == 0 {
				continue
			}

			buf = buf[:n]
			if len(buf) == 0 {
				continue
			}

			// write to buffer
			c := make([]byte, n)
			copy(c, buf)
			w.buffer.Add(c)
			// fmt.Fprintln(STDOUT, "buf:", string(buf))
			// fmt.Fprintln(STDOUT, "len:", len(w.buffer.GetAll()))
		}
	}()

	return writer
}

func (w *LogWriterS) readwriter() {
	t := time.NewTicker(w.cfg.BatchDuration)
	defer t.Stop()

	for range t.C {
		var data []byte
		for _, b := range w.buffer.Take() {
			data = append(data, b...)
		}

		if len(data) == 0 {
			continue
		}
		// fmt.Fprintln(STDOUT, "sending:", string(data))
		w.WriteBytes(data)
	}
}

func (w *LogWriterS) writeEntry(entry *Message) error {
	// fmt.Fprintf(STDOUT, "%+v\n", entry)
	if w.conn == nil {
		// fmt.Println("entry-point: CH connection is nil")
		return nil
	}
	err := w.conn.Send(entry)
	if err != nil || w.cfg.ShowOutput {
		fmt.Fprintln(STDOUT, string(entry.Data))
	}
	return err
}

func (w *LogWriterS) WriteBytes(data []byte) (int, error) {
	err := w.writeEntry(w.envelope(data))
	if err != nil {
		return 0, err
	}
	return len(data), nil
}
