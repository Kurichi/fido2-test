package router

import (
	"database/sql"
	"encoding/gob"
	"fido2-test/api/handler"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewRouter(db *sql.DB, w *webauthn.WebAuthn) *echo.Echo {
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339}:[${method}] ${status} ${path} ${error}` + "\n",
	}))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 2,
	}))
	e.Use(middleware.Recover())

	// health check
	e.GET("/healthz", handler.Healthz)

	// auth
	g := e.Group("/auth")
	g.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:          middleware.DefaultSkipper,
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowCredentials: true,
	}))
	g.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	gob.Register(&webauthn.SessionData{})

	// SignUp
	g.GET("/register/:name", handler.RegisterBegin(db, w))
	g.POST("/register/:name", handler.RegisterFinish(db, w))

	// SignIn

	return e
}
