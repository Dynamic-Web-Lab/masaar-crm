package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/maidulcu/masaar-crm/internal/api/middleware"
	"github.com/maidulcu/masaar-crm/internal/config"
	"github.com/maidulcu/masaar-crm/internal/domain"
	"github.com/maidulcu/masaar-crm/internal/email"
	"github.com/maidulcu/masaar-crm/internal/repo"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	users  *repo.UserRepo
	redis  *redis.Client
	config *config.Config
}

func NewAuthHandler(users *repo.UserRepo, rdb *redis.Client, cfg *config.Config) *AuthHandler {
	return &AuthHandler{users: users, redis: rdb, config: cfg}
}

// Login godoc
// @Summary      Login
// @Description  Authenticate with email and password. Returns JWT access + refresh tokens.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      object{email=string,password=string}  true  "Credentials"
// @Success      200   {object}  object{access_token=string,refresh_token=string,expires_in=int,user=object}
// @Failure      400   {object}  object{error=string}
// @Failure      401   {object}  object{error=string}
// @Failure      429   {object}  object{error=string}  "Too many login attempts"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	user, err := h.users.FindByEmail(c.Context(), body.Email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	access, refresh, err := h.generateTokenPair(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "token generation failed"})
	}

	// Store refresh token in Redis
	key := fmt.Sprintf("refresh:%s", refresh)
	ttl := time.Duration(h.config.JWTRefreshExpiryDays) * 24 * time.Hour
	if err := h.redis.Set(context.Background(), key, user.ID.String(), ttl).Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session error"})
	}

	return c.JSON(fiber.Map{
		"access_token":  access,
		"refresh_token": refresh,
		"expires_in":    h.config.JWTAccessExpiryMin * 60,
		"user": fiber.Map{
			"id":        user.ID,
			"name":      user.Name,
			"email":     user.Email,
			"role":      user.Role,
			"lang_pref": user.LangPref,
		},
	})
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Exchange a valid refresh token for a new access token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      object{refresh_token=string}  true  "Refresh token"
// @Success      200   {object}  object{access_token=string,expires_in=int}
// @Failure      400   {object}  object{error=string}
// @Failure      401   {object}  object{error=string}
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil || body.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "refresh_token required"})
	}

	key := fmt.Sprintf("refresh:%s", body.RefreshToken)
	userIDStr, err := h.redis.Get(context.Background(), key).Result()
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired refresh token"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}
	user, err := h.users.FindByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}

	access, _, err := h.generateTokenPair(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "token generation failed"})
	}

	return c.JSON(fiber.Map{
		"access_token": access,
		"expires_in":   h.config.JWTAccessExpiryMin * 60,
	})
}

// Logout godoc
// @Summary      Logout
// @Description  Invalidate the current session. Blacklists the access token and deletes the refresh token from Redis.
// @Tags         Auth
// @Accept       json
// @Param        body  body  object{refresh_token=string}  false  "Refresh token to invalidate"
// @Success      204
// @Security     BearerAuth
// @Router       /auth/logout [delete]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.BodyParser(&body)

	// Fail-closed on session invalidation
	if body.RefreshToken != "" {
		if err := h.redis.Del(context.Background(), fmt.Sprintf("refresh:%s", body.RefreshToken)).Err(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to revoke refresh token"})
		}
	}

	// Also invalidate current access token by bearer
	if token := middleware.BearerToken(c); token != "" {
		if err := h.redis.Set(context.Background(), "blacklist:"+token, "1", time.Duration(h.config.JWTAccessExpiryMin)*time.Minute).Err(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to revoke access token"})
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) generateTokenPair(user *domain.User) (access, refresh string, err error) {
	now := time.Now()

	accessClaims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"name": user.Name,
		"role": string(user.Role),
		"exp":  now.Add(time.Duration(h.config.JWTAccessExpiryMin) * time.Minute).Unix(),
		"iat":  now.Unix(),
	}
	access, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		return
	}

	refreshClaims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": now.Add(time.Duration(h.config.JWTRefreshExpiryDays) * 24 * time.Hour).Unix(),
		"iat": now.Unix(),
		"jti": uuid.New().String(),
	}
	refresh, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
		SignedString([]byte(h.config.JWTSecret))
	return
}

// ─── Magic Link Login ─────────────────────────────────────────────────────

// RequestMagicLink godoc
// @Summary      Request Magic Link
// @Description  Send a magic link to the user's email for passwordless login. Rate limited to 3 requests per email per hour.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      object{email=string,lang_pref=string}  true  "Email and preferred language (ar/en)"
// @Success      200   {object}  object{message=string}
// @Failure      400   {object}  object{error=string}
// @Failure      429   {object}  object{error=string}  "Rate limit exceeded"
// @Router       /auth/magic-link/request [post]
func (h *AuthHandler) RequestMagicLink(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		LangPref string `json:"lang_pref"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if body.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email required"})
	}

	if body.LangPref == "" {
		body.LangPref = "en"
	}

	ctx := context.Background()

	// Rate limiting: 3 requests per email per hour
	rateKey := fmt.Sprintf("magic_rate:%s", body.Email)
	count, err := h.redis.Get(ctx, rateKey).Int()
	if err != nil && err != redis.Nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	if count >= h.config.MagicLinkRateLimit {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "too many requests, please try again later",
		})
	}

	// Increment rate limit counter
	h.redis.Incr(ctx, rateKey)
	if count == 0 {
		h.redis.Expire(ctx, rateKey, 1*time.Hour)
	}

	// Generate magic token
	token := uuid.New().String()
	tokenKey := fmt.Sprintf("magic:%s", token)
	tokenData := fmt.Sprintf("%s|%s", body.Email, body.LangPref)

	// Store in Redis with expiry
	expiry := time.Duration(h.config.MagicLinkExpiryMin) * time.Minute
	err = h.redis.Set(ctx, tokenKey, tokenData, expiry).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}

	// Build magic link URL
	magicURL := fmt.Sprintf("%s/login?token=%s", h.config.MagicLinkBaseURL, token)

	// Send email (don't fail the request if email fails - graceful degradation)
	if emailService, ok := c.Locals("email_service").(*email.Service); ok && emailService.IsConfigured() {
		htmlBody, err := emailService.RenderMagicLinkTemplate(email.MagicLinkData{
			LoginURL:  magicURL,
			ExpiryMin: h.config.MagicLinkExpiryMin,
			Lang:      body.LangPref,
		})
		if err == nil {
			_ = emailService.Send(&domain.EmailHistory{
				ToEmail:   body.Email,
				Subject:   "Your Masaar CRM Login Link",
				HTMLBody:  htmlBody,
			})
		}
	}

	// Always return success to prevent email enumeration
	return c.JSON(fiber.Map{
		"message": "if the email exists, a magic link has been sent",
	})
}

// VerifyMagicLink godoc
// @Summary      Verify Magic Link
// @Description  Validate magic link token and issue JWT tokens. Auto-creates account if user doesn't exist.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      object{token=string}  true  "Magic link token"
// @Success      200   {object}  object{access_token=string,refresh_token=string,expires_in=int,user=object}
// @Failure      400   {object}  object{error=string}
// @Failure      401   {object}  object{error=string}
// @Router       /auth/magic-link/verify [post]
func (h *AuthHandler) VerifyMagicLink(c *fiber.Ctx) error {
	var body struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if body.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token required"})
	}

	ctx := context.Background()
	tokenKey := fmt.Sprintf("magic:%s", body.Token)

	// Get and delete token (single-use)
	tokenData, err := h.redis.Get(ctx, tokenKey).Result()
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
	}

	// Delete token to prevent reuse
	h.redis.Del(ctx, tokenKey)

	// Parse token data: email|lang_pref
	parts := split(tokenData, "|")
	if len(parts) != 2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token data"})
	}
	email := parts[0]
	langPref := parts[1]

	// Find or create user
	user, err := h.users.FindByEmail(ctx, email)
	if err != nil {
		// Auto-create account on first login
		user, err = h.users.CreateWithDefaults(ctx, email, langPref)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create account"})
		}
	}

	// Generate JWT tokens
	access, refresh, err := h.generateTokenPair(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "token generation failed"})
	}

	// Store refresh token in Redis
	key := fmt.Sprintf("refresh:%s", refresh)
	ttl := time.Duration(h.config.JWTRefreshExpiryDays) * 24 * time.Hour
	if err := h.redis.Set(ctx, key, user.ID.String(), ttl).Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session error"})
	}

	return c.JSON(fiber.Map{
		"access_token":  access,
		"refresh_token": refresh,
		"expires_in":    h.config.JWTAccessExpiryMin * 60,
		"user": fiber.Map{
			"id":        user.ID,
			"name":      user.Name,
			"email":     user.Email,
			"role":      user.Role,
			"lang_pref": user.LangPref,
		},
	})
}

// split is a simple string split helper
func split(s, sep string) []string {
	idx := len(s)
	for i := 0; i < len(s)-len(sep)+1; i++ {
		if s[i:i+len(sep)] == sep {
			idx = i
			break
		}
	}
	if idx == len(s) {
		return []string{s}
	}
	return []string{s[:idx], s[idx+len(sep):]}
}
