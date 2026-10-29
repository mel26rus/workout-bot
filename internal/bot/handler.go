package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"workout-bot/internal/dateutil"
	"workout-bot/internal/flow"
	"workout-bot/internal/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// пауза между отправками, у Telegram лимит около 30 сообщений в секунду
const sendDelay = 50 * time.Millisecond

type Handler struct {
	bot       *tgbotapi.BotAPI
	flow      *flow.MainFlow
	loc       *time.Location
	channelID int64
}

func NewHandler(bot *tgbotapi.BotAPI, flow *flow.MainFlow, loc *time.Location, channelID int64) *Handler {
	return &Handler{
		bot:       bot,
		flow:      flow,
		loc:       loc,
		channelID: channelID,
	}
}

func (h *Handler) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	u.AllowedUpdates = []string{"message", "callback_query", "channel_post", "edited_channel_post"}

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
			} else if update.ChannelPost != nil {
				h.handleChannelPost(ctx, update.ChannelPost)
			} else if update.EditedChannelPost != nil {
				h.handleChannelPost(ctx, update.EditedChannelPost)
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

	// убираем «часики» на кнопке, response.Toast показывается всплывающим сообщением
	_, err = h.bot.Request(tgbotapi.NewCallback(cb.ID, response.Toast))
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

// handleChannelPost запоминает пост канала тренера, из него потом собирается тренировка
func (h *Handler) handleChannelPost(ctx context.Context, msg *tgbotapi.Message) {
	const fn_name = "Handler.handleChannelPost"

	if msg.Chat.ID != h.channelID {
		slog.Warn(fn_name+" пост не из канала тренера, проверьте channel_id в config.yaml", "chatID", msg.Chat.ID, "title", msg.Chat.Title)
		return
	}

	err := h.flow.SaveChannelPost(ctx, channelPostFromMessage(msg))
	if err != nil {
		slog.Error(fn_name, "messageID", msg.MessageID, "Error", err)
	}
}

// channelPostFromMessage берёт из поста текст (или подпись) и сведения о файле для таблицы на странице тренировок
func channelPostFromMessage(msg *tgbotapi.Message) model.ChannelPost {
	p := model.ChannelPost{
		ChatID:    msg.Chat.ID,
		MessageID: msg.MessageID,
		Text:      msg.Text,
		MediaType: "other",
		PostDate:  msg.Time(),
	}
	if msg.Caption != "" {
		p.Text = msg.Caption
	}

	switch {
	case msg.Video != nil:
		p.MediaType = "video"
		p.FileName = msg.Video.FileName
		p.FileSize = int64(msg.Video.FileSize)
	case msg.Document != nil:
		p.MediaType = "document"
		p.FileName = msg.Document.FileName
		p.FileSize = int64(msg.Document.FileSize)
	case len(msg.Photo) > 0:
		p.MediaType = "photo"
		p.FileSize = int64(msg.Photo[len(msg.Photo)-1].FileSize)
	case msg.Text != "":
		p.MediaType = "text"
	}

	return p
}

// NotifyPaymentConfirmed сообщает человеку, что тренер подтвердила оплату
func (h *Handler) NotifyPaymentConfirmed(ctx context.Context, c model.PaymentConfirmation) error {
	sendTime, err := h.flow.GetSendTime(ctx)
	if err != nil {
		return err
	}

	text := fmt.Sprintf(
		"Оплата челленджа «%s» получена и подтверждена. Первая тренировка %s в %s.",
		c.DocWorkoutName, c.StartDate.Format("02.01.2006"), sendTime,
	)
	_, err = h.bot.Send(tgbotapi.NewMessage(c.TgID, text))
	if err != nil {
		return fmt.Errorf("send payment confirmation: %w", err)
	}

	return nil
}

// notifyAdmins сообщает всем админам о новом пользователе
func (h *Handler) notifyAdmins(ctx context.Context, user *tgbotapi.User) {
	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if user.UserName != "" {
		name += " (@" + user.UserName + ")"
	}

	h.NotifyAdmins(ctx, "Новый пользователь: "+name)
}

// NotifyAdmins отправляет текст всем админам, ошибки пишет в лог
func (h *Handler) NotifyAdmins(ctx context.Context, text string) {
	const fn_name = "Handler.NotifyAdmins"

	adminIDs, err := h.flow.GetAdminTgIDs(ctx)
	if err != nil {
		slog.Error(fn_name, "Error", err)
		return
	}

	for _, adminID := range adminIDs {
		_, err := h.bot.Send(tgbotapi.NewMessage(adminID, text))
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

	for _, w := range list {
		if ctx.Err() != nil {
			return
		}

		err := h.sendWorkout(w)
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

// sendWorkout копирует пост канала человеку: без пометки «переслано» и с защитой от пересылки и сохранения
func (h *Handler) sendWorkout(w model.DueWorkout) error {
	msg := tgbotapi.NewCopyMessage(w.TgID, w.SourceChatID, w.SourceMessageID)
	msg.ProtectContent = true
	msg.ReplyMarkup = h.flow.BuildWorkoutKeyboard(w)

	_, err := h.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("copy message: %w", err)
	}

	return nil
}
