package sdk

import (
	"context"
	"github.com/openzipkin/zipkin-go"
)

func SetLogFormat(ctx context.Context) {
	spanID := zipkin.SpanFromContext(ctx).Context().ID.String()
	traceID := zipkin.SpanFromContext(ctx).Context().TraceID.String()
	LOGGER.SetFormat("%{time} [%{module}] [%{level}] [" + traceID +  "," + spanID + "]  %{message}")
}
