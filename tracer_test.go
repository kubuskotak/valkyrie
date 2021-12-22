package valkyrie

import (
	"bytes"
	"context"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracer(t *testing.T) {
	tracer, cleanupTrace, err := Tracer("init app", "v1")
	assert.NoError(t, err)
	defer cleanupTrace(context.Background())

	ctx := context.Background()
	serverSpan := trace.SpanFromContext(ctx)
	if serverSpan == nil {
		// All we can do is create a new root span.
		_, serverSpan = tracer.Tracer("TestTracer").Start(ctx, "operationName")
	}
	defer serverSpan.End()

	trace.ContextWithSpan(ctx, serverSpan)
}

func ExampleTraceContextHook() {
	handleRequest := func(w http.ResponseWriter, req *http.Request) {
		logger := zerolog.Ctx(req.Context()).Hook(TraceContextHook(req.Context()))
		logger.Error().Msg("message")
	}
	http.HandleFunc("/", handleRequest)
}

func TestTraceContextHookNothing(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Hook(TraceContextHook(context.Background()))
	logger.Info().Msg("message")

	require.Equal(t, "{\"level\":\"info\",\"message\":\"message\"}\n", buf.String())
}
