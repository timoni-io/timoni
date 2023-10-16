package modes

import (
	"entry-point/global"
	"fmt"
	"os"
	"os/exec"
)

type StartEnder interface {
	// Start starts entry-point in specific mode (that implements Starter). It requires to be called after Setup.
	Start()
	// End ends entry-point in requested mode.
	End()
}

var (
	// Mode is the current mode of entry-point. It is set in init().
	Mode StartEnder
	// cmd is the command that will be executed.
	cmd *exec.Cmd

	// withContext defines if CommandContext is used.
	// withContext bool
	// ctx         context.Context
	// cancel      context.CancelFunc
)

func init() {

	if _, err := os.Stat("/data/sleep"); err == nil {
		fmt.Println("sleep mode")
		select {} // infinite loop
	}

	switch {
	case global.ActionToken != "":
		Mode = actionMode{}

	case global.CronExpression != "":
		Mode = cronMode{}

	default:
		Mode = serviceMode{}
	}
}

// Setup needs to be called before any other function.
func Setup() {
	// if withContext {
	// 	ctx, cancel = context.WithCancel(context.Background())
	// 	cmd = exec.CommandContext(ctx, global.ProcessCommand[0], global.ProcessCommand[1:]...)

	// } else {
	cmd = exec.Command(global.ProcessCommand[0], global.ProcessCommand[1:]...)
	// }
	cmd.Env = os.Environ()
	// if global.JournalProxyURL != "" {
	cmd.Stdout = global.LogWriter.Pipe()
	cmd.Stderr = global.LogWriter.Pipe()

	// } else {
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// }
}

func Reset() {
	// Copy previous pipes
	stdout, stderr := cmd.Stdout, cmd.Stderr

	// Recreate cmd
	cmd = exec.Command(global.ProcessCommand[0], global.ProcessCommand[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdout, cmd.Stderr = stdout, stderr
}
