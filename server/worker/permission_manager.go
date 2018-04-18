package worker

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/hashicorp/consul/api"
	"github.com/seefan/microgo/server/common"
	"strings"
	"sync"
	"time"
)

var empty = struct{}{}
//permisson manager

type PermissionManager struct {
	config     *common.Config
	admin      string
	consul     *api.Client
	group      map[string]map[string]interface{}
	service    map[string]string
	groupIndex *api.QueryOptions
	lock       sync.RWMutex
}

func NewPermissionManager() *PermissionManager {
	return &PermissionManager{
		//key->service id
		group: make(map[string]map[string]interface{}),
		groupIndex: &api.QueryOptions{
			WaitIndex: 0,
			WaitTime:  60 * time.Second,
		},
	}
}

func (p *PermissionManager) Init(cfg *common.Config, onlineService map[string]string) {
	p.config = cfg
	p.service = onlineService
	if p.config.Msr.Enabled {
		cfg := api.DefaultConfig()
		cfg.Address = fmt.Sprintf("%s:%d", p.config.Msr.Host, p.config.Msr.Port)
		if c, err := api.NewClient(cfg); err != nil {
			log.Error("connect consul error", err)
		} else {
			p.consul = c
			//加载超级管理员
			if kp, _, err := p.consul.KV().Get("config/admin", nil); err == nil && kp != nil {
				p.admin = string(kp.Value)
			}
			p.loadGroup()
			p.watch()
		}
	}
	if p.config.Debug {
		p.admin = "test"
	}
}
func (p *PermissionManager) watch() {
	go func() {
		for {
			if err := p.loadGroup(); err != nil {
				time.Sleep(time.Minute)
			}
		}
	}()
}

func (p *PermissionManager) loadGroup() error {
	//加载集群权限
	if ks, m, err := p.consul.KV().List("group", p.groupIndex); err == nil {
		if ks != nil {
			p.lock.Lock()
			defer p.lock.Unlock()
			for _, k := range ks {
				name := string(k.Key[6:])
				p.group[name] = make(map[string]interface{})
				ss := strings.Split(string(k.Value), ",")
				for _, u := range ss {
					p.group[name][u] = empty
				}
			}
			log.Debug("load group is ", p.group)
		}
		p.groupIndex.WaitIndex = m.LastIndex
		return nil
	} else {
		return err
	}
}
func (p *PermissionManager) Auth(sid, key string) int {
	//0000为shared service
	if sid == SDK_SERVICE {
		if _, ok := p.group[key]; ok {
			return SUCCESS
		}
	}
	p.lock.RLock()
	defer p.lock.RUnlock()
	////检查实例是否提供了服务
	if _, ok := p.service[sid]; !ok {
		log.Warn("auth failed 9001 ", sid, key)
		log.Debug("auth failed 9001 ", sid, key, p.service)
		return NOT_SERVICE
	}
	if key != p.admin { //如果是admin_key，不限制权限组
		if _, ok := p.group[key]; !ok {
			log.Warn("auth failed 9002,key not found", sid)
			log.Debug("auth failed 9002,key not found", sid, p.group)
			return NOT_USER
		} else if _, ok := p.group[key][sid]; !ok {
			log.Warn("auth failed 9002,service not found", sid, key)
			log.Debug("auth failed 9002,service not found", sid, key, p.group[sid])
			return AUTH_FAILED
		}
	}

	return SUCCESS
}
