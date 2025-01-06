package controllers

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Raymond9734/forum.git/BackEnd/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	DB *sql.DB
}

func NewAuthController(db *sql.DB) *AuthController {
	return &AuthController{DB: db}
}

func (ac *AuthController) RegisterUser(email, username, password string) error {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	_, err := ac.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, hashedPassword)
	if err != nil {
		fmt.Println(err)
		return errors.New("email or username already taken")
	}
	return nil
}

func (ac *AuthController) AuthenticateUser(username, password string) (*models.User, error) {
	user := &models.User{}
	err := ac.DB.QueryRow("SELECT id, email, username, password FROM users WHERE username = ?", username).
		Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("invalid username")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}
