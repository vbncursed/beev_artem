package handlers

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/artem13815/hr/api/http/presenter"
	"github.com/artem13815/hr/pkg/analysis"
	"github.com/artem13815/hr/pkg/nlp"
	"github.com/artem13815/hr/pkg/resume"
)

type AnalysisHandler struct {
	uc      analysis.UseCase
	resumes resume.Repository
}

func NewAnalysisHandler(uc analysis.UseCase, resumes resume.Repository) *AnalysisHandler {
	return &AnalysisHandler{uc: uc, resumes: resumes}
}

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
// @Failure 409 {object} presenter.ErrorResponse
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
		var notReady analysis.ErrProfileNotReady
		if errors.As(err, &notReady) {
			return presenter.Error(c, http.StatusConflict, notReady.Error())
		}
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

type candidatesResponse struct {
	Total int             `json:"total"`
	Items []candidateItem `json:"items"`
}

type candidateItem struct {
	AnalysisID string          `json:"analysisId"`
	ResumeID   string          `json:"resumeId"`
	Score      float32         `json:"score"`
	CreatedAt  string          `json:"createdAt"`
	Report     analysis.Report `json:"report"`
}

// CandidatesByVacancy возвращает "кандидатов" по вакансии: это анализы, отфильтрованные/отсортированные.
// @Summary Кандидаты по вакансии (фильтр/сортировка)
// @Description Возвращает список анализов по вакансии с фильтрацией по score и навыку, сортировкой и пагинацией.
// @Tags    Анализ
// @Produce json
// @Param   id path string true "ID вакансии (UUID)"
// @Param   minScore query number false "Минимальный score (0..1)"
// @Param   skill query string false "Навык (фильтр по resume profile skills; если профиля нет — fallback на report.matchedSkills)"
// @Param   sort query string false "Сортировка: score_desc|score_asc|created_desc|created_asc" default(score_desc)
// @Param   limit query int false "Лимит" default(50)
// @Param   offset query int false "Смещение" default(0)
// @Security BearerAuth
// @Success 200 {object} candidatesResponse
// @Failure 401 {object} presenter.ErrorResponse
// @Failure 404 {object} presenter.ErrorResponse
// @Router  /vacancies/{id}/candidates [get]
func (h *AnalysisHandler) CandidatesByVacancy(c *fiber.Ctx) error {
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

	// Query params
	minScore := float32(0)
	if v := strings.TrimSpace(c.Query("minScore")); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			minScore = float32(f)
		}
	}
	skillQuery := strings.TrimSpace(c.Query("skill"))
	skillNorm := nlp.NormalizeSkill(skillQuery)
	sortBy := strings.ToLower(strings.TrimSpace(c.Query("sort")))
	limit := 50
	offset := 0
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	// Fetch a bigger page, then filter/sort in-memory (быстро для MVP).
	const fetchLimit = 1000
	items, err := h.uc.ListByVacancy(c.Context(), actorID, isAdmin, vacancyID, fetchLimit, 0)
	if err != nil {
		return presenter.Error(c, http.StatusNotFound, "анализы не найдены")
	}

	// Cache profiles to avoid N+1 repeated reads per resume
	profileSkillsCache := make(map[uuid.UUID][]string, 128) // normalized skills list

	// Filter
	filtered := make([]analysis.Analysis, 0, len(items))
	for _, a := range items {
		if a.Score < minScore {
			continue
		}
		if skillNorm != "" {
			found := false
			// 1) Try profile skills (preferred)
			if h.resumes != nil {
				normSkills, ok := profileSkillsCache[a.ResumeID]
				if !ok {
					var rec resume.ProfileRecord
					var err error
					if isAdmin {
						rec, err = h.resumes.GetProfileAny(c.Context(), a.ResumeID)
					} else {
						rec, err = h.resumes.GetProfileForOwner(c.Context(), actorID, a.ResumeID)
					}
					if err == nil && rec.Status == resume.ProfileStatusOK {
						for _, s := range rec.Profile.Skills {
							ns := nlp.NormalizeSkill(strings.TrimSpace(s))
							if ns != "" {
								normSkills = append(normSkills, ns)
							}
						}
					}
					profileSkillsCache[a.ResumeID] = normSkills
				}
				for _, ns := range normSkills {
					if ns == skillNorm || nlp.ContainsPhrase(ns, skillNorm) || nlp.ContainsPhrase(skillNorm, ns) {
						found = true
						break
					}
				}
			}
			// 2) Fallback: matchedSkills from analysis report
			if !found {
				for _, s := range a.Report.MatchedSkills {
					ns := nlp.NormalizeSkill(strings.TrimSpace(s))
					if ns == skillNorm || nlp.ContainsPhrase(ns, skillNorm) || nlp.ContainsPhrase(skillNorm, ns) {
						found = true
						break
					}
				}
			}
			if !found {
				continue
			}
		}
		filtered = append(filtered, a)
	}

	// Sort
	switch sortBy {
	case "score_asc":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Score < filtered[j].Score })
	case "created_asc":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt.Before(filtered[j].CreatedAt) })
	case "created_desc":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	default: // score_desc
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Score > filtered[j].Score })
	}

	total := len(filtered)
	// Apply offset/limit
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	page := filtered[offset:end]

	out := make([]candidateItem, 0, len(page))
	for _, a := range page {
		out = append(out, candidateItem{
			AnalysisID: a.ID.String(),
			ResumeID:   a.ResumeID.String(),
			Score:      a.Score,
			CreatedAt:  a.CreatedAt.Format(time.RFC3339Nano),
			Report:     a.Report,
		})
	}
	return presenter.JSON(c, http.StatusOK, candidatesResponse{Total: total, Items: out})
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
