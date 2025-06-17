package database

import (
	"testing"

	"github.com/stretchr/testify/require"
	//"github.com/pashagolub/pgxmock"
)

// var (
// 	testInstance  *pgxpool.Pool
// 	testOnce      sync.Once
// 	testDBConnStr string = "postgres://dmitrij:Antirise1!@localhost:5432/metrics_db_test?sslmode=disable"
// )

func TestNewStorage(t *testing.T) {
	db := NewStorage()
	require.NotNil(t, db)
}

/*
func getTestDB() *pgxpool.Pool {
	testOnce.Do(func() {
		var err error

		testInstance, err = pgxpool.New(context.Background(), testDBConnStr)
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

func resetTestDB() error {
	if testInstance == nil {
		return nil
	}

	_, err := testInstance.Exec(context.Background(), "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		return fmt.Errorf("failed to reset test database: %w", err)
	}

	err = createTables(testInstance)
	if err != nil {
		return fmt.Errorf("unable to create test database tables: %w", err)
	}

	return nil
}

func closeTestDB() {
	if testInstance != nil {
		testInstance.Close()
	}
}
*/
