package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"snippets.adelh.dev/app/internal/db/mocks"
	"snippets.adelh.dev/app/internal/db/sqlc"
	"snippets.adelh.dev/app/internal/encryption"
)

//go:generate sh -c "cd ../../.. && mockery"

func TestSnippetService_GetSnippet(t *testing.T) {
	encryptionSvc, err := encryption.NewService("U2FsdGVkX1/K0w2X9/0jJeJk+nGGchmRtIpC/FP4YI0=")
	if err != nil {
		t.Fatal(err)
	}

	baseSnippet := sqlc.GetSnippetByPublicIDRow{
		ID:           1,
		PublicID:     "test-id",
		Title:        sql.NullString{String: "test", Valid: true},
		CreatedAt:    time.Now(),
		ExpiresAt:    sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
		EditToken:    "token",
		ViewCount:    0,
		LastEditedAt: sql.NullTime{},
		ContentType:  "text/plain",
	}

	plainContent := []byte("some content")
	encryptedContent, err := encryptionSvc.Encrypt(plainContent)
	if err != nil {
		t.Fatalf("failed to encrypt content: %s", err.Error())
	}

	password := "secret123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %s", err)
	}

	tests := []struct {
		name           string
		snippet        sqlc.GetSnippetByPublicIDRow
		params         GetSnippetParams
		expectedStatus int
		expectBody     bool
	}{
		{
			name: "Success - No Password",
			snippet: func() sqlc.GetSnippetByPublicIDRow {
				s := baseSnippet
				s.PasswordHash = sql.NullString{Valid: false}
				s.EncryptedContent = encryptedContent
				return s
			}(),
			params:         GetSnippetParams{},
			expectedStatus: http.StatusOK,
			expectBody:     true,
		},
		{
			name: "Success - With Correct Password",
			snippet: func() sqlc.GetSnippetByPublicIDRow {
				s := baseSnippet
				s.PasswordHash = sql.NullString{String: string(passwordHash), Valid: true}
				s.EncryptedContent = encryptedContent
				return s
			}(),
			params: GetSnippetParams{
				XSnippetPassword: &password,
			},
			expectedStatus: http.StatusOK,
			expectBody:     true,
		},
		{
			name: "Failure - Password Required",
			snippet: func() sqlc.GetSnippetByPublicIDRow {
				s := baseSnippet
				s.PasswordHash = sql.NullString{String: string(passwordHash), Valid: true}
				s.EncryptedContent = encryptedContent
				return s
			}(),
			params:         GetSnippetParams{},
			expectedStatus: http.StatusForbidden,
			expectBody:     false,
		},
		{
			name: "Failure - Invalid Password",
			snippet: func() sqlc.GetSnippetByPublicIDRow {
				s := baseSnippet
				s.PasswordHash = sql.NullString{String: string(passwordHash), Valid: true}
				s.EncryptedContent = encryptedContent
				return s
			}(),
			params: GetSnippetParams{
				XSnippetPassword: func() *string { s := "wrongpass"; return &s }(),
			},
			expectedStatus: http.StatusForbidden,
			expectBody:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			mockQuerier := mocks.NewMockQuerier(t)

			apiService := New(mockStore, encryptionSvc)

			mockStore.EXPECT().Replica().Return(mockQuerier)
			mockQuerier.EXPECT().GetSnippetByPublicID(mock.Anything, "test-id").Return(tc.snippet, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/snippets/test-id", nil)
			w := httptest.NewRecorder()

			apiService.GetSnippet(w, req, "test-id", tc.params)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			if tc.expectBody {
				var snippetResp SnippetResponse
				err := json.NewDecoder(resp.Body).Decode(&snippetResp)
				assert.NoError(t, err, "should decode response body")
				assert.Equal(t, "test-id", snippetResp.Id)
				assert.Equal(t, string(plainContent), snippetResp.Content)

			}
		})
	}
}

func TestSnippetService_DeleteSnippet(t *testing.T) {
	encryptionSvc, err := encryption.NewService("U2FsdGVkX1/K0w2X9/0jJeJk+nGGchmRtIpC/FP4YI0=")
	if err != nil {
		t.Fatal(err)
	}

	baseSnippet := sqlc.GetSnippetByPublicIDRow{
		ID:           1,
		PublicID:     "test-id",
		Title:        sql.NullString{String: "test", Valid: true},
		CreatedAt:    time.Now(),
		ExpiresAt:    sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
		EditToken:    "token",
		PasswordHash: sql.NullString{},
		ViewCount:    0,
		LastEditedAt: sql.NullTime{},
		ContentType:  "text/plain",
	}

	tests := []struct {
		name           string
		id             string
		params         DeleteSnippetParams
		snippet        sqlc.GetSnippetByPublicIDRow
		setupMocks     func(s sqlc.GetSnippetByPublicIDRow, store *mocks.MockStore, mockQuerier *mocks.MockQuerier)
		expectedStatus int
	}{
		{
			name:    "Delete Success",
			id:      "test-id",
			params:  DeleteSnippetParams{XEditToken: "token"},
			snippet: baseSnippet,
			setupMocks: func(s sqlc.GetSnippetByPublicIDRow, store *mocks.MockStore, mockQuerier *mocks.MockQuerier) {
				store.EXPECT().Primary().Return(mockQuerier)
				mockQuerier.EXPECT().GetSnippetByPublicID(mock.Anything, "test-id").Return(s, nil)
				mockQuerier.EXPECT().DeleteSnippetById(mock.Anything, s.ID).Return(1, nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:    "Delete 404",
			id:      "test-id",
			params:  DeleteSnippetParams{XEditToken: "token"},
			snippet: baseSnippet,
			setupMocks: func(s sqlc.GetSnippetByPublicIDRow, store *mocks.MockStore, mockQuerier *mocks.MockQuerier) {
				store.EXPECT().Primary().Return(mockQuerier)
				mockQuerier.EXPECT().GetSnippetByPublicID(mock.Anything, "test-id").Return(s, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "Delete 401",
			id:      "test-id",
			params:  DeleteSnippetParams{XEditToken: "wrong-token"},
			snippet: baseSnippet,
			setupMocks: func(s sqlc.GetSnippetByPublicIDRow, store *mocks.MockStore, mockQuerier *mocks.MockQuerier) {
				store.EXPECT().Primary().Return(mockQuerier)
				mockQuerier.EXPECT().GetSnippetByPublicID(mock.Anything, "test-id").Return(s, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQuerier := mocks.NewMockQuerier(t)
			store := mocks.NewMockStore(t)
			tt.setupMocks(tt.snippet, store, mockQuerier)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodDelete, "/api/snippets/"+tt.id, nil)
			s := &SnippetService{store: store, enc: encryptionSvc}
			s.DeleteSnippet(w, r, tt.id, tt.params)
			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
		})
	}
}
