package rpcserver

import (
	"reflect"
	"sync"
	"encoding/json"
    "fmt"
    "io"
    "log/slog"
    "net"
    "runtime/debug"
    "strings"
	"my-rpc/internal/framing"
    "my-rpc/internal/protocol"
    "my-rpc/internal/transport/tcp"
)

type MethodType struct {
	method    reflect.Value
	argType   reflect.Type
	replyType reflect.Type
}

type Service struct {
	name    string
	methods map[string]*MethodType
}
type Server struct {
	mu       sync.RWMutex
	services map[string]*Service
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]*Service),
	}
}

var _ tcp.Handler = (*Server)(nil)

func (s *Server) Handle(conn net.Conn,logger *slog.Logger){
	if logger == nil {
        logger = slog.Default()
    }

    logger.Info("RPC connection handler started")
	defer func() {
        logger.Info("RPC connection handler stopped")
    }()
	payload,err:=framing.ReadFrame(conn)
	if err!=nil{
		if err != io.EOF {
                logger.Error(
                    "failed to read RPC frame",
                    "error", err,
                )
            }
		return 


	}
	
	response, err := s.processRequest(payload)
        if err != nil {
            logger.Error(
                "failed to process RPC request",
                "error", err,
            )
            return
        }
	responsePayload, err := protocol.EncodeResponse(&response)
	if err != nil {
            logger.Error(
                "failed to encode RPC response",
                "error", err,
            )
            return
        }
	 if err := framing.WriteFrame(conn, responsePayload); err != nil {
            logger.Error(
                "failed to write RPC response",
                "error", err,
            )
            return
        }

        logger.Debug(
            "RPC response sent",
            "request_id", response.ID,
        )
}

func (s *Server) processRequest(
    payload []byte,
) (protocol.Response, error){
	req, err := protocol.DecodeRequest(payload)
    if err != nil {
        return newRequestError(payload, err)
    }
	serviceName, methodName, err := splitMethodName(req.Method)
    if err != nil {
        return newErrorResponse(
            req.ID,
            protocol.CodeInvalidRequest,
            "invalid method name",
        )
    }
	method, ok := s.lookupMethod(serviceName, methodName)
    if !ok {
        return newErrorResponse(
            req.ID,
            protocol.CodeMethodNotFound,
            "service or method not found",
        )
	}
	argument, err := decodeArgument(
        req.Params,
        method.argType,
    )
	 if err != nil {
        return newErrorResponse(
            req.ID,
            protocol.CodeInvalidRequest,
            "invalid method parameters",
        )
    }
	result, err := invokeMethod(method, argument)
    if err != nil {
        return newErrorResponse(
            req.ID,
            protocol.CodeInternalError,
            "method execution failed",
        )
    }
	response, err := protocol.NewResultResponse(
        req.ID,
        result,
    )
    if err != nil {
        return newErrorResponse(
            req.ID,
            protocol.CodeInternalError,
            "failed to encode method result",
        )
    }

    return response, nil

}

func splitMethodName(
    fullMethodName string,
) (string, string, error) {

    serviceName, methodName, found := strings.Cut(
        fullMethodName,
        ".",
    )

    if !found ||
        serviceName == "" ||
        methodName == "" ||
        strings.Contains(methodName, ".") {

        return "", "", fmt.Errorf(
            "invalid RPC method name %q",
            fullMethodName,
        )
    }

    return serviceName, methodName, nil
}

func decodeArgument(
    params json.RawMessage,
    argType reflect.Type,
) (reflect.Value, error) {

    argument := reflect.New(argType)

    if err := json.Unmarshal(params, argument.Interface()); err != nil {
        return reflect.Value{}, fmt.Errorf(
            "decode RPC method argument: %w",
            err,
        )
    }

    return argument.Elem(), nil
}


func invokeMethod(
    method *MethodType,
    argument reflect.Value,
) (result any, err error) {
	defer func() {
        if recovered := recover(); recovered != nil {
            err = fmt.Errorf(
                "RPC method panic: %v\n%s",
                recovered,
                debug.Stack(),
            )

            result = nil
        }
    }()
	outputs := method.method.Call(
        []reflect.Value{argument},
    )
	result = outputs[0].Interface()
	 if !outputs[1].IsNil() {
        return nil, outputs[1].Interface().(error)
    }

    return result, nil
}

func newErrorResponse(
    id uint64,
    code int,
    message string,
) (protocol.Response, error) {

    response, err := protocol.NewErrorResponse(
        id,
        code,
        message,
    )

    if err != nil {
        return protocol.Response{}, fmt.Errorf(
            "create RPC error response: %w",
            err,
        )
    }

    return response, nil
}

func newRequestError(
    payload []byte,
    cause error,
) (protocol.Response, error) {

    // Best-effort extraction of the request ID.
    var envelope struct {
        ID *uint64 `json:"id"`
    }

    var id uint64

    if err := json.Unmarshal(payload, &envelope); err == nil &&
        envelope.ID != nil {
        id = *envelope.ID
    }

    return newErrorResponse(
        id,
        protocol.CodeInvalidRequest,
        "invalid RPC request",
    )
}