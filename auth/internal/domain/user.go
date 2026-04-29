package domain

// IsAdmin reports whether the user holds the admin role. Used by handlers
// when deciding whether to allow a privileged action — keeps the role-string
// comparison in one place so a future rename or capitalisation change lives
// in domain, not scattered across transport.
func (u *User) IsAdmin() bool {
	if u == nil {
		return false
	}
	return u.Role == RoleAdmin
}
