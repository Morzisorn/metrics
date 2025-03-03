package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

var (
	testInstance  *pgx.Conn
	testOnce      sync.Once
	testDBConnStr string = "postgres://dmitrij:Antirise1!@localhost:5432/metrics_db_test?sslmode=disable"
)

func GetTestDB() *pgx.Conn {
	testOnce.Do(func() {
		var err error

		testInstance, err = pgx.Connect(context.Background(), testDBConnStr)
		if err != nil {
			fmt.Printf("Unable to connect to test database: %v", err)
		}

		err = createTables(testInstance)
		if err != nil {
			fmt.Printf("Unable to create test database tables: %v", err)
		}

	})
	return testInstance
}

func ResetTestDB() error {
	if testInstance == nil {
		return nil
	}

	_, err := testInstance.Exec(context.Background(), "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		return fmt.Errorf("failed to reset test database: %w", err)
	}

	err = createTables(testInstance)
	if err != nil {
		return fmt.Errorf("Unable to create test database tables: %w", err)
	}

	return nil
}

func CloseTestDB() {
	if testInstance != nil {
		err := testInstance.Close(context.Background())
		if err != nil {
			logger.Log.Panic("test DB close error: ", zap.Error(err))
		}
	}
}
