package adapter

import (
	"connectrpc.com/connect"
	"context"
	v1 "github.com/chushi-io/timber/gen/server/v1"
	"github.com/chushi-io/timber/gen/server/v1/serverv1connect"
	"github.com/chushi-io/timber/interceptor"
	"io"
	"net/http"
)

type Adapter struct {
	stream *connect.ClientStreamForClient[v1.StreamLogsRequest, v1.StreamLogsResponse]
	source string
}

func New(address string, authToken string, resource string) io.Writer {
	interceptors := connect.WithInterceptors(
		interceptor.NewClientAuthInterceptor(authToken),
	)
	logService := serverv1connect.NewLogsServiceClient(
		http.DefaultClient,
		address,
		interceptors,
	)
	return &Adapter{
		stream: logService.Forward(context.TODO()),
	}
}

func (a *Adapter) Write(p []byte) (int, error) {
	if err := a.stream.Send(&v1.StreamLogsRequest{
		Resource: a.source,
		Logs:     p,
	}); err != nil {
		return 0, err
	}
	return len(p), nil
}
