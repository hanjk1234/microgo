package test

import "context"

type HelloWorldImpl struct {
}

// Parameters:
//  - Name
func (h *HelloWorldImpl) SayHello(ctx context.Context, name string) (r string, err error) {
	return "hello " + name, nil
}
