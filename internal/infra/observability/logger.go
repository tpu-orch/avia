package observability

import "go.uber.org/zap"

type Logger interface {
	Info(message string, keyValues ...any)
	Error(message string, err error, keyValues ...any)
}

type zapLogger struct {
	logger *zap.SugaredLogger
}

func NewLogger(serviceName string) Logger {
	base, err := zap.NewProduction()
	if err != nil {
		return &zapLogger{logger: zap.NewNop().Sugar()}
	}

	return &zapLogger{logger: base.With(zap.String("service", serviceName)).Sugar()}
}

func (l *zapLogger) Info(message string, keyValues ...any) {
	l.logger.Infow(message, keyValues...)
}

func (l *zapLogger) Error(message string, err error, keyValues ...any) {
	if err != nil {
		keyValues = append([]any{"error", err}, keyValues...)
	}
	l.logger.Errorw(message, keyValues...)
}
