package modes

import (
	"entry-point/global"
	"fmt"
	"lib/tlog"
	"os"
	"strings"
	"time"
)

type serviceMode struct{}

func (s serviceMode) Start() {
	tlog.Info("Container starts in service mode with command: "+global.ProcessCommand[0]+" / args: "+strings.Join(global.ProcessCommand[1:], " "), tlog.Vars{"event": "true"})

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	if err == nil {
		tlog.Info(fmt.Sprintf("Command %s has ended correctly. Exit code %d", global.ProcessCommand, exitCode), tlog.Vars{"event": "true"})

	} else {
		tlog.Error(fmt.Sprintf("Command %s has ended with an error. Exit code %d, message: %s", global.ProcessCommand, exitCode, err.Error()), tlog.Vars{"event": "true"})
	}

}

func (s serviceMode) End() {

	for {
		Setup()
		s.Start()
		time.Sleep(time.Second * 5)

		if _, err := os.Stat("/data/sleep"); err == nil {
			fmt.Println("sleep mode")
			select {} // infinite loop
		}
	}
}
