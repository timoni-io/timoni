package utils

import (
	"bytes"
	"fmt"
	log "lib/tlog"
	"os"
	"os/exec"
	"strings"
	"time"
)

func PipelineCmds(cmds ...*exec.Cmd) (pipeLineOutput, collectedStandardError []byte, pipeLineError error) {
	// Require at least one command
	if len(cmds) < 1 {
		return nil, nil, nil
	}

	// Collect the output from the command(s)
	var output bytes.Buffer
	var stderr bytes.Buffer

	last := len(cmds) - 1
	for i, cmd := range cmds[:last] {
		var err error
		// Connect each command's stdin to the previous command's stdout
		if cmds[i+1].Stdin, err = cmd.StdoutPipe(); err != nil {
			return nil, nil, err
		}
		// Connect each command's stderr to a buffer
		cmd.Stderr = &stderr
	}

	// Connect the output and error for the last command
	cmds[last].Stdout, cmds[last].Stderr = &output, &stderr

	// Start each command
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}

	// Wait for each command to complete
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}

	// Return the pipeline output and the collected standard error
	return output.Bytes(), stderr.Bytes(), nil
}

func ExecCommand(cmd ...string) (string, *log.RecordS) {

	ts := time.Now()
	defer func() {
		t := time.Since(ts)
		if t.Seconds() > 5 {
			log.Warning("long command {{cmd}} {{time}}", log.Vars{
				"cmd":  cmd,
				"time": t.String(),
			})
		}
	}()

	exe := exec.Command(cmd[0], cmd[1:]...)

	exe.Env = []string{
		//"GIT_CURL_VERBOSE=1",
		"HOME=" + os.Getenv("HOME"),
		"USER=" + os.Getenv("USER"),
		"GIT_HTTP_CONNECT_TIMEOUT=2",
		"GIT_TERMINAL_PROMPT=0",
		"GIT_SSH_COMMAND=ssh -i /data/ssh_keys -o IdentitiesOnly=yes -o StrictHostKeyChecking=no",
	}

	out, err := exe.CombinedOutput()

	return string(out), log.Error(err, log.Vars{
		"cmd":    strings.Join(cmd, " "),
		"output": string(out),
	})
}

func ExecCommandEnv(cmd ...string) (string, error) {
	ts := time.Now()
	defer func() {
		t := time.Since(ts)
		if t.Seconds() > 5 {
			log.Warning("long command {{cmd}} {{time}}", log.Vars{
				"cmd":  cmd,
				"time": t.String(),
			})
		}
	}()

	exe := exec.Command(cmd[0], cmd[1:]...)
	exe.Env = os.Environ()

	out, err := exe.CombinedOutput()
	if err == nil {
		return string(out), nil
	}

	return string(out), fmt.Errorf("err: %s, CMD: %s", err, cmd)
}

func ExecCommandStdout(cmd ...string) *log.RecordS {

	ts := time.Now()
	cwd, _ := os.Getwd()
	defer func() {
		t := time.Since(ts)
		if t.Seconds() > 5 {
			log.Warning("long command {{cmd}} {{time}}", log.Vars{
				"cwd":  cwd,
				"cmd":  cmd,
				"time": t.String(),
			})
		}
	}()

	exe := exec.Command(cmd[0], cmd[1:]...)
	exe.Stdout = os.Stdout
	exe.Stderr = os.Stderr

	exe.Env = os.Environ()
	exe.Env = append(exe.Env, "GIT_HTTP_CONNECT_TIMEOUT=2")

	return log.Error(exe.Run(), log.Vars{
		"cwd":  cwd,
		"cmd":  strings.Join(cmd, " "),
		"time": time.Since(ts).String(),
	})
}
