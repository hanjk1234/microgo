package thriftworker

import (
	"encoding/json"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/cihub/seelog"
	"github.com/go-ini/ini"
	"github.com/hashicorp/consul/api"
	"github.com/seefan/microgo/global"
	"github.com/seefan/microgo/server/common"
	"io/ioutil"
)

type ServiceManager struct {
	config    *common.Config
	consul    *api.Client
	Processor map[string]thrift.TProcessor
	Service   map[string]string
}

func NewServiceManager(cfg *common.Config) *ServiceManager {
	return &ServiceManager{
		config:    cfg,
		Processor: make(map[string]thrift.TProcessor),
		Service:   make(map[string]string),
	}
}
func (s *ServiceManager) Init() {
	idService := make(map[string]string)
	if s.config.Msr.Enabled {
		cfg := api.DefaultConfig()
		cfg.Address = fmt.Sprintf("%s:%d", s.config.Msr.Host, s.config.Msr.Port)
		if c, err := api.NewClient(cfg); err != nil {
			log.Error("connect consul error", err)
		} else {
			s.consul = c
		}
	}
	serviceAll := s.getAllService()
	//id=>name
	for k, v := range serviceAll {
		idService[v] = k
	}

	//save id=>service
	if bs, err := json.MarshalIndent(serviceAll, "", "  "); err == nil {
		ioutil.WriteFile(common.Path(global.RuntimeRoot, "service_id.json"), bs, 0764)
	}

	serviceAll = make(map[string]string)
	//load all service
	for _, id := range s.config.Service {
		if name, ok := idService[id]; ok {
			if s.loadService(id, name) {
				s.Service[id] = name
			}
		}
	}

	log.Debug("service is load ", s.Service)
}
func (s *ServiceManager) getAllService() map[string]string {
	nameId := make(map[string]string)

	// cache
	if !common.NotExist(common.Path(global.RuntimeRoot, "service_id.json")) {
		if bs, err := ioutil.ReadFile(common.Path(global.RuntimeRoot, "service_id.json")); err == nil {
			if err := json.Unmarshal(bs, &nameId); err != nil {
				log.Errorf("service_id.json format error")
			}
		}
	}
	//merge to program register service id
	nameId = global.MergeServiceId(nameId)
	//load config from consul
	if s.consul != nil {
		if ks, _, err := s.consul.KV().List("service", nil); err == nil && ks != nil {
			for _, k := range ks {
				nameId[string(k.Value)] = string(k.Key[8:])
			}
		}
	}

	//load config form file
	if file, err := ini.Load(common.Path(global.RuntimeRoot, "serviceId.properties")); err == nil {
		for _, k := range file.Section("").KeyStrings() {
			if _, ok := nameId[k]; !ok {
				nameId[k] = file.Section("").Key(k).String()
			}
		}
	}
	return nameId
}
func (s *ServiceManager) loadService(id, name string) bool {
	if p, err := global.CreateThriftProcessor(name); err == nil {
		s.Processor[id] = p
		return true
	}
	return false
}
