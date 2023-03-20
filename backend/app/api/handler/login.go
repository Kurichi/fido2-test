package handler

import (
	"database/sql"
	"fido2-test/utils"
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

		// ユーザーの取得
		user, err := utils.GetUserWithCredentials(db, c.Request().Context(), name)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, map[string]string{"message": "User not found"})
			}
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
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

		user, err := utils.GetUserWithCredentials(db, c.Request().Context(), name)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, map[string]string{"message": "User not found"})
			}
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
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
