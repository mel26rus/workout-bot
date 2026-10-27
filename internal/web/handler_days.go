package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"workout-bot/internal/model"
)

const (
	formChallengePrefix = "challenge_"
	formDayPrefix       = "day_"
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

	posts, err := s.channelPostService.GetAllInfo(r.Context())
	if err != nil {
		slog.Error(fn_name+" GetAllInfo", "Error", err)
		http.Error(w, "Ошибка загрузки постов канала", http.StatusInternalServerError)
		return
	}

	sendTime, err := s.settingsService.GetSendTime(r.Context())
	if err != nil {
		slog.Error(fn_name+" GetSendTime", "Error", err)
		http.Error(w, "Ошибка загрузки времени рассылки", http.StatusInternalServerError)
		return
	}

	s.render(w, "days.html", pageData{
		Msg:        r.URL.Query().Get("msg"),
		TimeZone:   s.timeZone,
		SendTime:   sendTime,
		Challenges: challenges,
		Posts:      posts,
	})
}

// handleDaySave привязывает посты канала к дням челленджей: сохраняются только изменённые строки таблицы
func (s *Server) handleDaySave(w http.ResponseWriter, r *http.Request) {
	const fn_name = "Server.handleDaySave"
	slog.Debug("+" + fn_name)

	if err := r.ParseForm(); err != nil {
		redirectMsg(w, r, "/days", "Не удалось прочитать форму")
		return
	}

	posts, err := s.channelPostService.GetAllInfo(r.Context())
	if err != nil {
		slog.Error(fn_name+" GetAllInfo", "Error", err)
		redirectMsg(w, r, "/days", "Ошибка загрузки постов канала")
		return
	}

	var list []model.DocTabWorkout
	taken := make(map[model.DocTabWorkout]bool)

	for _, p := range posts {
		challengeValue := r.PostForm.Get(formChallengePrefix + strconv.FormatInt(p.ID, 10))
		dayValue := strings.TrimSpace(r.PostForm.Get(formDayPrefix + strconv.FormatInt(p.ID, 10)))
		if challengeValue == "" && dayValue == "" {
			continue
		}

		docWorkoutID, errChallenge := strconv.ParseInt(challengeValue, 10, 64)
		dayNumber, errDay := strconv.Atoi(dayValue)
		if errChallenge != nil || errDay != nil || dayNumber <= 0 {
			redirectMsg(w, r, "/days", fmt.Sprintf("Пост «%s»: выберите челлендж и укажите номер дня больше нуля", shortText(p.Text)))
			return
		}

		// строка не менялась
		if docWorkoutID == p.DocWorkoutID && dayNumber == p.DayNumber {
			continue
		}

		key := model.DocTabWorkout{DocWorkoutID: docWorkoutID, DayNumber: dayNumber}
		if taken[key] {
			redirectMsg(w, r, "/days", fmt.Sprintf("День %d выбран для нескольких постов сразу", dayNumber))
			return
		}
		taken[key] = true

		key.ChannelPostID = p.ID
		list = append(list, key)
	}

	if len(list) == 0 {
		redirectMsg(w, r, "/days", "Изменений нет")
		return
	}

	added, replaced, err := s.docTabWorkoutService.SaveAll(r.Context(), list)
	switch {
	case err == nil:
		redirectMsg(w, r, "/days", fmt.Sprintf("Сохранено: добавлено дней %d, заменено %d", added, replaced))
	case errors.Is(err, model.ErrDayInSchedule):
		redirectMsg(w, r, "/days", err.Error())
	default:
		slog.Error(fn_name+" SaveAll", "Error", err)
		redirectMsg(w, r, "/days", "Ошибка при сохранении тренировок")
	}
}

func (s *Server) handleSendTimeSave(w http.ResponseWriter, r *http.Request) {
	const fn_name = "Server.handleSendTimeSave"
	slog.Debug("+" + fn_name)

	value := strings.TrimSpace(r.FormValue("send_time"))
	err := s.settingsService.SetSendTime(r.Context(), value)
	switch {
	case err == nil:
		// шедулер пересчитывает ближайшую рассылку по новому времени
		s.onSendTimeChanged()
		redirectMsg(w, r, "/days", "Время рассылки сохранено: "+value)
	case errors.Is(err, model.ErrInvalidSendTime):
		redirectMsg(w, r, "/days", err.Error())
	default:
		slog.Error(fn_name, "Error", err)
		redirectMsg(w, r, "/days", "Ошибка при сохранении времени рассылки")
	}
}

// shortText обрезает подпись поста для сообщения об ошибке
func shortText(text string) string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) > 40 {
		return string(runes[:40]) + "…"
	}

	return string(runes)
}
