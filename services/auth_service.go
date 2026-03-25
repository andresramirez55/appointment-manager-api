package services

import (
	"context"
	"fmt"
	"time"

	"github.com/andresramirez/psych-appointments/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ProfessionalRepository define métodos para acceso a profesionales
type ProfessionalRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.Professional, error)
	FindByID(ctx context.Context, id int64) (*models.Professional, error)
	Create(ctx context.Context, professional *models.Professional) error
	Update(ctx context.Context, professional *models.Professional) error
}

type UpdateProfileRequest struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Specialty string `json:"specialty"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// AuthService maneja autenticación y generación de JWT
type AuthService struct {
	professionalRepo ProfessionalRepository
	jwtSecret        string
}

func NewAuthService(professionalRepo ProfessionalRepository, jwtSecret string) *AuthService {
	return &AuthService{
		professionalRepo: professionalRepo,
		jwtSecret:        jwtSecret,
	}
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Specialty string `json:"specialty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token        string               `json:"token"`
	Professional *models.Professional `json:"professional"`
}

func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*LoginResponse, error) {
	// Verificar que el email no esté en uso
	if _, err := s.professionalRepo.FindByEmail(ctx, req.Email); err == nil {
		return nil, fmt.Errorf("email already in use")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	professional := &models.Professional{
		Email:     req.Email,
		Password:  string(hash),
		Name:      req.Name,
		Phone:     req.Phone,
		Specialty: req.Specialty,
	}
	if err := s.professionalRepo.Create(ctx, professional); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	token, err := s.generateJWT(professional.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{Token: token, Professional: professional}, nil
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Buscar profesional por email
	professional, err := s.professionalRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verificar password
	if err := bcrypt.CompareHashAndPassword([]byte(professional.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generar JWT
	token, err := s.generateJWT(professional.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:        token,
		Professional: professional,
	}, nil
}

func (s *AuthService) GetProfile(ctx context.Context, professionalID int64) (*models.Professional, error) {
	return s.professionalRepo.FindByID(ctx, professionalID)
}

func (s *AuthService) UpdateProfile(ctx context.Context, professionalID int64, req *UpdateProfileRequest) (*models.Professional, error) {
	professional, err := s.professionalRepo.FindByID(ctx, professionalID)
	if err != nil {
		return nil, fmt.Errorf("professional not found")
	}
	professional.Name = req.Name
	professional.Phone = req.Phone
	professional.Specialty = req.Specialty
	if err := s.professionalRepo.Update(ctx, professional); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}
	return professional, nil
}

func (s *AuthService) UpdatePassword(ctx context.Context, professionalID int64, req *UpdatePasswordRequest) error {
	professional, err := s.professionalRepo.FindByID(ctx, professionalID)
	if err != nil {
		return fmt.Errorf("professional not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(professional.Password), []byte(req.CurrentPassword)); err != nil {
		return fmt.Errorf("contraseña actual incorrecta")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	professional.Password = string(hash)
	return s.professionalRepo.Update(ctx, professional)
}

func (s *AuthService) generateJWT(professionalID int64) (string, error) {
	claims := jwt.MapClaims{
		"professional_id": professionalID,
		"exp":             time.Now().Add(30 * 24 * time.Hour).Unix(), // Expira en 30 días
		"iat":             time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		professionalID := int64(claims["professional_id"].(float64))
		return professionalID, nil
	}

	return 0, fmt.Errorf("invalid token")
}
