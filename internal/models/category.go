package models

type Category struct {
	ID   int    `db:"id"`
	Slug string `db:"slug"`
	Name string `db:"name"`
}
