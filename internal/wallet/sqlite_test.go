package wallet

import (
	"context"
	"testing"

	"github.com/opinedajr/micro-investing/internal/infrastructure/database"
	"github.com/opinedajr/micro-investing/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteRepository_Create(t *testing.T) {
	t.Run("success - creates and retrieves wallet", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		err = gormDB.AutoMigrate(&Wallet{})
		require.NoError(t, err)

		repo := NewSQLiteRepository(gormDB)

		wallet := &Wallet{
			UserID:      shared.DefaultUserID,
			Name:        "Minha Carteira",
			Description: ptr("Ações e FIIs"),
		}

		err = repo.Create(context.Background(), wallet)
		assert.NoError(t, err)
		assert.NotEmpty(t, wallet.ID, "ID should be generated")

		found, err := repo.FindByID(context.Background(), wallet.ID)
		assert.NoError(t, err)
		assert.Equal(t, wallet.ID, found.ID)
		assert.Equal(t, wallet.Name, found.Name)
		assert.Equal(t, wallet.Description, found.Description)
	})

	t.Run("success - finds all wallets by user id", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		err = gormDB.AutoMigrate(&Wallet{})
		require.NoError(t, err)

		repo := NewSQLiteRepository(gormDB)
		userID := shared.DefaultUserID

		wallet1 := &Wallet{UserID: userID, Name: "Carteira 1", Description: ptr("Desc 1")}
		wallet2 := &Wallet{UserID: userID, Name: "Carteira 2", Description: ptr("Desc 2")}

		_ = repo.Create(context.Background(), wallet1)
		_ = repo.Create(context.Background(), wallet2)

		wallets, err := repo.FindAllByUserID(context.Background(), userID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(wallets), 2, "should find at least 2 wallets")
	})

	t.Run("success - finds wallet by name and user id", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		err = gormDB.AutoMigrate(&Wallet{})
		require.NoError(t, err)

		repo := NewSQLiteRepository(gormDB)
		userID := shared.DefaultUserID
		name := "Busca por Nome"

		wallet := &Wallet{UserID: userID, Name: name, Description: ptr("Desc")}
		_ = repo.Create(context.Background(), wallet)

		found, err := repo.FindByNameAndUserID(context.Background(), name, userID)
		assert.NoError(t, err)
		assert.Equal(t, name, found.Name)
		assert.Equal(t, userID, found.UserID)
	})

	t.Run("error - find by id returns nil for non-existent", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		err = gormDB.AutoMigrate(&Wallet{})
		require.NoError(t, err)

		repo := NewSQLiteRepository(gormDB)

		_, err = repo.FindByID(context.Background(), "non-existent-id")
		assert.Error(t, err)
	})

	t.Run("success - deletes wallet", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		err = gormDB.AutoMigrate(&Wallet{})
		require.NoError(t, err)

		repo := NewSQLiteRepository(gormDB)

		wallet := &Wallet{UserID: shared.DefaultUserID, Name: "To Delete", Description: ptr("Will be deleted")}
		err = repo.Create(context.Background(), wallet)
		require.NoError(t, err)

		err = repo.Delete(context.Background(), wallet.ID)
		assert.NoError(t, err)

		_, err = repo.FindByID(context.Background(), wallet.ID)
		assert.Error(t, err)
	})

	t.Run("error - delete non-existent wallet returns error", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		err = gormDB.AutoMigrate(&Wallet{})
		require.NoError(t, err)

		repo := NewSQLiteRepository(gormDB)

		err = repo.Delete(context.Background(), "non-existent-id")
		assert.Error(t, err)
	})
}

func ptr(s string) *string {
	return &s
}
