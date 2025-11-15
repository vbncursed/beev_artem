package handlers

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/artem13815/hr/api/http/presenter"
	"github.com/artem13815/hr/pkg/vacancy"
)

type VacancyHandler struct {
	uc vacancy.UseCase
}

func NewVacancyHandler(uc vacancy.UseCase) *VacancyHandler { return &VacancyHandler{uc: uc} }

type skillDTO struct {
	Skill  string  `json:"skill"`
	Weight float32 `json:"weight"`
}

type createVacancyRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Skills      []skillDTO `json:"skills"`
}

// @Summary Создать вакансию
// @Description Создаёт вакансию с эталонными навыками и их весами.
// @Tags        Вакансии
// @Accept      json
// @Produce     json
// @Param       input body createVacancyRequest true "Данные вакансии"
// @Security    BearerAuth
// @Success     201 {object} map[string]any
// @Failure     400 {object} presenter.ErrorResponse
// @Router      /vacancies [post]
func (h *VacancyHandler) Create(c *fiber.Ctx) error {
	var req createVacancyRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Error(c, http.StatusBadRequest, "невалидный JSON")
	}
	userIDStr, _ := c.Locals("userId").(string)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Description) == "" {
		return presenter.Error(c, http.StatusBadRequest, "title и description обязательны")
	}
	var sw []vacancy.SkillWeight
	for _, s := range req.Skills {
		sw = append(sw, vacancy.SkillWeight{Skill: s.Skill, Weight: s.Weight})
	}
	v := vacancy.Vacancy{
		ID:          uuid.New(),
		OwnerID:     uid,
		Title:       req.Title,
		Description: req.Description,
		Skills:      sw,
	}
	v, err = h.uc.Create(c.Context(), v)
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, err.Error())
	}
	return presenter.JSON(c, http.StatusCreated, fiber.Map{
		"id":    v.ID.String(),
		"title": v.Title,
	})
}

// @Summary Получить вакансию по ID
// @Tags    Вакансии
// @Produce json
// @Param   id path string true "ID вакансии (UUID)"
// @Security BearerAuth
// @Success 200 {object} vacancy.Vacancy
// @Failure 404 {object} presenter.ErrorResponse
// @Router  /vacancies/{id} [get]
func (h *VacancyHandler) GetByID(c *fiber.Ctx) error {
	userIDStr, _ := c.Locals("userId").(string)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "невалидный UUID")
	}
	var v vacancy.Vacancy
	if isAdmin {
		v, err = h.uc.GetByIDAdmin(c.Context(), id)
	} else {
		v, err = h.uc.GetByID(c.Context(), uid, id)
	}
	if err != nil {
		return presenter.Error(c, http.StatusNotFound, "вакансия не найдена")
	}
	return presenter.JSON(c, http.StatusOK, v)
}

// @Summary Список вакансий
// @Tags    Вакансии
// @Produce json
// @Security BearerAuth
// @Router  /vacancies [get]
func (h *VacancyHandler) List(c *fiber.Ctx) error {
	userIDStr, _ := c.Locals("userId").(string)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)
	var vs []vacancy.Vacancy
	if isAdmin {
		vs, err = h.uc.ListAdmin(c.Context(), 50, 0)
	} else {
		vs, err = h.uc.List(c.Context(), uid, 50, 0)
	}
	if err != nil {
		return presenter.Error(c, http.StatusInternalServerError, "не удалось получить список")
	}
	return presenter.JSON(c, http.StatusOK, vs)
}

type updateSkillsRequest struct {
	Skills []skillDTO `json:"skills"`
}

// @Summary Обновить навыки и веса вакансии
// @Tags    Вакансии
// @Accept  json
// @Produce json
// @Param   id path string true "ID вакансии (UUID)"
// @Param   input body updateSkillsRequest true "Список навыков с весами"
// @Security BearerAuth
// @Success 204 {object} nil
// @Failure 400 {object} presenter.ErrorResponse
// @Failure 404 {object} presenter.ErrorResponse
// @Router  /vacancies/{id}/skills [put]
func (h *VacancyHandler) UpdateSkills(c *fiber.Ctx) error {
	userIDStr, _ := c.Locals("userId").(string)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "невалидный UUID")
	}
	var req updateSkillsRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Error(c, http.StatusBadRequest, "невалидный JSON")
	}
	var sw []vacancy.SkillWeight
	for _, s := range req.Skills {
		sw = append(sw, vacancy.SkillWeight{Skill: s.Skill, Weight: s.Weight})
	}
	if err := h.uc.UpdateSkills(c.Context(), uid, id, sw); err != nil {
		return presenter.Error(c, http.StatusNotFound, "вакансия не найдена")
	}
	return c.SendStatus(http.StatusNoContent)
}

// @Summary Удалить вакансию
// @Tags    Вакансии
// @Produce json
// @Param   id path string true "ID вакансии (UUID)"
// @Security BearerAuth
// @Success 204 {object} nil
// @Failure 401 {object} presenter.ErrorResponse
// @Failure 404 {object} presenter.ErrorResponse
// @Router  /vacancies/{id} [delete]
func (h *VacancyHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "невалидный UUID")
	}
	userIDStr, _ := c.Locals("userId").(string)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)
	if isAdmin {
		err = h.uc.DeleteAdmin(c.Context(), id)
	} else {
		err = h.uc.Delete(c.Context(), uid, id)
	}
	if err != nil {
		return presenter.Error(c, http.StatusNotFound, "вакансия не найдена")
	}
	return c.SendStatus(http.StatusNoContent)
}
