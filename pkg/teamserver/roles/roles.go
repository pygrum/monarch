package roles

import (
	"path/filepath"
	"slices"
)

const (
	RoleAdmin  Role = "admin"
	RolePlayer Role = "player"
)

type Role string
type Roles []Role

var (
	AdminOnly = []string{
		"/Install",
		"/Uninstall",

		"/HttpOpen",
		"/HttpsOpen",

		"/HttpClose",
		"/HttpsClose",

		"/TcpOpen",
		"/TcpClose",
	}
)

func (rs Roles) String() []string {
	var s []string
	for _, r := range rs {
		s = append(s, string(r))
	}
	return s
}

// All returns all roles
func All() Roles {
	return Roles{
		RoleAdmin,
		RolePlayer,
	}
}

// ValidRole reports whether the specified role is a valid existing role
func ValidRole(r string) bool {
	return slices.Contains(All(), Role(r))
}

// AuthorizedEndpoint reports whether a player with the provided role is allowed to access the provided endpoint
func AuthorizedEndpoint(r Role, endpoint string) bool {
	if r == RoleAdmin {
		return true
	}
	if r == RolePlayer {
		return !slices.Contains(AdminOnly, "/"+filepath.Base(endpoint))
	}
	return false
}
