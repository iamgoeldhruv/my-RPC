package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Request struct {
	ID     uint64          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type Response struct {
	ID     uint64          `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (r Request) Validate() error {
	if strings.TrimSpace(r.Method) == "" {
		return errors.New("request method cannot be empty")
	}

	dot := strings.LastIndex(r.Method, ".")
	if dot <= 0 || dot == len(r.Method)-1 {
		return fmt.Errorf("invalid method name %q: expected Service.Method", r.Method)
	}

	if strings.TrimSpace(r.Method[:dot]) == "" ||
		strings.TrimSpace(r.Method[dot+1:]) == "" {
		return fmt.Errorf("invalid method name %q", r.Method)
	}

	params := bytes.TrimSpace(r.Params)

	if len(params) == 0 {
		return errors.New("request params cannot be empty")
	}

	if !json.Valid(params) {
		return errors.New("request params contain invalid JSON")
	}

	if params[0] != '{' {
		return errors.New("request params must be a JSON object")
	}

	return nil
}

func (r Response) Validate() error {
	hasResult := len(r.Result) > 0
	hasError := r.Error != nil

	if hasResult == hasError {
		return errors.New("response must contain exactly one of result or error")
	}

	if hasResult && !json.Valid(r.Result) {
		return errors.New("response result contains invalid JSON")
	}

	if hasError {
		if r.Error.Code == 0 {
			return errors.New("response error code cannot be zero")
		}

		if strings.TrimSpace(r.Error.Message) == "" {
			return errors.New("response error message cannot be empty")
		}
	}

	return nil
}

func DecodeRequest(data []byte) (Request, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return Request{}, fmt.Errorf("decode RPC request fields: %w", err)
	}
	_, ok := fields["id"]
	if !ok {
		return Request{}, errors.New("request is missing id")
	}

	var req Request

	if err := json.Unmarshal(data, &req); err != nil {
		return Request{}, fmt.Errorf("decode RPC request: %w", err)
	}

	if err := req.Validate(); err != nil {
		return Request{}, fmt.Errorf("validate RPC request: %w", err)
	}

	return req, nil
}

func EncodeRequest(req Request) ([]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validate RPC request: %w", err)
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode RPC request: %w", err)
	}

	return data, nil
}

func EncodeResponse(resp Response) ([]byte, error) {
	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("validate RPC response: %w", err)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("encode RPC response: %w", err)
	}

	return data, nil
}

func NewResultResponse(id uint64, result any) (Response, error) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return Response{}, fmt.Errorf("marshal RPC result: %w", err)
	}

	resp := Response{
		ID:     id,
		Result: resultJSON,
	}

	if err := resp.Validate(); err != nil {
		return Response{}, err
	}

	return resp, nil
}

func NewErrorResponse(id uint64, code int, message string) (Response, error) {
	resp := Response{
		ID: id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}

	if err := resp.Validate(); err != nil {
		return Response{}, err
	}

	return resp, nil
}
