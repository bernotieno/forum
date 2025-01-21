package Test

import (
	"database/sql"
	"testing"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	// Initialize the logger for tests
	logger.Init()
}

func TestAuthController_RegisterUser(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	type fields struct {
		DB *sql.DB
	}
	type args struct {
		email    string
		username string
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "Successful Registration",
			fields: fields{
				DB: db,
			},
			args: args{
				email:    "test@example.com",
				username: "testuser",
				password: "password123",
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Duplicate Email or Username",
			fields: fields{
				DB: db,
			},
			args: args{
				email:    "test@example.com",
				username: "testuser2",
				password: "password123",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Internal Server Error - Hashing Password",
			fields: fields{
				DB: db,
			},
			args: args{
				email:    "test2@example.com",
				username: "testuser2",
				password: "",
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &controllers.AuthController{
				DB: tt.fields.DB,
			}
			got, err := ac.RegisterUser(tt.args.email, tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthController.RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AuthController.RegisterUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

// InsertTestUser inserts a test user into the database.
func InsertTestUser(db *sql.DB, email, username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)",
		email, username, hashedPassword)
	return err
}

func TestAuthController_AuthenticateUser(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Insert a test user into the database
	err = InsertTestUser(db, "test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	type fields struct {
		DB *sql.DB
	}
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.User
		wantErr bool
	}{
		{
			name: "Successful Authentication",
			fields: fields{
				DB: db,
			},
			args: args{
				username: "testuser",
				password: "password123",
			},
			want: &models.User{
				ID:       1,
				Email:    "test@example.com",
				Username: "testuser",
			},
			wantErr: false,
		},
		{
			name: "Invalid Username",
			fields: fields{
				DB: db,
			},
			args: args{
				username: "nonexistentuser",
				password: "password123",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid Password",
			fields: fields{
				DB: db,
			},
			args: args{
				username: "testuser",
				password: "wrongpassword",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &controllers.AuthController{
				DB: tt.fields.DB,
			}
			got, err := ac.AuthenticateUser(tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthController.AuthenticateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				// If an error is expected, we don't need to check the returned user
				return
			}
			if got.ID != tt.want.ID || got.Email != tt.want.Email || got.Username != tt.want.Username {
				t.Errorf("AuthController.AuthenticateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthController_IsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "Valid Email",
			email: "test@example.com",
			want:  true,
		},
		{
			name:  "Valid Email with Subdomain",
			email: "test@sub.example.com",
			want:  true,
		},
		{
			name:  "Valid Email with Plus Sign",
			email: "test+alias@example.com",
			want:  true,
		},
		{
			name:  "Invalid Email - Missing @",
			email: "testexample.com",
			want:  false,
		},
		{
			name:  "Invalid Email - Missing Domain",
			email: "test@",
			want:  false,
		},
		{
			name:  "Invalid Email - Missing Username",
			email: "@example.com",
			want:  false,
		},
		{
			name:  "Invalid Email - Invalid Characters",
			email: "test@exa mple.com",
			want:  false,
		},
		{
			name:  "Invalid Email - Empty String",
			email: "",
			want:  false,
		},
	}

	ac := &controllers.AuthController{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ac.IsValidEmail(tt.email)
			if got != tt.want {
				t.Errorf("AuthController.IsValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthController_IsValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{
			name:     "Valid Username - Alphanumeric",
			username: "testuser123",
			want:     true,
		},
		{
			name:     "Valid Username - Underscore",
			username: "test_user",
			want:     true,
		},
		{
			name:     "Valid Username - Minimum Length",
			username: "abc",
			want:     true,
		},
		{
			name:     "Valid Username - Maximum Length",
			username: "abcdefghijklmnopqrst",
			want:     true,
		},
		{
			name:     "Invalid Username - Too Short",
			username: "ab",
			want:     false,
		},
		{
			name:     "Invalid Username - Too Long",
			username: "abcdefghijklmnopqrstu",
			want:     false,
		},
		{
			name:     "Invalid Username - Invalid Characters",
			username: "test-user",
			want:     false,
		},
		{
			name:     "Invalid Username - Empty String",
			username: "",
			want:     false,
		},
	}

	ac := &controllers.AuthController{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ac.IsValidUsername(tt.username)
			if got != tt.want {
				t.Errorf("AuthController.IsValidUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthController_IsValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "Valid Password - Meets All Requirements",
			password: "Password123!",
			want:     true,
		},
		{
			name:     "Valid Password - Special Characters",
			password: "P@ssw0rd",
			want:     true,
		},
		{
			name:     "Invalid Password - Too Short",
			password: "Pass1!",
			want:     false,
		},
		{
			name:     "Invalid Password - Missing Uppercase",
			password: "password123!",
			want:     false,
		},
		{
			name:     "Invalid Password - Missing Lowercase",
			password: "PASSWORD123!",
			want:     false,
		},
		{
			name:     "Invalid Password - Missing Number",
			password: "Password!",
			want:     false,
		},
		{
			name:     "Invalid Password - Missing Special Character",
			password: "Password123",
			want:     false,
		},
		{
			name:     "Invalid Password - Empty String",
			password: "",
			want:     false,
		},
	}

	ac := &controllers.AuthController{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ac.IsValidPassword(tt.password)
			if got != tt.want {
				t.Errorf("AuthController.IsValidPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUsernameByID(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Insert a test user into the database
	err = InsertTestUser(db, "test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Get the ID of the inserted user
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username = ?", "testuser").Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to retrieve user ID: %v", err)
	}

	tests := []struct {
		name   string
		userID int
		want   string
	}{
		{
			name:   "Valid User ID",
			userID: userID,
			want:   "testuser",
		},
		{
			name:   "Invalid User ID - Non-existent",
			userID: 999, // Non-existent user ID
			want:   "",
		},
		{
			name:   "Invalid User ID - Negative Number",
			userID: -1, // Negative user ID
			want:   "",
		},
		{
			name:   "Invalid User ID - Zero",
			userID: 0, // Zero user ID
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := controllers.GetUsernameByID(db, tt.userID)
			if got != tt.want {
				t.Errorf("GetUsernameByID() = %v, want %v", got, tt.want)
			}
		})
	}
}
