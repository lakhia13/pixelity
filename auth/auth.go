package auth

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

var (
	JWTSecret = []byte("your-secret-key") // Change this in production
)

type User struct {
	gorm.Model
	Username       string `gorm:"uniqueIndex;not null"`
	Email          string `gorm:"uniqueIndex;not null"`
	Password       string `gorm:"not null"`
	Albums         []Album
	IsActive       bool      `gorm:"default:true"`
	LastLoginAt    time.Time
	ProfilePicture string
}

type Album struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	UserID      uint
	User        User
	MediaItems  []MediaItem `gorm:"many2many:album_media_items;"`
}

type MediaItem struct {
	gorm.Model
	Filename      string `gorm:"not null"`
	Path          string `gorm:"not null"`
	ThumbnailPath string
	Size          int64
	MimeType      string
	UserID        uint
	User          User
	Albums        []Album `gorm:"many2many:album_media_items;"`
}

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) RegisterUser(username, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	result := s.db.Create(&user)
	return result.Error
}

func (s *AuthService) LoginUser(username, password string) (string, error) {
	var user User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Update last login
	s.db.Model(&user).UpdateColumn("last_login_at", time.Now())

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) GetUserByID(id uint) (*User, error) {
	var user User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) UpdateUser(user *User) error {
	return s.db.Save(user).Error
}

func (s *AuthService) ValidateToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	userID := uint(claims["user_id"].(float64))
	return userID, nil
}