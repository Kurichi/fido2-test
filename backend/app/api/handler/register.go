package handler

import (
	"database/sql"
	"encoding/json"
	"fido2-test/model"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func RegisterBegin(db *sql.DB, w *webauthn.WebAuthn) echo.HandlerFunc {
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
			// ユーザーが存在しない場合は新規作成
			if err == sql.ErrNoRows {
				stmt, err := conn.PrepareContext(ctx, "INSERT INTO users (name, display_name) VALUES ($1,$2) RETURNING id;")
				if err != nil {
					c.Logger().Error(err)
					return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
				}
				defer stmt.Close()

				user.Name = name
				user.DisplayName = strings.Split(name, "@")[0]
				if err := stmt.QueryRowContext(ctx, user.Name, user.DisplayName).Scan(&user.ID); err != nil {
					c.Logger().Error(err)
					return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
				}
			} else {
				c.Logger().Error(err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
			}
		}

		creation, sessionData, err := w.BeginRegistration(user)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		sess, err := session.Get("session", c)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}
		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: true,
		}
		sess.Values["sessionData"] = *sessionData
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		return c.JSON(http.StatusOK, creation)
	}
}

func RegisterFinish(db *sql.DB, w *webauthn.WebAuthn) echo.HandlerFunc {
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

		user := &model.User{}
		if err := stmt.QueryRowContext(ctx, name).Scan(&user.ID, &user.Name, &user.DisplayName); err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": `User(name is ${name}) not found`})
			} else {
				c.Logger().Error(err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
			}
		}

		sess, err := session.Get("session", c)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		sessionData := sess.Values["sessionData"]
		if sessionData == nil {
			fmt.Println("sessionData is nil")
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
		}

		credential, err := w.FinishRegistration(user, *(sessionData.(*webauthn.SessionData)), c.Request())
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}
		// fmt.Println("credential", credential)

		stmt, err = conn.PrepareContext(ctx, "INSERT INTO credentials (user_id, credential) VALUES ($1,$2);")
		defer stmt.Close()
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		b, err := json.Marshal(credential)
		if _, err = stmt.ExecContext(ctx, user.ID, string(b)); err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Registration success"})

	}
}
