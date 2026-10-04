package core_postgres_gorm

import (
	"context"
	"fmt"
	"time"

	core_logger "github.com/Rics69/x-net/internal/core/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
	opTimeout time.Duration
}

func NewDB(ctx context.Context, config Config, log *core_logger.Logger) (*DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newGormLogger(log),

		// без этого постгресовые ошибки приходят как *pgconn.PgError,
		// с этим - как gorm.ErrDuplicatedKey / ErrForeignKeyViolated / ErrCheckConstraintViolated,
		// и в репозиториях можно проверять через errors.Is не завязываясь на драйвер
		TranslateError: true,

		// по умолчанию GORM оборачивает каждый одиночный Create/Update/Delete в транзакцию,
		// для одного запроса это лишний BEGIN/COMMIT. Где нужна транзакция - делаем явно через db.Transaction
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open gorm: %w", err)
	}

	// GORM работает поверх database/sql, пул соединений настраивается там
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB from gorm: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &DB{
		DB:        gormDB,
		opTimeout: config.Timeout,
	}, nil
}

func (d *DB) OpTimeout() time.Duration {
	return d.opTimeout
}

func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB from gorm: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close sql.DB: %w", err)
	}

	return nil
}
