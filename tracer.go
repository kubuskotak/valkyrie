package valkyrie

import (
	"fmt"

	"github.com/opentracing/opentracing-go"
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
