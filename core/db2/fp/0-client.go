package fp

import (
	"context"
	"core/config"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/lukx33/lwhelper/out"
	"github.com/quic-go/quic-go"
)

type clientS struct {
	serverAddress string
	serverDomain  string
	connection    quic.Connection
}

// ---

var client = &clientS{
	serverAddress: config.FocalPointAddr(),
	serverDomain:  "fp.timoni.io",
}

func (client *clientS) Connect() out.Info {

	if client.connection != nil {

		if client.connection.Context().Err() == nil {
			// already connected
			return out.NewSuccess()
		}

		// reconnect needed
	}

	var err error
	client.connection, err = quic.DialAddr(
		context.Background(),
		client.serverAddress,
		&tls.Config{
			ServerName: client.serverDomain,
			NextProtos: []string{"h3", "h3-32", "h3-31", "h3-30", "h3-29"},
		},
		&quic.Config{
			KeepAlivePeriod:       15 * time.Second,
			HandshakeIdleTimeout:  2 * time.Second,
			MaxIdleTimeout:        3 * time.Second,
			MaxIncomingUniStreams: -1,
		},
	)
	if err != nil {
		return out.New(err)
	}

	return out.NewSuccess()
}

// ---

func (client *clientS) call(action string, args any, response out.Info) {

	request := &requestS{
		Action: action,
		// TODO: UserID: config.Token(), // kto wykonuje ta akcje, autoryzacja operacji
	}
	request.Payload, _ = json.Marshal(args)

	// out.PrintJSON(struct {
	// 	Args    interface{}
	// 	Request interface{}
	// 	Trace   interface{}
	// }{
	// 	Args:    args,
	// 	Request: request,
	// 	Trace:   out.Trace(1),
	// })

	// ---

	for i := 0; i < 3; i++ {

		// reconnect gdy serwer zerwie polaczenie
		if client.innterCall(request, response) {
			return // success
		}

		// reset connection and retry
		client.connection = nil
		time.Sleep(time.Duration(i * 10000))
	}
}

func (client *clientS) innterCall(request *requestS, response out.Info) bool {

	if connInfo := client.Connect(); connInfo.NotValid() {
		response.InfoAddCause(connInfo)
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1*time.Second))
	defer cancel()

	stream, err := client.connection.OpenStreamSync(ctx)
	if out.CatchError(response, err).NotValid() {
		return false
	}

	// ---

	writer := cbor.NewEncoder(stream)
	if out.CatchError(response, writer.Encode(request)).NotValid() {
		return false
	}

	reader := cbor.NewDecoder(stream)
	return !out.CatchError(response, reader.Decode(response)).NotValid()
}

// ---

type testPushS struct {
	Ala  string
	Ma   int
	Kota bool
}

func (client *clientS) pull(args any, response out.Info) {

	request := &requestS{
		Action: "testPush",
		// TODO: UserID: config.Token(), // kto wykonuje ta akcje, autoryzacja operacji
	}
	request.Payload, _ = json.Marshal(args)

	out.PrintJSON(string(request.Payload))
	out.PrintJSON(request)

	// ---

	// TODO: reconnect gdy serwer zrestaruje polaczenie

	if connInfo := client.Connect(); connInfo.NotValid() {
		response.InfoAddCause(connInfo)
		client.connection = nil
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1*time.Second))
	defer cancel()

	stream, err := client.connection.OpenStreamSync(ctx)
	if out.CatchError(response, err).NotValid() {
		client.connection = nil
		return
	}

	// ---

	writer := cbor.NewEncoder(stream)
	if out.CatchError(response, writer.Encode(request)).NotValid() {
		return
	}

	reader := cbor.NewDecoder(stream)

	for {
		response := new(testPushS)
		err = reader.Decode(response)
		if err != nil {
			fmt.Println(err)
			return
		}
		out.PrintJSON(response)
	}
}

func TestPull() {
	fmt.Println("*** TestPull - begin")
	client.pull(
		req_listQueryS{
			Where: "query test...",
		},
		&out.DontUseMeInfoS{},
	)
	fmt.Println("*** TestPull - end")
}

func SendEmail(toEmail, subject, htmlMessage string) out.Info {

	response := new(out.DontUseMeInfoS)
	client.call(
		"SendEmail",
		req_sendMailS{
			ToEmail:     toEmail,
			Subject:     subject,
			HtmlMessage: htmlMessage,
		},
		response,
	)
	return response
}
