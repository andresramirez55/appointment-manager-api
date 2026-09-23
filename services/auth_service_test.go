package services

import (
	"context"
	"errors"
	"testing"

	"github.com/andresramirez/psych-appointments/domain"
	"github.com/andresramirez/psych-appointments/identity"
	"github.com/andresramirez/psych-appointments/models"
)

type authRepoFake struct {
	profile *models.Professional
	created int
	failure error
}

func (r *authRepoFake) FindByEmail(context.Context, string) (*models.Professional, error) {
	if r.failure != nil {
		return nil, r.failure
	}
	if r.profile != nil {
		return r.profile, nil
	}
	return nil, domain.ErrProfessionalNotFound
}
func (r *authRepoFake) FindByID(context.Context, int64) (*models.Professional, error) {
	return r.profile, nil
}
func (r *authRepoFake) FindByAuthUserID(_ context.Context, id string) (*models.Professional, error) {
	if r.failure != nil {
		return nil, r.failure
	}
	if r.profile != nil && r.profile.AuthUserID != nil && *r.profile.AuthUserID == id {
		return r.profile, nil
	}
	return nil, domain.ErrProfessionalNotFound
}
func (r *authRepoFake) Create(_ context.Context, p *models.Professional) error {
	r.created++
	p.ID = 42
	r.profile = p
	return nil
}
func (r *authRepoFake) Update(context.Context, *models.Professional) error { return nil }

type providerFake struct {
	revoked int
	logins  int
}

func (p *providerFake) Authenticate(context.Context, bool, string, string, string) (*identity.Session, error) {
	p.logins++
	return &identity.Session{User: identity.User{ID: "auth-user", Email: "doctor@example.com", Name: "Doctor"}, Tokens: identity.Tokens{AccessToken: "access", RefreshToken: "refresh"}}, nil
}
func (*providerFake) Validate(context.Context, string) (string, error) { return "auth-user", nil }
func (*providerFake) Refresh(context.Context, string) (*identity.Tokens, error) {
	return &identity.Tokens{AccessToken: "access", RefreshToken: "rotated"}, nil
}
func (p *providerFake) Logout(context.Context, string) error { p.revoked++; return nil }

func TestLoginNeverLinksExistingProfileByEmail(t *testing.T) {
	repo := &authRepoFake{profile: &models.Professional{ID: 7, Email: "doctor@example.com", Password: "legacy-hash"}}
	provider := &providerFake{}
	service := NewAuthService(repo, provider)
	_, err := service.Login(context.Background(), &LoginRequest{})
	if !errors.Is(err, ErrAccountLinkRequired) || repo.created != 0 || repo.profile.AuthUserID != nil || provider.revoked != 1 {
		t.Fatalf("unsafe account linkage: %v", err)
	}
	if _, err := service.ValidateToken(context.Background(), "access"); err == nil {
		t.Fatal("identity without linked profile authorized")
	}
}
func TestNewAccountProfileAndStableAuthorization(t *testing.T) {
	repo := &authRepoFake{}
	provider := &providerFake{}
	service := NewAuthService(repo, provider)
	response, err := service.Register(context.Background(), &RegisterRequest{Email: "doctor@example.com", Password: "secure-password", Phone: "123", Specialty: "Psychology"})
	if err != nil {
		t.Fatal(err)
	}
	if response.Token != "access" || repo.profile.Password != "" || repo.profile.Phone != "123" || repo.created != 1 {
		t.Fatal("incorrect profile creation")
	}
	response, err = service.Login(context.Background(), &LoginRequest{})
	if err != nil || response.Professional.ID != 42 || repo.created != 1 {
		t.Fatalf("profile identity changed: %v", err)
	}
	id, err := service.ValidateToken(context.Background(), "access")
	if err != nil || id != 42 {
		t.Fatalf("wrong professional authorization: %v", err)
	}
}
func TestDatabaseFailureDoesNotCreateIdentity(t *testing.T) {
	repo := &authRepoFake{failure: errors.New("database unavailable")}
	provider := &providerFake{}
	_, err := NewAuthService(repo, provider).Register(context.Background(), &RegisterRequest{Password: "secure-password"})
	if err == nil || provider.logins != 0 {
		t.Fatal("registered identity despite database failure")
	}
}
