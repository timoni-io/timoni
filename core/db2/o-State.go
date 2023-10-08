package db2

// This file is automatically generated, manual editing is not recommended.

type StateT uint16

const (
	State_suspended StateT = 1
	State_new       StateT = 2
	State_deploying StateT = 3
	State_error     StateT = 4
	State_ready     StateT = 5
	State_offline   StateT = 6
)

var translationMapEN_State = map[StateT]string{
	1: "suspended",
	2: "new",
	3: "deploying",
	4: "error",
	5: "ready",
	6: "offline",
}

func (o StateT) EN() string { return translationMapEN_State[o] }

var translationMapPL_State = map[StateT]string{
	1: "zatrzymana",
	2: "nowa",
	3: "wdrażanie",
	4: "błąd",
	5: "sprawna",
	6: "offline",
}

func (o StateT) PL() string { return translationMapPL_State[o] }
