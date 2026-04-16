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
	if body.RefreshToken != "" {
		if err := h.redis.Del(context.Background(), fmt.Sprintf("refresh:%s", body.RefreshToken)).Err(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to logout"})
		}
	}
	// Also invalidate current access token by bearer
	if token := middleware.BearerToken(c); token != "" {
		if err := h.redis.Set(context.Background(), "blacklist:"+token, "1",
			time.Duration(h.config.JWTAccessExpiryMin)*time.Minute).Err(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to logout"})
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
