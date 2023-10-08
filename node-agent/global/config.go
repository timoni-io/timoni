package global

import "os"

var (
	GitTag      = "???"
	TimoniURL   = os.Getenv("TIMONI_URL")
	VMAgentAddr = os.Getenv("VM_AGENT_ADDR") // eg. 127.0.0.1:4242
)
