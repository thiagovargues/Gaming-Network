package main

import (
	"context"
	"log"
	"time"

	"backend/internal/config"
	"backend/internal/repo"
	"backend/internal/repo/sqlite"
	"backend/pkg/migrate"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()
	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := migrate.Apply(cfg.DBPath); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	pass, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ctx := context.Background()
	aliceID, _ := repo.CreateUser(ctx, db, "alice@example.com", string(pass), "Alice", "Doe", "1990-01-01", nil, ptr("alice"), ptr("Hello"))
	bobID, _ := repo.CreateUser(ctx, db, "bob@example.com", string(pass), "Bob", "Smith", "1991-02-02", nil, ptr("bob"), nil)
	carolID, _ := repo.CreateUser(ctx, db, "carol@example.com", string(pass), "Carol", "Lee", "1992-03-03", nil, ptr("carol"), nil)

	_ = repo.CreateFollow(ctx, db, bobID, aliceID)
	_ = repo.CreateFollow(ctx, db, carolID, aliceID)

	group, _ := repo.CreateGroup(ctx, db, aliceID, "Gamers", "Group for gaming fans")
	_ = repo.AddGroupMember(ctx, db, group.ID, bobID, "member")

	_, _ = repo.CreatePost(ctx, db, aliceID, "Welcome to the network!", "public", nil, nil, nil)
	_, _ = repo.CreatePost(ctx, db, aliceID, "Followers-only update", "followers", nil, nil, nil)
	_, _ = repo.CreatePost(ctx, db, aliceID, "Group post", "group", nil, nil, &group.ID)

	log.Printf("seed complete: users=%d group=%d", 3, group.ID)
	time.Sleep(100 * time.Millisecond)
}

func ptr(v string) *string { return &v }
