package grpcclient

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandleGRPCError(t *testing.T) {
	testCases := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "InvalidArgument maps to 400",
			err:            status.Error(codes.InvalidArgument, "bad field"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthenticated maps to 401",
			err:            status.Error(codes.Unauthenticated, "bad token"),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "PermissionDenied maps to 403",
			err:            status.Error(codes.PermissionDenied, "forbidden"),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "NotFound maps to 404",
			err:            status.Error(codes.NotFound, "missing item"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "AlreadyExists maps to 409",
			err:            status.Error(codes.AlreadyExists, "duplicate email"),
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "DeadlineExceeded maps to 504",
			err:            status.Error(codes.DeadlineExceeded, "timeout"),
			expectedStatus: http.StatusGatewayTimeout,
		},
		{
			name:           "Unavailable maps to 503",
			err:            status.Error(codes.Unavailable, "service down"),
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "Non-gRPC error maps to 500",
			err:            errors.New("random internal error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			HandleGRPCError(c, tc.err, "operation failed")

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}
