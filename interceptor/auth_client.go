package interceptor

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
)

const tokenHeader = "Authorization"

type ClientAuthInterceptor struct {
	token string
}

func NewClientAuthInterceptor(token string) *ClientAuthInterceptor {
	return &ClientAuthInterceptor{token}
}

func (i *ClientAuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	// Same as previous UnaryInterceptorFunc.
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		req.Header().Set(tokenHeader, fmt.Sprintf("Bearer %s", i.token))
		return next(ctx, req)
	})
}

func (i *ClientAuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(
		ctx context.Context,
		spec connect.Spec,
	) connect.StreamingClientConn {
		conn := next(ctx, spec)
		conn.RequestHeader().Set(tokenHeader, fmt.Sprintf("Bearer %s", i.token))
		return conn
	})
}

func (i *ClientAuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(
		ctx context.Context,
		conn connect.StreamingHandlerConn,
	) error {
		return next(ctx, conn)
	})
}
