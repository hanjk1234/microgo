/*
@Time : 2018/5/4 10:46 
@Author : seefan
@File : register
@Software: microgo
*/
package consul

import (
	"github.com/hashicorp/consul/api"
	"fmt"
	"github.com/seefan/microgo/server/common"
	log "github.com/cihub/seelog"
)

//service register
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

//register all service
func (r *Register) RegisterService(name, address string, port int, serviceName []string) {
	if r.consul == nil {
		return
	}
	reg := new(api.AgentServiceRegistration)
	reg.ID = fmt.Sprintf("%s:%s:%d", name, address, port)
	reg.Name = name
	reg.Port = port
	reg.Tags = append(reg.Tags, serviceName...)

	reg.Address = address
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

	if err := r.consul.Agent().ServiceRegister(reg); err != nil {
		log.Errorf("global onlineService %s failed. host is %s:%d。", reg.ID, address, port, err)
	} else {
		log.Debugf("global onlineService %s success.tag is %v", reg.ID, reg.Tags)
		//Languages for the implementation of reputation services
		r.consul.KV().Put(&api.KVPair{
			Key:   "node/platform/" + reg.ID,
			Value: []byte("Go"),
		}, nil)
	}
}

//unregister
func (r *Register) RemoveService(name, address string, port int) {
	if r.consul == nil {
		return
	}
	id := fmt.Sprintf("%s:%s:%d", name, address, port)
	if err := r.consul.Agent().ServiceDeregister(id); err != nil {
		log.Warnf("unregister onlineService %s failed", id, err)
	} else {
		log.Debugf("unregister onlineService %s success", id)
	}
}
