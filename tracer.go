package valkyrie

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rs/zerolog"
	"github.com/uber/jaeger-client-go"
	"io"
	"time"

	"github.com/opentracing/opentracing-go"
	spanLog "github.com/opentracing/opentracing-go/log"
	"github.com/rs/zerolog/log"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
)

func Tracer(service, version string) (opentracing.Tracer, func(), error) {
	cfg, err := jaegerCfg.FromEnv()
	if err != nil {
		log.Error().Err(err).Msg("Could not parse Jaeger env vars")
		return nil, nil, err
	}

	cfg.Sampler.Param = 1

	if service != "" {
		cfg.ServiceName = service
	}
	jMetricsFactory := metrics.NullFactory

	tracer, closer, err := cfg.NewTracer(
		jaegerCfg.Metrics(jMetricsFactory),
		jaegerCfg.Tag(fmt.Sprintf("%.version", service), version),
		jaegerCfg.MaxTagValueLength(2048),
	)
	if err != nil {
		log.Error().Err(err).Msg("Could not initialize jaeger tracer")
		return nil, nil, err
	}

	cleanup := func() {
		_ = closer.Close()
	}

	return tracer, cleanup, err
}

const (
	// SpanIDFieldName is the field name for the span ID.
	SpanIDFieldName = "span.id"

	// SpanContext is the field name for the span context.
	SpanContext = "span.context"

	// TraceIDFieldName is the field name for the trace ID.
	TraceIDFieldName = "trace.id"
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
	if span := opentracing.SpanFromContext(t.ctx); span != nil {
		if sc, ok := span.Context().(jaeger.SpanContext); ok {
			buf := new(bytes.Buffer)
			if err := span.Tracer().Inject(sc, opentracing.Binary, buf); err != nil {
				span.LogFields(spanLog.Error(err))
			}

			e.Hex(SpanContext, buf.Bytes())
			e.Str(TraceIDFieldName, sc.TraceID().String())
			e.Str(SpanIDFieldName, sc.SpanID().String())
		}
	}
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

	var logRecord logRecord
	events, err := logRecord.decode(bytes.NewReader(p))
	if err != nil {
		return 0, err
	}

	if logRecord.spanContext != nil {
		span := opentracing.StartSpan(
			"zero-logger",
			ext.RPCServerOption(logRecord.spanContext),
		)
		defer span.Finish()

		fields := make([]spanLog.Field, 0)
		for key, value := range events {
			if key == SpanContext {
				continue
			}
			if key == zerolog.LevelFieldName {
				w.levelTo(value.(string), span)
				fields = append(fields, spanLog.String(key, value.(string)))
				span.SetTag(key, value.(string))
				span.SetTag(zerolog.MessageFieldName, events[zerolog.MessageFieldName])
				continue
			}

			switch v := value.(type) {
			case string:
				fields = append(fields, spanLog.String(key, v))
			case json.Number:
				fields = append(fields, spanLog.String(key, fmt.Sprint(v)))
			default:
				b, err := json.Marshal(v)
				if err != nil {
					fields = append(fields, spanLog.String(key, fmt.Sprintf("[error: %v]", err)))
				} else {
					fields = append(fields, spanLog.String(key, string(b)))
				}
			}
		}
		span.LogFields(fields...)
	}

	return len(p), nil
}

func (w *ZeroWriter) levelTo(level string, span opentracing.Span) {
	lvl, _ := zerolog.ParseLevel(level)
	switch lvl {
	case zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel:
		span.SetTag("error", true)
	}
}

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

	if strVal, ok := m[SpanContext].(string); ok {
		b, _ := hex.DecodeString(strVal)
		buf := new(bytes.Buffer)
		buf.Write(b)

		sc, err := opentracing.GlobalTracer().
			Extract(opentracing.Binary, buf)
		if err != nil {
			return events, err
		}
		l.spanContext = sc
	}

	if strVal, ok := m[SpanIDFieldName].(string); ok {
		l.spanId = strVal
	}

	if strVal, ok := m[TraceIDFieldName].(string); ok {
		l.traceId = strVal
	}
	return m, nil
}
