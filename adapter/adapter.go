package adapter

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	v1 "github.com/chushi-io/timber/gen/server/v1"
	"github.com/chushi-io/timber/gen/server/v1/serverv1connect"
	"github.com/chushi-io/timber/interceptor"
	"io"
	"net/http"
)

type Adapter struct {
	logService serverv1connect.LogsServiceClient
	stream     *connect.ClientStreamForClient[v1.StreamLogsRequest, v1.StreamLogsResponse]
	source     string
}

func New(address string, authToken string, resource string) io.Writer {
	interceptors := connect.WithInterceptors(
		interceptor.NewClientAuthInterceptor(authToken),
	)
	return &Adapter{
		logService: serverv1connect.NewLogsServiceClient(
			http.DefaultClient,
			address,
			interceptors,
		),
	}
}

func (a *Adapter) Write(p []byte) (int, error) {
	if a.stream == nil {
		fmt.Println("opening stream")
		a.stream = a.logService.Forward(context.TODO())
	}
	fmt.Println("sending log message")
	fmt.Println(string(p))
	if err := a.stream.Send(&v1.StreamLogsRequest{
		Resource: a.source,
		Logs:     p,
	}); err != nil {
		fmt.Println(err)
		return 0, err
	}
	fmt.Println("log message sent")
	return len(p), nil
}

func (a *Adapter) Flush() error {
	_, err := a.stream.CloseAndReceive()
	return err
}
