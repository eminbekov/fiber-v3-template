package cache

import "github.com/gofrs/uuid/v5"

func UserByIDKey(id uuid.UUID) string {
	return "user:" + id.String()
}

func UserByUsernameKey(username string) string {
	return "user:username:" + username
}

func UserListKey(filterHash string) string {
	return "user:list:" + filterHash
}

func PermissionsByUserIDKey(userID uuid.UUID) string {
	return "permissions:user:" + userID.String()
}
