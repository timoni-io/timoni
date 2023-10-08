package kube

import (
	"bufio"
	"fmt"
	"io"
	log "lib/tlog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

type DomainPathS struct {
	ElementName string `toml:"element"`
	Port        int32  `toml:"port"`
	Prefix      string `toml:"prefix"`
	Label       string `toml:"label"`
}

func cmdRunAndPrintOutput(args ...string) (success bool) {
	cmdString := strings.Join(args, " ")
	log.Info("cmd running: " + cmdString)

	cmd := exec.Command(args[0], args[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	multi := io.MultiReader(stdout, stderr)
	out := bufio.NewScanner(multi)

	for out.Scan() {
		fmt.Println(out.Text())
	}

	cmd.Wait()
	return cmd.ProcessState.ExitCode() == 0
}
