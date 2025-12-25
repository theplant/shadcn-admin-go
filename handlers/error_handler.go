package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	api "github.com/sunfmin/shadcn-admin-go/api/gen/admin"
)

// hideErrorDetails controls whether error details are included in responses
var hideErrorDetails bool

// SetHideErrorDetails configures whether to hide error details in responses
// Should be set to true in production environments
func SetHideErrorDetails(hide bool) {
	hideErrorDetails = hide
}

// OgenErrorHandler implements ogenerrors.ErrorHandler for ogen servers
// It maps service sentinel errors to user-friendly HTTP responses
func OgenErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	errCode := mapServiceError(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errCode.HTTPStatus)

	resp := api.ErrorResponse{
		Code:    errCode.Code,
		Message: errCode.Message,
	}

	// Include details in development only
	if !hideErrorDetails && err != nil {
		resp.Details.SetTo(err.Error())
	}

	json.NewEncoder(w).Encode(resp)
}

// mapServiceError finds the matching ErrorCode for a service error
func mapServiceError(err error) ErrorCode {
	// Check context errors first
	if errors.Is(err, context.Canceled) {
		return Errors.RequestCancelled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return Errors.RequestTimeout
	}

	// Check service sentinel errors
	for _, errCode := range AllErrors() {
		if errCode.ServiceErr != nil && errors.Is(err, errCode.ServiceErr) {
			return errCode
		}
	}

	return Errors.InternalError
}
