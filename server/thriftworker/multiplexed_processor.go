package thriftworker

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/cihub/seelog"
	"strings"
	"time"
	"context"
	"github.com/seefan/microgo/server/worker"
)

type MultiplexedProcessor struct {
	serviceProcessorMap map[string]thrift.TProcessor
	DefaultProcessor    thrift.TProcessor
	auth                *worker.PermissionManager
}

func (t *MultiplexedProcessor) RegisterDefault(processor thrift.TProcessor) {
	t.DefaultProcessor = processor
}

func (t *MultiplexedProcessor) RegisterProcessor(name string, processor thrift.TProcessor) {
	if t.serviceProcessorMap == nil {
		t.serviceProcessorMap = make(map[string]thrift.TProcessor)
	}
	t.serviceProcessorMap[name] = processor
}

func newMultiplexedProcessor() *MultiplexedProcessor {
	return &MultiplexedProcessor{
		serviceProcessorMap: make(map[string]thrift.TProcessor),
	}
}
func (t *MultiplexedProcessor) Process(ctx context.Context, in, out thrift.TProtocol) (bool, thrift.TException) {

	name, typeId, seqid, err := in.ReadMessageBegin()
	if err != nil {
		return false, err
	}
	if typeId != thrift.CALL && typeId != thrift.ONEWAY {
		return false, fmt.Errorf("Unexpected message type %v", typeId)
	}
	//extract the service name
	v := strings.SplitN(name, thrift.MULTIPLEXED_SEPARATOR, 3)
	if len(v) != 3 || v[0] == "" || v[1] == "" || v[2] == "" {
		if t.DefaultProcessor != nil {
			smb := thrift.NewStoredMessageProtocol(in, name, typeId, seqid)
			return t.DefaultProcessor.Process(ctx, smb, out)
		}
		log.Warnf("service name not found in message name: %s.  Did you forget to use a TMultiplexProtocol in your client?", name)
		return t.processFailed(ctx, in, out, name, seqid, fmt.Sprintf("%s not found", name), worker.NOT_SERVICE)
	}
	if t.auth != nil {
		code := t.auth.Auth(v[0], v[1])
		if code != worker.SUCCESS {
			log.Warnf("AuthFailed %s %s %d", v[0], v[1], code)
			return t.processFailed(ctx, in, out, v[2], seqid, "Authentication failed", int32(code))
		}
	}
	actualProcessor, ok := t.serviceProcessorMap[v[0]]
	if !ok {
		log.Warnf("service name not found: %s.  Did you forget to call registerProcessor()?", v[0])
		return t.processFailed(ctx, in, out, v[2], seqid, fmt.Sprintf("%s not found", v[0]), worker.NOT_SERVICE)
	}
	smb := newStoredMessageProtocol(in, v[2], typeId, seqid)

	return t.processMethod(ctx, actualProcessor, smb, out, v[0], v[2])
}

func (t *MultiplexedProcessor) processMethod(ctx context.Context, actualProcessor thrift.TProcessor, smb *storedMessageProtocol, out thrift.TProtocol,
	serviceName, methodName string) (bool, thrift.TException) {
	now := time.Now()
	re, err := actualProcessor.Process(ctx, smb, out)
	if err == nil {
		log.Debugf("MethodExecOk %s %s %v", serviceName, methodName, time.Since(now).Seconds())
	} else {
		log.Warnf("MethodExecErr %s %s %v", serviceName, methodName, time.Since(now).Seconds())
	}
	return re, err
}
func (t *MultiplexedProcessor) processFailed(ctx context.Context, in, out thrift.TProtocol, name string, seqId int32, err string, code int32) (bool, thrift.TException) {
	in.Skip(thrift.STRUCT)
	in.ReadMessageEnd()
	x5 := thrift.NewTApplicationException(code, err)
	out.WriteMessageBegin(name, thrift.EXCEPTION, seqId)
	x5.Write(out)
	out.WriteMessageEnd()
	out.Flush()
	return false, x5
}

//Protocol that use stored message for ReadMessageBegin
type storedMessageProtocol struct {
	thrift.TProtocol
	name   string
	typeId thrift.TMessageType
	seqid  int32
}

func newStoredMessageProtocol(protocol thrift.TProtocol, name string, typeId thrift.TMessageType, seqid int32) *storedMessageProtocol {
	return &storedMessageProtocol{protocol, name, typeId, seqid}
}

func (s *storedMessageProtocol) ReadMessageBegin() (name string, typeId thrift.TMessageType, seqid int32, err error) {
	return s.name, s.typeId, s.seqid, nil
}
