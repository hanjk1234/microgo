package main

import (
	"github.com/seefan/microgo/server"
	"github.com/seefan/microgo/server/thriftworker"
	"github.com/seefan/microgo/server/worker"
)

func main() {
	var run worker.Worker = new(thriftworker.ThriftWorker)
	server.Run(run)
}
