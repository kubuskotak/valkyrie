package valkyrie

import (
	"context"
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
