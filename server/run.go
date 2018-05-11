package server

import (
	"os"
	"github.com/seefan/microgo/server/worker"
	"github.com/seefan/microgo/global"
	"github.com/golangteam/function/run"
	"syscall"
	"github.com/seefan/microgo/server/common"
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
		start(w, true)
	} else if CMD[1] == cmd {
		worker.Stop()
	} else if CMD[2] == cmd {
		worker.Stop() //restart with last config
		start(w, false)
	} else {
		println("Usage: ./you_file {start|stop|restart} [options]")
		if err := worker.Start(w, true); err != nil {
			println(err.Error())
		}
	}
}
func start(w worker.Worker, reloadConfig bool) {
	f, _ := os.Create(common.Path(global.RuntimeRoot, "nohup.log"))
	run.Nohup(func() error {
		return worker.Start(w, true)
	}, func(e error) {
		w.Stop()
	}, f, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
}
