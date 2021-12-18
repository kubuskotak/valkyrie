package valkyrie

import (
	"go.opentelemetry.io/otel/baggage"
	"testing"
)

func TestTestHarness(t *testing.T) {
	ctx := Test(nil, t)
	t.Log("🖳")
	Log(ctx).Msg("🖳")
	Log(ctx).Msg("🆒")
}

func TestDeduplication(t *testing.T) {
	ctx := Test(nil, t)
	t.Log("make sure keys aren't repeated")
	m, _ := baggage.NewMember("x", "5")
	b, _ := baggage.FromContext(ctx).SetMember(m)
	ctx = baggage.ContextWithBaggage(ctx, b)
	Log(ctx).Msg("5?")
	t.Log("should be 5")
	m, _ = baggage.NewMember("x", "6")
	b, _ = baggage.FromContext(ctx).SetMember(m)
	ctx = baggage.ContextWithBaggage(ctx, b)
	Log(ctx).Msg("6?")
	t.Log("should be 6")
}

func TestSub(t *testing.T) {
	ctx := Test(nil, t)
	t.Log("make sure subtests are intelligible")
	m, _ := baggage.NewMember("outer", "test")
	b, _ := baggage.FromContext(ctx).SetMember(m)
	ctx = baggage.ContextWithBaggage(ctx, b)
	t.Run("a", func(t *testing.T) {
		ctx := Test(ctx, t)
		m, _ := baggage.NewMember("inner", "a")
		b, _ := baggage.FromContext(ctx).SetMember(m)
		ctx = baggage.ContextWithBaggage(ctx, b)
		Log(ctx).Msg("hello")
	})
	t.Run("b", func(t *testing.T) {
		ctx := Test(ctx, t)
		m, _ := baggage.NewMember("inner", "b")
		b, _ := baggage.FromContext(ctx).SetMember(m)
		ctx = baggage.ContextWithBaggage(ctx, b)
		Log(ctx).Msg("hello")
	})
}
