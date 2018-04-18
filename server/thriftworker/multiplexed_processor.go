package thriftworker

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/cihub/seelog"
	"strings"
	"time"
	"context"
)

type MultiplexedProcessor struct {
	thrift.TMultiplexedProcessor
	serviceProcessorMap map[string]thrift.TProcessor
}

func NewMultiplexedProcessor() *MultiplexedProcessor {
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
	v := strings.SplitN(name, thrift.MULTIPLEXED_SEPARATOR, 4)
	if len(v) != 4 || v[0] == "" || v[1] == "" || v[2] == "" || v[3] == "" {
		if t.DefaultProcessor != nil {
			smb := thrift.NewStoredMessageProtocol(in, name, typeId, seqid)
			return t.DefaultProcessor.Process(ctx, smb, out)
		}
		log.Warnf("service name not found in message name: %s.  Did you forget to use a TMultiplexProtocol in your client?", name)
		return t.processFailed(ctx, in, out, v[3], seqid, fmt.Sprintf("%s not found", v[0]), NOT_SERVICE)
	}
	//if t.Auth != nil {
	//	code := t.Auth.Auth(v[0], v[1], v[2])
	//	if code != SUCCESS {
	//		log.Warnf("AuthFailed %s %s %d", v[1], v[2], code)
	//		return t.processFailed(in, out, v[3], seqid, "Authentication failed", int32(code))
	//	}
	//}
	actualProcessor, ok := t.serviceProcessorMap[v[0]]
	if !ok {
		log.Warnf("service name not found: %s.  Did you forget to call registerProcessor()?", v[0])
		return t.processFailed(ctx, in, out, v[3], seqid, fmt.Sprintf("%s not found", v[0]), NOT_SERVICE)
	}
	smb := NewStoredMessageProtocol(in, v[3], typeId, seqid)

	return t.processMethod(ctx, actualProcessor, smb, out, v[0], v[3])
}

func (t *MultiplexedProcessor) processMethod(ctx context.Context, actualProcessor thrift.TProcessor, smb *storedMessageProtocol, out thrift.TProtocol,
	serviceName, methodName string) (bool, thrift.TException) {
	now := time.Now()
	re, err := actualProcessor.Process(ctx, smb, out)
	if err == nil {
		log.Infof("MethodExecOk %s %s %v", serviceName, methodName, time.Since(now).Seconds())
	} else {
		log.Infof("MethodExecErr %s %s %v", serviceName, methodName, time.Since(now).Seconds())
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

func NewStoredMessageProtocol(protocol thrift.TProtocol, name string, typeId thrift.TMessageType, seqid int32) *storedMessageProtocol {
	return &storedMessageProtocol{protocol, name, typeId, seqid}
}

func (s *storedMessageProtocol) ReadMessageBegin() (name string, typeId thrift.TMessageType, seqid int32, err error) {
	return s.name, s.typeId, s.seqid, nil
}
