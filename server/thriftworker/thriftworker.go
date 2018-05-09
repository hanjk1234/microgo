package thriftworker

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/seefan/microgo/server/common"
	"github.com/seefan/microgo/global"
	"github.com/seefan/microgo/server/worker"
	"time"
)

type thriftWorker struct {
	serviceManager *worker.ServiceManager
	register       *worker.RegisterManager
	mProcessor     *MultiplexedProcessor

	config            *common.Config
	permissionManager *worker.PermissionManager
	//public
	TransportFactory thrift.TTransportFactory
	ProtocolFactory  thrift.TProtocolFactory
	isRun            bool
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
		t.register = worker.NewRegisterManager(t.config)
	}
	processor := make(map[string]thrift.TProcessor)
	t.serviceManager.Init(t.config, func(name string) error {
		if p, err := createThriftProcessor(name); err == nil {
			processor[name] = p
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
func (t *thriftWorker) AppendPermissionCheck(p worker.PermissionCheck) {
	t.permissionManager.Append(p)
}
func (t *thriftWorker) registerService() {
	if t.register != nil {
		go func() {
			for {
				<-time.After(time.Second)
				if t.isRun {
					var sns []string
					for sn := range t.serviceManager.OnlineService {
						sns = append(sns, sn)
					}
					t.register.RegisterService(sns)
					break
				}
			}
		}()
	}
}
