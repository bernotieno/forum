package models

type User struct {
	ID       int
	Email    string
	Username string
	Password string
}
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}
