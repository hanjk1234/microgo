package worker

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/hashicorp/consul/api"
	"github.com/seefan/microgo/server/common"
	"github.com/seefan/microgo/global"
)

type Register struct {
	config *common.Config
	consul *api.Client
}

func NewRegister(cfg *common.Config) *Register {
	r := &Register{
		config: cfg,
	}
	apiCfg := api.DefaultConfig()
	apiCfg.Address = fmt.Sprintf("%s:%d", cfg.Msr.Host, cfg.Msr.Port)
	if c, err := api.NewClient(apiCfg); err != nil {
		log.Error("connect consul error", err)
	} else {
		r.consul = c
	}
	return r
}

func (r *Register) Register(service map[string]string) {
	reg := new(api.AgentServiceRegistration)
	reg.ID = fmt.Sprintf("%s:%s:%d", r.config.WorkerType, r.config.Host, r.config.Port)
	if global.RuntimeTest {
		reg.Name = "TEST"
	} else {
		reg.Name = r.config.WorkerType
	}
	reg.Port = r.config.Port
	for k := range service {
		reg.Tags = append(reg.Tags, k)
	}
	reg.Address = r.config.Host
	if r.config.Msr.Check.Enabled {
		if r.config.Msr.Check.RemoveTime < 1 {
			r.config.Msr.Check.RemoveTime = 1
		}
		reg.Check = new(api.AgentServiceCheck)
		reg.Check.DeregisterCriticalServiceAfter = fmt.Sprintf("%dm", r.config.Msr.Check.RemoveTime)
		reg.Check.Timeout = fmt.Sprintf("%ds", r.config.Msr.Check.Timeout)
		reg.Check.Interval = fmt.Sprintf("%ds", r.config.Msr.Check.Interval)
		reg.Check.TCP = fmt.Sprintf("%s:%d", r.config.Host, r.config.Port)
	}
	r.consul.KV().Put(&api.KVPair{
		Key:   "node/platform/" + reg.ID,
		Value: []byte("Go"),
	}, nil)
	if err := r.consul.Agent().ServiceRegister(reg); err != nil {
		log.Errorf("global service %s failed. host is %s:%d。", reg.ID, r.config.Msr.Host, r.config.Msr.Port, err)
	} else {
		log.Debugf("global service %s success.tag is %v", reg.ID, reg.Tags)
	}
}
func (r *Register) DisRegister() {
	id := fmt.Sprintf("%s:%s:%d", r.config.WorkerType, r.config.Host, r.config.Port)
	if err := r.consul.Agent().ServiceDeregister(id); err != nil {
		log.Warnf("disregister service %s failed", id, err)
	} else {
		log.Debugf("disregister service %s success", id)
	}
}
