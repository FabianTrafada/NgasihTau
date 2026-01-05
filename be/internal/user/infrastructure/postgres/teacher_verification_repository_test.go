// Package postgres contains unit tests for the TeacherVerificationRepository.
// Tests CRUD operations and status updates.
package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"ngasihtau/internal/user/domain"
)

// mockRow implements pgx.Row for testing
type mockRow struct {
	scanFunc func(dest ...any) error
}

func (m *mockRow) Scan(dest ...any) error {
	return m.scanFunc(dest...)
}

// mockRows implements pgx.Rows for testing
type mockRows struct {
	data    [][]any
	current int
	closed  bool
	err     error
}

func (m *mockRows) Close()                        { m.closed = true }
func (m *mockRows) Err() error                    { return m.err }
func (m *mockRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}
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
			case *string:
				*v = row[i].(string)
			case *domain.CredentialType:
				*v = row[i].(domain.CredentialType)
			case *domain.VerificationStatus:
				*v = row[i].(domain.VerificationStatus)
			case **uuid.UUID:
				if row[i] != nil {
					val := row[i].(uuid.UUID)
					*v = &val
				}
			case **time.Time:
				if row[i] != nil {
					val := row[i].(time.Time)
					*v = &val
				}
			case **string:
				if row[i] != nil {
					val := row[i].(string)
					*v = &val
				}
			case *time.Time:
				*v = row[i].(time.Time)
			case *int:
				*v = row[i].(int)
			case *bool:
				*v = row[i].(bool)
			}
		}
	}
	return nil
}
func (m *mockRows) Values() ([]any, error) { return nil, nil }
func (m *mockRows) RawValues() [][]byte    { return nil }
func (m *mockRows) Conn() *pgx.Conn        { return nil }

// mockDBTX implements DBTX interface for testing
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
	return &mockRows{}, nil
}

func (m *mockDBTX) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, sql, args...)
	}
	return &mockRow{scanFunc: func(dest ...any) error { return pgx.ErrNoRows }}
}

// TestTeacherVerificationRepository_Create tests the Create method
func TestTeacherVerificationRepository_Create(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *mockDBTX
		verification *domain.TeacherVerification
		wantErr      bool
	}{
		{
			name: "successful creation",
			setup: func() *mockDBTX {
				return &mockDBTX{
					execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
						return pgconn.NewCommandTag("INSERT 0 1"), nil
					},
				}
			},
			verification: domain.NewTeacherVerification(
				uuid.New(),
				"John Doe",
				"1234567890123456",
				domain.CredentialTypeGovernmentID,
				"ref-doc-123",
			),
			wantErr: false,
		},
		{
			name: "creation with educator card credential",
			setup: func() *mockDBTX {
				return &mockDBTX{
					execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
						return pgconn.NewCommandTag("INSERT 0 1"), nil
					},
				}
			},
			verification: domain.NewTeacherVerification(
				uuid.New(),
				"Jane Smith",
				"EDU-2024-001",
				domain.CredentialTypeEducatorCard,
				"https://storage.example.com/docs/edu-card.pdf",
			),
			wantErr: false,
		},
		{
			name: "creation with professional cert credential",
			setup: func() *mockDBTX {
				return &mockDBTX{
					execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
						return pgconn.NewCommandTag("INSERT 0 1"), nil
					},
				}
			},
			verification: domain.NewTeacherVerification(
				uuid.New(),
				"Bob Wilson",
				"BNSP-2024-12345",
				domain.CredentialTypeProfessionalCert,
				"/documents/bnsp-cert.pdf",
			),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			repo := NewTeacherVerificationRepository(db)

			err := repo.Create(context.Background(), tt.verification)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify ID was set
				if tt.verification.ID == uuid.Nil {
					t.Error("Expected ID to be set")
				}
				// Verify timestamps were set
				if tt.verification.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
				if tt.verification.UpdatedAt.IsZero() {
					t.Error("Expected UpdatedAt to be set")
				}
				// Verify status is pending
				if tt.verification.Status != domain.VerificationStatusPending {
					t.Errorf("Expected status pending, got %s", tt.verification.Status)
				}
			}
		})
	}
}

// TestTeacherVerificationRepository_FindByID tests the FindByID method
func TestTeacherVerificationRepository_FindByID(t *testing.T) {
	verificationID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name    string
		setup   func() *mockDBTX
		id      uuid.UUID
		wantErr bool
	}{
		{
			name: "found verification",
			setup: func() *mockDBTX {
				return &mockDBTX{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFunc: func(dest ...any) error {
								*(dest[0].(*uuid.UUID)) = verificationID
								*(dest[1].(*uuid.UUID)) = userID
								*(dest[2].(*string)) = "John Doe"
								*(dest[3].(*string)) = "1234567890"
								*(dest[4].(*domain.CredentialType)) = domain.CredentialTypeGovernmentID
								*(dest[5].(*string)) = "ref-123"
								*(dest[6].(*domain.VerificationStatus)) = domain.VerificationStatusPending
								*(dest[7].(**uuid.UUID)) = nil
								*(dest[8].(**time.Time)) = nil
								*(dest[9].(**string)) = nil
								*(dest[10].(*time.Time)) = now
								*(dest[11].(*time.Time)) = now
								return nil
							},
						}
					},
				}
			},
			id:      verificationID,
			wantErr: false,
		},
		{
			name: "not found",
			setup: func() *mockDBTX {
				return &mockDBTX{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFunc: func(dest ...any) error {
								return pgx.ErrNoRows
							},
						}
					},
				}
			},
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			repo := NewTeacherVerificationRepository(db)

			result, err := repo.FindByID(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result != nil {
				if result.ID != verificationID {
					t.Errorf("Expected ID %s, got %s", verificationID, result.ID)
				}
				if result.UserID != userID {
					t.Errorf("Expected UserID %s, got %s", userID, result.UserID)
				}
			}
		})
	}
}

// TestTeacherVerificationRepository_FindByUserID tests the FindByUserID method
func TestTeacherVerificationRepository_FindByUserID(t *testing.T) {
	verificationID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name    string
		setup   func() *mockDBTX
		userID  uuid.UUID
		wantErr bool
	}{
		{
			name: "found verification by user ID",
			setup: func() *mockDBTX {
				return &mockDBTX{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFunc: func(dest ...any) error {
								*(dest[0].(*uuid.UUID)) = verificationID
								*(dest[1].(*uuid.UUID)) = userID
								*(dest[2].(*string)) = "John Doe"
								*(dest[3].(*string)) = "1234567890"
								*(dest[4].(*domain.CredentialType)) = domain.CredentialTypeGovernmentID
								*(dest[5].(*string)) = "ref-123"
								*(dest[6].(*domain.VerificationStatus)) = domain.VerificationStatusPending
								*(dest[7].(**uuid.UUID)) = nil
								*(dest[8].(**time.Time)) = nil
								*(dest[9].(**string)) = nil
								*(dest[10].(*time.Time)) = now
								*(dest[11].(*time.Time)) = now
								return nil
							},
						}
					},
				}
			},
			userID:  userID,
			wantErr: false,
		},
		{
			name: "not found by user ID",
			setup: func() *mockDBTX {
				return &mockDBTX{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFunc: func(dest ...any) error {
								return pgx.ErrNoRows
							},
						}
					},
				}
			},
			userID:  uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			repo := NewTeacherVerificationRepository(db)

			result, err := repo.FindByUserID(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByUserID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result != nil {
				if result.UserID != userID {
					t.Errorf("Expected UserID %s, got %s", userID, result.UserID)
				}
			}
		})
	}
}

// TestTeacherVerificationRepository_UpdateStatus tests the UpdateStatus method
func TestTeacherVerificationRepository_UpdateStatus(t *testing.T) {
	verificationID := uuid.New()
	reviewerID := uuid.New()
	rejectionReason := "Invalid document"

	tests := []struct {
		name       string
		setup      func() *mockDBTX
		id         uuid.UUID
		status     domain.VerificationStatus
		reviewedBy uuid.UUID
		reason     *string
		wantErr    bool
	}{
		{
			name: "approve verification",
			setup: func() *mockDBTX {
				return &mockDBTX{
					execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
						return pgconn.NewCommandTag("UPDATE 1"), nil
					},
				}
			},
			id:         verificationID,
			status:     domain.VerificationStatusApproved,
			reviewedBy: reviewerID,
			reason:     nil,
			wantErr:    false,
		},
		{
			name: "reject verification with reason",
			setup: func() *mockDBTX {
				return &mockDBTX{
					execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
						return pgconn.NewCommandTag("UPDATE 1"), nil
					},
				}
			},
			id:         verificationID,
			status:     domain.VerificationStatusRejected,
			reviewedBy: reviewerID,
			reason:     &rejectionReason,
			wantErr:    false,
		},
		{
			name: "verification not found",
			setup: func() *mockDBTX {
				return &mockDBTX{
					execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
						return pgconn.NewCommandTag("UPDATE 0"), nil
					},
				}
			},
			id:         uuid.New(),
			status:     domain.VerificationStatusApproved,
			reviewedBy: reviewerID,
			reason:     nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			repo := NewTeacherVerificationRepository(db)

			err := repo.UpdateStatus(context.Background(), tt.id, tt.status, tt.reviewedBy, tt.reason)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTeacherVerificationRepository_Update tests the Update method
func TestTeacherVerificationRepository_Update(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *mockDBTX
		verification *domain.TeacherVerification
		wantErr      bool
	}{
		{
			name: "successful update",
			setup: func() *mockDBTX {
				return &mockDBTX{
					execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
						return pgconn.NewCommandTag("UPDATE 1"), nil
					},
				}
			},
			verification: &domain.TeacherVerification{
				ID:             uuid.New(),
				UserID:         uuid.New(),
				FullName:       "Updated Name",
				IDNumber:       "9876543210",
				CredentialType: domain.CredentialTypeEducatorCard,
				DocumentRef:    "updated-ref",
				Status:         domain.VerificationStatusPending,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			wantErr: false,
		},
		{
			name: "verification not found",
			setup: func() *mockDBTX {
				return &mockDBTX{
					execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
						return pgconn.NewCommandTag("UPDATE 0"), nil
					},
				}
			},
			verification: &domain.TeacherVerification{
				ID:             uuid.New(),
				UserID:         uuid.New(),
				FullName:       "Test",
				IDNumber:       "123",
				CredentialType: domain.CredentialTypeGovernmentID,
				DocumentRef:    "ref",
				Status:         domain.VerificationStatusPending,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			repo := NewTeacherVerificationRepository(db)

			err := repo.Update(context.Background(), tt.verification)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTeacherVerificationRepository_ExistsByUserID tests the ExistsByUserID method
func TestTeacherVerificationRepository_ExistsByUserID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *mockDBTX
		userID  uuid.UUID
		want    bool
		wantErr bool
	}{
		{
			name: "exists",
			setup: func() *mockDBTX {
				return &mockDBTX{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFunc: func(dest ...any) error {
								*(dest[0].(*bool)) = true
								return nil
							},
						}
					},
				}
			},
			userID:  uuid.New(),
			want:    true,
			wantErr: false,
		},
		{
			name: "does not exist",
			setup: func() *mockDBTX {
				return &mockDBTX{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFunc: func(dest ...any) error {
								*(dest[0].(*bool)) = false
								return nil
							},
						}
					},
				}
			},
			userID:  uuid.New(),
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			repo := NewTeacherVerificationRepository(db)

			got, err := repo.ExistsByUserID(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExistsByUserID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ExistsByUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTeacherVerificationRepository_ExistsPendingByUserID tests the ExistsPendingByUserID method
func TestTeacherVerificationRepository_ExistsPendingByUserID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *mockDBTX
		userID  uuid.UUID
		want    bool
		wantErr bool
	}{
		{
			name: "pending exists",
			setup: func() *mockDBTX {
				return &mockDBTX{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFunc: func(dest ...any) error {
								*(dest[0].(*bool)) = true
								return nil
							},
						}
					},
				}
			},
			userID:  uuid.New(),
			want:    true,
			wantErr: false,
		},
		{
			name: "no pending exists",
			setup: func() *mockDBTX {
				return &mockDBTX{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFunc: func(dest ...any) error {
								*(dest[0].(*bool)) = false
								return nil
							},
						}
					},
				}
			},
			userID:  uuid.New(),
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			repo := NewTeacherVerificationRepository(db)

			got, err := repo.ExistsPendingByUserID(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExistsPendingByUserID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ExistsPendingByUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTeacherVerificationRepository_FindPending tests the FindPending method
func TestTeacherVerificationRepository_FindPending(t *testing.T) {
	verificationID1 := uuid.New()
	verificationID2 := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name      string
		setup     func() *mockDBTX
		limit     int
		offset    int
		wantCount int
		wantTotal int
		wantErr   bool
	}{
		{
			name: "found pending verifications",
			setup: func() *mockDBTX {
				return &mockDBTX{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFunc: func(dest ...any) error {
								*(dest[0].(*int)) = 2
								return nil
							},
						}
					},
					queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
						return &mockRows{
							data: [][]any{
								{verificationID1, userID1, "John Doe", "123", domain.CredentialTypeGovernmentID, "ref1", domain.VerificationStatusPending, nil, nil, nil, now, now},
								{verificationID2, userID2, "Jane Smith", "456", domain.CredentialTypeEducatorCard, "ref2", domain.VerificationStatusPending, nil, nil, nil, now, now},
							},
						}, nil
					},
				}
			},
			limit:     10,
			offset:    0,
			wantCount: 2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name: "no pending verifications",
			setup: func() *mockDBTX {
				return &mockDBTX{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFunc: func(dest ...any) error {
								*(dest[0].(*int)) = 0
								return nil
							},
						}
					},
					queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
						return &mockRows{data: [][]any{}}, nil
					},
				}
			},
			limit:     10,
			offset:    0,
			wantCount: 0,
			wantTotal: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			repo := NewTeacherVerificationRepository(db)

			results, total, err := repo.FindPending(context.Background(), tt.limit, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindPending() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(results) != tt.wantCount {
				t.Errorf("FindPending() count = %v, want %v", len(results), tt.wantCount)
			}
			if total != tt.wantTotal {
				t.Errorf("FindPending() total = %v, want %v", total, tt.wantTotal)
			}
		})
	}
}
