package tracing

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func WrapHandlerFunc(handlerFunc http.HandlerFunc, operation string) http.Handler {
	return otelhttp.NewHandler(handlerFunc, operation)
}
