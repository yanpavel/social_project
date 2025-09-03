package main

import (
	"log"

	"github.com/yanpavel/social_project/internal/db"
	"github.com/yanpavel/social_project/internal/env"
	"github.com/yanpavel/social_project/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:6404/socialnetwork?sslmode=disable")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	store := store.NewStorage(conn)

	db.Seed(store)
}
