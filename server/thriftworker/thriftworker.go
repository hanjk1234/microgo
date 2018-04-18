package thriftworker

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/seefan/microgo/server/common"
	"time"

	"github.com/seefan/microgo/global"
	"github.com/seefan/microgo/server/worker"
)

type ThriftWorker struct {
	serviceManager    *worker.ServiceManager
	register          *worker.Register
	mProcessor        *MultiplexedProcessor
	server            thrift.TServer
	config            *common.Config
	permissionManager *worker.PermissionManager
	//public
	TransportFactory thrift.TTransportFactory
	ProtocolFactory  thrift.TProtocolFactory
}

func NewThriftWorker() *ThriftWorker {
	return &ThriftWorker{
		config:            common.NewConfig(),
		mProcessor:        NewMultiplexedProcessor(),
		serviceManager:    worker.NewServiceManager(),
		permissionManager: worker.NewPermissionManager(),
		TransportFactory:  thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory()),
		ProtocolFactory:   thrift.NewTBinaryProtocolFactoryDefault(),
	}
}

func (t *ThriftWorker) Start() error {
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
	socket, err := thrift.NewTServerSocketTimeout(fmt.Sprintf("%s:%d", t.config.Host, t.config.Port), time.Duration(t.config.Timeout)*time.Second)
	if err != nil {
		return err
	}

	t.server = thrift.NewTSimpleServer4(t.mProcessor, socket, t.TransportFactory, t.ProtocolFactory)
	if t.register != nil {
		t.register.Register(t.serviceManager.OnlineService)
	}
	return t.server.Serve()
}
func (t *ThriftWorker) Stop() error {
	if t.register != nil {
		t.register.DisRegister()
	}
	if t.server != nil {
		return t.server.Stop()
	}
	return nil
}
