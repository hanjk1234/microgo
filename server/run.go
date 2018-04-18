package server

import (
	"os"
	"github.com/seefan/microgo/server/worker"
	"github.com/seefan/microgo/server/thriftworker"
)

var (
	CMD = []string{"start", "stop", "restart"}
)

func Run(worker worker.Worker) {
	cmd := "start"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	if CMD[0] == cmd {
		thriftworker.Start(worker, true)
	} else if CMD[1] == cmd {
		thriftworker.Stop()
	} else if CMD[2] == cmd {
		thriftworker.Stop() //restart with last config
		thriftworker.Start(worker, false)
	} else {
		println("Usage: ./you_file {start|stop|restart}")
	}
}
