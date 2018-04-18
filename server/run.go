package server

import (
	"os"
	"github.com/seefan/microgo/server/worker"
)

var (
	CMD = []string{"start", "stop", "restart"}
)

func Run(w worker.Worker) {
	cmd := "start"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	if CMD[0] == cmd {
		worker.Start(w, true)
	} else if CMD[1] == cmd {
		worker.Stop()
	} else if CMD[2] == cmd {
		worker.Stop() //restart with last config
		worker.Start(w, false)
	} else {
		println("Usage: ./you_file {start|stop|restart}")
	}
}
