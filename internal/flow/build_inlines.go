package flow

import (
	"fmt"

	"workout-bot/internal/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	constWorkoutSelect = "workout:"
	constWorkoutDone   = "done:"
)

// всплывающее сообщение после нажатия «Выполнил»
const constDoneToast = "🔥 Засчитано! Так держать!"

// buildChallengesInline: по кнопке на челлендж, выбранный отмечен галочкой
func buildChallengesInline(list []model.DocWorkout, selectedID int64) *tgbotapi.InlineKeyboardMarkup {

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, c := range list {
		title := c.Name
		if c.ID == selectedID {
			title = "✅ " + title
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(
			title,
			fmt.Sprintf("%s%d", constWorkoutSelect, c.ID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	ikb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return &ikb
}

func buildDoneInline(docTabUserWorkoutID int64) *tgbotapi.InlineKeyboardMarkup {
	ikb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💪 Выполнил", fmt.Sprintf("%s%d", constWorkoutDone, docTabUserWorkoutID)),
		),
	)
	return &ikb
}

// buildEmptyInline: пустая клавиатура, по ней кнопки у сообщения убираются
func buildEmptyInline() *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{}}
}
