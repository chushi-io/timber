package interceptor

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"strings"
)

type ServerAuthInterceptor struct{}

var (
	errNoToken      = errors.New("no token provided")
	errInvalidToken = errors.New("invalid token")
)

func NewServerAuthInterceptor() *ServerAuthInterceptor {
	return &ServerAuthInterceptor{}
}

func (i *ServerAuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	// Same as previous UnaryInterceptorFunc.
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		if req.Header().Get(tokenHeader) == "" {
			return nil, connect.NewError(connect.CodeUnauthenticated, errNoToken)
		}

		chunks := strings.Split(req.Header().Get(tokenHeader), " ")
		if len(chunks) != 2 {
			return nil, connect.NewError(connect.CodeUnauthenticated, errInvalidToken)
		}

		// TODO: We need to actually verify the tokens
		return next(ctx, req)
	})
}

func (*ServerAuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(
		ctx context.Context,
		spec connect.Spec,
	) connect.StreamingClientConn {
		conn := next(ctx, spec)
		conn.RequestHeader().Set(tokenHeader, "sample")
		return conn
	})
}

func (i *ServerAuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(
		ctx context.Context,
		conn connect.StreamingHandlerConn,
	) error {
		if conn.RequestHeader().Get(tokenHeader) == "" {
			return connect.NewError(connect.CodeUnauthenticated, errNoToken)
		}

		chunks := strings.Split(conn.RequestHeader().Get(tokenHeader), " ")
		if len(chunks) != 2 {
			return connect.NewError(connect.CodeUnauthenticated, errInvalidToken)
		}

		// TODO: We actually need to validate the token
		return next(ctx, conn)
	})
}
