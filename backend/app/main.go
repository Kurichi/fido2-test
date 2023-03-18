package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type RegisterResponse struct {
	Challenge string `json:"challenge"`
	Rp        struct {
		Name string `json:"name"`
	} `json:"rp"`
	User        User   `json:"user"`
	Attestation string `json:"attestation"`
}

func RandomStr(digit int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 乱数を生成
	b := make([]byte, digit)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("unexpected error...")
	}

	// letters からランダムに取り出して文字列を生成
	var result string
	for _, v := range b {
		// index が letters の長さに収まるように調整
		result += string(letters[int(v)%len(letters)])
	}
	return result, nil
}

func main() {
	USER := os.Getenv("POSTGRES_USER")
	PASS := os.Getenv("POSTGRES_PASSWORD")
	HOST := os.Getenv("POSTGRES_HOST")
	DBNAME := os.Getenv("POSTGRES_DB")

	CONNECT := fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
		HOST, USER, PASS, DBNAME)

	db, err := sql.Open("postgres", CONNECT)
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(10 * time.Second)

	if err := db.Ping(); err != nil {
		panic(err)
	}

	e := echo.New()

	e.POST("/register", func(c echo.Context) error {
		user := &User{}
		if err := c.Bind(&user); err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
		}
		ctx := c.Request().Context()

		conn, err := db.Conn(ctx)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}
		defer conn.Close()

		stmt, err := conn.PrepareContext(ctx, "insert into users (uid, email) values ($1, $2);")
		if err != nil {
			fmt.Println("ERROR")
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}
		defer stmt.Close()

		user.ID = uuid.New().String()
		user.DisplayName = "Kurichi"
		if _, err := stmt.ExecContext(ctx, user.ID, user.Name); err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		p, err := conn.PrepareContext(ctx, "INSERT INTO challenges (user_id, challenge) VALUES ($1,$2);")
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}
		defer p.Close()

		cha, err := RandomStr(64)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		if _, err := p.ExecContext(ctx, user.ID, cha); err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		}

		return c.JSON(http.StatusOK, &RegisterResponse{
			Challenge: cha,
			Rp: struct {
				Name string `json:"name"`
			}{
				Name: "Kurichi",
			},
			User:        *user,
			Attestation: "none",
		})

	})

	e.Logger.Fatal(e.Start(":8080"))
}
