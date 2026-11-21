package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDecodeRequest(t *testing.T) {
	data := []byte(`{
		"id": 1,
		"method": "Calculator.Add",
		"params": {
			"a": 2,
			"b": 3
		}
	}`)

	req, err := DecodeRequest(data)
	if err != nil {
		t.Fatalf("DecodeRequest() error = %v", err)
	}

	if req.ID != 1 {
		t.Errorf("ID = %d, want 1", req.ID)
	}

	if req.Method != "Calculator.Add" {
		t.Errorf("Method = %q, want Calculator.Add", req.Method)
	}

	var params struct {
		A int `json:"a"`
		B int `json:"b"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		t.Fatalf("decode params: %v", err)
	}

	if params.A != 2 || params.B != 3 {
		t.Errorf("params = %+v, want A=2 and B=3", params)
	}
}

func TestEncodeRequest(t *testing.T) {
	req := Request{
		ID:     1,
		Method: "Calculator.Add",
		Params: json.RawMessage(`{"a":2,"b":3}`),
	}

	data, err := EncodeRequest(req)
	if err != nil {
		t.Fatalf("EncodeRequest() error = %v", err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.ID != req.ID {
		t.Errorf("ID = %d, want %d", decoded.ID, req.ID)
	}

	if decoded.Method != req.Method {
		t.Errorf("Method = %q, want %q", decoded.Method, req.Method)
	}
}

func TestNewResultResponse(t *testing.T) {
	resp, err := NewResultResponse(1, 5)
	if err != nil {
		t.Fatalf("NewResultResponse() error = %v", err)
	}

	data, err := EncodeResponse(resp)
	if err != nil {
		t.Fatalf("EncodeResponse() error = %v", err)
	}

	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.ID != 1 {
		t.Errorf("ID = %d, want 1", decoded.ID)
	}

	if string(decoded.Result) != "5" {
		t.Errorf("Result = %s, want 5", decoded.Result)
	}

	if decoded.Error != nil {
		t.Errorf("Error = %+v, want nil", decoded.Error)
	}
}

func TestNewErrorResponse(t *testing.T) {
	resp, err := NewErrorResponse(
		1,
		CodeMethodNotFound,
		"method not found",
	)
	if err != nil {
		t.Fatalf("NewErrorResponse() error = %v", err)
	}

	data, err := EncodeResponse(resp)
	if err != nil {
		t.Fatalf("EncodeResponse() error = %v", err)
	}

	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Error == nil {
		t.Fatal("expected an error response")
	}

	if decoded.Error.Code != CodeMethodNotFound {
		t.Errorf("Error.Code = %d, want %d",
			decoded.Error.Code, CodeMethodNotFound)
	}

	if decoded.Result != nil {
		t.Errorf("Result = %s, want nil", decoded.Result)
	}
}

func TestResultZeroIsNotOmitted(t *testing.T) {
	resp, err := NewResultResponse(1, 0)
	if err != nil {
		t.Fatalf("NewResultResponse() error = %v", err)
	}

	data, err := EncodeResponse(resp)
	if err != nil {
		t.Fatalf("EncodeResponse() error = %v", err)
	}

	if !strings.Contains(string(data), `"result":0`) {
		t.Errorf("response %s does not contain result 0", data)
	}
}

func TestRequestValidation(t *testing.T) {
	tests := []struct {
		name string
		req  Request
	}{
		{
			name: "empty method",
			req: Request{
				ID:     1,
				Method: "",
				Params: json.RawMessage(`{}`),
			},
		},
		{
			name: "method without service",
			req: Request{
				ID:     1,
				Method: ".Add",
				Params: json.RawMessage(`{}`),
			},
		},
		{
			name: "method without method name",
			req: Request{
				ID:     1,
				Method: "Calculator.",
				Params: json.RawMessage(`{}`),
			},
		},
		{
			name: "missing params",
			req: Request{
				ID:     1,
				Method: "Calculator.Add",
			},
		},
		{
			name: "invalid JSON params",
			req: Request{
				ID:     1,
				Method: "Calculator.Add",
				Params: json.RawMessage(`{invalid}`),
			},
		},
		{
			name: "params must be object",
			req: Request{
				ID:     1,
				Method: "Calculator.Add",
				Params: json.RawMessage(`[1,2,3]`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestResponseRejectsResultAndErrorTogether(t *testing.T) {
	resp := Response{
		ID:     1,
		Result: json.RawMessage(`5`),
		Error: &RPCError{
			Code:    CodeInternalError,
			Message: "failure",
		},
	}

	if err := resp.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestResponseRejectsNeitherResultNorError(t *testing.T) {
	resp := Response{
		ID: 1,
	}

	if err := resp.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestNilResultIsJSONNull(t *testing.T) {
	resp, err := NewResultResponse(1, nil)
	if err != nil {
		t.Fatalf("NewResultResponse() error = %v", err)
	}

	data, err := EncodeResponse(resp)
	if err != nil {
		t.Fatalf("EncodeResponse() error = %v", err)
	}

	if !strings.Contains(string(data), `"result":null`) {
		t.Errorf("response %s does not contain result null", data)
	}
}
func TestDecodeRequestRejectsMissingID(t *testing.T) {
	data := []byte(`{
        "method": "Calculator.Add",
        "params": {"a": 2, "b": 3}
    }`)

	_, err := DecodeRequest(data)
	if err == nil {
		t.Fatal("expected error for missing request ID")
	}
}
