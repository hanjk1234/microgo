package worker

import (
	"github.com/seefan/microgo/server/common"
	"github.com/seefan/microgo/global"
)

type Register interface {
	RegisterService(name, address string, port int, serviceName []string)
	RemoveService(name, address string, port int)
}
type RegisterManager struct {
	regs   []Register
	config *common.Config
}

func NewRegisterManager(cfg *common.Config) *RegisterManager {
	return &RegisterManager{
		config: cfg,
	}
}
func (r *RegisterManager) RegisterService(service []string) {
	var name string
	if global.RuntimeTest {
		name = "TEST"
	} else {
		name = r.config.WorkerType
	}
	for _, s := range r.regs {
		s.RegisterService(name, r.config.Host, r.config.Port, service)
	}
}
func (r *RegisterManager) RemoveService() {
	var name string
	if global.RuntimeTest {
		name = "TEST"
	} else {
		name = r.config.WorkerType
	}
	for _, s := range r.regs {
		s.RemoveService(name, r.config.Host, r.config.Port)
	}
}
