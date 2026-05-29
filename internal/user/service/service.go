package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"lumiere/internal/models"
	userrepo "lumiere/internal/user/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string
	Password string
	Name     string
}

type RegisterResponse struct {
	Token string
	User  models.PublicUser
}

type LoginRequest struct {
	Username string
	Password string
}

type LoginResponse struct {
	Token string
	User  models.PublicUser
}

type Service struct {
	repo      userrepo.UserRepo
	jwtSecret string
}

func New(repo userrepo.UserRepo, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	if _, err := s.repo.GetByUsername(ctx, req.Username); err == nil {
		return nil, errors.New("username already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username: req.Username,
		Password: string(hashed),
		Name:     req.Name,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	user.Password = ""

	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	pub := user.Public()
	return &RegisterResponse{Token: signed, User: pub}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("username or password incorrect")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("username or password incorrect")
	}

	user.Password = ""

	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	pub := user.Public()
	return &LoginResponse{Token: signed, User: pub}, nil
}

// QuickLogin validates the provided JWT token and returns the public user if valid.
func (s *Service) QuickLogin(ctx context.Context, tokenStr string) (*models.PublicUser, error) {
	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	sub, ok := claims["sub"]
	if !ok {
		return nil, errors.New("missing subject in token")
	}

	var uid uint
	switch v := sub.(type) {
	case float64:
		uid = uint(v)
	case string:
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, errors.New("invalid subject in token")
		}
		uid = uint(n)
	default:
		return nil, errors.New("invalid subject type in token")
	}

	user, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	pub := user.Public()
	return &pub, nil
}
