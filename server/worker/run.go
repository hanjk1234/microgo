package worker

import (
	"fmt"
	"github.com/seefan/microgo/server/common"
	"github.com/seefan/microgo/global"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	log "github.com/cihub/seelog"
	"github.com/hashicorp/consul/api"
	"github.com/seefan/to"
)

func Start(run Worker, reloadConfig bool) {
	defer common.PrintErr()

	global.RuntimeRoot = os.Args[0] + "_runtime"

	if common.NotExist(global.RuntimeRoot) {
		os.MkdirAll(global.RuntimeRoot, 0764)
		os.Mkdir(common.Path(global.RuntimeRoot, "logs"), 0764)
	}
	common.InitLog(common.Path(global.RuntimeRoot, "log.xml"), common.Path(global.RuntimeRoot, "logs", "micro_go.log"))
	defer log.Flush()
	if reloadConfig {
		buildConfig()
	}
	if err := common.SavePid(common.Path(global.RuntimeRoot, "worker.pid")); err != nil {
		log.Errorf("can not save pid file")
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	go func() {
		log.Info("worker starting")
		if err := run.Start(); err != nil {
			log.Infof("worker start failed:%s", err)
			println(err.Error())
			sig <- syscall.SIGABRT
		}
	}()

	s := <-sig

	log.Info("received signal:", s)
	if err := run.Stop(); err != nil {
		log.Error("worker close has error", err)
	} else {
		log.Info("worker is closed")
	}
}
func buildConfig() {
	cfg := common.NewConfig()
	cfg.Load("conf.ini")
	cfg.LoadArgs(os.Args)
	//cfg.Service = append(cfg.Service, "0001")
	cfg.Service = common.RemoveDuplicatesAndEmpty(cfg.Service)
	if cfg.Msr.Enabled {
		apiCfg := api.DefaultConfig()
		apiCfg.Address = fmt.Sprintf("%s:%d", cfg.Msr.Host, cfg.Msr.Port)
		if consul, err := api.NewClient(apiCfg); err != nil {
			log.Error("connect consul error", err)
		} else {
			if cfg.Host == "" || cfg.Host == "0.0.0.0" {
				if self, err := consul.Agent().Self(); err == nil {
					cfg.Host = to.String(self["Config"]["BindAddr"])
				}
			}
			if ks, _, err := consul.KV().List("config/system_node_test", nil); err == nil {
				for _, k := range ks {
					if string(k.Key[24:]) == cfg.Host {
						global.RuntimeTest = true
						cfg.IsTesting = true
						break
					}
				}
			}
		}
	}

	cfg.SaveJson(common.Path(global.RuntimeRoot, "conf.json"))
}
func Stop() {
	defer common.PrintErr()

	global.RuntimeRoot = os.Args[0] + "_runtime"

	cfg := common.NewConfig()
	cfg.LoadJson(global.RuntimeRoot + "/conf.json")

	common.InitLog(common.Path(global.RuntimeRoot, "log.xml"), common.Path(global.RuntimeRoot, "logs", "micro_go.log"))
	defer log.Flush()
	if cfg.Msr.Enabled {
		register := NewRegisterManager(cfg)
		register.RemoveService()
	}
	pid, err := common.GetPid(common.Path(global.RuntimeRoot, "worker.pid"))

	if err == nil {
		checkCmd := exec.Command("kill", "-s", "0", pid)
		killCmd := exec.Command("kill", "-s", "USR2", pid)
		now := time.Now()
		if err := killCmd.Run(); err == nil {
			for {
				if err := checkCmd.Run(); err != nil {
					break
				}
				time.Sleep(time.Millisecond * 100)
				if time.Since(now).Seconds() > 30 {
					break
				}
			}
		} else {
			log.Error("worker stop error", err)
		}
	} else {
		println("pid file not found")
	}
}
