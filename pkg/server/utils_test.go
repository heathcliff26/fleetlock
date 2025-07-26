package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heathcliff26/fleetlock/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestReadUserIP(t *testing.T) {
	tests := []struct {
		name         string
		headers      map[string]string
		remoteAddr   string
		expectedIP   string
	}{
		{
			name: "XRealIP",
			headers: map[string]string{
				"x-real-ip": "192.168.1.100",
			},
			remoteAddr: "10.0.0.1:8080",
			expectedIP: "192.168.1.100",
		},
		{
			name: "XForwardedFor",
			headers: map[string]string{
				"x-forwarded-for": "203.0.113.10",
			},
			remoteAddr: "10.0.0.1:8080",
			expectedIP: "203.0.113.10",
		},
		{
			name:       "RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "10.0.0.1:8080",
			expectedIP: "10.0.0.1:8080",
		},
		{
			name: "XRealIPPrecedence",
			headers: map[string]string{
				"x-real-ip":       "192.168.1.100",
				"x-forwarded-for": "203.0.113.10",
			},
			remoteAddr: "10.0.0.1:8080",
			expectedIP: "192.168.1.100",
		},
		{
			name: "XForwardedForFallback",
			headers: map[string]string{
				"x-forwarded-for": "203.0.113.10",
			},
			remoteAddr: "10.0.0.1:8080",
			expectedIP: "203.0.113.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			
			result := ReadUserIP(req)
			assert.Equal(t, tt.expectedIP, result)
		})
	}
}

func TestSendResponse(t *testing.T) {
	tests := []struct {
		name           string
		response       interface{}
		expectOutput   bool
		expectedStatus int
	}{
		{
			name: "ValidResponse",
			response: api.FleetLockResponse{
				Kind:  "success",
				Value: "test message",
			},
			expectOutput:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name: "HealthResponse",
			response: api.FleetlockHealthResponse{
				Status: "ok",
				Error:  "",
			},
			expectOutput:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "EmptyStruct",
			response:       struct{}{},
			expectOutput:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "NilResponse",
			response:       nil,
			expectOutput:   true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			
			sendResponse(rr, tt.response)
			
			if tt.expectOutput {
				assert.NotEmpty(t, rr.Body.String(), "Response body should not be empty")
				// Check that it's valid JSON
				body := rr.Body.String()
				if tt.response != nil {
					assert.Contains(t, body, "{", "Response should contain JSON object")
				} else {
					// nil response marshals to "null"
					assert.Contains(t, body, "null", "Nil response should marshal to null")
				}
			}
		})
	}
}

func TestSendResponseWithUnmarshalableData(t *testing.T) {
	// Test with data that cannot be marshaled to JSON
	rr := httptest.NewRecorder()
	
	// Functions cannot be marshaled to JSON
	unmarshalable := struct {
		Func func()
	}{
		Func: func() {},
	}
	
	sendResponse(rr, unmarshalable)
	
	// The function should handle the error gracefully and not write anything
	assert.Empty(t, rr.Body.String(), "Response body should be empty when marshaling fails")
}