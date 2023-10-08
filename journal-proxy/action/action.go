package action

import ws "lib/ws/server"

func Register(server *ws.Server) {
	ws.RegisterAction[Get](server, "Get")
	ws.RegisterSubscription[Live](server, "Live")
	ws.RegisterAction[Insert](server, "Insert")
}
