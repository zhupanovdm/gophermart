package logging

const (
	// CorrelationIDKey is used to track unique request.
	CorrelationIDKey = "cid"

	// ComponentKey is used to track concrete component side effects.
	ComponentKey = "service"

	// CorrelationIDHeader is used to transport Correlation ID context value via the HTTP header.
	CorrelationIDHeader = "X-CorrelationID"
)

//type LogCtxProvider interface {
//	LoggerCtx(ctx zerolog.Context) zerolog.Context
//}
//
//var _ LogCtxProvider = (LoggerCtxUpdate)(nil)
//
//type LoggerCtxUpdate func(ctx zerolog.Context) zerolog.Context
//
//func (upd LoggerCtxUpdate) LoggerCtx(ctx zerolog.Context) zerolog.Context {
//	if upd != nil {
//		return upd(ctx)
//	}
//	return ctx
//}

//func LogCtxUpdateWith(ctx zerolog.Context, providers ...LogCtxProvider) zerolog.Context {
//	for _, p := range providers {
//		ctx = p.LoggerCtx(ctx)
//	}
//	return ctx
//}

//func LogCtxFrom(providers ...LogCtxProvider) LoggerCtxUpdate {
//	return func(ctx zerolog.Context) zerolog.Context {
//		return LogCtxUpdateWith(ctx, providers...)
//	}
//}
//
//func LogCtxKeyStr(key string, value string) LoggerCtxUpdate {
//	return func(ctx zerolog.Context) zerolog.Context {
//		return ctx.Str(key, value)
//	}
//}
