package handlers

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/artem13815/hr/api/http/presenter"
	"github.com/artem13815/hr/pkg/auth"
)

type AuthHandler struct {
	useCase auth.AuthUseCase
}

func NewAuthHandler(useCase auth.AuthUseCase) *AuthHandler {
	return &AuthHandler{useCase: useCase}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles user registration.
// @Summary Register user
// @Tags    auth
// @Accept  json
// @Produce json
// @Param   input body registerRequest true "registration payload"
// @Success 201 {object} map[string]any
// @Failure 400 {object} presenter.ErrorResponse
// @Failure 409 {object} presenter.ErrorResponse
// @Router  /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Error(c, http.StatusBadRequest, "invalid JSON payload")
	}
	if strings.TrimSpace(req.Email) == "" || req.Password == "" {
		return presenter.Error(c, http.StatusBadRequest, "email and password are required")
	}

	result, err := h.useCase.Register(c.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case auth.ErrUserAlreadyExists:
			return presenter.Error(c, http.StatusConflict, "user already exists")
		default:
			return presenter.Error(c, http.StatusInternalServerError, "failed to register user")
		}
	}

	return presenter.JSON(c, http.StatusCreated, fiber.Map{
		"id":        result.User.ID.String(),
		"email":     result.User.Email,
		"createdAt": result.User.CreatedAt,
		"token":     result.Token,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles user login.
// @Summary Login
// @Tags    auth
// @Accept  json
// @Produce json
// @Param   input body loginRequest true "login payload"
// @Success 200 {object} map[string]any
// @Failure 400 {object} presenter.ErrorResponse
// @Failure 401 {object} presenter.ErrorResponse
// @Router  /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Error(c, http.StatusBadRequest, "invalid JSON payload")
	}
	if strings.TrimSpace(req.Email) == "" || req.Password == "" {
		return presenter.Error(c, http.StatusBadRequest, "email and password are required")
	}

	result, err := h.useCase.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			return presenter.Error(c, http.StatusUnauthorized, "invalid credentials")
		}
		return presenter.Error(c, http.StatusInternalServerError, "failed to login")
	}

	return presenter.JSON(c, http.StatusOK, fiber.Map{
		"id":    result.User.ID.String(),
		"email": result.User.Email,
		"token": result.Token,
	})
}
