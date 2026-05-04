package handler

import (
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maidulcu/masaar-crm/internal/api/middleware"
	"github.com/maidulcu/masaar-crm/internal/domain"
	"github.com/maidulcu/masaar-crm/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

// validatePasswordStrength returns an error message if the password doesn't meet
// minimum requirements: 8+ chars, at least one uppercase, one digit.
func validatePasswordStrength(p string) string {
	if len(p) < 8 {
		return "password must be at least 8 characters"
	}
	var hasUpper, hasDigit bool
	for _, r := range p {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasUpper {
		return "password must contain at least one uppercase letter"
	}
	if !hasDigit {
		return "password must contain at least one digit"
	}
	return ""
}

type UserHandler struct {
	users *repo.UserRepo
	audit *repo.AuditLogRepo
}

func NewUserHandler(users *repo.UserRepo, audit *repo.AuditLogRepo) *UserHandler {
	return &UserHandler{users: users, audit: audit}
}

// GetMe godoc
// @Summary      Get current user
// @Description  Returns the authenticated user's profile.
// @Tags         Users
// @Produce      json
// @Success      200  {object}  object{id=string,name=string,email=string,role=string,lang_pref=string}
// @Security     BearerAuth
// @Router       /users/me [get]
func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	claims := middleware.ClaimsFromCtx(c)
	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
	}

	user, err := h.users.FindByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(fiber.Map{
		"id":        user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"role":      user.Role,
		"lang_pref": user.LangPref,
	})
}

// ChangePassword godoc
// @Summary      Change password
// @Description  Change the authenticated user's password. Requires current password for verification.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body  body  object{current_password=string,new_password=string}  true  "Passwords"
// @Success      204
// @Failure      400  {object}  object{error=string}
// @Failure      401  {object}  object{error=string}
// @Security     BearerAuth
// @Router       /users/me/password [patch]
func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if body.CurrentPassword == "" || body.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "current_password and new_password are required"})
	}
	if msg := validatePasswordStrength(body.NewPassword); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	claims := middleware.ClaimsFromCtx(c)
	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
	}

	user, err := h.users.FindByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.CurrentPassword)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "current password is incorrect"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
	}

	if err := h.users.UpdatePassword(c.Context(), userID, string(hash)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update password"})
	}

	h.audit.Log(c.Context(), userID, repo.AuditPasswordChange, repo.AuditUser, userID, nil)
	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateLang godoc
// @Summary      Update language preference
// @Description  Update the authenticated user's language preference (ar or en).
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body  body  object{lang=string}  true  "Language"
// @Success      200   {object}  object{lang_pref=string}
// @Failure      400   {object}  object{error=string}
// @Security     BearerAuth
// @Router       /users/me/lang [patch]
func (h *UserHandler) UpdateLang(c *fiber.Ctx) error {
	var body struct {
		Lang string `json:"lang"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if body.Lang != "ar" && body.Lang != "en" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "lang must be 'ar' or 'en'"})
	}

	claims := middleware.ClaimsFromCtx(c)
	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
	}

	if err := h.users.UpdateLangPref(c.Context(), userID, body.Lang); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update language"})
	}

	return c.JSON(fiber.Map{"lang_pref": body.Lang})
}

// ListUsers godoc
// @Summary      List users
// @Description  Returns all users in the company. Admin only.
// @Tags         Users
// @Produce      json
// @Success      200  {array}   object{id=string,name=string,email=string,role=string,lang_pref=string}
// @Security     BearerAuth
// @Router       /users [get]
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.users.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch users"})
	}

	out := make([]fiber.Map, 0, len(users))
	for _, u := range users {
		out = append(out, fiber.Map{
			"id":        u.ID,
			"name":      u.Name,
			"email":     u.Email,
			"role":      u.Role,
			"lang_pref": u.LangPref,
			"wa_number": u.WANumber,
			"created_at": u.CreatedAt,
		})
	}
	return c.JSON(out)
}

// CreateUser godoc
// @Summary      Create user
// @Description  Create a new user account. Admin only. Password must be 8+ chars with uppercase and digit.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body  body  object{name=string,email=string,password=string,role=string,lang_pref=string}  true  "User data"
// @Success      201   {object}  object{id=string,name=string,email=string,role=string}
// @Failure      400   {object}  object{error=string}
// @Security     BearerAuth
// @Router       /users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var body struct {
		Name     string      `json:"name"`
		Email    string      `json:"email"`
		Password string      `json:"password"`
		Role     domain.Role `json:"role"`
		LangPref string      `json:"lang_pref"`
		WANumber string      `json:"wa_number"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if strings.TrimSpace(body.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}
	if strings.TrimSpace(body.Email) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
	}
	if msg := validatePasswordStrength(body.Password); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	validRoles := map[domain.Role]bool{
		domain.RoleAdmin:  true,
		domain.RoleAgent:  true,
		domain.RoleViewer: true,
	}
	if !validRoles[body.Role] {
		body.Role = domain.RoleAgent
	}
	if body.LangPref != "ar" && body.LangPref != "en" {
		body.LangPref = "ar"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
	}

	user := &domain.User{
		Name:         body.Name,
		Email:        body.Email,
		PasswordHash: string(hash),
		Role:         body.Role,
		LangPref:     body.LangPref,
		WANumber:     body.WANumber,
	}
	if err := h.users.Create(c.Context(), user); err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already in use"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":        user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"role":      user.Role,
		"lang_pref": user.LangPref,
	})
}
