package handler

import (
	"database/sql"
	"fido2-test/utils"
	"net/http"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
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

		user, err := utils.GetUserWithCredentials(db, c.Request().Context(), name)
		if err != nil {
			if err == sql.ErrNoRows {
				user, err = utils.CreateUser(db, c.Request().Context(), name, strings.Split(name, "@")[0])
				if err != nil {
					c.Logger().Error(err)
					return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
				}
			} else {
				c.Logger().Error(err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
			}
		}

		registerOptions := func(credCreationOpts *protocol.PublicKeyCredentialCreationOptions) {
			credCreationOpts.CredentialExcludeList = user.CredentialExcludeList()
		}

		creation, sessionData, err := w.BeginRegistration(user, registerOptions)
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
		sess.Values["sessionData"] = sessionData
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

		user, err := utils.GetUser(db, c.Request().Context(), name)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		sess, err := session.Get("session", c)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		sessionData := sess.Values["sessionData"].(*webauthn.SessionData)
		if sessionData == nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
		}

		credential, err := w.FinishRegistration(user, *sessionData, c.Request())
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		if err := utils.AddCredential(db, c.Request().Context(), user.ID, credential); err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Registration success"})

	}
}
