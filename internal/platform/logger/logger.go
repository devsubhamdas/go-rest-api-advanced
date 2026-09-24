package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/config"
)

// prettyJSONWriter re-indents each JSON log line written by slog.
type prettyJSONWriter struct {
	w io.Writer
}

func (p prettyJSONWriter) Write(b []byte) (int, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, bytes.TrimSpace(b), "", "  "); err != nil {
		// Not valid JSON for some reason; write it through untouched.
		return p.w.Write(b)
	}
	buf.WriteByte('\n')
	if _, err := p.w.Write(buf.Bytes()); err != nil {
		return 0, err
	}
	return len(b), nil
}

func NewJSONLogger(cfg *config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}

	var out io.Writer = os.Stdout
	if cfg.AppEnv != "production" {
		out = prettyJSONWriter{w: os.Stdout}
	}

	return slog.New(slog.NewJSONHandler(out, opts))
}
