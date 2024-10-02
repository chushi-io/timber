package interceptor

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"fmt"
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
			fmt.Println(errNoToken.Error())
			return nil, connect.NewError(connect.CodeUnauthenticated, errNoToken)
		}

		chunks := strings.Split(req.Header().Get(tokenHeader), " ")
		if len(chunks) != 2 {
			fmt.Println(errInvalidToken.Error())
			return nil, connect.NewError(connect.CodeUnauthenticated, errInvalidToken)
		}

		// TODO: We need to actually verify the tokens
		fmt.Println("Authenticated request")
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
			fmt.Println(errNoToken.Error())
			return connect.NewError(connect.CodeUnauthenticated, errNoToken)
		}

		chunks := strings.Split(conn.RequestHeader().Get(tokenHeader), " ")
		if len(chunks) != 2 {
			fmt.Println(errInvalidToken.Error())
			return connect.NewError(connect.CodeUnauthenticated, errInvalidToken)
		}

		// TODO: We actually need to validate the token
		fmt.Println("authenticated request")
		return next(ctx, conn)
	})
}
