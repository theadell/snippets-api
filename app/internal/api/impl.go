package api

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"snippets.adelh.dev/app/internal/cache"
	"snippets.adelh.dev/app/internal/db"
	"snippets.adelh.dev/app/internal/db/sqlc"
	"snippets.adelh.dev/app/internal/encryption"
)

//go:generate go tool oapi-codegen -config cfg.yaml ../../../openapi-spec/openapi.yaml

type SnippetService struct {
	store      db.Store
	redisCache *cache.RedisCache
	enc        *encryption.Service
}

var _ ServerInterface = (*SnippetService)(nil)

func New(store db.Store, encryptionService *encryption.Service, redisCache *cache.RedisCache) *SnippetService {
	return &SnippetService{
		enc:        encryptionService,
		store:      store,
		redisCache: redisCache,
	}
}

func (s *SnippetService) GetSnippet(w http.ResponseWriter, r *http.Request, id string, params GetSnippetParams) {
	snippet, err := s.getAndValidateSnippet(w, r, id)
	if err != nil {
		return
	}

	if snippet.PasswordHash.Valid {
		if params.XSnippetPassword == nil {
			forbiddenError(w, r, "Password required")
			return
		}
		if passwordErr := bcrypt.CompareHashAndPassword([]byte(snippet.PasswordHash.String), []byte(*params.XSnippetPassword)); passwordErr != nil {
			forbiddenError(w, r, "Invalid password")
			return
		}
	}
	content, err := s.enc.Decrypt(snippet.EncryptedContent)
	if err != nil {
		internalServerError(w, r, fmt.Errorf("failed to decrypt snippet: %w", err))
		return
	}

	snippetDTO := SnippetResponse{
		Title:       stringPtr(snippet.Title),
		ContentType: &snippet.ContentType,
		Content:     string(content),
		CreatedAt:   snippet.CreatedAt,
		ExpiresAt:   &snippet.ExpiresAt.Time,
		Id:          snippet.PublicID,
	}

	ok(w, snippetDTO)
}

func (s *SnippetService) CreateSnippet(w http.ResponseWriter, r *http.Request) {
	var req SnippetCreateRequest

	if err := readJSON(w, r, &req); err != nil {
		badRequestError(w, r, err.Error())
		return
	}

	password := toNullString(req.Password)
	if password.Valid {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			internalServerError(w, r, fmt.Errorf("failed to hash password: %w", err))
			return
		}
		password.String = string(hash)
	}

	encryptedData, err := s.enc.Encrypt([]byte(req.Content))
	if err != nil {
		internalServerError(w, r, fmt.Errorf("failed to encrypt content: %w", err))
		return
	}

	title := toNullString(req.Title)

	expiresAt, err := parseExpiresIn(req.ExpiresIn)
	if err != nil {
		badRequestError(w, r, err.Error())
		return
	}

	contentType := stringValue(req.ContentType, "text/plain")

	result, err := s.store.Primary().CreateSnippet(r.Context(), sqlc.CreateSnippetParams{
		Title:            title,
		ExpiresAt:        expiresAt,
		PasswordHash:     password,
		ContentType:      contentType,
		EncryptedContent: encryptedData,
		EditToken:        generateEditToken(),
	})
	if err != nil {
		slog.Error("failed to create snippet", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := SnippetCreateResponse{
		ExpiresAt: &expiresAt.Time,
		Id:        result.PublicID,
		EditToken: &result.EditToken,
	}
	ok(w, response)
}

func (s *SnippetService) UpdateSnippet(w http.ResponseWriter, r *http.Request, id string, params UpdateSnippetParams) {
	snippet, err := s.getAndValidateSnippet(w, r, id)
	if err != nil {
		return
	}

	if params.XEditToken != snippet.EditToken {
		unauthorizedError(w, r, "Invalid edit token")
		return
	}

	var req SnippetCreateRequest
	if err := readJSON(w, r, &req); err != nil {
		badRequestError(w, r, err.Error())
		return
	}
	title := toNullString(req.Title)

	expiresAt, err := parseExpiresIn(req.ExpiresIn)
	if err != nil {
		badRequestError(w, r, err.Error())
		return
	}

	encryptedData, err := s.enc.Encrypt([]byte(req.Content))
	if err != nil {
		internalServerError(w, r, fmt.Errorf("failed to encrypt content: %w", err))
		return
	}

	contentType := stringValue(req.ContentType, snippet.ContentType)

	err = s.store.WithTx(r.Context(), func(q sqlc.Querier) error {
		updateParams := sqlc.UpdateSnippetParams{
			ID:        snippet.ID,
			Title:     title,
			ExpiresAt: expiresAt,
		}

		_, err := q.UpdateSnippet(r.Context(), updateParams)
		if err != nil {
			return fmt.Errorf("failed to update snippet metadata: %w", err)
		}

		contentParams := sqlc.UpdateSnippetContentParams{
			SnippetID:        snippet.ID,
			ContentType:      contentType,
			EncryptedContent: encryptedData,
		}

		err = q.UpdateSnippetContent(r.Context(), contentParams)
		if err != nil {
			return fmt.Errorf("failed to update snippet content: %w", err)
		}
		s.redisCache.Delete(r.Context(), fmt.Sprintf("snippet:%s", id))
		return nil
	})
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	updatedSnippet, err := s.store.Primary().GetSnippetByPublicID(r.Context(), id)
	if err != nil {
		internalServerError(w, r, fmt.Errorf("failed to retrieve updated snippet: %w", err))
		return
	}

	content, err := s.enc.Decrypt(updatedSnippet.EncryptedContent)
	if err != nil {
		internalServerError(w, r, fmt.Errorf("failed to decrypt content: %w", err))
		return
	}

	snippetDTO := SnippetResponse{
		Title:       stringPtr(snippet.Title),
		ContentType: &snippet.ContentType,
		Content:     string(content),
		CreatedAt:   snippet.CreatedAt,
		ExpiresAt:   &snippet.ExpiresAt.Time,
		Id:          snippet.PublicID,
	}
	ok(w, snippetDTO)
}

func (s *SnippetService) DeleteSnippet(w http.ResponseWriter, r *http.Request, id string, params DeleteSnippetParams) {
	snippet, err := s.getAndValidateSnippet(w, r, id)
	if err != nil {
		return
	}

	if params.XEditToken != snippet.EditToken {
		unauthorizedError(w, r, "Invalid edit token")
		return
	}
	_, err = s.store.Primary().DeleteSnippetById(r.Context(), snippet.ID)
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getAndValidateSnippet retrieves a snippet and checks if it's valid and not expired
// If primary is true, it uses the primary database, otherwise it uses a replica
func (s *SnippetService) getAndValidateSnippet(w http.ResponseWriter, r *http.Request, publicID string) (*sqlc.GetSnippetByPublicIDRow, error) {
	var snippet sqlc.GetSnippetByPublicIDRow
	var err error
	var cacheHit bool
	cacheKey := fmt.Sprintf("snippet:%s", publicID)

	cacheHit = s.redisCache.Get(r.Context(), cacheKey, &snippet)

	if !cacheHit {
		snippet, err = s.store.Replica().GetSnippetByPublicID(r.Context(), publicID)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFoundError(w, r, "Snippet not found")
			return nil, err
		}
		internalServerError(w, r, fmt.Errorf("failed to retrieve snippet: %w", err))
		return nil, err
	}

	if snippet.ExpiresAt.Valid && snippet.ExpiresAt.Time.Before(time.Now()) {
		notFoundError(w, r, "Snippet has expired")
		return nil, fmt.Errorf("snippet has expired")
	}

	if !cacheHit {
		s.redisCache.Set(r.Context(), cacheKey, snippet)
	}

	return &snippet, nil
}
