package main

import (
	"github.com/seefan/microgo/server"
	"github.com/seefan/microgo/server/thriftworker"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/seefan/microgo/test/gen-go/test"
	test2 "github.com/seefan/microgo/test"
)

func main() {

	run := thriftworker.NewHttpWorker()
	run.RegisterThriftProcessor("test.HelloWorld", func() thrift.TProcessor {
		return test.NewHelloWorldProcessor(&test2.HelloWorldImpl{})
	})
	run.TransportFactory = thrift.NewTHttpClientTransportFactory("/abc")
	run.ProtocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	server.RegisterServiceId("test.HelloWorld", "1002")
	server.Run(run)
}
