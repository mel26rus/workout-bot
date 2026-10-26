package web

import (
	"log/slog"
	"net/http"
	"strings"

	"workout-bot/internal/model"
)

func (s *Server) handleChallenges(w http.ResponseWriter, r *http.Request) {
	const fn_name = "Server.handleChallenges"
	slog.Debug("+" + fn_name)

	challenges, err := s.docWorkoutService.GetAll(r.Context())
	if err != nil {
		slog.Error(fn_name, "Error", err)
		http.Error(w, "Ошибка загрузки челленджей", http.StatusInternalServerError)
		return
	}

	s.render(w, "challenges.html", pageData{
		Msg:        r.URL.Query().Get("msg"),
		Challenges: challenges,
	})
}

func (s *Server) handleChallengeCreate(w http.ResponseWriter, r *http.Request) {
	const fn_name = "Server.handleChallengeCreate"
	slog.Debug("+" + fn_name)

	if err := r.ParseForm(); err != nil {
		redirectMsg(w, r, "/challenges", "Не удалось прочитать форму")
		return
	}

	name := strings.TrimSpace(r.PostFormValue("name"))
	if name == "" {
		redirectMsg(w, r, "/challenges", "Укажите название")
		return
	}

	_, err := s.docWorkoutService.Create(r.Context(), model.DocWorkout{
		Name:        name,
		Description: strings.TrimSpace(r.PostFormValue("description")),
		PaymentURL:  strings.TrimSpace(r.PostFormValue("payment_url")),
	})
	if err != nil {
		slog.Error(fn_name, "Error", err)
		redirectMsg(w, r, "/challenges", "Ошибка при создании челленджа")
		return
	}

	redirectMsg(w, r, "/challenges", "Челлендж создан")
}
