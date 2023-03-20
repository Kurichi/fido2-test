package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"fido2-test/model"

	"github.com/go-webauthn/webauthn/webauthn"
)

func GetUser(db *sql.DB, ctx context.Context, username string) (*model.User, error) {
	const QUERY = `SELECT id, name, display_name FROM users WHERE name = $1;`

	conn, err := db.Conn(ctx)
	defer conn.Close()
	if err != nil {
		return nil, err
	}

	stmt, err := conn.PrepareContext(ctx, QUERY)
	defer stmt.Close()
	if err != nil {
		return nil, err
	}

	// ユーザーの取得
	user := &model.User{}
	if err := stmt.QueryRowContext(ctx, username).Scan(&user.ID, &user.Name, &user.DisplayName); err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserWithCredentials(db *sql.DB, ctx context.Context, username string) (*model.User, error) {
	const QUERY = `
	SELECT id, name, display_name, (
		SELECT JSON_AGG(credential)
		FROM credentials c
		WHERE u.id = c.user_id
	) credentials
	FROM users u
	WHERE name = $1;`

	conn, err := db.Conn(ctx)
	defer conn.Close()
	if err != nil {
		return nil, err
	}

	stmt, err := conn.PrepareContext(ctx, QUERY)
	defer stmt.Close()
	if err != nil {
		return nil, err
	}

	// ユーザーの取得
	user := &model.User{}
	var s sql.NullString
	if err := stmt.QueryRowContext(ctx, username).Scan(&user.ID, &user.Name, &user.DisplayName, &s); err != nil {
		return nil, err
	}
	if s.Valid {
		json.Unmarshal([]byte(s.String), &user.Credentials)
	}

	return user, nil
}

func CreateUser(db *sql.DB, ctx context.Context, name, display_name string) (*model.User, error) {
	const QUERY = `INSERT INTO users (name, display_name) VALUES ($1,$2) RETURNING id;`

	conn, err := db.Conn(ctx)
	defer conn.Close()
	if err != nil {
		return nil, err
	}

	stmt, err := conn.PrepareContext(ctx, QUERY)
	defer stmt.Close()
	if err != nil {
		return nil, err
	}

	user := &model.User{}
	user.Name = name
	user.DisplayName = display_name
	if err := stmt.QueryRowContext(ctx, user.Name, user.DisplayName).Scan(&user.ID); err != nil {
		return nil, err
	}

	return user, nil
}

func AddCredential(db *sql.DB, ctx context.Context, id string, credential *webauthn.Credential) error {
	const QUERY = `INSERT INTO credentials (user_id, credential) VALUES ($1, $2);`

	conn, err := db.Conn(ctx)
	defer conn.Close()
	if err != nil {
		return err
	}

	stmt, err := conn.PrepareContext(ctx, QUERY)
	defer stmt.Close()
	if err != nil {
		return err
	}

	b, err := json.Marshal(credential)
	if _, err = stmt.ExecContext(ctx, id, string(b)); err != nil {
		return err
	}

	return nil
}
