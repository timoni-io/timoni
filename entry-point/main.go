package main

import (
	"entry-point/global"
	"entry-point/modes"
	"entry-point/server"
	"fmt"
	"os"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func main() {

	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(global.GitTag)
		os.Exit(0)
	}

	fmt.Fprintln(global.STDOUT, "Entry-point "+global.GitTag+" is starting... "+global.EnvironmentID+" / "+global.ElementName)

	if _, err := os.Stat("/data/sleep"); err == nil {
		fmt.Println("sleep mode")
		select {} // infinite loop
	}

	if len(global.ProcessCommand) == 0 {
		fmt.Fprintln(global.STDOUT, "Container entry-point starts, but no command was given to run, so I am waiting in an endless loop...")
		select {}
	}

	modes.ApplyVariablesOnFiles()
	go modes.TailDir()
	go server.ServeStaticFiles()

	modes.Setup()
	modes.Mode.Start()
	modes.Mode.End()
}
