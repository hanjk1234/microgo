package main

import (
	"github.com/seefan/microgo/server"
	"github.com/seefan/microgo/server/thriftworker"
	"github.com/seefan/microgo/server/worker"
	"github.com/seefan/microgo/global"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/seefan/microgo/test/gen-go/test"
	test2 "github.com/seefan/microgo/test"
)

func main() {
	global.RegisterServiceId( "test.HelloWorld","1002")

	var run worker.Worker = thriftworker.NewThriftWorker()
	thriftworker.RegisterThriftProcessor("test.HelloWorld", func() thrift.TProcessor {
		return test.NewHelloWorldProcessor(&test2.HelloWorldImpl{})
	})
	server.Run(run)
}
