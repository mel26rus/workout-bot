package web

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"workout-bot/internal/model"
)

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	const fn_name = "Server.handleUsers"
	slog.Debug("+" + fn_name)

	participants, err := s.userService.GetParticipants(r.Context())
	if err != nil {
		slog.Error(fn_name, "Error", err)
		http.Error(w, "Ошибка загрузки списка", http.StatusInternalServerError)
		return
	}

	s.render(w, "users.html", pageData{
		Msg:          r.URL.Query().Get("msg"),
		Participants: participants,
	})
}

func (s *Server) handleUserPaid(w http.ResponseWriter, r *http.Request) {
	const fn_name = "Server.handleUserPaid"
	slog.Debug("+" + fn_name)

	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		redirectMsg(w, r, "/users", "Неверный id пользователя")
		return
	}

	confirmation, err := s.docUserWorkoutService.ConfirmPayment(r.Context(), userID)
	switch {
	case err == nil:
		msg := "Оплата подтверждена, старт завтра. Человеку отправлено сообщение"
		if err := s.notifier.NotifyPaymentConfirmed(r.Context(), confirmation); err != nil {
			slog.Error(fn_name+" NotifyPaymentConfirmed", "userID", userID, "Error", err)
			msg = "Оплата подтверждена, старт завтра. Сообщение человеку отправить не удалось, подробности в логе"
		}
		redirectMsg(w, r, "/users", msg)
	case errors.Is(err, model.ErrUserNotFound),
		errors.Is(err, model.ErrNoChallengeSelected),
		errors.Is(err, model.ErrAlreadyPaid):
		redirectMsg(w, r, "/users", err.Error())
	default:
		slog.Error(fn_name, "userID", userID, "Error", err)
		redirectMsg(w, r, "/users", "Ошибка при подтверждении оплаты")
	}
}
