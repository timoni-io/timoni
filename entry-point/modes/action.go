package modes

import (
	"crypto/tls"
	"entry-point/global"
	"fmt"
	"io"
	"lib/tlog"
	"net/http"
	"strings"
)

type ElementActionStatusT string

const (
	ElementActionStatusPending     ElementActionStatusT = "Pending"
	ElementActionStatusRunning     ElementActionStatusT = "Running"
	ElementActionStatusSucceeded   ElementActionStatusT = "Succeeded"
	ElementActionStatusFailed      ElementActionStatusT = "Failed"
	ElementActionStatusTerminating ElementActionStatusT = "Terminating"
)

type actionMode struct{}

func (a actionMode) Start() {
	tlog.Info("Container starts in action mode with command: " + strings.Join(global.ProcessCommand, " "))
	a.sendStatus(ElementActionStatusRunning)

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	if err != nil {
		a.sendStatus(ElementActionStatusFailed)
		tlog.Error(fmt.Sprintf("Command %s has ended with an error. Exit code %d, message: %s", global.ProcessCommand, exitCode, err.Error()))
		return
	}

	a.sendStatus(ElementActionStatusSucceeded)
	tlog.Info(fmt.Sprintf("Command %s has ended correctly. Exit code %d", global.ProcessCommand, exitCode))
}

func (a actionMode) End() {
	select {}
}

func (a actionMode) sendStatus(status ElementActionStatusT) {

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}
	response, err := client.Get(fmt.Sprintf(
		"https://%s/api/entry-point-actions-status?envID=%s&elementName=%s&actionToken=%s&status=%s",
		global.TimoniURL,
		global.EnvironmentID,
		global.ElementName,
		global.ActionToken,
		status,
	))
	if err != nil {
		fmt.Fprintf(global.STDERR, "Error sending action status: %s\n", err.Error())
		return
	}

	if response.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Fprintf(global.STDERR, "Error reading action status response: %s\n", err.Error())
			return
		}
		defer response.Body.Close()

		fmt.Fprintf(global.STDERR, "Received error while sending status: %s, %s\n", response.Status, string(body))
		return
	}
}
