package worker

import (
	log "github.com/cihub/seelog"
	"github.com/seefan/microgo/server/common"
)

type ServiceManager struct {
	config        *common.Config
	OnlineService map[string]interface{}
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		OnlineService: make(map[string]interface{}),
	}
}
func (s *ServiceManager) Init(cfg *common.Config, f func(name string) error) {
	s.config = cfg
	//check all onlineService
	for _, name := range s.config.Service {
			if err := f(name); err == nil {
				s.OnlineService[name] = new(struct{})
			} else {
				log.Errorf("load onlineService %s error", name)
			}
	}

	log.Debug("onlineService config is load ", s.OnlineService)
}