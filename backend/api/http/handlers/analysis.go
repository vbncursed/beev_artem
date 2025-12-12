package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/artem13815/hr/api/http/presenter"
	"github.com/artem13815/hr/pkg/analysis"
)

type AnalysisHandler struct {
	uc analysis.UseCase
}

func NewAnalysisHandler(uc analysis.UseCase) *AnalysisHandler { return &AnalysisHandler{uc: uc} }

type createAnalysisRequest struct {
	ResumeID  string `json:"resumeId"`
	VacancyID string `json:"vacancyId"`
}

// Create создаёт анализ резюме относительно вакансии, считает score и формирует отчёт.
// @Summary Создать анализ резюме по вакансии
// @Tags    Анализ
// @Accept  json
// @Produce json
// @Param   input body createAnalysisRequest true "Пара resumeId + vacancyId"
// @Security BearerAuth
// @Success 201 {object} analysis.Analysis
// @Failure 400 {object} presenter.ErrorResponse
// @Failure 401 {object} presenter.ErrorResponse
// @Failure 404 {object} presenter.ErrorResponse
// @Failure 500 {object} presenter.ErrorResponse
// @Router  /analyses [post]
func (h *AnalysisHandler) Create(c *fiber.Ctx) error {
	var req createAnalysisRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Error(c, http.StatusBadRequest, "невалидный JSON")
	}
	resumeID, err := uuid.Parse(req.ResumeID)
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "невалидный resumeId")
	}
	vacancyID, err := uuid.Parse(req.VacancyID)
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "невалидный vacancyId")
	}
	userIDStr, _ := c.Locals("userId").(string)
	actorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)

	out, err := h.uc.Create(c.Context(), actorID, isAdmin, resumeID, vacancyID)
	if err != nil {
		// упрощённо: если не нашли/нет доступа — 404
		return presenter.Error(c, http.StatusNotFound, err.Error())
	}
	return presenter.JSON(c, http.StatusCreated, out)
}

// Get возвращает анализ по id (владелец/админ).
// @Summary Получить анализ
// @Tags    Анализ
// @Produce json
// @Param   id path string true "ID анализа (UUID)"
// @Security BearerAuth
// @Success 200 {object} analysis.Analysis
// @Failure 401 {object} presenter.ErrorResponse
// @Failure 404 {object} presenter.ErrorResponse
// @Router  /analyses/{id} [get]
func (h *AnalysisHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "невалидный id")
	}
	userIDStr, _ := c.Locals("userId").(string)
	actorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)
	out, err := h.uc.Get(c.Context(), actorID, isAdmin, id)
	if err != nil {
		return presenter.Error(c, http.StatusNotFound, "анализ не найден")
	}
	return presenter.JSON(c, http.StatusOK, out)
}

// List возвращает список анализов текущего пользователя (админ — все).
// @Summary Список анализов
// @Tags    Анализ
// @Produce json
// @Security BearerAuth
// @Success 200 {array} analysis.Analysis
// @Failure 401 {object} presenter.ErrorResponse
// @Router  /analyses [get]
func (h *AnalysisHandler) List(c *fiber.Ctx) error {
	userIDStr, _ := c.Locals("userId").(string)
	actorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)
	items, err := h.uc.List(c.Context(), actorID, isAdmin, 50, 0)
	if err != nil {
		return presenter.Error(c, http.StatusInternalServerError, "не удалось получить список")
	}
	return presenter.JSON(c, http.StatusOK, items)
}

// ListByVacancy возвращает анализы по вакансии (только владелец вакансии/админ).
// @Summary Список анализов по вакансии
// @Tags    Анализ
// @Produce json
// @Param   id path string true "ID вакансии (UUID)"
// @Security BearerAuth
// @Success 200 {array} analysis.Analysis
// @Failure 401 {object} presenter.ErrorResponse
// @Failure 404 {object} presenter.ErrorResponse
// @Router  /vacancies/{id}/analyses [get]
func (h *AnalysisHandler) ListByVacancy(c *fiber.Ctx) error {
	vacancyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "невалидный id")
	}
	userIDStr, _ := c.Locals("userId").(string)
	actorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)
	items, err := h.uc.ListByVacancy(c.Context(), actorID, isAdmin, vacancyID, 50, 0)
	if err != nil {
		return presenter.Error(c, http.StatusNotFound, "анализы не найдены")
	}
	return presenter.JSON(c, http.StatusOK, items)
}

// Delete удаляет анализ (владелец/админ).
// @Summary Удалить анализ
// @Tags    Анализ
// @Param   id path string true "ID анализа (UUID)"
// @Security BearerAuth
// @Success 204 {object} nil
// @Failure 401 {object} presenter.ErrorResponse
// @Failure 404 {object} presenter.ErrorResponse
// @Router  /analyses/{id} [delete]
func (h *AnalysisHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "невалидный id")
	}
	userIDStr, _ := c.Locals("userId").(string)
	actorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)
	if err := h.uc.Delete(c.Context(), actorID, isAdmin, id); err != nil {
		return presenter.Error(c, http.StatusNotFound, "анализ не найден")
	}
	return c.SendStatus(http.StatusNoContent)
}


