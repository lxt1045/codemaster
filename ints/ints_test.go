package ints

import (
	"fmt"
	"testing"

	"golang.org/x/net/context"
)

type Handler func(ctx context.Context, req interface{}) (rsp interface{}, err error)
type Interceptor func(ctx context.Context, req interface{}, handler Handler) (rsp interface{}, err error)

func getChainInterceptorHandler(ctx context.Context, ints []Interceptor, idx int, h Handler) Handler {
	if idx == len(ints)-1 {
		return h
	}
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return ints[idx+1](ctx, req, h)
	}
}

func processHandler(ctx context.Context, req interface{}) (interface{}, error) {
	fmt.Println("process request")
	return nil, nil
}

func int1(ctx context.Context, req interface{}, handler Handler) (interface{}, error) {
	fmt.Println("int1 begin")
	handler(ctx, req)
	fmt.Println("int1 end")
	return nil, nil
}

func int2(ctx context.Context, req interface{}, handler Handler) (interface{}, error) {
	fmt.Println("int2 begin")
	handler(ctx, req)
	fmt.Println("int2 end")
	return nil, nil
}

func TestInterceptors(t *testing.T) {

	ints := []Interceptor{int1, int2}

	var firstChainedInterceptor Interceptor
	firstChainedInterceptor = func(ctx context.Context, req interface{}, h Handler) (rsp interface{}, err error) {
		return ints[0](ctx, req, h)
	}

	ctx := context.TODO()
	req := struct{}{}
	firstChainedInterceptor(ctx, req, getChainInterceptorHandler(ctx, ints, 0, processHandler))
}