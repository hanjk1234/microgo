/**
* config/admin=test is administrator,can access to all services
* can use plug-ins to extend permission limits
*/
package worker

import (
	log "github.com/cihub/seelog"
	"github.com/seefan/microgo/server/common"
)

//extended check
type PermissionCheck interface {
	// check service name and token
	//
	// @param serviceName string  service name
	// @param token string user token
	// @return int 0 or 1000 for success and others for failure
	Check(serviceName, token string) int
}

//permission manager

type PermissionManager struct {
	config *common.Config
	admin  string
	//online service
	onlineService map[string]interface{}
	//plug check
	extCheck []PermissionCheck
}

func NewPermissionManager() *PermissionManager {
	return &PermissionManager{
		onlineService: make(map[string]interface{}),
	}
}

func (p *PermissionManager) Init(cfg *common.Config, onlineService map[string]interface{}) {
	p.config = cfg
	p.onlineService = onlineService
	if p.config.Debug {
		p.admin = "test"
	}
}

func (p *PermissionManager) Auth(sid, key string) int {
	//
	if _, ok := p.onlineService[sid]; !ok {
		log.Warn("auth failed  ", NOT_SERVICE, sid, key)
		return NOT_SERVICE
	}
	if key != p.admin {
		return AUTH_FAILED
	}
	if p.extCheck != nil {
		for _, pc := range p.extCheck {
			if r := pc.Check(sid, key); r != 0 && r != SUCCESS {
				return r
			}
		}
	}
	return SUCCESS
}
