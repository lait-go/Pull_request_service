package http

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	"revivers/pkg/logger"

	"github.com/google/uuid"
)

// contextKey используется для типизированных ключей контекста
type contextKey string

const requestIDKey contextKey = "request_id"

// RecoveryMiddleware обрабатывает паники и предотвращает падение приложения
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered", nil,
					"error", err,
					"path", r.URL.Path,
					"method", r.Method,
					"stack", string(debug.Stack()),
				)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware добавляет уникальный ID к каждому запросу для трейсинга
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Добавляем request ID в заголовок ответа
		w.Header().Set("X-Request-ID", requestID)

		// Добавляем request ID в контекст
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware логирует все HTTP запросы
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем ResponseWriter для отслеживания статуса
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		// Получаем request ID из контекста
		requestID := GetRequestID(r.Context())

		logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.statusCode,
			"duration_ms", duration.Milliseconds(),
			"request_id", requestID,
			"remote_addr", r.RemoteAddr,
		)
	})
}

// responseWriter обертка для http.ResponseWriter для отслеживания статуса ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetRequestID извлекает request ID из контекста
func GetRequestID(ctx context.Context) string {
	if id := ctx.Value(requestIDKey); id != nil {
		return id.(string)
	}
	return ""
}
