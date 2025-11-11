package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/artem13815/hr/api/http/presenter"
	"github.com/artem13815/hr/pkg/llm/openrouter"
	"github.com/artem13815/hr/pkg/resume"
)

type ResumeHandler struct {
	svc resume.AnalysisService
	llm *openrouter.Client // kept for exposing model name in response only
	// Limit uploaded file size read into memory (bytes)
	maxBytes int64
}

func NewResumeHandler(svc resume.AnalysisService, llm *openrouter.Client) *ResumeHandler {
	return &ResumeHandler{svc: svc, llm: llm, maxBytes: 15 << 20} // 15MB
}

// Analyze обрабатывает загруженное резюме (PDF/DOCX), извлекает текст
// и отправляет его в LLM для получения рекомендаций по улучшению.
// @Summary Анализ резюме и рекомендации по улучшению
// @Description Принимает файл резюме в формате PDF или DOCX, извлекает текст и запрашивает у LLM рекомендации.
// @Tags    Резюме
// @Accept  multipart/form-data
// @Produce json
// @Param   file formData file true "Файл резюме (PDF или DOCX)"
// @Success 200 {object} map[string]any "Успех: рекомендации и служебная информация"
// @Failure 400 {object} presenter.ErrorResponse "Ошибка валидации или чтения файла"
// @Failure 500 {object} presenter.ErrorResponse "Внутренняя ошибка сервиса"
// @Router  /resume/analyze [post]
func (h *ResumeHandler) Analyze(c *fiber.Ctx) error {
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
	result, err := h.svc.Analyze(c.Context(), fh.Filename, data)
	if err != nil {
		return presenter.Error(c, http.StatusBadRequest, fmt.Sprintf("analysis failed: %v", err))
	}
	return presenter.JSON(c, http.StatusOK, fiber.Map{
		"model":     h.llm.Model,
		"result":    result.Answer,
		"sizeB":     len(data),
		"filename":  result.Filename,
		"charsUsed": result.CharsUsed,
		"excerpted": result.Excerpted,
	})
}

func readAtMost(f multipart.File, max int64) ([]byte, error) {
	limited := io.LimitReader(f, max+1)
	b, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if int64(len(b)) > max {
		return nil, fmt.Errorf("file too large: limit is %d bytes", max)
	}
	return b, nil
}
