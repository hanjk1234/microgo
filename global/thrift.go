package global

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"fmt"
)

var (
	//
	thriftServiceProcessor = map[string]func() thrift.TProcessor{}
)

func RegisterThriftProcessor(name string, proc func() thrift.TProcessor) {
	thriftServiceProcessor[name] = proc
}
func CreateThriftProcessor(name string) (thrift.TProcessor, error) {
	if f, ok := thriftServiceProcessor[name]; ok {
		return f(), nil
	} else {
		return nil, fmt.Errorf("empty processor func")
	}
}


