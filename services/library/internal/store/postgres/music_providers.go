package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

const providerConnectionColumns = `provider_id, status, account_id, display_name, avatar_url, membership,
	membership_expire_time, encrypted_credentials, create_time, update_time`

const providerAttemptColumns = `uid, provider_id, status, qr_image, qr_content_type, qr_payload, authorization_url,
	state_hash, encrypted_state, create_time, update_time, expire_time, consumed_time`

// ListProviderConnections returns all configured external accounts without decrypting them.
func (s *Store) ListProviderConnections(ctx context.Context) ([]domain.ProviderConnection, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+providerConnectionColumns+" FROM music_provider_connections ORDER BY provider_id")
	if err != nil {
		return nil, fmt.Errorf("list music provider connections: %w", err)
	}
	defer rows.Close()
	connections := make([]domain.ProviderConnection, 0, 3)
	for rows.Next() {
		connection, err := scanProviderConnection(rows)
		if err != nil {
			return nil, err
		}
		connections = append(connections, connection)
	}
	return connections, rows.Err()
}

// GetProviderConnection returns one external account including encrypted credentials.
func (s *Store) GetProviderConnection(ctx context.Context, provider domain.MusicProvider) (domain.ProviderConnection, error) {
	connection, err := scanProviderConnection(s.pool.QueryRow(ctx,
		"SELECT "+providerConnectionColumns+" FROM music_provider_connections WHERE provider_id = $1", provider))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ProviderConnection{}, fmt.Errorf("%w: music provider connection", domain.ErrNotFound)
	}
	return connection, err
}

// UpsertProviderConnection replaces the single account attached to a provider.
func (s *Store) UpsertProviderConnection(ctx context.Context, connection domain.ProviderConnection) (domain.ProviderConnection, error) {
	query := `INSERT INTO music_provider_connections (provider_id, status, account_id, display_name, avatar_url,
		membership, membership_expire_time, encrypted_credentials, create_time, update_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (provider_id) DO UPDATE SET status = EXCLUDED.status, account_id = EXCLUDED.account_id,
		display_name = EXCLUDED.display_name, avatar_url = EXCLUDED.avatar_url, membership = EXCLUDED.membership,
		membership_expire_time = EXCLUDED.membership_expire_time,
		encrypted_credentials = EXCLUDED.encrypted_credentials, update_time = EXCLUDED.update_time
		RETURNING ` + providerConnectionColumns
	stored, err := scanProviderConnection(s.pool.QueryRow(ctx, query, connection.Provider, connection.Status,
		connection.AccountID, connection.DisplayName, connection.AvatarURL, connection.Membership,
		connection.MembershipExpireTime, connection.EncryptedCredentials, connection.CreateTime, connection.UpdateTime))
	if err != nil {
		return domain.ProviderConnection{}, fmt.Errorf("upsert music provider connection: %w", err)
	}
	return stored, nil
}

// DeleteProviderConnection removes all long-lived credentials for one source.
func (s *Store) DeleteProviderConnection(ctx context.Context, provider domain.MusicProvider) error {
	if _, err := s.pool.Exec(ctx, "DELETE FROM music_provider_connections WHERE provider_id = $1", provider); err != nil {
		return fmt.Errorf("delete music provider connection: %w", err)
	}
	return nil
}

// CreateProviderConnectionAttempt stores one interactive login operation.
func (s *Store) CreateProviderConnectionAttempt(ctx context.Context, attempt domain.ProviderConnectionAttempt) (domain.ProviderConnectionAttempt, error) {
	query := `INSERT INTO music_provider_connection_attempts (uid, provider_id, status, qr_image, qr_content_type,
		qr_payload, authorization_url, state_hash, encrypted_state, create_time, update_time, expire_time, consumed_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING ` + providerAttemptColumns
	stored, err := scanProviderAttempt(s.pool.QueryRow(ctx, query, attempt.UID, attempt.Provider, attempt.Status,
		nullBytes(attempt.QRImage), attempt.QRContentType, attempt.QRPayload, attempt.AuthorizationURL, nullBytes(attempt.StateHash),
		attempt.EncryptedState, attempt.CreateTime, attempt.UpdateTime, attempt.ExpireTime, attempt.ConsumedTime))
	if err != nil {
		return domain.ProviderConnectionAttempt{}, fmt.Errorf("create music provider connection attempt: %w", err)
	}
	return stored, nil
}

// GetProviderConnectionAttempt returns one attempt by its opaque UID.
func (s *Store) GetProviderConnectionAttempt(ctx context.Context, uid string) (domain.ProviderConnectionAttempt, error) {
	attempt, err := scanProviderAttempt(s.pool.QueryRow(ctx,
		"SELECT "+providerAttemptColumns+" FROM music_provider_connection_attempts WHERE uid = $1", uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ProviderConnectionAttempt{}, fmt.Errorf("%w: music provider connection attempt", domain.ErrNotFound)
	}
	return attempt, err
}

// GetProviderConnectionAttemptByStateHash returns one live OAuth attempt.
func (s *Store) GetProviderConnectionAttemptByStateHash(ctx context.Context, stateHash []byte) (domain.ProviderConnectionAttempt, error) {
	query := "SELECT " + providerAttemptColumns + ` FROM music_provider_connection_attempts
		WHERE state_hash = $1 AND consumed_time IS NULL AND expire_time > now()`
	attempt, err := scanProviderAttempt(s.pool.QueryRow(ctx, query, stateHash))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ProviderConnectionAttempt{}, fmt.Errorf("%w: live music provider connection attempt", domain.ErrNotFound)
	}
	return attempt, err
}

// UpdateProviderConnectionAttempt persists provider polling state.
func (s *Store) UpdateProviderConnectionAttempt(ctx context.Context, attempt domain.ProviderConnectionAttempt) (domain.ProviderConnectionAttempt, error) {
	query := `UPDATE music_provider_connection_attempts SET status = $2, qr_image = $3, qr_content_type = $4,
		qr_payload = $5, authorization_url = $6, encrypted_state = $7, update_time = $8, expire_time = $9, consumed_time = $10
		WHERE uid = $1 RETURNING ` + providerAttemptColumns
	stored, err := scanProviderAttempt(s.pool.QueryRow(ctx, query, attempt.UID, attempt.Status,
		nullBytes(attempt.QRImage), attempt.QRContentType, attempt.QRPayload, attempt.AuthorizationURL, attempt.EncryptedState,
		attempt.UpdateTime, attempt.ExpireTime, attempt.ConsumedTime))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ProviderConnectionAttempt{}, fmt.Errorf("%w: music provider connection attempt", domain.ErrNotFound)
	}
	if err != nil {
		return domain.ProviderConnectionAttempt{}, fmt.Errorf("update music provider connection attempt: %w", err)
	}
	return stored, nil
}

// ConsumeProviderConnectionAttempt atomically invalidates one callback state.
func (s *Store) ConsumeProviderConnectionAttempt(ctx context.Context, uid string, consumedTime time.Time) error {
	result, err := s.pool.Exec(ctx, `UPDATE music_provider_connection_attempts SET consumed_time = $2,
		update_time = $2 WHERE uid = $1 AND consumed_time IS NULL AND expire_time > $2`, uid, consumedTime)
	if err != nil {
		return fmt.Errorf("consume music provider connection attempt: %w", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("%w: live music provider connection attempt", domain.ErrNotFound)
	}
	return nil
}

// PurgeExpiredProviderConnectionAttempts removes expired or consumed transient credentials.
func (s *Store) PurgeExpiredProviderConnectionAttempts(ctx context.Context, before time.Time) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM music_provider_connection_attempts
		WHERE expire_time < $1 OR (consumed_time IS NOT NULL AND consumed_time < $1)`, before)
	if err != nil {
		return fmt.Errorf("purge music provider connection attempts: %w", err)
	}
	return nil
}

func scanProviderConnection(row rowScanner) (domain.ProviderConnection, error) {
	var connection domain.ProviderConnection
	var membershipExpire pgtype.Timestamptz
	err := row.Scan(&connection.Provider, &connection.Status, &connection.AccountID, &connection.DisplayName,
		&connection.AvatarURL, &connection.Membership, &membershipExpire, &connection.EncryptedCredentials,
		&connection.CreateTime, &connection.UpdateTime)
	if err != nil {
		return domain.ProviderConnection{}, err
	}
	if membershipExpire.Valid {
		value := membershipExpire.Time.UTC()
		connection.MembershipExpireTime = &value
	}
	return connection, nil
}

func scanProviderAttempt(row rowScanner) (domain.ProviderConnectionAttempt, error) {
	var attempt domain.ProviderConnectionAttempt
	var consumed pgtype.Timestamptz
	err := row.Scan(&attempt.UID, &attempt.Provider, &attempt.Status, &attempt.QRImage, &attempt.QRContentType, &attempt.QRPayload,
		&attempt.AuthorizationURL, &attempt.StateHash, &attempt.EncryptedState, &attempt.CreateTime,
		&attempt.UpdateTime, &attempt.ExpireTime, &consumed)
	if err != nil {
		return domain.ProviderConnectionAttempt{}, err
	}
	if consumed.Valid {
		value := consumed.Time.UTC()
		attempt.ConsumedTime = &value
	}
	return attempt, nil
}

func nullBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}
