package model

import "errors"

var (
	ErrUserNotFound        = errors.New("пользователь не найден")
	ErrNoChallengeSelected = errors.New("человек ещё не выбрал челлендж")
	ErrAlreadyPaid         = errors.New("оплата по этому челленджу уже подтверждена")
	ErrChallengeNotFound   = errors.New("челлендж не найден")
	ErrWorkoutDayNotFound  = errors.New("тренировка дня не найдена")
	ErrInvalidSendTime     = errors.New("время рассылки должно быть в формате ЧЧ:ММ")
	ErrDayInSchedule       = errors.New("тренировка уже в расписании участников, переместить её нельзя")
)
