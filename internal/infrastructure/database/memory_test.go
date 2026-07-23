package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestUser struct {
	ID    uint
	Name  string
	Email string
}

func TestNewMemoryDatabase(t *testing.T) {
	t.Run("success - creates memory database instance", func(t *testing.T) {
		memoryDB := NewMemoryDatabase(t)

		assert.NotNil(t, memoryDB)
		assert.NotNil(t, memoryDB.t)
		assert.Nil(t, memoryDB.db)
		assert.Nil(t, memoryDB.sqlDB)
	})
}

func TestMemoryDatabase_Connect(t *testing.T) {
	t.Run("success - connects to in-memory database", func(t *testing.T) {
		ctx := context.Background()
		memoryDB := NewMemoryDatabase(t)

		db, err := memoryDB.Connect(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, db)
		assert.NotNil(t, memoryDB.db)
		assert.NotNil(t, memoryDB.sqlDB)
	})

	t.Run("success - connect is idempotent (returns same instance)", func(t *testing.T) {
		ctx := context.Background()
		memoryDB := NewMemoryDatabase(t)

		db1, err1 := memoryDB.Connect(ctx)
		assert.NoError(t, err1)

		db2, err2 := memoryDB.Connect(ctx)
		assert.NoError(t, err2)

		assert.Equal(t, db1, db2)
	})
}

func TestMemoryDatabase_Migrate(t *testing.T) {
	t.Run("success - migrates single model", func(t *testing.T) {
		ctx := context.Background()
		memoryDB := NewMemoryDatabase(t)

		_, err := memoryDB.Connect(ctx)
		assert.NoError(t, err)

		err = memoryDB.Migrate(&TestUser{})

		assert.NoError(t, err)

		var tableName string
		result := memoryDB.db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name='test_users'").Scan(&tableName)
		assert.NoError(t, result.Error)
		assert.Equal(t, "test_users", tableName)
	})

	t.Run("success - migrates multiple models", func(t *testing.T) {
		ctx := context.Background()
		memoryDB := NewMemoryDatabase(t)

		_, err := memoryDB.Connect(ctx)
		assert.NoError(t, err)

		type TestProduct struct {
			ID    uint
			Name  string
			Price float64
		}

		err = memoryDB.Migrate(&TestUser{}, &TestProduct{})

		assert.NoError(t, err)
	})

	t.Run("error - migrate without connection", func(t *testing.T) {
		memoryDB := NewMemoryDatabase(t)

		err := memoryDB.Migrate(&TestUser{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database not connected")
	})

	t.Run("error - migrate with no models", func(t *testing.T) {
		ctx := context.Background()
		memoryDB := NewMemoryDatabase(t)

		_, err := memoryDB.Connect(ctx)
		assert.NoError(t, err)

		err = memoryDB.Migrate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no models provided for migration")
	})
}

func TestMemoryDatabase_Close(t *testing.T) {
	t.Run("success - closes connection", func(t *testing.T) {
		ctx := context.Background()
		memoryDB := NewMemoryDatabase(t)

		_, err := memoryDB.Connect(ctx)
		assert.NoError(t, err)

		err = memoryDB.Close()

		assert.NoError(t, err)
	})

	t.Run("success - close without connection returns nil", func(t *testing.T) {
		memoryDB := NewMemoryDatabase(t)

		err := memoryDB.Close()

		assert.NoError(t, err)
	})
}

func TestMemoryDatabase_IntegrationWorkflow(t *testing.T) {
	t.Run("success - full workflow (connect -> migrate -> query -> close)", func(t *testing.T) {
		ctx := context.Background()
		memoryDB := NewMemoryDatabase(t)

		db, err := memoryDB.Connect(ctx)
		assert.NoError(t, err)

		err = memoryDB.Migrate(&TestUser{})
		assert.NoError(t, err)

		testUser := TestUser{Name: "John Doe", Email: "john@example.com"}
		result := db.Create(&testUser)
		assert.NoError(t, result.Error)
		assert.NotZero(t, testUser.ID)

		var retrievedUser TestUser
		result = db.First(&retrievedUser, testUser.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, "John Doe", retrievedUser.Name)
		assert.Equal(t, "john@example.com", retrievedUser.Email)

		err = memoryDB.Close()
		assert.NoError(t, err)
	})
}
