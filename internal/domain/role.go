package domain

type Role struct {
	ID          int64        `db:"id"`
	Name        string       `db:"name"`
	Description string       `db:"description"`
	Permissions []Permission `db:"-"`
}
