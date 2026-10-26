package web

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"workout-bot/internal/model"
)

func (s *Server) handleDays(w http.ResponseWriter, r *http.Request) {
	const fn_name = "Server.handleDays"
	slog.Debug("+" + fn_name)

	challenges, err := s.docWorkoutService.GetAll(r.Context())
	if err != nil {
		slog.Error(fn_name+" GetAll", "Error", err)
		http.Error(w, "Ошибка загрузки челленджей", http.StatusInternalServerError)
		return
	}

	days, err := s.docTabWorkoutService.GetAllInfo(r.Context())
	if err != nil {
		slog.Error(fn_name+" GetAllInfo", "Error", err)
		http.Error(w, "Ошибка загрузки тренировок", http.StatusInternalServerError)
		return
	}

	s.render(w, "days.html", pageData{
		Msg:        r.URL.Query().Get("msg"),
		Challenges: challenges,
		Days:       days,
	})
}

func (s *Server) handleDaySave(w http.ResponseWriter, r *http.Request) {
	const fn_name = "Server.handleDaySave"
	slog.Debug("+" + fn_name)

	maxBytes := int64(s.cfg.MaxFileMB) << 20
	// запас на текстовые поля формы
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes+(1<<20))

	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			redirectMsg(w, r, "/days", fmt.Sprintf("Файл больше %d МБ", s.cfg.MaxFileMB))
			return
		}
		slog.Error(fn_name+" ParseMultipartForm", "Error", err)
		redirectMsg(w, r, "/days", "Не удалось прочитать форму")
		return
	}
	defer r.MultipartForm.RemoveAll()

	docWorkoutID, err := strconv.ParseInt(r.FormValue("doc_workout_id"), 10, 64)
	if err != nil {
		redirectMsg(w, r, "/days", "Выберите челлендж")
		return
	}

	dayNumber, err := strconv.Atoi(r.FormValue("day_number"))
	if err != nil || dayNumber <= 0 {
		redirectMsg(w, r, "/days", "Номер дня должен быть больше нуля")
		return
	}

	workout := model.DocTabWorkout{
		DocWorkoutID: docWorkoutID,
		DayNumber:    dayNumber,
		Description:  strings.TrimSpace(r.FormValue("description")),
	}

	f, header, err := r.FormFile("file")
	switch {
	case err == nil:
		defer f.Close()
		if header.Size > maxBytes {
			redirectMsg(w, r, "/days", fmt.Sprintf("Файл больше %d МБ", s.cfg.MaxFileMB))
			return
		}
		workout.File, err = io.ReadAll(f)
		if err != nil {
			slog.Error(fn_name+" ReadAll", "Error", err)
			redirectMsg(w, r, "/days", "Не удалось прочитать файл")
			return
		}
		workout.FileName = filepath.Base(header.Filename)
		workout.MimeType = header.Header.Get("Content-Type")
		if workout.MimeType == "" {
			workout.MimeType = http.DetectContentType(workout.File)
		}
	case errors.Is(err, http.ErrMissingFile):
		// файл не приложен, это допустимо
	default:
		slog.Error(fn_name+" FormFile", "Error", err)
		redirectMsg(w, r, "/days", "Не удалось прочитать файл")
		return
	}

	replaced, err := s.docTabWorkoutService.Save(r.Context(), workout)
	if err != nil {
		slog.Error(fn_name+" Save", "Error", err)
		redirectMsg(w, r, "/days", "Ошибка при сохранении тренировки")
		return
	}

	if replaced {
		redirectMsg(w, r, "/days", fmt.Sprintf("Тренировка дня %d перезаписана", dayNumber))
		return
	}
	redirectMsg(w, r, "/days", fmt.Sprintf("Тренировка дня %d добавлена", dayNumber))
}
