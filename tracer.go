package valkyrie

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	semConv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"io"
	"runtime"
	"time"

	"github.com/opentracing/opentracing-go"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

func Tracer(service, version string) (*sdkTrace.TracerProvider, func(ctx context.Context), error) {
	exp, err := jaeger.New(jaeger.WithAgentEndpoint())
	if err != nil {
		log.Error().Err(err).Msg("Could not parse Jaeger env vars")
		return nil, nil, err
	}
	tracer := sdkTrace.NewTracerProvider(
		// Always be sure to batch in production.
		sdkTrace.WithBatcher(exp),
		// Record information about this application in a Resource.
		sdkTrace.WithResource(resource.NewWithAttributes(
			semConv.SchemaURL,
			semConv.ServiceNameKey.String(service),
			attribute.String("version", version),
		)),
	)

	cleanup := func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tracer.Shutdown(ctx); err != nil {
			log.Fatal().Err(err)
		}
	}

	return tracer, cleanup, err
}

func StartSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	var tracer trace.Tracer
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		tracer = span.TracerProvider().Tracer(operationName)
	} else {
		tracer = otel.GetTracerProvider().Tracer(operationName)
	}
	pc, _, _, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	return tracer.Start(ctx, details.Name())

}

const (
	// SpanIDFieldName is the field name for the span ID.
	SpanIDFieldName = "span.id"

	// SpanContext is the field name for the span context.
	SpanContext = "span.context"

	// TraceIDFieldName is the field name for the trace ID.
	TraceIDFieldName = "trace.id"

	traceID64bitsWidth  = 64 / 4
	traceID128bitsWidth = 128 / 4
	spanIDWidth         = 64 / 4

	traceIDPadding = "0000000000000000"
)

// TraceContextHook returns a zerolog.Hook that will add any trace context
// contained in ctx to log events.
func TraceContextHook(ctx context.Context) zerolog.Hook {
	return traceContextHook{ctx}
}

type traceContextHook struct {
	ctx context.Context
}

func (t traceContextHook) Run(e *zerolog.Event, level zerolog.Level, message string) {
	sc := trace.SpanFromContext(t.ctx).SpanContext()
	if !sc.TraceID().IsValid() || !sc.SpanID().IsValid() {
		return
	}
	e.Str(TraceIDFieldName, sc.TraceID().String())
	e.Str(SpanIDFieldName, sc.TraceID().String())
}

type ZeroWriter struct {
	// MinLevel holds the minimum level of logs to send to
	//
	// MinLevel must be greater than or equal to zerolog.ErrorLevel.
	// If it is less than this, zerolog.ErrorLevel will be used as
	// the minimum instead.
	MinLevel zerolog.Level
}

func (w *ZeroWriter) minLevel() zerolog.Level {
	minLevel := w.MinLevel
	if minLevel < zerolog.DebugLevel {
		minLevel = zerolog.DebugLevel
	}
	return minLevel
}

// Write is a no-op.
func (*ZeroWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// WriteLevel decodes the JSON-encoded log record in p, and reports it as an error using w.Tracer.
func (w *ZeroWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	if level < w.minLevel() || level >= zerolog.NoLevel {
		return len(p), nil
	}

	var (
		scc = trace.SpanContextConfig{}
		err error
	)

	var logRecord logRecord
	events, err := logRecord.decode(bytes.NewReader(p))
	if err != nil {
		return 0, err
	}

	if logRecord.traceId != "" {
		id := logRecord.spanId
		log.Info().Str(TraceIDFieldName, id).Send()
		if len(id) != traceID128bitsWidth && len(id) != traceID64bitsWidth {
			return 0, errInvalidTraceIDLength
		}
		// padding when length is 16
		if len(id) == traceID64bitsWidth {
			id = traceIDPadding + id
		}
		scc.TraceID, err = trace.TraceIDFromHex(id)
		if err != nil {
			return 0, errMalformedTraceID
		}
	}

	if logRecord.spanId != "" {
		id := logRecord.spanId
		log.Info().Str(SpanIDFieldName, id).Send()
		if len(id) != spanIDWidth {
			return 0, errInvalidSpanIDLength
		}
		scc.SpanID, err = trace.SpanIDFromHex(id)
		if err != nil {
			return 0, errMalformedSpanID
		}
	}

	ctx := trace.ContextWithRemoteSpanContext(context.Background(), trace.NewSpanContext(scc))
	span := trace.SpanFromContext(ctx)
	defer span.End()

	for key, value := range events {
		if key == SpanContext {
			continue
		}
		if key == zerolog.LevelFieldName {
			w.levelTo(value.(string), span)
			span.SetAttributes(attribute.String(key, value.(string)))
			buf := &bytes.Buffer{}
			enc := json.NewEncoder(buf)
			enc.SetEscapeHTML(true)
			if err := enc.Encode(events[zerolog.MessageFieldName]); err != nil {
				continue
			}
			span.SetAttributes(attribute.String(zerolog.MessageFieldName, buf.String()))
			continue
		}

		switch v := value.(type) {
		case string:
			span.SetAttributes(attribute.String(key, v))
		case json.Number:
			span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
		default:
			b, err := json.Marshal(v)
			if err != nil {
				span.RecordError(err)
			} else {
				span.SetAttributes(attribute.String(key, fmt.Sprintf("%s", b)))
			}
		}
	}

	return len(p), nil
}

func (w *ZeroWriter) levelTo(level string, span trace.Span) {
	lvl, _ := zerolog.ParseLevel(level)
	switch lvl {
	case zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel:
		span.SetStatus(codes.Error, "logging")
	}
}

var (
	errInvalidTraceIDLength = errors.New("invalid trace id length, must be either 16 or 32")
	errMalformedTraceID     = errors.New("cannot decode trace id from header, should be a string of hex, lowercase trace id can't be all zero")
	errInvalidSpanIDLength  = errors.New("invalid span id length, must be 16")
	errMalformedSpanID      = errors.New("cannot decode span id from header, should be a string of hex, lowercase span id can't be all zero")
)

type logRecord struct {
	message         string
	timestamp       time.Time
	traceId, spanId string
	spanContext     opentracing.SpanContext
}

func (l *logRecord) decode(r io.Reader) (events map[string]interface{}, err error) {
	m := make(map[string]interface{})
	d := json.NewDecoder(r)
	d.UseNumber()
	if err := d.Decode(&m); err != nil {
		return events, err
	}

	l.message, _ = m[zerolog.MessageFieldName].(string)
	if strVal, ok := m[zerolog.TimestampFieldName].(string); ok {
		if t, err := time.Parse(zerolog.TimeFieldFormat, strVal); err == nil {
			l.timestamp = t.UTC()
		}
	}

	if strVal, ok := m[SpanIDFieldName].(string); ok {
		l.spanId = strVal
	}

	if strVal, ok := m[TraceIDFieldName].(string); ok {
		l.traceId = strVal
	}
	return m, nil
}
