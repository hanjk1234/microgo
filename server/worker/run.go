package worker

import (
	"fmt"
	"github.com/seefan/microgo/server/common"
	"github.com/seefan/microgo/global"
	"os"
	log "github.com/cihub/seelog"
	"github.com/hashicorp/consul/api"
	"github.com/seefan/to"
)

func Start(w Worker, reloadConfig bool) error {
	defer common.PrintErr()

	global.RuntimeRoot = os.Args[0] + "_runtime"

	if common.NotExist(global.RuntimeRoot) {
		os.MkdirAll(global.RuntimeRoot, 0764)
	}
	if err := common.SavePid(common.Path(global.RuntimeRoot, "worker.pid")); err != nil {
		println("can not save pid file")
	}
	if common.NotExist(common.Path(global.RuntimeRoot, "logs")) {
		os.Mkdir(common.Path(global.RuntimeRoot, "logs"), 0764)
	}
	common.InitLog(common.Path(global.RuntimeRoot, "log.xml"), common.Path(global.RuntimeRoot, "logs", "micro_go.log"))
	defer log.Flush()
	if reloadConfig {
		buildConfig()
	}
	log.Info("worker starting")
	return w.Start()
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
	if pid, err := common.GetPid(common.Path(global.RuntimeRoot, "worker.pid")); err == nil {
		common.Kill(pid)
	}

}
