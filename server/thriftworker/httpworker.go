package thriftworker

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"net/http"
	log "github.com/cihub/seelog"
)

type HttpWorker struct {
	thriftWorker
	server *http.Server
}

func NewHttpWorker() *HttpWorker {
	return &HttpWorker{
		thriftWorker: newThriftWorker(),
	}
}

func (t *HttpWorker) Start() (err error) {
	if err = t.init(); err != nil {
		return
	}
	addr := fmt.Sprintf("%s:%d", t.config.Host, t.config.Port)
	t.server = &http.Server{Addr: addr}
	log.Debugf("host on %s:%d", t.config.Host, t.config.Port)
	handler := thrift.NewThriftHandlerFunc(t.mProcessor, t.ProtocolFactory, t.ProtocolFactory)

	http.HandleFunc("/", handler)
	if t.register != nil {
		t.register.Register(t.serviceManager.OnlineService)
	}
	//return t.server.Serve()
	return t.server.ListenAndServe()
}
func (t *HttpWorker) Stop() error {
	if t.register != nil {
		t.register.DisRegister()
	}
	if t.server != nil {
		return t.server.Close()
	}
	return nil
}