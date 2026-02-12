package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/pranaovs/qashare/config"
)

var logger = slog.Default()

// ANSI color codes for log level backgrounds
const (
	colorReset  = "\033[0m"
	colorDebug  = "\033[46;97m" // Cyan bg, black text
	colorInfo   = "\033[42;97m" // Green bg, black text
	colorWarn   = "\033[43;97m" // Yellow bg, black text
	colorError  = "\033[41;97m" // Red bg, white text
	colorSource = "\033[2m"     // Dim text for source location
)

// prettyHandler is a custom slog.Handler that outputs colored, human-readable logs.
type prettyHandler struct {
	opts  slog.HandlerOptions
	mu    *sync.Mutex
	out   io.Writer
	attrs []slog.Attr
	group string
}

func newPrettyHandler(out io.Writer, opts *slog.HandlerOptions) *prettyHandler {
	h := &prettyHandler{
		out: out,
		mu:  &sync.Mutex{},
	}
	if opts != nil {
		h.opts = *opts
	}
	return h
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	// Time: HH:MM:SS
	timeStr := r.Time.Format("2006/01/02 - 15:04:05")

	// Level with color
	var levelStr string
	switch {
	case r.Level >= slog.LevelError:
		levelStr = fmt.Sprintf("%s ERROR %s", colorError, colorReset)
	case r.Level >= slog.LevelWarn:
		levelStr = fmt.Sprintf("%s WARN %s", colorWarn, colorReset)
	case r.Level >= slog.LevelInfo:
		levelStr = fmt.Sprintf("%s INFO %s", colorInfo, colorReset)
	default:
		levelStr = fmt.Sprintf("%s DEBUG %s", colorDebug, colorReset)
	}

	// Source location from PC
	sourceStr := ""
	if r.PC != 0 {
		fn := runtime.FuncForPC(r.PC)
		if fn != nil {
			file, line := fn.FileLine(r.PC)
			// Show parent_dir/file.go:line for brevity
			short := filepath.Join(filepath.Base(filepath.Dir(file)), filepath.Base(file))
			sourceStr = fmt.Sprintf("%s[%s:%d]%s ", colorSource, short, line, colorReset)
		}
	}

	// Build attrs string
	var attrStr strings.Builder
	for _, a := range h.attrs {
		attrStr.WriteString(" " + formatAttr(h.group, a))
	}
	r.Attrs(func(a slog.Attr) bool {
		attrStr.WriteString(" " + formatAttr(h.group, a))
		return true
	})

	line := fmt.Sprintf("%s |%s| %s%s%s\n", timeStr, levelStr, sourceStr, r.Message, attrStr.String())

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write([]byte(line))
	return err
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs), len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	newAttrs = append(newAttrs, attrs...)
	return &prettyHandler{
		opts:  h.opts,
		mu:    h.mu,
		out:   h.out,
		attrs: newAttrs,
		group: h.group,
	}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	newGroup := name
	if h.group != "" {
		newGroup = h.group + "." + name
	}
	return &prettyHandler{
		opts:  h.opts,
		mu:    h.mu,
		out:   h.out,
		attrs: h.attrs,
		group: newGroup,
	}
}

func formatAttr(group string, a slog.Attr) string {
	key := a.Key
	if group != "" {
		key = group + "." + key
	}
	return fmt.Sprintf("%s=%v", key, a.Value)
}

// InitDefaultLogger sets up the pretty logger with default settings (INFO level).
// Call this early in main() before config is loaded so that all startup logs
// are formatted consistently.
func InitDefaultLogger() {
	handler := newPrettyHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})
	logger = slog.New(handler)
	slog.SetDefault(logger)

	// Redirect standard log package through slog
	log.SetOutput(&slogWriter{level: slog.LevelInfo})
	log.SetFlags(0)
}

// InitLogger re-initializes the logger with the provided config.
// Call this after config is loaded to apply debug level if configured.
func InitLogger(cfg *config.Config) {
	level := slog.LevelInfo
	if cfg.App.Debug {
		level = slog.LevelDebug
	}

	handler := newPrettyHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})
	logger = slog.New(handler)
	slog.SetDefault(logger)

	// Redirect standard log package through slog
	log.SetOutput(&slogWriter{level: slog.LevelInfo})
	log.SetFlags(0)
}

// slogWriter adapts slog to be used as an io.Writer for the standard log package.
type slogWriter struct {
	level slog.Level
}

func (w *slogWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	slog.Log(context.Background(), w.level, msg)
	return len(p), nil
}

// Logger returns the global structured logger
func Logger() *slog.Logger {
	return logger
}

// LogError logs an error with context
func LogError(ctx context.Context, msg string, err error, attrs ...any) {
	allAttrs := append([]any{"error", err}, attrs...)
	logger.ErrorContext(ctx, msg, allAttrs...)
}

// LogInfo logs an informational message
func LogInfo(ctx context.Context, msg string, attrs ...any) {
	logger.InfoContext(ctx, msg, attrs...)
}

// LogDebug logs a debug message
func LogDebug(ctx context.Context, msg string, attrs ...any) {
	logger.DebugContext(ctx, msg, attrs...)
}

// LogWarn logs a warning message
func LogWarn(ctx context.Context, msg string, attrs ...any) {
	logger.WarnContext(ctx, msg, attrs...)
}
