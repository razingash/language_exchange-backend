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

	baseQuery := `
		SELECT 
			u.id, u.email, u.full_name,
			ul.language_id, ul.type
		FROM users u
	`

	joins := []string{}
	conditions := []string{"u.id != $1"}
	args := []interface{}{userID}
	argIndex := 2

	if native != "" {
		joins = append(joins, fmt.Sprintf(`
			JOIN user_languages ul_native 
				ON ul_native.user_id = u.id 
				AND ul_native.type = 'native' 
				AND ul_native.language_id = $%d
		`, argIndex))
		args = append(args, native)
		argIndex++
	}

	if target != "" {
		joins = append(joins, fmt.Sprintf(`
			JOIN user_languages ul_target 
				ON ul_target.user_id = u.id 
				AND ul_target.type = 'target' 
				AND ul_target.language_id = $%d
		`, argIndex))
		args = append(args, target)
		argIndex++
	}

	joins = append(joins, "LEFT JOIN user_languages ul ON ul.user_id = u.id")

	query := baseQuery + strings.Join(joins, "\n") + "\nWHERE " + strings.Join(conditions, " AND ") + "\nORDER BY u.id DESC"

	return db.DB.Query(ctx, query, args...)
}

func UpdateSelectedLanguages(userID string, langs Languages) error {
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
