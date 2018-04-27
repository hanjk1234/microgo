package thriftworker

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/seefan/microgo/server/common"
	"github.com/seefan/microgo/global"
	"github.com/seefan/microgo/server/worker"
)

type thriftWorker struct {
	serviceManager *worker.ServiceManager
	register       *worker.Register
	mProcessor     *MultiplexedProcessor

	config            *common.Config
	permissionManager *worker.PermissionManager
	//public
	TransportFactory thrift.TTransportFactory
	ProtocolFactory  thrift.TProtocolFactory
}

func newThriftWorker() thriftWorker {
	return thriftWorker{
		config:            common.NewConfig(),
		mProcessor:        newMultiplexedProcessor(),
		serviceManager:    worker.NewServiceManager(),
		permissionManager: worker.NewPermissionManager(),
		TransportFactory:  thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory()),
		ProtocolFactory:   thrift.NewTBinaryProtocolFactoryDefault(),
	}
}
func (t *thriftWorker) init() error {
	t.config.LoadJson(global.RuntimeRoot + "/conf.json")

	if t.config.Msr.Enabled {
		t.register = worker.NewRegister(t.config)
	}
	processor := make(map[string]thrift.TProcessor)
	t.serviceManager.Init(t.config, func(id, name string) error {
		if p, err := createThriftProcessor(name); err == nil {
			processor[id] = p
			return nil
		} else {
			return err
		}
	})
	if len(processor) == 0 {
		return fmt.Errorf("empty service")
	}
	t.permissionManager.Init(t.config, t.serviceManager.OnlineService)
	for id, processor := range processor {
		t.mProcessor.RegisterProcessor(id, processor)
	}
	return nil
}

func (t *thriftWorker) RegisterThriftProcessor(name string, proc func() thrift.TProcessor) {
	thriftServiceProcessor[name] = proc
}

func (t *thriftWorker) RegisterServiceId(serviceName, id string) {
	global.ServiceId[serviceName] = id
}
