package server

import (
	"connectrpc.com/connect"
	"context"
	v1 "github.com/chushi-io/timber/gen/server/v1"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

type LogService struct {
	logDirectory string
	logger       *zap.Logger
}

func New(logDirectory string, logger *zap.Logger) *LogService {
	return &LogService{logDirectory, logger}
}

func (l *LogService) Forward(ctx context.Context, stream *connect.ClientStream[v1.StreamLogsRequest]) (*connect.Response[v1.StreamLogsResponse], error) {
	l.logger.Debug("forward request received", zap.String("resource", stream.Msg().Resource))
	var logFile *os.File
	var err error

	defer func() {
		if logFile != nil {
			l.logger.Debug("closing file", zap.String("log", logFile.Name()))
			logFile.Close()
		}
	}()
	for stream.Receive() {
		l.logger.Debug("received log stream", zap.String("resource", stream.Msg().Resource))
		if logFile == nil {
			l.logger.Debug("opening file", zap.String("log", logFile.Name()))
			logFile, err = os.OpenFile(filepath.Join(l.logDirectory, stream.Msg().Resource), os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				l.logger.Error("failed opening file", zap.Error(err), zap.String("log", logFile.Name()))
				return nil, connect.NewError(connect.CodeUnknown, err)
			}
		}

		if _, err = logFile.WriteString(string(stream.Msg().Logs)); err != nil {
			l.logger.Error("failed writing to file", zap.Error(err), zap.String("log", logFile.Name()))
			return nil, connect.NewError(connect.CodeUnknown, err)
		}
	}

	return connect.NewResponse(&v1.StreamLogsResponse{}), nil
}
