package domain

type Permission struct {
	ID       int64  `db:"id"`
	Resource string `db:"resource"`
	Action   string `db:"action"`
}

func (permission Permission) String() string {
	return permission.Resource + "." + permission.Action
}
