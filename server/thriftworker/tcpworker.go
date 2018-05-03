package thriftworker

import (
	"fmt"
	"time"
	log "github.com/cihub/seelog"
	"git.apache.org/thrift.git/lib/go/thrift"
)

type TcpWorker struct {
	thriftWorker

	server thrift.TServer
}

func NewTcpWorker() *TcpWorker {
	return &TcpWorker{
		thriftWorker: newThriftWorker(),
	}
}
func (t *TcpWorker) Start() (err error) {
	if err = t.init(); err != nil {
		return
	}
	var socket *thrift.TServerSocket
	if t.config.Timeout > 0 {
		if socket, err = thrift.NewTServerSocketTimeout(fmt.Sprintf("%s:%d", t.config.Host, t.config.Port), time.Duration(t.config.Timeout)*time.Second); err != nil {
			return err
		}
	} else {
		if socket, err = thrift.NewTServerSocket(fmt.Sprintf("%s:%d", t.config.Host, t.config.Port)); err != nil {
			return err
		}
	}

	log.Debugf("host on %s:%d", t.config.Host, t.config.Port)
	t.server = thrift.NewTSimpleServer4(t.mProcessor, socket, t.TransportFactory, t.ProtocolFactory)
	t.isRun = true
	t.registerService()
	err = t.server.Serve()
	if err != nil {
		t.isRun = false
	}
	return err
}

func (t *TcpWorker) Stop() error {
	if t.register != nil {
		t.register.UnRegister()
	}
	if t.server != nil {
		return t.server.Stop()
	}
	return nil
}
