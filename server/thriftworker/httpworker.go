package thriftworker

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"net/http"
	log "github.com/cihub/seelog"
	"context"
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

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Access-Control-Allow-Origin", "*")
		writer.Header().Add("Access-Control-Allow-Headers", "content-type")
		handler(writer, request)
	})
	t.registerService()
	t.isRun = true
	err = t.server.ListenAndServe()
	if err != nil {
		t.isRun = false
	}
	return err
}
func (t *HttpWorker) Stop() error {
	if t.register != nil {
		t.register.UnRegister()
	}
	if t.server != nil {
		return t.server.Shutdown(context.Background())
	}
	return nil
}
