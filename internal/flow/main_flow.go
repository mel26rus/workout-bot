package flow

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"workout-bot/internal/model"
	"workout-bot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MainFlow struct {
	userService              *service.UserService
	docWorkoutService        *service.DocWorkoutService
	docTabWorkoutService     *service.DocTabWorkoutService
	docTabUserWorkoutService *service.DocTabUserWorkoutService
}

// Response: что бот должен показать.
// Text пустой, Keyboard есть: меняем только кнопки сообщения.
// Оба пустые: ничего не меняем.
type Response struct {
	Text     string
	Keyboard *tgbotapi.InlineKeyboardMarkup
}

func NewMainFlow(
	user *service.UserService,
	docWorkout *service.DocWorkoutService,
	docTabWorkout *service.DocTabWorkoutService,
	docTabUserWorkout *service.DocTabUserWorkoutService,
) *MainFlow {
	return &MainFlow{
		userService:              user,
		docWorkoutService:        docWorkout,
		docTabWorkoutService:     docTabWorkout,
		docTabUserWorkoutService: docTabUserWorkout,
	}
}

// HandleStart регистрирует человека и присылает выбор челленджа, isNew == true для нового пользователя
func (f *MainFlow) HandleStart(ctx context.Context, from *tgbotapi.User) (Response, bool, error) {
	isNew, err := f.userService.GetOrCreate(ctx, model.User{
		TgID:         from.ID,
		Username:     from.UserName,
		FirstName:    from.FirstName,
		Surname:      from.LastName,
		LanguageCode: from.LanguageCode,
	})
	if err != nil {
		return Response{}, false, err
	}

	list, err := f.docWorkoutService.GetAll(ctx)
	if err != nil {
		return Response{}, isNew, err
	}
	if len(list) == 0 {
		return Response{Text: "Пока нет доступных челленджей"}, isNew, nil
	}

	text := fmt.Sprintf("Привет, %s! Выбери челлендж:", html.EscapeString(from.FirstName))
	return Response{Text: text, Keyboard: buildChallengesInline(list, 0)}, isNew, nil
}

func (f *MainFlow) HandleCallback(ctx context.Context, tgUserID int64, data string) (Response, error) {
	slog.Debug("MainFlow.HandleCallback: Got callback", "tgUserID", tgUserID, "data", data)

	switch {

	case data == constNoop:
		return Response{}, nil

	case strings.HasPrefix(data, constWorkoutSelect):
		id, err := strconv.ParseInt(strings.TrimPrefix(data, constWorkoutSelect), 10, 64)
		if err != nil {
			return Response{}, err
		}
		return f.selectChallenge(ctx, tgUserID, id)

	case strings.HasPrefix(data, constWorkoutDone):
		id, err := strconv.ParseInt(strings.TrimPrefix(data, constWorkoutDone), 10, 64)
		if err != nil {
			return Response{}, err
		}
		if err := f.docTabUserWorkoutService.MarkDone(ctx, id, tgUserID); err != nil {
			return Response{}, err
		}
		return Response{Keyboard: buildDoneStubInline()}, nil
	}

	return Response{}, fmt.Errorf("unknown callback: %s", data)
}

// selectChallenge запоминает выбор и показывает описание челленджа со ссылкой на оплату
func (f *MainFlow) selectChallenge(ctx context.Context, tgUserID int64, docWorkoutID int64) (Response, error) {
	if err := f.userService.SetDocWorkout(ctx, tgUserID, docWorkoutID); err != nil {
		return Response{}, err
	}

	w, err := f.docWorkoutService.GetByID(ctx, docWorkoutID)
	if err != nil {
		return Response{}, err
	}

	list, err := f.docWorkoutService.GetAll(ctx)
	if err != nil {
		return Response{}, err
	}

	text := fmt.Sprintf("<b>%s</b>", html.EscapeString(w.Name))
	if w.Description != "" {
		text += "\n\n" + html.EscapeString(w.Description)
	}
	if w.PaymentURL != "" {
		text += "\n\nОплата: " + html.EscapeString(w.PaymentURL)
	}

	return Response{Text: text, Keyboard: buildChallengesInline(list, docWorkoutID)}, nil
}

// BuildWorkoutResponse собирает сообщение с тренировкой дня и кнопкой «Выполнил»
func (f *MainFlow) BuildWorkoutResponse(w model.DueWorkout) Response {
	text := w.Description
	if strings.TrimSpace(text) == "" {
		text = fmt.Sprintf("Тренировка дня %d", w.DayNumber)
	}

	return Response{Text: text, Keyboard: buildDoneInline(w.ID)}
}

func (f *MainFlow) GetAdminTgIDs(ctx context.Context) ([]int64, error) {
	return f.userService.GetAdminTgIDs(ctx)
}

func (f *MainFlow) GetDueWorkouts(ctx context.Context, date time.Time) ([]model.DueWorkout, error) {
	return f.docTabUserWorkoutService.GetDue(ctx, date)
}

func (f *MainFlow) GetWorkoutFile(ctx context.Context, docTabWorkoutID int64) ([]byte, error) {
	return f.docTabWorkoutService.GetFile(ctx, docTabWorkoutID)
}

func (f *MainFlow) SaveWorkoutFileID(ctx context.Context, docTabWorkoutID int64, tgFileID string) error {
	return f.docTabWorkoutService.SetTgFileID(ctx, docTabWorkoutID, tgFileID)
}

func (f *MainFlow) MarkWorkoutSent(ctx context.Context, docTabUserWorkoutID int64) error {
	return f.docTabUserWorkoutService.MarkSent(ctx, docTabUserWorkoutID)
}
