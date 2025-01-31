// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"time"
)

type Feed struct {
	ID        int32
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Url       string
	UserID    int32
}

type FeedsFollow struct {
	ID        int32
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    int32
	FeedsID   int32
}

type User struct {
	ID        int32
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
}
