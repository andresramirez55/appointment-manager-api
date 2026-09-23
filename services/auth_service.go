package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/andresramirez/psych-appointments/domain"
	"github.com/andresramirez/psych-appointments/identity"
	"github.com/andresramirez/psych-appointments/models"
)

var ErrAccountLinkRequired = errors.New("esta cuenta requiere vinculación con el servicio de autenticación; contactá al administrador")

type ProfessionalRepository interface {
	FindByEmail(context.Context, string) (*models.Professional, error)
	FindByID(context.Context, int64) (*models.Professional, error)
	FindByAuthUserID(context.Context, string) (*models.Professional, error)
	Create(context.Context, *models.Professional) error
	Update(context.Context, *models.Professional) error
}
type IdentityProvider interface {
	Authenticate(context.Context, bool, string, string, string) (*identity.Session, error)
	Validate(context.Context, string) (string, error)
	Refresh(context.Context, string) (*identity.Tokens, error)
	Logout(context.Context, string) error
}
type AuthService struct {
	professionalRepo ProfessionalRepository
	identity         IdentityProvider
}

func NewAuthService(repo ProfessionalRepository, provider IdentityProvider) *AuthService {
	return &AuthService{repo, provider}
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
type UpdateProfileRequest struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Specialty string `json:"specialty"`
}
type LoginResponse struct {
	Token        string               `json:"token"`
	RefreshToken string               `json:"refresh_token,omitempty"`
	Professional *models.Professional `json:"professional"`
}

func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*LoginResponse, error) {
	if len(req.Password) < 12 {
		return nil, fmt.Errorf("la contraseña debe tener al menos 12 caracteres")
	}
	// Check the app before creating an identity; never adopt an existing profile by email.
	_, err := s.professionalRepo.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err == nil {
		return nil, ErrAccountLinkRequired
	}
	if !errors.Is(err, domain.ErrProfessionalNotFound) {
		return nil, err
	}
	session, err := s.identity.Authenticate(ctx, true, req.Email, req.Password, req.Name)
	if err != nil {
		return nil, err
	}
	return s.completeLogin(ctx, session, req.Phone, req.Specialty)
}
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	session, err := s.identity.Authenticate(ctx, false, req.Email, req.Password, "")
	if err != nil {
		return nil, err
	}
	return s.completeLogin(ctx, session, "", "")
}
func (s *AuthService) completeLogin(ctx context.Context, session *identity.Session, phone, specialty string) (*LoginResponse, error) {
	p, err := s.professionalRepo.FindByAuthUserID(ctx, session.User.ID)
	if errors.Is(err, domain.ErrProfessionalNotFound) {
		_, emailErr := s.professionalRepo.FindByEmail(ctx, session.User.Email)
		if emailErr == nil {
			err = ErrAccountLinkRequired
		} else if !errors.Is(emailErr, domain.ErrProfessionalNotFound) {
			err = emailErr
		} else {
			id := session.User.ID
			p = &models.Professional{AuthUserID: &id, Email: session.User.Email, Name: session.User.Name, Phone: phone, Specialty: specialty}
			err = s.professionalRepo.Create(ctx, p)
			// A concurrent login may have created this same identity's profile.
			if err != nil {
				if existing, e := s.professionalRepo.FindByAuthUserID(ctx, id); e == nil {
					p, err = existing, nil
				}
			}
		}
	}
	if err != nil {
		_ = s.identity.Logout(ctx, session.Tokens.RefreshToken)
		return nil, err
	}
	return &LoginResponse{Token: session.Tokens.AccessToken, RefreshToken: session.Tokens.RefreshToken, Professional: p}, nil
}
func (s *AuthService) ValidateToken(ctx context.Context, token string) (int64, error) {
	sub, err := s.identity.Validate(ctx, token)
	if err != nil {
		return 0, err
	}
	p, err := s.professionalRepo.FindByAuthUserID(ctx, sub)
	if err != nil {
		return 0, err
	}
	return p.ID, nil
}
func (s *AuthService) Refresh(ctx context.Context, refresh string) (*identity.Tokens, error) {
	pair, err := s.identity.Refresh(ctx, refresh)
	if err != nil {
		return nil, err
	}
	if _, err := s.ValidateToken(ctx, pair.AccessToken); err != nil {
		_ = s.identity.Logout(ctx, pair.RefreshToken)
		return nil, identity.ErrRejected
	}
	return pair, nil
}
func (s *AuthService) Logout(ctx context.Context, refresh string) error {
	return s.identity.Logout(ctx, refresh)
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
