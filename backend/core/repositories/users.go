package repositories

import (
	"backend/core/db"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type UserInfo struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type Languages struct {
	Native []int `json:"native"`
	Target []int `json:"target"`
}

func SelectUserInfo(userID string) (UserInfo, error) {
	var user UserInfo

	err := db.DB.QueryRow(context.Background(), `
		SELECT id, email, full_name
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.FullName)

	return user, err
}

func SelectUserLanguages(userID string) (pgx.Rows, error) {
	rows, err := db.DB.Query(context.Background(), `
		SELECT language_id, type FROM user_languages WHERE user_id = $1
	`, userID)

	return rows, err
}

func SelectTargetedUsers(target string, native string, userID string) (pgx.Rows, error) {
	ctx := context.Background()
	var rows pgx.Rows
	var err error

	if native != "" && target != "" {
		rows, err = db.DB.Query(ctx, `
			SELECT u.id, u.email, u.full_name
			FROM users u
			JOIN user_languages ul1 ON ul1.user_id = u.id AND ul1.type = 'native' AND ul1.language_id = $1
			JOIN user_languages ul2 ON ul2.user_id = u.id AND ul2.type = 'target' AND ul2.language_id = $2
			WHERE u.id != $3
		`, target, native, userID)
	} else {
		rows, err = db.DB.Query(ctx, `
			SELECT id, email, full_name FROM users WHERE id != $1
		`, userID)
	}
	return rows, err
}

func UpdateSelectedLanguages(userID string) error {
	var langs Languages
	ctx := context.Background()
	var (
		values []string
		args   []interface{}
		argPos = 1
	)

	for _, langID := range langs.Native {
		values = append(values, fmt.Sprintf("($%d, $%d, 'native')", argPos, argPos+1))
		args = append(args, userID, langID)
		argPos += 2
	}

	for _, langID := range langs.Target {
		values = append(values, fmt.Sprintf("($%d, $%d, 'target')", argPos, argPos+1))
		args = append(args, userID, langID)
		argPos += 2
	}

	if len(values) > 0 {
		query := fmt.Sprintf(`
			INSERT INTO user_languages (user_id, language_id, type)
			VALUES %s
			ON CONFLICT (user_id, language_id, type) DO NOTHING
		`, strings.Join(values, ", "))
		_, err := db.DB.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	return nil
}
