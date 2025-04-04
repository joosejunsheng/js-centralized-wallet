package trace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
)

type keyType int

const (
	keyTrace keyType = iota
)

const (
	logKeyTracePath = "trace_path"
)

type trace struct {
	l    *slog.Logger
	path string
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		logger = slog.Default()
	}

	return context.WithValue(ctx, keyTrace, trace{l: logger})
}

func Logger(ctx context.Context) (context.Context, *slog.Logger) {
	t, ok := ctx.Value(keyTrace).(trace)
	if !ok {
		t = trace{l: slog.Default()}
		ctx = context.WithValue(ctx, keyTrace, t)
	}

	if t.path == "" {
		t.path = randStr(16)
	} else {
		t.path += "/" + randStr(8)
	}

	return context.WithValue(ctx, keyTrace, t),
		t.l.With(logKeyTracePath, t.path)
}

func randStr(n int) string {
	r := make([]byte, n)
	_, _ = rand.Read(r)
	return hex.EncodeToString(r)
}
