package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/artem13815/hr/api/http/presenter"
	"github.com/artem13815/hr/pkg/resume"
)

type ResumesHandler struct {
	repo     resume.Repository
	profiles resume.ProfileUseCase
	maxBytes int64
	baseDir  string
}

func NewResumesHandler(repo resume.Repository, profiles resume.ProfileUseCase) *ResumesHandler {
	return &ResumesHandler{
		repo:     repo,
		profiles: profiles,
		maxBytes: 15 << 20, // 15MB
		baseDir:  "uploads",
	}
}

// Upload загружает файл резюме, сохраняет его на диск и извлекает текст.
// @Summary Загрузить резюме
// @Description Принимает PDF/DOCX, сохраняет файл и извлекает текст для дальнейшего анализа.
// @Tags        Резюме
// @Accept      multipart/form-data
// @Produce     json
// @Param       file formData file true "Файл резюме (PDF/DOCX)"
// @Security    BearerAuth
// @Success     201 {object} map[string]any
// @Failure     401 {object} presenter.ErrorResponse
// @Failure     400 {object} presenter.ErrorResponse
// @Failure     500 {object} presenter.ErrorResponse
// @Router      /resumes [post]
func (h *ResumesHandler) Upload(c *fiber.Ctx) error {
	fh, err := c.FormFile("file")
	if err != nil || fh == nil {
		return presenter.Error(c, http.StatusBadRequest, "file is required (pdf or docx)")
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext != ".pdf" && ext != ".docx" {
		return presenter.Error(c, http.StatusBadRequest, "unsupported file format: only pdf and docx are allowed")
	}
	file, err := fh.Open()
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "failed to open uploaded file")
	}
	defer file.Close()
	data, err := readAtMost(file, h.maxBytes)
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, err.Error())
	}
	// Save to disk
	if err := os.MkdirAll(h.baseDir, 0o755); err != nil {
		return presenter.Error(c, http.StatusInternalServerError, "failed to prepare storage")
	}
	id := uuid.New()
	dst := filepath.Join(h.baseDir, id.String()+ext)
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return presenter.Error(c, http.StatusInternalServerError, "failed to store file")
	}
	// Extract text
	txt, err := resume.ParseResumeText(fh.Filename, data)
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, fmt.Sprintf("failed to parse resume: %v", err))
	}
	ownerIDStr, _ := c.Locals("userId").(string)
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	meta := resume.Resume{
		ID:         id,
		OwnerID:    ownerID,
		Filename:   fh.Filename,
		MimeType:   fh.Header.Get("Content-Type"),
		Size:       fh.Size,
		StorageURI: dst,
	}
	if err := h.repo.Create(c.Context(), meta); err != nil {
		return presenter.Error(c, http.StatusInternalServerError, "failed to save metadata")
	}
	if err := h.repo.SaveParsed(c.Context(), resume.Parsed{ResumeID: id, Text: txt}); err != nil {
		return presenter.Error(c, http.StatusInternalServerError, "failed to save parsed text")
	}
	// Build and store structured profile (best-effort)
	var profileStatus any
	if h.profiles != nil {
		rec := h.profiles.BuildAndSave(c.Context(), id, txt)
		profileStatus = rec.Status
	}
	return presenter.JSON(c, http.StatusCreated, fiber.Map{
		"id":            id.String(),
		"filename":      fh.Filename,
		"sizeB":         fh.Size,
		"profileStatus": profileStatus,
	})
}

// List возвращает список резюме пользователя (или все, если админ).
// @Summary Список резюме
// @Tags    Резюме
// @Produce json
// @Param   limit query int false "Лимит" default(50)
// @Param   offset query int false "Смещение" default(0)
// @Security BearerAuth
// @Success 200 {array} resume.Resume
// @Failure 401 {object} presenter.ErrorResponse
// @Failure 500 {object} presenter.ErrorResponse
// @Router  /resumes [get]
func (h *ResumesHandler) List(c *fiber.Ctx) error {
	isAdmin, _ := c.Locals("isAdmin").(bool)
	userIDStr, _ := c.Locals("userId").(string)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	limit, offset := parseLimitOffset(c, 50)
	var items []resume.Resume
	if isAdmin {
		items, err = h.repo.ListAll(c.Context(), limit, offset)
	} else {
		items, err = h.repo.ListByOwner(c.Context(), uid, limit, offset)
	}
	if err != nil {
		return presenter.Error(c, http.StatusInternalServerError, "failed to list resumes")
	}
	return presenter.JSON(c, http.StatusOK, items)
}

// Get возвращает метаданные и (опционально) распарсенный текст.
// @Summary Получить резюме
// @Tags    Резюме
// @Produce json
// @Param   id path string true "ID резюме (UUID)"
// @Security BearerAuth
// @Success 200 {object} map[string]any
// @Failure 401 {object} presenter.ErrorResponse
// @Failure 404 {object} presenter.ErrorResponse
// @Router  /resumes/{id} [get]
func (h *ResumesHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "invalid id")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)
	userIDStr, _ := c.Locals("userId").(string)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	var meta resume.Resume
	if isAdmin {
		meta, err = h.repo.GetMetaAny(c.Context(), id)
	} else {
		meta, err = h.repo.GetMetaForOwner(c.Context(), uid, id)
	}
	if err != nil {
		return presenter.Error(c, http.StatusNotFound, "resume not found")
	}
	parsed, _ := h.repo.GetParsed(c.Context(), id) // may be empty if not parsed
	return presenter.JSON(c, http.StatusOK, fiber.Map{
		"meta":   meta,
		"parsed": parsed.Text,
	})
}

// Profile возвращает структурированный профиль резюме (skills/experience/education).
// @Summary Получить профиль резюме (структура)
// @Tags    Резюме
// @Produce json
// @Param   id path string true "ID резюме (UUID)"
// @Security BearerAuth
// @Success 200 {object} resume.ProfileRecord
// @Failure 401 {object} presenter.ErrorResponse
// @Failure 404 {object} presenter.ErrorResponse
// @Router  /resumes/{id}/profile [get]
func (h *ResumesHandler) Profile(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "invalid id")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)
	userIDStr, _ := c.Locals("userId").(string)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	var rec resume.ProfileRecord
	if isAdmin {
		rec, err = h.repo.GetProfileAny(c.Context(), id)
	} else {
		rec, err = h.repo.GetProfileForOwner(c.Context(), uid, id)
	}
	if err != nil {
		return presenter.Error(c, http.StatusNotFound, "profile not found")
	}
	return presenter.JSON(c, http.StatusOK, rec)
}

// RebuildProfile принудительно переизвлекает профиль резюме (например после 429/failed).
// @Summary Переизвлечь профиль резюме
// @Tags    Резюме
// @Produce json
// @Param   id path string true "ID резюме (UUID)"
// @Security BearerAuth
// @Success 200 {object} resume.ProfileRecord
// @Failure 401 {object} presenter.ErrorResponse
// @Failure 404 {object} presenter.ErrorResponse
// @Failure 500 {object} presenter.ErrorResponse
// @Router  /resumes/{id}/profile/rebuild [post]
func (h *ResumesHandler) RebuildProfile(c *fiber.Ctx) error {
	if h.profiles == nil {
		return presenter.Error(c, http.StatusInternalServerError, "profile extractor is not configured")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "invalid id")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)
	userIDStr, _ := c.Locals("userId").(string)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	// Access check via meta
	if isAdmin {
		_, err = h.repo.GetMetaAny(c.Context(), id)
	} else {
		_, err = h.repo.GetMetaForOwner(c.Context(), uid, id)
	}
	if err != nil {
		return presenter.Error(c, http.StatusNotFound, "resume not found")
	}
	parsed, err := h.repo.GetParsed(c.Context(), id)
	if err != nil {
		return presenter.Error(c, http.StatusNotFound, "parsed text not found")
	}
	rec := h.profiles.BuildAndSave(c.Context(), id, parsed.Text)
	return presenter.JSON(c, http.StatusOK, rec)
}

// Download скачивает исходный файл резюме.
// @Summary Скачать файл резюме
// @Tags    Резюме
// @Produce application/octet-stream
// @Param   id path string true "ID резюме (UUID)"
// @Security BearerAuth
// @Success 200 {file} file
// @Failure 401 {object} presenter.ErrorResponse
// @Failure 404 {object} presenter.ErrorResponse
// @Router  /resumes/{id}/file [get]
func (h *ResumesHandler) Download(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "invalid id")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)
	userIDStr, _ := c.Locals("userId").(string)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	var meta resume.Resume
	if isAdmin {
		meta, err = h.repo.GetMetaAny(c.Context(), id)
	} else {
		meta, err = h.repo.GetMetaForOwner(c.Context(), uid, id)
	}
	if err != nil {
		return presenter.Error(c, http.StatusNotFound, "resume not found")
	}
	return c.Download(meta.StorageURI, meta.Filename)
}

// Delete удаляет резюме и сопутствующие данные, а также файл на диске.
// @Summary Удалить резюме
// @Tags    Резюме
// @Param   id path string true "ID резюме (UUID)"
// @Security BearerAuth
// @Success 204 {object} nil
// @Failure 401 {object} presenter.ErrorResponse
// @Failure 404 {object} presenter.ErrorResponse
// @Router  /resumes/{id} [delete]
func (h *ResumesHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, "invalid id")
	}
	isAdmin, _ := c.Locals("isAdmin").(bool)
	userIDStr, _ := c.Locals("userId").(string)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Error(c, http.StatusUnauthorized, "не удалось определить пользователя")
	}
	var meta resume.Resume
	if isAdmin {
		meta, err = h.repo.DeleteAny(c.Context(), id)
	} else {
		meta, err = h.repo.DeleteForOwner(c.Context(), uid, id)
	}
	if err != nil {
		return presenter.Error(c, http.StatusNotFound, "resume not found")
	}
	_ = os.Remove(meta.StorageURI)
	return c.SendStatus(http.StatusNoContent)
}
