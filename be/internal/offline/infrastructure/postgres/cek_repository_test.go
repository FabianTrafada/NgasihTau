package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ngasihtau/internal/offline/domain"
)

// mockDBTX is a mock implementation of DBTX for testing.
type mockDBTX struct {
	execFunc     func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	queryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (m *mockDBTX) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, arguments...)
	}
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (m *mockDBTX) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, sql, args...)
	}
	return nil, nil
}

func (m *mockDBTX) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, sql, args...)
	}
	return &mockRow{}
}

// mockRow implements pgx.Row for testing.
type mockRow struct {
	scanFunc func(dest ...any) error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.scanFunc != nil {
		return m.scanFunc(dest...)
	}
	return pgx.ErrNoRows
}

// mockRows implements pgx.Rows for testing.
type mockRows struct {
	data    [][]any
	current int
	closed  bool
}

func (m *mockRows) Close()                                       { m.closed = true }
func (m *mockRows) Err() error                                   { return nil }
func (m *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *mockRows) RawValues() [][]byte                          { return nil }
func (m *mockRows) Conn() *pgx.Conn                              { return nil }
func (m *mockRows) Values() ([]any, error)                       { return nil, nil }

func (m *mockRows) Next() bool {
	if m.current < len(m.data) {
		m.current++
		return true
	}
	return false
}

func (m *mockRows) Scan(dest ...any) error {
	if m.current == 0 || m.current > len(m.data) {
		return pgx.ErrNoRows
	}
	row := m.data[m.current-1]
	for i, d := range dest {
		if i < len(row) {
			switch v := d.(type) {
			case *uuid.UUID:
				*v = row[i].(uuid.UUID)
			case *[]byte:
				*v = row[i].([]byte)
			case *int:
				*v = row[i].(int)
			case *time.Time:
				*v = row[i].(time.Time)
			}
		}
	}
	return nil
}

func TestNewCEKRepository(t *testing.T) {
	db := &mockDBTX{}
	repo := NewCEKRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestCEKRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful create", func(t *testing.T) {
		db := &mockDBTX{
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				assert.Contains(t, sql, "INSERT INTO offline_ceks")
				assert.Len(t, arguments, 7)
				return pgconn.NewCommandTag("INSERT 0 1"), nil
			},
		}

		repo := NewCEKRepository(db)
		cek := domain.NewContentEncryptionKey(
			uuid.New(),
			uuid.New(),
			uuid.New(),
			[]byte("encrypted-key"),
			1,
		)

		err := repo.Create(ctx, cek)
		require.NoError(t, err)
	})

	t.Run("create error", func(t *testing.T) {
		db := &mockDBTX{
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, assert.AnError
			},
		}

		repo := NewCEKRepository(db)
		cek := domain.NewContentEncryptionKey(
			uuid.New(),
			uuid.New(),
			uuid.New(),
			[]byte("encrypted-key"),
			1,
		)

		err := repo.Create(ctx, cek)
		assert.Error(t, err)
	})
}

func TestCEKRepository_FindByID(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		expectedID := uuid.New()
		expectedUserID := uuid.New()
		expectedMaterialID := uuid.New()
		expectedDeviceID := uuid.New()
		expectedKey := []byte("encrypted-key")
		expectedVersion := 1
		expectedCreatedAt := time.Now()

		db := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				assert.Contains(t, sql, "SELECT")
				assert.Contains(t, sql, "FROM offline_ceks")
				assert.Contains(t, sql, "WHERE id = $1")
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*uuid.UUID) = expectedID
						*dest[1].(*uuid.UUID) = expectedUserID
						*dest[2].(*uuid.UUID) = expectedMaterialID
						*dest[3].(*uuid.UUID) = expectedDeviceID
						*dest[4].(*[]byte) = expectedKey
						*dest[5].(*int) = expectedVersion
						*dest[6].(*time.Time) = expectedCreatedAt
						return nil
					},
				}
			},
		}

		repo := NewCEKRepository(db)
		cek, err := repo.FindByID(ctx, expectedID)

		require.NoError(t, err)
		require.NotNil(t, cek)
		assert.Equal(t, expectedID, cek.ID)
		assert.Equal(t, expectedUserID, cek.UserID)
		assert.Equal(t, expectedMaterialID, cek.MaterialID)
		assert.Equal(t, expectedDeviceID, cek.DeviceID)
		assert.Equal(t, expectedKey, cek.EncryptedKey)
		assert.Equal(t, expectedVersion, cek.KeyVersion)
	})

	t.Run("not found", func(t *testing.T) {
		db := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						return pgx.ErrNoRows
					},
				}
			},
		}

		repo := NewCEKRepository(db)
		cek, err := repo.FindByID(ctx, uuid.New())

		require.NoError(t, err)
		assert.Nil(t, cek)
	})
}

func TestCEKRepository_FindByComposite(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		expectedID := uuid.New()
		userID := uuid.New()
		materialID := uuid.New()
		deviceID := uuid.New()

		db := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				assert.Contains(t, sql, "WHERE user_id = $1 AND material_id = $2 AND device_id = $3")
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*uuid.UUID) = expectedID
						*dest[1].(*uuid.UUID) = userID
						*dest[2].(*uuid.UUID) = materialID
						*dest[3].(*uuid.UUID) = deviceID
						*dest[4].(*[]byte) = []byte("key")
						*dest[5].(*int) = 1
						*dest[6].(*time.Time) = time.Now()
						return nil
					},
				}
			},
		}

		repo := NewCEKRepository(db)
		cek, err := repo.FindByComposite(ctx, userID, materialID, deviceID)

		require.NoError(t, err)
		require.NotNil(t, cek)
		assert.Equal(t, expectedID, cek.ID)
	})

	t.Run("not found", func(t *testing.T) {
		db := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						return pgx.ErrNoRows
					},
				}
			},
		}

		repo := NewCEKRepository(db)
		cek, err := repo.FindByComposite(ctx, uuid.New(), uuid.New(), uuid.New())

		require.NoError(t, err)
		assert.Nil(t, cek)
	})
}

func TestCEKRepository_DeleteByDeviceID(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		deviceID := uuid.New()
		db := &mockDBTX{
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				assert.Contains(t, sql, "DELETE FROM offline_ceks WHERE device_id = $1")
				assert.Equal(t, deviceID, arguments[0])
				return pgconn.NewCommandTag("DELETE 3"), nil
			},
		}

		repo := NewCEKRepository(db)
		err := repo.DeleteByDeviceID(ctx, deviceID)
		require.NoError(t, err)
	})
}

func TestCEKRepository_DeleteByMaterialID(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		materialID := uuid.New()
		db := &mockDBTX{
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				assert.Contains(t, sql, "DELETE FROM offline_ceks WHERE material_id = $1")
				assert.Equal(t, materialID, arguments[0])
				return pgconn.NewCommandTag("DELETE 5"), nil
			},
		}

		repo := NewCEKRepository(db)
		err := repo.DeleteByMaterialID(ctx, materialID)
		require.NoError(t, err)
	})
}

func TestCEKRepository_UpdateKeyVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		cekID := uuid.New()
		newKey := []byte("new-encrypted-key")
		newVersion := 2

		db := &mockDBTX{
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				assert.Contains(t, sql, "UPDATE offline_ceks")
				assert.Contains(t, sql, "SET encrypted_key = $2, key_version = $3")
				assert.Equal(t, cekID, arguments[0])
				assert.Equal(t, newKey, arguments[1])
				assert.Equal(t, newVersion, arguments[2])
				return pgconn.NewCommandTag("UPDATE 1"), nil
			},
		}

		repo := NewCEKRepository(db)
		err := repo.UpdateKeyVersion(ctx, cekID, newKey, newVersion)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		db := &mockDBTX{
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				return pgconn.NewCommandTag("UPDATE 0"), nil
			},
		}

		repo := NewCEKRepository(db)
		err := repo.UpdateKeyVersion(ctx, uuid.New(), []byte("key"), 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cek not found")
	})
}
