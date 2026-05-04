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
	"github.com/maidulcu/masaar-crm/internal/repo"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	users  *repo.UserRepo
	redis  *redis.Client
	config *config.Config
	audit  *repo.AuditLogRepo
}

func NewAuthHandler(users *repo.UserRepo, rdb *redis.Client, cfg *config.Config, audit *repo.AuditLogRepo) *AuthHandler {
	return &AuthHandler{users: users, redis: rdb, config: cfg, audit: audit}
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

	// Check lockout before touching the database
	lockKey := fmt.Sprintf("lockout:%s", body.Email)
	failKey := fmt.Sprintf("loginfail:%s", body.Email)
	ctx := context.Background()

	if locked, _ := h.redis.Exists(ctx, lockKey).Result(); locked > 0 {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "account temporarily locked due to too many failed attempts; try again in 15 minutes",
		})
	}

	user, err := h.users.FindByEmail(c.Context(), body.Email)
	if err != nil {
		// Increment failure counter even on unknown email to prevent enumeration timing
		h.redis.Incr(ctx, failKey)
		h.redis.Expire(ctx, failKey, 15*time.Minute)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		count, _ := h.redis.Incr(ctx, failKey).Result()
		h.redis.Expire(ctx, failKey, 15*time.Minute)
		if count >= 5 {
			h.redis.Set(ctx, lockKey, "1", 15*time.Minute)
			h.redis.Del(ctx, failKey)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "account locked for 15 minutes after too many failed attempts",
			})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	// Successful login — clear failure counter
	h.redis.Del(ctx, failKey, lockKey)

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

	h.audit.Log(c.Context(), user.ID, repo.AuditLogin, repo.AuditUser, user.ID, fiber.Map{
		"ip": c.IP(),
	})

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
	ctx := context.Background()

	// Single-use: delete the old token atomically before issuing a new one.
	// If the same token is used twice (replay attack), the second call gets 401.
	userIDStr, err := h.redis.GetDel(ctx, key).Result()
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired refresh token"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}

	// Verify user still exists and is active (admin may have deleted the account)
	user, err := h.users.FindByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account not found or deactivated"})
	}

	// Issue a new token pair (rotated refresh token)
	access, newRefresh, err := h.generateTokenPair(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "token generation failed"})
	}

	// Store rotated refresh token
	ttl := time.Duration(h.config.JWTRefreshExpiryDays) * 24 * time.Hour
	if err := h.redis.Set(ctx, fmt.Sprintf("refresh:%s", newRefresh), user.ID.String(), ttl).Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session error"})
	}

	return c.JSON(fiber.Map{
		"access_token":  access,
		"refresh_token": newRefresh,
		"expires_in":    h.config.JWTAccessExpiryMin * 60,
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

	// Log logout — best effort, user ID from JWT claims
	if sub, ok := middleware.ClaimsFromCtx(c)["sub"].(string); ok {
		if userID, err := uuid.Parse(sub); err == nil {
			h.audit.Log(c.Context(), userID, repo.AuditLogout, repo.AuditUser, userID, nil)
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
