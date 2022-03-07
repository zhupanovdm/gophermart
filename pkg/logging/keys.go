package logging

const (
	// CorrelationIDKey is used to track unique request.
	CorrelationIDKey = "cid"

	// ServiceKey is used to track concrete app service side effects.
	ServiceKey = "service"

	// CorrelationIDHeader is used to transport Correlation ID context value via the HTTP header.
	CorrelationIDHeader = "X-CorrelationID"
)
