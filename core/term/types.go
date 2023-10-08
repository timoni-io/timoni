package term

import (
	"io"

	"k8s.io/client-go/tools/remotecommand"
)

const BufferSize = 16 << 10

type termSize struct {
	Cols uint16
	Rows uint16
}

type termResizer struct {
	Buffer chan *remotecommand.TerminalSize
}

func (w *termResizer) Push(size *termSize) {
	w.Buffer <- &remotecommand.TerminalSize{
		Width:  size.Cols,
		Height: size.Rows,
	}
}

func (w *termResizer) Next() *remotecommand.TerminalSize {
	return <-w.Buffer
}

type termIO struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Resizer remotecommand.TerminalSizeQueue
}
