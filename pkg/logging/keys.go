package logging

const (
	// CorrelationIDKey is used to track unique request.
	CorrelationIDKey = "cid"

	// ServiceKey is used to track concrete app service side effects.
	ServiceKey = "service"

	UserLoginKey = "user_login"

	UserIdKey = "user_id"

	OrderNumberKey = "order_number"

	// CorrelationIDHeader is used to transport Correlation ID context value via the HTTP header.
	CorrelationIDHeader = "X-CorrelationID"
)
