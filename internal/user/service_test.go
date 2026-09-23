package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/errorsx"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/user/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockRepository is a testify mock implementing the Repository interface,
// so Service tests never touch a real database.
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Create(ctx context.Context, u *User) (*User, error) {
	args := m.Called(ctx, u)
	out, _ := args.Get(0).(*User)
	return out, args.Error(1)
}

func (m *mockRepository) Update(ctx context.Context, u *User) (*User, error) {
	args := m.Called(ctx, u)
	out, _ := args.Get(0).(*User)
	return out, args.Error(1)
}

func (m *mockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	args := m.Called(ctx, id)
	out, _ := args.Get(0).(*User)
	return out, args.Error(1)
}

func (m *mockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	out, _ := args.Get(0).(*User)
	return out, args.Error(1)
}

func (m *mockRepository) GetMany(ctx context.Context) ([]User, error) {
	args := m.Called(ctx)
	out, _ := args.Get(0).([]User)
	return out, args.Error(1)
}

// NOTE: dto.CreateUserInput.Validate() / UpdateUserInput.Validate() weren't
// shown to me, so the "invalid input" cases below use an empty struct and
// only assert that Validate() rejects it and the repo is never called. If
// your Validate() happens to accept an empty struct, those cases will fail
// loudly (the mock has no expectation set up) rather than silently pass —
// swap in whatever input your real validation rejects.

func TestService_Create(t *testing.T) {
	validInput := &dto.CreateUserInput{
		Name:     "  Subham  ",
		Email:    "  Subham@Example.com  ",
		Password: "SecurePassword123!",
	}

	tests := []struct {
		name      string
		input     *dto.CreateUserInput
		mockSetup func(m *mockRepository)
		wantErr   error // matched with errors.Is; nil means "just require.Error"
		wantErrIs bool  // set true when wantErr should be checked with errors.Is
		check     func(t *testing.T, resp *dto.UserResponseData)
	}{
		{
			name:  "invalid input never reaches the repo",
			input: &dto.CreateUserInput{},
		},
		{
			name:  "trims name, lowercases email, hashes password, maps response",
			input: validInput,
			mockSetup: func(m *mockRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(u *User) bool {
					return u.Name == "Subham" &&
						u.Email == "subham@example.com" &&
						u.Password != "" &&
						u.Password != validInput.Password // must be hashed, not the raw password
				})).Return(&User{
					ID:        uuid.New(),
					Name:      "Subham",
					Email:     "subham@example.com",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
			},
			check: func(t *testing.T, resp *dto.UserResponseData) {
				assert.Equal(t, "Subham", resp.Name)
				assert.Equal(t, "subham@example.com", resp.Email)
				assert.NotEqual(t, uuid.Nil, resp.ID)
			},
		},
		{
			name:  "repo error is propagated",
			input: validInput,
			mockSetup: func(m *mockRepository) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(nil, errorsx.ErrEmailAlreadyExists)
			},
			wantErr:   errorsx.ErrEmailAlreadyExists,
			wantErrIs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}
			svc := NewService(repo)

			resp, err := svc.Create(context.Background(), tt.input)

			if tt.wantErrIs {
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, tt.wantErr)
				repo.AssertExpectations(t)
				return
			}

			if tt.mockSetup == nil {
				// invalid-input case: Validate() must have failed before the repo call
				require.Error(t, err)
				assert.Nil(t, resp)
				repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			if tt.check != nil {
				tt.check(t, resp)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	validID := uuid.New()
	validInput := &dto.UpdateUserInput{Name: "  New Name  ", Email: "  New@Example.com  "}

	tests := []struct {
		name      string
		id        string
		input     *dto.UpdateUserInput
		mockSetup func(m *mockRepository)
		wantErr   error
		check     func(t *testing.T, resp *dto.UserResponseData)
	}{
		{
			name:    "invalid id never reaches validation or the repo",
			id:      "not-a-uuid",
			input:   validInput,
			wantErr: errorsx.ErrInvalidID,
		},
		{
			name:  "invalid input never reaches the repo",
			id:    validID.String(),
			input: &dto.UpdateUserInput{},
		},
		{
			name:  "trims name, lowercases email, maps response",
			id:    validID.String(),
			input: validInput,
			mockSetup: func(m *mockRepository) {
				m.On("Update", mock.Anything, mock.MatchedBy(func(u *User) bool {
					return u.ID == validID && u.Name == "New Name" && u.Email == "new@example.com"
				})).Return(&User{
					ID:        validID,
					Name:      "New Name",
					Email:     "new@example.com",
					UpdatedAt: time.Now(),
				}, nil)
			},
			check: func(t *testing.T, resp *dto.UserResponseData) {
				assert.Equal(t, validID, resp.ID)
				assert.Equal(t, "New Name", resp.Name)
				assert.Equal(t, "new@example.com", resp.Email)
			},
		},
		{
			name:  "repo error is propagated",
			id:    validID.String(),
			input: validInput,
			mockSetup: func(m *mockRepository) {
				m.On("Update", mock.Anything, mock.Anything).
					Return(nil, errorsx.ErrNotFound)
			},
			wantErr: errorsx.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}
			svc := NewService(repo)

			resp, err := svc.Update(context.Background(), tt.id, tt.input)

			switch {
			case tt.wantErr != nil:
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, tt.wantErr)
				repo.AssertExpectations(t)
			case tt.mockSetup == nil:
				// invalid-input case
				require.Error(t, err)
				assert.Nil(t, resp)
				repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			default:
				require.NoError(t, err)
				require.NotNil(t, resp)
				if tt.check != nil {
					tt.check(t, resp)
				}
				repo.AssertExpectations(t)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name      string
		id        string
		mockSetup func(m *mockRepository)
		wantErr   error
	}{
		{
			name:    "invalid id never reaches the repo",
			id:      "not-a-uuid",
			wantErr: errorsx.ErrInvalidID,
		},
		{
			name: "success",
			id:   validID.String(),
			mockSetup: func(m *mockRepository) {
				m.On("Delete", mock.Anything, validID).Return(nil)
			},
		},
		{
			name: "repo error is propagated",
			id:   validID.String(),
			mockSetup: func(m *mockRepository) {
				m.On("Delete", mock.Anything, validID).Return(errorsx.ErrNotFound)
			},
			wantErr: errorsx.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}
			svc := NewService(repo)

			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name      string
		id        string
		mockSetup func(m *mockRepository)
		wantErr   error
		check     func(t *testing.T, resp *dto.UserResponseData)
	}{
		{
			name:    "invalid id never reaches the repo",
			id:      "not-a-uuid",
			wantErr: errorsx.ErrInvalidID,
		},
		{
			name: "success maps response",
			id:   validID.String(),
			mockSetup: func(m *mockRepository) {
				m.On("GetByID", mock.Anything, validID).Return(&User{
					ID:    validID,
					Name:  "Found",
					Email: "found@example.com",
				}, nil)
			},
			check: func(t *testing.T, resp *dto.UserResponseData) {
				assert.Equal(t, validID, resp.ID)
				assert.Equal(t, "Found", resp.Name)
			},
		},
		{
			name: "repo error is propagated",
			id:   validID.String(),
			mockSetup: func(m *mockRepository) {
				m.On("GetByID", mock.Anything, validID).Return(nil, errorsx.ErrNotFound)
			},
			wantErr: errorsx.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}
			svc := NewService(repo)

			resp, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, tt.wantErr)
				repo.AssertExpectations(t)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			if tt.check != nil {
				tt.check(t, resp)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestService_GetByEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		mockSetup func(m *mockRepository)
		wantErr   error
		check     func(t *testing.T, resp *dto.UserResponseData)
	}{
		{
			name:  "success maps response",
			email: "found@example.com",
			mockSetup: func(m *mockRepository) {
				m.On("GetByEmail", mock.Anything, "found@example.com").Return(&User{
					ID:    uuid.New(),
					Name:  "Found",
					Email: "found@example.com",
				}, nil)
			},
			check: func(t *testing.T, resp *dto.UserResponseData) {
				assert.Equal(t, "found@example.com", resp.Email)
			},
		},
		{
			name:  "repo error is propagated",
			email: "missing@example.com",
			mockSetup: func(m *mockRepository) {
				m.On("GetByEmail", mock.Anything, "missing@example.com").
					Return(nil, errorsx.ErrNotFound)
			},
			wantErr: errorsx.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			tt.mockSetup(repo)
			svc := NewService(repo)

			resp, err := svc.GetByEmail(context.Background(), tt.email)

			if tt.wantErr != nil {
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, tt.wantErr)
				repo.AssertExpectations(t)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			if tt.check != nil {
				tt.check(t, resp)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestService_GetMany(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(m *mockRepository)
		wantErr   bool
		wantLen   int
	}{
		{
			name: "empty repo returns empty slice",
			mockSetup: func(m *mockRepository) {
				m.On("GetMany", mock.Anything).Return([]User{}, nil)
			},
			wantLen: 0,
		},
		{
			name: "maps every user",
			mockSetup: func(m *mockRepository) {
				m.On("GetMany", mock.Anything).Return([]User{
					{ID: uuid.New(), Name: "A", Email: "a@example.com"},
					{ID: uuid.New(), Name: "B", Email: "b@example.com"},
				}, nil)
			},
			wantLen: 2,
		},
		{
			name: "repo error is propagated",
			mockSetup: func(m *mockRepository) {
				m.On("GetMany", mock.Anything).Return(nil, errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			tt.mockSetup(repo)
			svc := NewService(repo)

			resp, err := svc.GetMany(context.Background())

			if tt.wantErr {
				assert.Nil(t, resp)
				require.Error(t, err)
				repo.AssertExpectations(t)
				return
			}

			require.NoError(t, err)
			assert.Len(t, resp, tt.wantLen)
			repo.AssertExpectations(t)
		})
	}
}
