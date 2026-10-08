// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"time"

	"github.com/twigex/twigex/model"
)

type oidcProviderRepository struct {
	Db *sql.DB
}

func NewOIDCProviderRepository(Db *sql.DB) *oidcProviderRepository {
	return &oidcProviderRepository{Db: Db}
}

func (r *oidcProviderRepository) Create(p *model.OIDCProvider) (*model.OIDCProvider, error) {
	if p.ID == "" {
		p.ID = model.NewID()
	}

	p.CreatedAt = time.Now().Unix()

	_, err := r.Db.Exec(`
        INSERT INTO oidc_providers
            (id, name, enabled, discovery_url, client_id, client_secret, scopes, button_text, button_color, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID,
		p.Name,
		p.Enabled,
		p.DiscoveryURL,
		p.ClientID,
		p.ClientSecret,
		p.Scopes,
		p.ButtonText,
		p.ButtonColor,
		p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *oidcProviderRepository) GetAll() ([]model.OIDCProvider, error) {
	rows, err := r.Db.Query(`
		SELECT id, name, enabled, discovery_url, client_id, client_secret,
		       scopes,
		       COALESCE(button_text, ''),
		       COALESCE(button_color, ''),
		       created_at
		FROM oidc_providers
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	providers := make([]model.OIDCProvider, 0)
	for rows.Next() {
		var p model.OIDCProvider
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Enabled,
			&p.DiscoveryURL,
			&p.ClientID,
			&p.ClientSecret,
			&p.Scopes,
			&p.ButtonText,
			&p.ButtonColor,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		providers = append(providers, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return providers, nil
}

// GetAllEnabled returns only enabled providers
func (r *oidcProviderRepository) GetAllEnabled() ([]model.OIDCProvider, error) {
	rows, err := r.Db.Query(`
		SELECT id, name, enabled, discovery_url, client_id, client_secret,
		       scopes,
		       COALESCE(button_text, ''),
		       COALESCE(button_color, ''),
		       created_at
		FROM oidc_providers
		WHERE enabled = 1
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	providers := make([]model.OIDCProvider, 0)
	for rows.Next() {
		var p model.OIDCProvider
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Enabled,
			&p.DiscoveryURL,
			&p.ClientID,
			&p.ClientSecret,
			&p.Scopes,
			&p.ButtonText,
			&p.ButtonColor,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		providers = append(providers, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return providers, nil
}

func (r *oidcProviderRepository) GetByID(id string) (*model.OIDCProvider, error) {
	p := &model.OIDCProvider{}
	err := r.Db.QueryRow(`
		SELECT id, name, enabled, discovery_url, client_id, client_secret,
		       scopes,
		       COALESCE(button_text, ''),
		       COALESCE(button_color, ''),
		       created_at
		FROM oidc_providers
		WHERE id = ?`, id).Scan(
		&p.ID,
		&p.Name,
		&p.Enabled,
		&p.DiscoveryURL,
		&p.ClientID,
		&p.ClientSecret,
		&p.Scopes,
		&p.ButtonText,
		&p.ButtonColor,
		&p.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *oidcProviderRepository) Update(p *model.OIDCProvider) (*model.OIDCProvider, error) {
	_, err := r.Db.Exec(`
		UPDATE oidc_providers
		SET name=?, enabled=?, discovery_url=?, client_id=?,
		    scopes=?, button_text=?, button_color=?
		WHERE id=?`,
		p.Name,
		p.Enabled,
		p.DiscoveryURL,
		p.ClientID,
		p.Scopes,
		p.ButtonText,
		p.ButtonColor,
		p.ID,
	)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *oidcProviderRepository) UpdateSecret(id, encryptedSecret string) error {
	_, err := r.Db.Exec(`
		UPDATE oidc_providers SET client_secret=? WHERE id=?`,
		encryptedSecret, id,
	)
	return err
}

func (r *oidcProviderRepository) Delete(id string) error {
	_, err := r.Db.Exec(`DELETE FROM oidc_providers WHERE id=?`, id)
	return err
}

func (r *oidcProviderRepository) Count() (int, error) {
	var count int
	err := r.Db.QueryRow(`SELECT COUNT(*) FROM oidc_providers`).Scan(&count)
	return count, err
}
