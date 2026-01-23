package repo

import (
	"context"
	"database/sql"
	"strings"
)

func CreateUser(ctx context.Context, db *sql.DB, email, passwordHash, firstName, lastName, dob string, avatar, nickname, about *string) (int64, error) {
	result, err := db.ExecContext(
		ctx,
		`INSERT INTO users (email, password_hash, first_name, last_name, dob, avatar_path, nickname, about, is_public)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		email,
		passwordHash,
		firstName,
		lastName,
		dob,
		nullableString(avatar),
		nullableString(nickname),
		nullableString(about),
	)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func CreateOAuthUser(ctx context.Context, db *sql.DB, email, firstName, lastName string, avatar *string) (int64, error) {
	result, err := db.ExecContext(
		ctx,
		`INSERT INTO users (email, password_hash, first_name, last_name, dob, avatar_path, nickname, about, is_public)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		email,
		"",
		firstName,
		lastName,
		"1970-01-01",
		nullableString(avatar),
		sql.NullString{},
		sql.NullString{},
	)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func GetUserByEmail(ctx context.Context, db *sql.DB, email string) (int64, *string, error) {
	var id int64
	var hash sql.NullString
	row := db.QueryRowContext(ctx, "SELECT id, password_hash FROM users WHERE email = ?", email)
	if err := row.Scan(&id, &hash); err != nil {
		return 0, nil, err
	}
	if !hash.Valid {
		return id, nil, nil
	}
	return id, &hash.String, nil
}

func GetUserIDByEmail(ctx context.Context, db *sql.DB, email string) (int64, bool, error) {
	var id int64
	row := db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = ?", email)
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}

func GetUserByID(ctx context.Context, db *sql.DB, id int64) (UserProfile, bool, error) {
	var profile UserProfile
	var dob sql.NullString
	var avatar sql.NullString
	var nickname sql.NullString
	var about sql.NullString
	var sex sql.NullString
	var age sql.NullInt64
	var isPublic int
	var showFirst int
	var showLast int
	var showAge int
	var showSex int
	var showNickname int
	var showAbout int

	row := db.QueryRowContext(ctx, `SELECT id, email, first_name, last_name, dob, avatar_path, nickname, about, sex, age,
		show_first_name, show_last_name, show_age, show_sex, show_nickname, show_about, is_public, created_at
		FROM users WHERE id = ?`, id)
	if err := row.Scan(
		&profile.ID,
		&profile.Email,
		&profile.FirstName,
		&profile.LastName,
		&dob,
		&avatar,
		&nickname,
		&about,
		&sex,
		&age,
		&showFirst,
		&showLast,
		&showAge,
		&showSex,
		&showNickname,
		&showAbout,
		&isPublic,
		&profile.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return UserProfile{}, false, nil
		}
		return UserProfile{}, false, err
	}

	profile.Avatar = nullableStringPtr(avatar)
	profile.Nickname = nullableStringPtr(nickname)
	profile.About = nullableStringPtr(about)
	profile.Sex = nullableStringPtr(sex)
	if age.Valid {
		v := int(age.Int64)
		profile.Age = &v
	}
	profile.ShowFirstName = showFirst == 1
	profile.ShowLastName = showLast == 1
	profile.ShowAge = showAge == 1
	profile.ShowSex = showSex == 1
	profile.ShowNickname = showNickname == 1
	profile.ShowAbout = showAbout == 1
	if dob.Valid {
		profile.DOB = strings.TrimSpace(dob.String)
	} else {
		profile.DOB = ""
	}
	profile.IsPublic = isPublic == 1
	return profile, true, nil
}

func UpdateMe(ctx context.Context, db *sql.DB, userID int64, isPublic *bool, avatar, nickname, about, firstName, lastName, dob, sex *string, age *int, showFirst, showLast, showAge, showSex, showNickname, showAbout *bool) (UserProfile, error) {
	if isPublic != nil {
		value := 0
		if *isPublic {
			value = 1
		}
		if _, err := db.ExecContext(ctx, "UPDATE users SET is_public = ? WHERE id = ?", value, userID); err != nil {
			return UserProfile{}, err
		}
	}

	if avatar != nil {
		if _, err := db.ExecContext(ctx, "UPDATE users SET avatar_path = ? WHERE id = ?", nullableString(avatar), userID); err != nil {
			return UserProfile{}, err
		}
	}
	if nickname != nil {
		if _, err := db.ExecContext(ctx, "UPDATE users SET nickname = ? WHERE id = ?", nullableString(nickname), userID); err != nil {
			return UserProfile{}, err
		}
	}
	if about != nil {
		if _, err := db.ExecContext(ctx, "UPDATE users SET about = ? WHERE id = ?", nullableString(about), userID); err != nil {
			return UserProfile{}, err
		}
	}
	if firstName != nil {
		if _, err := db.ExecContext(ctx, "UPDATE users SET first_name = ? WHERE id = ?", nullableString(firstName), userID); err != nil {
			return UserProfile{}, err
		}
	}
	if lastName != nil {
		if _, err := db.ExecContext(ctx, "UPDATE users SET last_name = ? WHERE id = ?", nullableString(lastName), userID); err != nil {
			return UserProfile{}, err
		}
	}
	if dob != nil {
		if _, err := db.ExecContext(ctx, "UPDATE users SET dob = ? WHERE id = ?", nullableString(dob), userID); err != nil {
			return UserProfile{}, err
		}
	}
	if sex != nil {
		if _, err := db.ExecContext(ctx, "UPDATE users SET sex = ? WHERE id = ?", nullableString(sex), userID); err != nil {
			return UserProfile{}, err
		}
	}
	if age != nil {
		if _, err := db.ExecContext(ctx, "UPDATE users SET age = ? WHERE id = ?", *age, userID); err != nil {
			return UserProfile{}, err
		}
	}
	if showFirst != nil {
		val := 0
		if *showFirst {
			val = 1
		}
		if _, err := db.ExecContext(ctx, "UPDATE users SET show_first_name = ? WHERE id = ?", val, userID); err != nil {
			return UserProfile{}, err
		}
	}
	if showLast != nil {
		val := 0
		if *showLast {
			val = 1
		}
		if _, err := db.ExecContext(ctx, "UPDATE users SET show_last_name = ? WHERE id = ?", val, userID); err != nil {
			return UserProfile{}, err
		}
	}
	if showAge != nil {
		val := 0
		if *showAge {
			val = 1
		}
		if _, err := db.ExecContext(ctx, "UPDATE users SET show_age = ? WHERE id = ?", val, userID); err != nil {
			return UserProfile{}, err
		}
	}
	if showSex != nil {
		val := 0
		if *showSex {
			val = 1
		}
		if _, err := db.ExecContext(ctx, "UPDATE users SET show_sex = ? WHERE id = ?", val, userID); err != nil {
			return UserProfile{}, err
		}
	}
	if showNickname != nil {
		val := 0
		if *showNickname {
			val = 1
		}
		if _, err := db.ExecContext(ctx, "UPDATE users SET show_nickname = ? WHERE id = ?", val, userID); err != nil {
			return UserProfile{}, err
		}
	}
	if showAbout != nil {
		val := 0
		if *showAbout {
			val = 1
		}
		if _, err := db.ExecContext(ctx, "UPDATE users SET show_about = ? WHERE id = ?", val, userID); err != nil {
			return UserProfile{}, err
		}
	}

	profile, _, err := GetUserByID(ctx, db, userID)
	return profile, err
}

func IsUserPublic(ctx context.Context, db *sql.DB, userID int64) (bool, bool, error) {
	var isPublic int
	row := db.QueryRowContext(ctx, "SELECT is_public FROM users WHERE id = ?", userID)
	if err := row.Scan(&isPublic); err != nil {
		if err == sql.ErrNoRows {
			return false, false, nil
		}
		return false, false, err
	}
	return isPublic == 1, true, nil
}

func nullableString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: trimmed, Valid: true}
}

func nullableStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	val := strings.TrimSpace(value.String)
	if val == "" {
		return nil
	}
	return &val
}
