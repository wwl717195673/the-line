package domain

const (
	RoleAdmin = "admin"
)

type Actor struct {
	PersonID uint64
	RoleType string
}

func (a Actor) IsAdmin() bool {
	return a.RoleType == RoleAdmin
}
