package thriftworker

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/seefan/microgo/server/common"
	"time"

	"github.com/seefan/microgo/global"
)

type ThriftWorker struct {
	serviceManager *ServiceManager
	register       *Register
	mProcessor     *MultiplexedProcessor
	server         thrift.TServer
	config         *common.Config
}

func (t *ThriftWorker) Init() {
	t.config = common.NewConfig()
	t.config.LoadJson(global.RuntimeRoot + "/conf.json")
	t.serviceManager = NewServiceManager(t.config)

	if t.config.Msr.Enabled {
		t.register = NewRegister(t.config)
	}
}
func (t *ThriftWorker) Start() error {
	t.Init()
	t.serviceManager.Init()
	if len(t.serviceManager.Processor) == 0 {
		//return fmt.Errorf("empty service")
	}
	t.mProcessor = NewMultiplexedProcessor()
	for id, processor := range t.serviceManager.Processor {
		t.mProcessor.RegisterProcessor(id, processor)
	}
	socket, err := thrift.NewTServerSocketTimeout(fmt.Sprintf("%s:%d", t.config.Host, t.config.Port), time.Duration(t.config.Timeout)*time.Second)
	if err != nil {
		return err
	}
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	t.server = thrift.NewTSimpleServer4(t.mProcessor, socket, transportFactory, protocolFactory)
	if t.register != nil {
		t.register.Register(t.serviceManager.Service)
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
