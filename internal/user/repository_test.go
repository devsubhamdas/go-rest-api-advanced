package user

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/errorsx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB

// TestMain starts one Postgres container for the whole package.
// NOTE: do NOT set gorm.Config{TranslateError: true} here (or in prod) —
// the repository matches on *pgconn.PgError, and translation would hide it.
func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := tcpostgres.Run(ctx, "postgres:18-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres container: %v\n", err)
		os.Exit(1)
	}

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection string: %v\n", err)
		_ = pg.Terminate(ctx)
		os.Exit(1)
	}

	testDB, err = gorm.Open(gormpostgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "open gorm: %v\n", err)
		_ = pg.Terminate(ctx)
		os.Exit(1)
	}

	if err := testDB.AutoMigrate(&User{}); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		_ = pg.Terminate(ctx)
		os.Exit(1)
	}

	code := m.Run()

	_ = pg.Terminate(ctx)
	os.Exit(code)
}

// newTestRepo gives every test a clean users table.
func newTestRepo(t *testing.T) Repository {
	t.Helper()
	require.NoError(t, testDB.Exec("TRUNCATE TABLE users").Error)
	return NewRepository(testDB)
}

func seedUser(t *testing.T, repo Repository, name, email string) *User {
	t.Helper()
	u, err := repo.Create(context.Background(), &User{
		Name:     name,
		Email:    email,
		Password: "1234",
	})
	require.NoError(t, err)
	return u
}

// assertMappedErr checks err against exactly one of: a specific errorsx
// sentinel (wantErr), or "some other, unmapped error" (wantOtherErr).
func assertMappedErr(t *testing.T, err error, wantErr error, wantOtherErr bool) {
	t.Helper()

	if wantErr != nil {
		assert.ErrorIs(t, err, wantErr)
		return
	}

	if wantOtherErr {
		require.Error(t, err)
		assert.NotErrorIs(t, err, errorsx.ErrEmailAlreadyExists)
		assert.NotErrorIs(t, err, errorsx.ErrDuplicateKey)
		assert.NotErrorIs(t, err, errorsx.ErrNotFound)
		return
	}

	assert.NoError(t, err)
}

func TestRepository_Create(t *testing.T) {
	tests := []struct {
		name string
		// setup seeds any prerequisite rows and returns the user passed to Create.
		setup func(t *testing.T, repo Repository) *User
		// cancelCtx passes an already-cancelled context to Create.
		cancelCtx bool
		// wantErr is matched with errors.Is; nil + wantOtherErr=false means success.
		wantErr      error
		wantOtherErr bool
		// check runs extra assertions on the returned user (success cases only).
		check func(t *testing.T, got *User)
	}{
		{
			name: "success generates id and timestamps",
			setup: func(t *testing.T, repo Repository) *User {
				return &User{Name: "Test User", Email: "test@example.com", Password: "1234"}
			},
			check: func(t *testing.T, got *User) {
				assert.NotEqual(t, uuid.Nil, got.ID)
				assert.False(t, got.CreatedAt.IsZero())
				assert.False(t, got.UpdatedAt.IsZero())
			},
		},
		{
			name: "duplicate email returns ErrEmailAlreadyExists",
			setup: func(t *testing.T, repo Repository) *User {
				seedUser(t, repo, "First", "dup@example.com")
				return &User{Name: "Second", Email: "dup@example.com", Password: "1234"}
			},
			wantErr: errorsx.ErrEmailAlreadyExists,
		},
		{
			name: "duplicate primary key returns ErrDuplicateKey",
			setup: func(t *testing.T, repo Repository) *User {
				first := seedUser(t, repo, "First", "one@example.com")
				return &User{ID: first.ID, Name: "Second", Email: "two@example.com", Password: "hash"}
			},
			wantErr: errorsx.ErrDuplicateKey,
		},
		{
			name: "unmapped db error is returned as-is",
			setup: func(t *testing.T, repo Repository) *User {
				return &User{Name: "X", Email: "x@example.com", Password: "1234"}
			},
			cancelCtx:    true,
			wantOtherErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestRepo(t)

			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			input := tt.setup(t, repo)
			got, err := repo.Create(ctx, input)

			if tt.wantErr != nil || tt.wantOtherErr {
				assert.Nil(t, got)
				assertMappedErr(t, err, tt.wantErr, tt.wantOtherErr)
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestRepository_Update(t *testing.T) {
	tests := []struct {
		name string
		// setup seeds prerequisite rows and returns the user passed to Update.
		setup        func(t *testing.T, repo Repository) *User
		wantErr      error
		wantOtherErr bool
		// check runs extra assertions by re-fetching the row (success cases only).
		check func(t *testing.T, repo Repository, id uuid.UUID)
	}{
		{
			name: "updates name and email but not password",
			setup: func(t *testing.T, repo Repository) *User {
				u := seedUser(t, repo, "Old Name", "old@example.com")
				u.Name = "New Name"
				u.Email = "new@example.com"
				u.Password = "should-not-be-saved"
				return u
			},
			check: func(t *testing.T, repo Repository, id uuid.UUID) {
				got, err := repo.GetByID(context.Background(), id)
				require.NoError(t, err)
				assert.Equal(t, "New Name", got.Name)
				assert.Equal(t, "new@example.com", got.Email)
				assert.Equal(t, "1234", got.Password, "Select(name, email) must leave password untouched")
			},
		},
		{
			name: "duplicate email returns ErrEmailAlreadyExists",
			setup: func(t *testing.T, repo Repository) *User {
				first := seedUser(t, repo, "First", "first@example.com")
				second := seedUser(t, repo, "Second", "second@example.com")
				second.Email = first.Email
				return second
			},
			wantErr: errorsx.ErrEmailAlreadyExists,
		},
		{
			name: "non-existent user returns ErrNotFound",
			setup: func(t *testing.T, repo Repository) *User {
				return &User{ID: uuid.New(), Name: "Ghost", Email: "ghost@example.com"}
			},
			wantErr: errorsx.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestRepo(t)

			input := tt.setup(t, repo)
			id := input.ID
			got, err := repo.Update(context.Background(), input)

			if tt.wantErr != nil || tt.wantOtherErr {
				assert.Nil(t, got)
				assertMappedErr(t, err, tt.wantErr, tt.wantOtherErr)
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, repo, id)
			}
		})
	}
}

func TestRepository_Delete(t *testing.T) {
	tests := []struct {
		name string
		// setup seeds prerequisite rows and returns the id passed to Delete.
		setup   func(t *testing.T, repo Repository) uuid.UUID
		wantErr error
		// check runs extra assertions after Delete (e.g. other rows untouched).
		check func(t *testing.T, repo Repository)
	}{
		{
			name: "removes the row",
			setup: func(t *testing.T, repo Repository) uuid.UUID {
				u := seedUser(t, repo, "Del", "del@example.com")
				return u.ID
			},
		},
		{
			name: "non-existent id returns no error",
			setup: func(t *testing.T, repo Repository) uuid.UUID {
				return uuid.New()
			},
		},
		{
			name: "only deletes the target user",
			setup: func(t *testing.T, repo Repository) uuid.UUID {
				target := seedUser(t, repo, "Target", "target@example.com")
				seedUser(t, repo, "Other", "other@example.com")
				return target.ID
			},
			check: func(t *testing.T, repo Repository) {
				users, err := repo.GetMany(context.Background())
				require.NoError(t, err)
				require.Len(t, users, 1)
				assert.Equal(t, "other@example.com", users[0].Email)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestRepo(t)

			id := tt.setup(t, repo)
			err := repo.Delete(context.Background(), id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)

			_, getErr := repo.GetByID(context.Background(), id)
			assert.ErrorIs(t, getErr, errorsx.ErrNotFound)

			if tt.check != nil {
				tt.check(t, repo)
			}
		})
	}
}

func TestRepository_GetByID(t *testing.T) {
	tests := []struct {
		name string
		// setup seeds prerequisite rows and returns the id to look up.
		setup   func(t *testing.T, repo Repository) uuid.UUID
		wantErr error
		check   func(t *testing.T, got *User, id uuid.UUID)
	}{
		{
			name: "found",
			setup: func(t *testing.T, repo Repository) uuid.UUID {
				return seedUser(t, repo, "Found", "found@example.com").ID
			},
			check: func(t *testing.T, got *User, id uuid.UUID) {
				assert.Equal(t, id, got.ID)
				assert.Equal(t, "Found", got.Name)
				assert.Equal(t, "found@example.com", got.Email)
			},
		},
		{
			name: "not found returns ErrNotFound",
			setup: func(t *testing.T, repo Repository) uuid.UUID {
				return uuid.New()
			},
			wantErr: errorsx.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestRepo(t)

			id := tt.setup(t, repo)
			got, err := repo.GetByID(context.Background(), id)

			if tt.wantErr != nil {
				assert.Nil(t, got)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, got, id)
			}
		})
	}
}

func TestRepository_GetByEmail(t *testing.T) {
	tests := []struct {
		name string
		// setup seeds prerequisite rows and returns the email to look up.
		setup   func(t *testing.T, repo Repository) string
		wantErr error
		check   func(t *testing.T, got *User)
	}{
		{
			name: "found",
			setup: func(t *testing.T, repo Repository) string {
				seedUser(t, repo, "Mail", "mail@example.com")
				return "mail@example.com"
			},
			check: func(t *testing.T, got *User) {
				assert.Equal(t, "mail@example.com", got.Email)
			},
		},
		{
			name: "not found returns ErrNotFound",
			setup: func(t *testing.T, repo Repository) string {
				return "nobody@example.com"
			},
			wantErr: errorsx.ErrNotFound,
		},
		{
			// Documents current behavior: lookup is case-sensitive, so
			// normalize (lowercase) emails in the service layer.
			name: "lookup is case-sensitive",
			setup: func(t *testing.T, repo Repository) string {
				seedUser(t, repo, "Case", "case@example.com")
				return "CASE@example.com"
			},
			wantErr: errorsx.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestRepo(t)

			email := tt.setup(t, repo)
			got, err := repo.GetByEmail(context.Background(), email)

			if tt.wantErr != nil {
				assert.Nil(t, got)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestRepository_GetMany(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, repo Repository) []uuid.UUID
	}{
		{
			name: "empty table returns empty slice",
			setup: func(t *testing.T, repo Repository) []uuid.UUID {
				return nil
			},
		},
		{
			name: "returns all users",
			setup: func(t *testing.T, repo Repository) []uuid.UUID {
				a := seedUser(t, repo, "A", "a@example.com")
				b := seedUser(t, repo, "B", "b@example.com")
				return []uuid.UUID{a.ID, b.ID}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestRepo(t)

			want := tt.setup(t, repo)
			users, err := repo.GetMany(context.Background())

			require.NoError(t, err)
			require.Len(t, users, len(want))

			var got []uuid.UUID
			for _, u := range users {
				got = append(got, u.ID)
			}
			assert.ElementsMatch(t, want, got)
		})
	}
}
