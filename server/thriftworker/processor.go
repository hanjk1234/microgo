package thriftworker

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
)

var (
	thriftServiceProcessor = make(map[string]func() thrift.TProcessor)
)

func createThriftProcessor(name string) (thrift.TProcessor, error) {
	if f, ok := thriftServiceProcessor[name]; ok {
		return f(), nil
	} else {
		return nil, fmt.Errorf("empty processor func")
	}
}
