/*
@Time : 2018/5/9 15:38 
@Author : seefan
@File : http_processor
@Software: microgo
*/
package thriftworker

import (
	"time"
	"fmt"
	"github.com/seefan/microgo/server/worker"
	"git.apache.org/thrift.git/lib/go/thrift"
	"context"
	log "github.com/cihub/seelog"
	"github.com/golangteam/function/errors"
)

type HttpProcessor struct {
	processor thrift.TProcessor
	auth      func(string, string) int
	urls      []string
}

func newHttpProcessor(url []string, auth func(string, string) int) *HttpProcessor {
	return &HttpProcessor{
		auth: auth,
		urls: url,
	}
}
func (h *HttpProcessor) getProcessorName() (string, error) {
	if len(h.urls) < 3 {
		return "", errors.New("protocol error")
	} else {
		return h.urls[1], nil
	}
}

func (h *HttpProcessor) Process(ctx context.Context, in, out thrift.TProtocol) (bool, thrift.TException) {

	name, typeId, seqid, err := in.ReadMessageBegin()
	if err != nil {
		return false, err
	}
	if typeId != thrift.CALL && typeId != thrift.ONEWAY {
		return false, fmt.Errorf("Unexpected message type %v", typeId)
	}

	if len(h.urls) < 3 {
		log.Warnf("service name not found in message name: %s.  Did you forget to use a TMultiplexProtocol in your client?", name)
		return h.processFailed(ctx, in, out, name, seqid, fmt.Sprintf("%s not found", name), worker.NOT_SERVICE)
	}
	serviceName := h.urls[1]
	token := h.urls[2]
	if h.auth != nil {
		code := h.auth(serviceName, token)
		if code != worker.SUCCESS {
			log.Warnf("AuthFailed %s %s %d", serviceName, token, code)
			return h.processFailed(ctx, in, out, name, seqid, "Authentication failed", int32(code))
		}
	}
	if h.processor == nil {
		log.Warnf("service name not found: %s.  Did you forget to call registerProcessor()?", serviceName)
		return h.processFailed(ctx, in, out, name, seqid, fmt.Sprintf("%s not found", serviceName), worker.NOT_SERVICE)
	}
	smb := newStoredMessageProtocol(in, name, typeId, seqid)
	return h.processMethod(ctx, h.processor, smb, out, serviceName, name)
}

func (h *HttpProcessor) processMethod(ctx context.Context, actualProcessor thrift.TProcessor, in, out thrift.TProtocol, serviceName, methodName string) (bool, thrift.TException) {
	now := time.Now()
	re, err := actualProcessor.Process(ctx, in, out)
	if err == nil {
		log.Debugf("MethodExecOk %s %s %v", serviceName, methodName, time.Since(now).Seconds())
	} else {
		log.Warnf("MethodExecErr %s %s %v", serviceName, methodName, time.Since(now).Seconds())
	}
	return re, err
}
func (h *HttpProcessor) processFailed(ctx context.Context, in, out thrift.TProtocol, name string, seqId int32, err string, code int32) (bool, thrift.TException) {
	in.Skip(thrift.STRUCT)
	in.ReadMessageEnd()
	x5 := thrift.NewTApplicationException(code, err)
	out.WriteMessageBegin(name, thrift.EXCEPTION, seqId)
	x5.Write(out)
	out.WriteMessageEnd()
	out.Flush()
	return false, x5
}
