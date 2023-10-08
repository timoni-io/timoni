package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
	"unsafe"
)

type Command struct {
	Name string
	Args []string

	cmd *exec.Cmd
}

func NewCommand(name string, args ...string) *Command {
	return &Command{
		Name: name,
		Args: args,
	}
}

type RunOptions struct {
	Dir string
	Env map[string]string

	Stdin io.Reader
	// When using Stdout, it's impossible to use PipeReader
	Stdout io.Writer
	// When using Stderr, it's impossible to use PipeReader
	Stderr io.Writer
	// When using PipeReader, it's impossible to use Stdout or/and Stderr
	PipeReader func(stdout, stderr io.ReadCloser) `json:"-"`

	DisablePager bool

	Timeout time.Duration
}

func (c *Command) Run(opts *RunOptions) error {
	if opts == nil {
		opts = &RunOptions{}
	}

	var ctx context.Context
	var cancel context.CancelFunc

	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), opts.Timeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	if opts.Env != nil {
		for k, v := range opts.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	} else {
		cmd.Env = os.Environ()
	}

	if opts.DisablePager {
		cmd.Env = append(cmd.Env, "PAGER='cat'")
	}

	cmd.Dir = opts.Dir
	cmd.Stdin = opts.Stdin
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr

	if opts.PipeReader != nil {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		go opts.PipeReader(stdout, stderr)
	}

	c.cmd = cmd
	if err := cmd.Run(); err != nil && ctx.Err() != context.DeadlineExceeded {
		return err
	}
	return ctx.Err()
}

func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b)) // that's what Golang's strings.Builder.String() does (go/src/strings/builder.go)
}

// RunStdBytes runs the command with options and returns stdout/stderr as bytes
func (c *Command) RunStdBytes(opts *RunOptions) (stdout, stderr []byte, err error) {
	if opts == nil {
		opts = &RunOptions{}
	}
	if opts.Stdout != nil || opts.Stderr != nil {
		// must panic here, otherwise there would be bugs if developers set Stdin/Stderr by mistake, and it would be very difficult to debug
		panic("stdout and stderr field must be nil when using RunStdBytes")
	}

	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	opts.Stdout = stdoutBuf
	opts.Stderr = stderrBuf
	if err = c.Run(opts); err != nil {
		// return error and stderr output
		return nil, stderrBuf.Bytes(), err
	}

	// even if there is no err, there could still be some stderr output
	return bytes.TrimSpace(stdoutBuf.Bytes()),
		bytes.TrimSpace(stderrBuf.Bytes()),
		nil
}

// RunStdString runs the command with options and returns stdout/stderr as string
func (c *Command) RunStdString(opts *RunOptions) (stdout, stderr string, err error) {
	stdoutBytes, stderrBytes, err := c.RunStdBytes(opts)
	return bytesToString(stdoutBytes), bytesToString(stderrBytes), err
}

func (c *Command) ExitCode() int {
	state := c.cmd.ProcessState
	for state == nil {
		time.Sleep(time.Millisecond)
	}
	return state.ExitCode()
}
