package core_postgres_gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	core_logger "github.com/Rics69/x-net/internal/core/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

const slowQueryThreshold = 200 * time.Millisecond

// адаптер, чтобы GORM писал в наш zap, а не в stdout своим форматом.
// логгер берём из ctx - тогда у SQL-запросов будет request_id от middleware
type gormLogger struct {
	log *core_logger.Logger
}

func newGormLogger(log *core_logger.Logger) *gormLogger {
	return &gormLogger{
		log: log,
	}
}

// уровень логирования решает zap (LOGGER_LEVEL), поэтому LogMode игнорим
func (l *gormLogger) LogMode(gorm_logger.LogLevel) gorm_logger.Interface {
	return l
}

func (l *gormLogger) Info(ctx context.Context, msg string, args ...any) {
	l.fromContext(ctx).Info(fmt.Sprintf(msg, args...))
}

func (l *gormLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.fromContext(ctx).Warn(fmt.Sprintf(msg, args...))
}

func (l *gormLogger) Error(ctx context.Context, msg string, args ...any) {
	l.fromContext(ctx).Error(fmt.Sprintf(msg, args...))
}

func (l *gormLogger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	log := l.fromContext(ctx)
	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("elapsed", elapsed),
	}

	switch {
	// RecordNotFound - обычный сценарий (нет юзера), не ошибка уровня БД.
	// Ошибки запросов и так всплывут наверх и залогируются в ErrorResponse
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		log.Debug("gorm query failed", append(fields, zap.Error(err))...)
	case elapsed > slowQueryThreshold:
		log.Warn("gorm slow query", fields...)
	default:
		log.Debug("gorm query", fields...)
	}
}

func (l *gormLogger) fromContext(ctx context.Context) *core_logger.Logger {
	return core_logger.FromContextOr(ctx, l.log)
}
