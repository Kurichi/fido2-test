package handler

import (
	"database/sql"
	"encoding/json"
	"fido2-test/model"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func LoginBegin(db *sql.DB, w *webauthn.WebAuthn) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.Param("name")
		if name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Path parameter 'name' is required"})
		}

		ctx := c.Request().Context()
		conn, err := db.Conn(ctx)
		defer conn.Close()
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		stmt, err := conn.PrepareContext(ctx, "SELECT id, name, display_name FROM users WHERE name = $1;")
		defer stmt.Close()
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		// ユーザーの取得
		user := &model.User{}
		if err := stmt.QueryRowContext(ctx, name).Scan(&user.ID, &user.Name, &user.DisplayName); err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, map[string]string{"message": "User not found"})
			}
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		stmt, err = conn.PrepareContext(ctx, "SELECT credential FROM credentials WHERE user_id = $1;")
		defer stmt.Close()
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		row, err := stmt.QueryContext(ctx, user.ID)
		defer row.Close()
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		for row.Next() {
			var s string
			if err := row.Scan(&s); err != nil {
				c.Logger().Error(err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
			}
			cred := webauthn.Credential{}
			json.Unmarshal([]byte(s), &cred)
			user.Credentials = append(user.Credentials, cred)
		}

		opt, s, err := w.BeginLogin(user)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		// セッションに保存
		sess, err := session.Get("session", c)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}
		sess.Values["sessionData"] = s

		if err := sess.Save(c.Request(), c.Response()); err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		return c.JSON(http.StatusOK, opt)
	}
}

func LoginFinish(db *sql.DB, w *webauthn.WebAuthn) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.Param("name")
		if name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Path parameter 'name' is required"})
		}

		ctx := c.Request().Context()
		conn, err := db.Conn(ctx)
		defer conn.Close()
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		stmt, err := conn.PrepareContext(ctx, "SELECT id, name, display_name FROM users WHERE name = $1;")
		defer stmt.Close()
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		// ユーザーの取得
		user := &model.User{}
		if err := stmt.QueryRowContext(ctx, name).Scan(&user.ID, &user.Name, &user.DisplayName); err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, map[string]string{"message": "User not found"})
			}
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}
		stmt, err = conn.PrepareContext(ctx, "SELECT credential FROM credentials WHERE user_id = $1;")
		defer stmt.Close()
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		row, err := stmt.QueryContext(ctx, user.ID)
		defer row.Close()
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		for row.Next() {
			var s string
			if err := row.Scan(&s); err != nil {
				c.Logger().Error(err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
			}
			cred := webauthn.Credential{}
			json.Unmarshal([]byte(s), &cred)
			user.Credentials = append(user.Credentials, cred)
		}

		sess, err := session.Get("session", c)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		// セッションから取得
		sessionData := sess.Values["sessionData"].(*webauthn.SessionData)

		if _, err := w.FinishLogin(user, *sessionData, c.Request()); err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "authentication failed"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

}

func credDecode(s string) (*webauthn.Credential, error) {
	cred := &webauthn.Credential{}
	if err := json.Unmarshal([]byte(s), &cred); err != nil {
		return nil, err
	}
	return cred, nil
}
