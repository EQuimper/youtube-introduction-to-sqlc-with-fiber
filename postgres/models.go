// Code generated by sqlc. DO NOT EDIT.

package postgres

import (
	"database/sql"
)

type Todo struct {
	ID        int64
	Name      string
	Completed sql.NullBool
}