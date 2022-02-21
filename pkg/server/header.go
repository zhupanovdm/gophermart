package server

import (
	"fmt"
	"net/http"
	"strings"
)

type Header http.Header

func (h Header) String() string {
	var builder strings.Builder
	for k, v := range h {
		if builder.Len() != 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(fmt.Sprintf("%s:%s", k, strings.Join(v, ";")))
	}
	return builder.String()
}
