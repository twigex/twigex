// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"strings"
	"time"

	"github.com/twigex/twigex/model"
)

type roleRepository struct {
	Db *sql.DB
}

func NewRoleRepository(Db *sql.DB) (*roleRepository, error) {
	repo := &roleRepository{}
	repo.Db = Db
	return repo, nil
}

func (r *roleRepository) scanRole(row interface{ Scan(...interface{}) error }) (model.Role, string, error) {
	role := model.Role{}
	var permissions string
	err := row.Scan(&role.ID, &role.Name, &role.DisplayName, &role.Description, &permissions, &role.CreatedAt, &role.UpdatedAt, &role.BuiltIn)
	return role, permissions, err
}

func (r *roleRepository) GetAll() ([]model.Role, error) {
	results, err := r.Db.Query("SELECT id, name, displayname, description, permissions, created_at, updated_at, built_in FROM roles")
	if err != nil {
		return nil, err
	}

	defer results.Close()

	roles := make([]model.Role, 0)
	for results.Next() {
		role, permissions, err := r.scanRole(results)
		if err != nil {
			return nil, err
		}

		role.Permissions = strings.Split(permissions, " ")
		roles = append(roles, role)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *roleRepository) GetByID(id string) (*model.Role, error) {
	row := r.Db.QueryRow("SELECT id, name, displayname, description, permissions, created_at, updated_at, built_in FROM roles WHERE id=?", id)

	role, permissions, err := r.scanRole(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	role.Permissions = strings.Split(permissions, " ")
	return &role, nil
}

func (r *roleRepository) GetByName(name string) (*model.Role, error) {
	row := r.Db.QueryRow("SELECT id, name, displayname, description, permissions, created_at, updated_at, built_in FROM roles WHERE name=?", name)

	role, permissions, err := r.scanRole(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	role.Permissions = strings.Split(permissions, " ")
	return &role, nil
}

func (r *roleRepository) GetByNames(roleNames []string) ([]model.Role, error) {
	if len(roleNames) == 0 {
		empty := []model.Role{}
		return empty, nil
	}

	args := make([]interface{}, len(roleNames))
	for i, name := range roleNames {
		args[i] = name
	}

	results, err := r.Db.Query(
		`SELECT id, name, displayname, description, permissions, created_at, updated_at, built_in FROM roles WHERE name IN (?`+strings.Repeat(",?", len(args)-1)+`)`,
		args...,
	)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	roles := make([]model.Role, 0)
	for results.Next() {
		role, permissions, err := r.scanRole(results)
		if err != nil {
			return nil, err
		}

		role.Permissions = strings.Split(permissions, " ")
		roles = append(roles, role)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *roleRepository) CreateOrUpdate(role model.Role) (*model.Role, error) {
	if role.ID == "" {
		return r.createRole(role)
	}

	permissions := strings.Join(role.Permissions, " ")
	updated := time.Now().Unix()
	_, err := r.Db.Exec(
		"UPDATE roles SET name=?, displayname=?, description=?, permissions=?, built_in=?, updated_at=? WHERE id=?",
		role.Name, role.DisplayName, role.Description, permissions, role.BuiltIn, updated, role.ID,
	)
	if err != nil {
		return nil, err
	}

	role.UpdatedAt = updated
	return &role, nil
}

func (r *roleRepository) createRole(role model.Role) (*model.Role, error) {
	id := model.NewID()
	now := time.Now().Unix()
	permissions := strings.Join(role.Permissions, " ")

	_, err := r.Db.Exec(
		"INSERT INTO roles (id, name, displayname, description, permissions, created_at, updated_at, built_in) VALUES(?,?,?,?,?,?,?,?)",
		id, role.Name, role.DisplayName, role.Description, permissions, now, now, role.BuiltIn,
	)
	if err != nil {
		return nil, err
	}

	return &model.Role{
		ID:          id,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		Permissions: role.Permissions,
		BuiltIn:     role.BuiltIn,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (r *roleRepository) Delete(id string) error {
	_, err := r.Db.Exec("DELETE FROM roles WHERE id=?", id)
	return err
}
