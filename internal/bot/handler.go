package bot

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"workout-bot/internal/dateutil"
	"workout-bot/internal/flow"
	"workout-bot/internal/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// лимит подписи к файлу в Telegram
const captionLimit = 1024

// пауза между отправками, у Telegram лимит около 30 сообщений в секунду
const sendDelay = 50 * time.Millisecond

type Handler struct {
	bot  *tgbotapi.BotAPI
	flow *flow.MainFlow
	loc  *time.Location
}

func NewHandler(bot *tgbotapi.BotAPI, flow *flow.MainFlow, loc *time.Location) *Handler {
	return &Handler{
		bot:  bot,
		flow: flow,
		loc:  loc,
	}
}

func (h *Handler) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := h.bot.GetUpdatesChan(u)

	for {
		select {

		case <-ctx.Done():
			slog.Info("Stopping telegram handler...")
			h.bot.StopReceivingUpdates()
			return

		case update, ok := <-updates:
			if !ok {
				return
			}

			if update.CallbackQuery != nil {
				h.handleCallback(ctx, update.CallbackQuery)
			} else if update.Message != nil {
				h.handleMessage(ctx, update.Message)
			}
		}
	}
}

func (h *Handler) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	const fn_name = "Handler.handleMessage"

	if msg.From == nil || !msg.IsCommand() || msg.Command() != "start" {
		return
	}

	response, isNew, err := h.flow.HandleStart(ctx, msg.From)
	if err != nil {
		slog.Error(fn_name, "Error", err)
		_, err = h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Что-то пошло не так, попробуйте ещё раз"))
		if err != nil {
			slog.Error(fn_name+" error_1", "Error", err)
		}
		return
	}

	newMessage := tgbotapi.NewMessage(msg.Chat.ID, response.Text)
	newMessage.ParseMode = tgbotapi.ModeHTML
	if response.Keyboard != nil {
		newMessage.ReplyMarkup = response.Keyboard
	}
	_, err = h.bot.Send(newMessage)
	if err != nil {
		slog.Error(fn_name+" error_2", "Error", err)
	}

	if isNew {
		h.notifyAdmins(ctx, msg.From)
	}
}

func (h *Handler) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	const fn_name = "Handler.handleCallback"

	slog.Debug(fn_name, "userTgId", cb.From.ID, "callback data", cb.Data)

	response, err := h.flow.HandleCallback(ctx, cb.From.ID, cb.Data)
	if err != nil {
		slog.Error(fn_name, "Error", err)
		_, err = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Не получилось, нажмите /start и попробуйте ещё раз"))
		if err != nil {
			slog.Error(fn_name+" error_1", "Error", err)
		}
		return
	}

	// убираем «часики» на кнопке
	_, err = h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
	if err != nil {
		slog.Error(fn_name+" error_2", "Error", err)
	}

	if cb.Message == nil || (response.Text == "" && response.Keyboard == nil) {
		return
	}

	var edit tgbotapi.Chattable
	if response.Text != "" {
		editMessage := tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, response.Text)
		editMessage.ParseMode = tgbotapi.ModeHTML
		editMessage.ReplyMarkup = response.Keyboard
		edit = editMessage
	} else {
		// у сообщения с видео меняем только кнопки
		edit = tgbotapi.NewEditMessageReplyMarkup(cb.Message.Chat.ID, cb.Message.MessageID, *response.Keyboard)
	}

	_, err = h.bot.Request(edit)
	if err != nil && !strings.Contains(err.Error(), "message is not modified") {
		slog.Error(fn_name+" error_3", "Error", err)
	}
}

// notifyAdmins сообщает всем админам о новом пользователе
func (h *Handler) notifyAdmins(ctx context.Context, user *tgbotapi.User) {
	const fn_name = "Handler.notifyAdmins"

	adminIDs, err := h.flow.GetAdminTgIDs(ctx)
	if err != nil {
		slog.Error(fn_name, "Error", err)
		return
	}

	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if user.UserName != "" {
		name += " (@" + user.UserName + ")"
	}
	text := "Новый пользователь: " + html.EscapeString(name)

	for _, adminID := range adminIDs {
		msg := tgbotapi.NewMessage(adminID, text)
		msg.ParseMode = tgbotapi.ModeHTML
		_, err := h.bot.Send(msg)
		if err != nil {
			slog.Error(fn_name+" error_1", "adminID", adminID, "Error", err)
		}
	}
}

// HandleSendWorkouts рассылает тренировки, у которых send_date равна сегодняшней
func (h *Handler) HandleSendWorkouts(ctx context.Context) {
	const fn_name = "Handler.HandleSendWorkouts"
	slog.Debug("+" + fn_name)

	date := dateutil.Today(h.loc)
	list, err := h.flow.GetDueWorkouts(ctx, date)
	if err != nil {
		slog.Error(fn_name, "Error", err)
		return
	}
	slog.Debug(fn_name, "date", date.Format("2006-01-02"), "count", len(list))

	// file_id, полученные в этой рассылке, чтобы не грузить один и тот же файл каждому
	fileIDs := make(map[int64]string)

	for _, w := range list {
		if ctx.Err() != nil {
			return
		}

		err := h.sendWorkout(ctx, w, fileIDs)
		if err != nil {
			slog.Error(fn_name+" sendWorkout", "tgID", w.TgID, "docTabUserWorkoutID", w.ID, "Error", err)
			continue
		}

		err = h.flow.MarkWorkoutSent(ctx, w.ID)
		if err != nil {
			slog.Error(fn_name+" MarkWorkoutSent", "docTabUserWorkoutID", w.ID, "Error", err)
		}

		time.Sleep(sendDelay)
	}

	slog.Debug("-" + fn_name)
}

func (h *Handler) sendWorkout(ctx context.Context, w model.DueWorkout, fileIDs map[int64]string) error {
	response := h.flow.BuildWorkoutResponse(w)

	if !w.HasFile {
		return h.sendText(w.TgID, response)
	}

	var file tgbotapi.RequestFileData
	fileID := fileIDs[w.DocTabWorkoutID]
	if fileID == "" {
		fileID = w.TgFileID
	}
	if fileID != "" {
		file = tgbotapi.FileID(fileID)
	} else {
		data, err := h.flow.GetWorkoutFile(ctx, w.DocTabWorkoutID)
		if err != nil {
			return err
		}
		file = tgbotapi.FileBytes{Name: w.FileName, Bytes: data}
	}

	// длинный текст не влезает в подпись: файл без подписи, текст отдельным сообщением
	caption := response.Text
	separateText := utf8.RuneCountInString(caption) > captionLimit
	if separateText {
		caption = ""
	}

	var sent tgbotapi.Message
	var err error
	var newFileID string

	if strings.HasPrefix(w.MimeType, "video/") {
		video := tgbotapi.NewVideo(w.TgID, file)
		video.Caption = caption
		video.ProtectContent = true
		if !separateText {
			video.ReplyMarkup = response.Keyboard
		}
		sent, err = h.bot.Send(video)
		if err == nil && sent.Video != nil {
			newFileID = sent.Video.FileID
		}
	} else {
		document := tgbotapi.NewDocument(w.TgID, file)
		document.Caption = caption
		document.ProtectContent = true
		if !separateText {
			document.ReplyMarkup = response.Keyboard
		}
		sent, err = h.bot.Send(document)
		if err == nil && sent.Document != nil {
			newFileID = sent.Document.FileID
		}
	}
	if err != nil {
		return err
	}

	// запоминаем file_id, дальше файл уходит по нему
	if fileID == "" && newFileID != "" {
		fileIDs[w.DocTabWorkoutID] = newFileID
		err = h.flow.SaveWorkoutFileID(ctx, w.DocTabWorkoutID, newFileID)
		if err != nil {
			slog.Error("Handler.sendWorkout SaveWorkoutFileID", "docTabWorkoutID", w.DocTabWorkoutID, "Error", err)
		}
	}

	if separateText {
		return h.sendText(w.TgID, response)
	}

	return nil
}

func (h *Handler) sendText(chatID int64, response flow.Response) error {
	msg := tgbotapi.NewMessage(chatID, response.Text)
	msg.ProtectContent = true
	if response.Keyboard != nil {
		msg.ReplyMarkup = response.Keyboard
	}
	_, err := h.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send text: %w", err)
	}

	return nil
}
