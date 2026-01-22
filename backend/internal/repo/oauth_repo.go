package repo

import (
	"context"
	"database/sql"
)

func GetUserIDByOAuth(ctx context.Context, db *sql.DB, provider, subject string) (int64, bool, error) {
	var userID int64
	row := db.QueryRowContext(ctx, "SELECT user_id FROM oauth_accounts WHERE provider = ? AND subject = ?", provider, subject)
	if err := row.Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}
	return userID, true, nil
}

func CreateOAuthAccount(ctx context.Context, db *sql.DB, userID int64, provider, subject, email string) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO oauth_accounts (user_id, provider, subject, email)
		 VALUES (?, ?, ?, ?)`,
		userID,
		provider,
		subject,
		email,
	)
	return err
}
