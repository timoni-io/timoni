package ws

import (
	"lib/terrors"
)

type ResponseS struct {
	RequestID string
	Code      terrors.Error
	Data      any
}

func extractValue(data any) any {
	switch data := data.(type) {
	case error:
		return data.Error()
	default:
		return data
	}
}

func Response(r *Request, code terrors.Error, data any) *ResponseS {
	data = extractValue(data)

	// switch code {
	// case codes.Success:
	// default:
	// log.Errorw("code: "+code.String(), "data", data)
	// }

	return &ResponseS{
		RequestID: r.RequestID,
		Code:      code,
		Data:      data,
	}
}
