package thriftworker

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"net/http"
	log "github.com/cihub/seelog"
	"context"
	"strings"
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

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Access-Control-Allow-Origin", "*")
		writer.Header().Add("Access-Control-Allow-Headers", "content-type")
		hp := newHttpProcessor(strings.Split(request.RequestURI, "/"), t.mProcessor.auth)
		handler := thrift.NewThriftHandlerFunc(hp, t.ProtocolFactory, t.ProtocolFactory)
		var pr thrift.TProcessor
		if serviceName, err := hp.getProcessorName(); err == nil {
			if pr, err = t.mProcessor.getProcessor(serviceName); err == nil {
				hp.processor = pr
			}
		}
		handler(writer,request)
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
	if t.config.Msr.Enabled {
		t.register.RemoveService()
	}
	if t.server != nil {
		return t.server.Shutdown(context.Background())
	}
	return nil
}
