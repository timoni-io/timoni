package modes

import (
	"context"
	"entry-point/global"
	"lib/tlog"
	"strings"
	"time"

	"github.com/adhocore/gronx"
	"github.com/adhocore/gronx/pkg/tasker"
)

var taskCompleted = true

type cronMode struct{}

func (c cronMode) Start() {
	if cron := gronx.New(); !cron.IsValid(global.CronExpression) {
		tlog.Fatal("Cron expression is not valid")
	}

	tlog.Info("Container starts in cronjob mode with command: " + strings.Join(global.ProcessCommand, " "))

	taskr := tasker.New(tasker.Option{
		Shell: "/bin/sh",
		// Verbose: true,
	})

	taskr.Task(global.CronExpression, func(ctx context.Context) (int, error) {
		defer func() {
			Reset()
			taskCompleted = true
		}()
		if !taskCompleted {
			tlog.Warning("Previous cronjob is still running, skipping execution")
			return 2, nil
		}
		taskCompleted = false
		err := cmd.Run()
		if err != nil {
			taskr.Log.Println(err)
			tlog.Error(err.Error())
			return 1, err
		}
		return 0, nil
	})

	if global.CronUntilMinutes != 0 {
		taskr.Until(time.Duration(global.CronUntilMinutes) * time.Minute)
	}

	taskr.Run()
}

func (c cronMode) End() {
	tlog.Warning("Cronjob has ended")
	select {}
}
