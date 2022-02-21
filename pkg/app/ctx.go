package app

import "fmt"

// ctxKeyPrefix defines global context key prefix to prevent collisions with dependency packages context keys.
const ctxKeyPrefix = "GopherMart-"

// CtxKey is used to retrieve values from context by its key.
type ContextKey string

func (c ContextKey) String() string {
	return fmt.Sprintf("%s%s", ctxKeyPrefix, string(c))
}
