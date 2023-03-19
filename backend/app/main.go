package main

import (
	"fido2-test/api/router"
	"fido2-test/db"

	"github.com/go-webauthn/webauthn/webauthn"
	_ "github.com/lib/pq"
)

func main() {
	db, err := db.New()
	if err != nil {
		panic(err)
	}

	w, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "Kurichi WebAuthn Demo",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:3000"},
	})
	if err != nil {
		panic(err)
	}

	e := router.NewRouter(db, w)

	e.Logger.Fatal(e.Start(":8080"))
}
