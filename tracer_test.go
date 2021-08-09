package valkyrie

import (
	"bytes"
	"context"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

func TestTracer(t *testing.T) {
	tracer, cleanupTrace, err := Tracer("init app", "v1")
	assert.NoError(t, err)
	defer cleanupTrace()

	ctx := context.Background()
	serverSpan := opentracing.SpanFromContext(ctx)
	if serverSpan == nil {
		// All we can do is create a new root span.
		serverSpan = tracer.StartSpan("operationName")
	} else {
		serverSpan.SetOperationName("operationName")
	}
	defer serverSpan.Finish()

	opentracing.ContextWithSpan(ctx, serverSpan)
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
