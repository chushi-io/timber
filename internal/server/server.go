package server

import (
	"connectrpc.com/connect"
	"context"
	v1 "github.com/chushi-io/timber/gen/server/v1"
	"log"
	"os"
)

type LogService struct {
}

func New() *LogService {
	return &LogService{}
}

func (l *LogService) Forward(ctx context.Context, stream *connect.ClientStream[v1.StreamLogsRequest]) (*connect.Response[v1.StreamLogsResponse], error) {
	log.Println("Request headers: ", stream.RequestHeader())
	var logFile *os.File
	var err error

	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()
	for stream.Receive() {
		if logFile == nil {
			logFile, err = os.OpenFile(stream.Msg().Resource, os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnknown, err)
			}
		}

		if _, err = logFile.WriteString(string(stream.Msg().Logs)); err != nil {
			return nil, connect.NewError(connect.CodeUnknown, err)
		}
	}

	return connect.NewResponse(&v1.StreamLogsResponse{}), nil
}
