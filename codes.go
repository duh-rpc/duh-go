package duh

//const (
//	CodeOK                  = 200
//	CodeBadRequest          = 400
//	CodeUnauthorized        = 401
//	CodeRequestFailed       = 402
//	CodeMethodNotAllowed    = 403
//	CodeConflict            = 409
//	CodeClientError         = 428
//	CodeTooManyRequests     = 429
//	CodeInternalError       = 500
//	CodeInfrastructureError = 502
//)

// IsInfrastructureError returns true if the code should be an infrastructure class error
func IsInfrastructureError(code int64) bool {
	return code < 500
}
