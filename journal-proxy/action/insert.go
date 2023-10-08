package action

import (
	"journal-proxy/global"
	"journal-proxy/wsb"
	"lib/terrors"
	ws "lib/ws/server"
)

type Insert struct {
	Log *global.Entry
}

func (a Insert) Handle(r *ws.Request) (code terrors.Error, data any) {
	conn := wsb.ConnPool.GetNoWait()
	if conn == nil {
		return terrors.DatabaseError, nil
	}
	defer wsb.ConnPool.Add(conn)

	err := conn.InsertOne(a.Log)
	if err != nil {
		return terrors.InternalServerError, err
	}
	return terrors.Success, nil
}
