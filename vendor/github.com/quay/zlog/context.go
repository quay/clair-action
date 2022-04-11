package zlog

import (
	"context"

	"go.opentelemetry.io/otel/baggage"
)

// ContextWithValues is a helper for the go.opentelemetry.io/otel/baggage v1
// API. It takes pairs of strings and adds them to the Context via the baggage
// package.
//
// Any trailing value is silently dropped.
func ContextWithValues(ctx context.Context, pairs ...string) context.Context {
	var err error
	b := baggage.FromContext(ctx)
	pairs = pairs[:len(pairs)-len(pairs)%2]
	ms := make([]baggage.Member, 0, len(pairs)/2)
	for i := 0; i < len(pairs)/2; i = i + 2 {
		k, v := pairs[i], pairs[i+1]
		m, err := baggage.NewMember(k, v)
		if err != nil {
			Warn(ctx).
				Err(err).
				Str("key", k).
				Msg("failed to create baggage member")
			continue
		}
		ms = append(ms, m)
	}
	b, err = baggage.New(append(b.Members(), ms...)...)
	if err != nil {
		Warn(ctx).
			Err(err).
			Msg("failed to create baggage")
		return ctx
	}
	return baggage.ContextWithBaggage(ctx, b)
}
