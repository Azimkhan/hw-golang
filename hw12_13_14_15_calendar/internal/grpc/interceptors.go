package grpc

import (
	"context"
	"fmt"
	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/app"
	"google.golang.org/grpc"
	"time"
)

func UnaryLoggingInterceptor(logger app.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		res, err := handler(ctx, req)
		dur := time.Since(start)
		logger.Info(fmt.Sprintf("[gRPC] %s, request=%v, response=%v, err=%v, duraton=%s", info.FullMethod, req, res, err, dur))
		return res, err
	}
}
