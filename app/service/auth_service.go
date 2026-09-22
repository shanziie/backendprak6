package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type AuthService struct {
	userRepo   repository.UserRepository
	tokenRepo  repository.TokenRepository
	jwtManager *helper.JWTManager
	perms      *helper.PermissionSet
	refreshTTL time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwtManager *helper.JWTManager,
	perms *helper.PermissionSet,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
		perms:      perms,
		refreshTTL: refreshTTL,
	}
}

func (s *AuthService) Register(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if errStr := CheckPasswordStrength(req.Password); errStr != "" {
		return helper.FailValidation(c, map[string]string{"password": errStr})
	}

	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses password")
	}

	newUser, err := s.userRepo.Create(ctx, model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "user",
		IsActive: true,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(c, fiber.StatusConflict, "username atau email sudah terdaftar")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mendaftarkan pengguna")
	}

	return helper.Success(c, fiber.StatusCreated, "registrasi berhasil", newUser)
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	user, err := s.userRepo.FindByUsername(ctx, strings.TrimSpace(req.Username))
	if err != nil {
		helper.VerifyDummyPassword(req.Password)
		return helper.Fail(c, fiber.StatusUnauthorized, "username atau password salah")
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Fail(c, fiber.StatusUnauthorized, "username atau password salah")
	}

	accessToken, err := s.jwtManager.GenerateAccess(user)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token")
	}

	rawRefreshToken, err := helper.RandomToken(32)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat refresh token")
	}

	tokenHash := helper.SHA256Hex(rawRefreshToken)
	expiresAt := time.Now().Add(s.refreshTTL)

	if err := s.tokenRepo.Save(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan session")
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwtManager.AccessTTL().Seconds()),
	})
}

func (s *AuthService) Refresh(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	tokenHash := helper.SHA256Hex(strings.TrimSpace(req.RefreshToken))
	storedToken, err := s.tokenRepo.FindByHash(ctx, tokenHash)
	if err != nil || storedToken.RevokedAt != nil || time.Now().After(storedToken.ExpiresAt) {
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid atau telah kedaluwarsa")
	}

	_ = s.tokenRepo.Revoke(ctx, tokenHash)

	user, err := s.userRepo.FindByID(ctx, storedToken.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "user tidak ditemukan")
	}

	newAccessToken, err := s.jwtManager.GenerateAccess(user)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token baru")
	}

	newRawRefreshToken, err := helper.RandomToken(32)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat refresh token")
	}

	newTokenHash := helper.SHA256Hex(newRawRefreshToken)
	newExpiresAt := time.Now().Add(s.refreshTTL)
	_ = s.tokenRepo.Save(ctx, user.ID, newTokenHash, newExpiresAt)

	return helper.Success(c, fiber.StatusOK, "token berhasil diperbarui", model.TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRawRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwtManager.AccessTTL().Seconds()),
	})
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RefreshRequest
	if err := c.BodyParser(&req); err == nil && req.RefreshToken != "" {
		tokenHash := helper.SHA256Hex(strings.TrimSpace(req.RefreshToken))
		_ = s.tokenRepo.Revoke(ctx, tokenHash)
	}

	return helper.Success(c, fiber.StatusOK, "logout berhasil", nil)
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	user, err := s.userRepo.FindByID(ctx, current.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusNotFound, "user tidak ditemukan")
	}

	return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", fiber.Map{
		"user":        user,
		"permissions": s.perms.PermissionsOf(user.Role),
	})
}