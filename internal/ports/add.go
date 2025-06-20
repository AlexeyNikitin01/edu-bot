package ports

import (
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/repo/edu"
)

type QuestionDraft struct {
	Step     int
	Question string
	Tag      string
	Answers  []string
}

var drafts = make(map[int64]*QuestionDraft)

func add() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		tgUser := ctx.Sender()
		userID := tgUser.ID

		u, err := edu.Users(edu.UserWhere.TGUserID.EQ(userID)).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Send("Вы не зарегистрированы.")
		}

		msg := ctx.Message().Text

		if msg == "/add" {
			drafts[userID] = &QuestionDraft{Step: 1}
			return ctx.Send("Введите текст вопроса:")
		}

		draft, exists := drafts[userID]
		if !exists {
			return ctx.Send("Начните с команды /add")
		}

		switch draft.Step {
		case 1:
			draft.Question = msg
			draft.Step++
			return ctx.Send("Теперь введите тэг:")
		case 2:
			draft.Tag = msg
			draft.Step++
			return ctx.Send("Теперь введите правильный ответ:")
		case 3:
			draft.Answers = append(draft.Answers, msg) // правильный
			draft.Step++
			return ctx.Send("Введите неправильный ответ 1 (или /done, чтобы завершить):")
		case 4:
			if msg == "/done" {
				goto Save
			}
			draft.Answers = append(draft.Answers, msg)
			return ctx.Send("Введите ещё неправильный ответ (или /done):")
		}

	Save:
		q := &edu.Question{
			Question: draft.Question,
			Tag:      draft.Tag,
		}
		err = q.Insert(GetContext(ctx), boil.GetContextDB(), boil.Infer())
		if err != nil {
			delete(drafts, userID)
			return ctx.Send("Ошибка при сохранении вопроса.")
		}

		for i, answer := range draft.Answers {
			a := edu.Answer{
				QuestionID: q.ID,
				Answer:     answer,
				IsCorrect:  i == 0,
			}
			if err = a.Insert(GetContext(ctx), boil.GetContextDB(), boil.Infer()); err != nil {
				delete(drafts, userID)
				return ctx.Send("Ошибка при сохранении ответа.")
			}
		}

		uq := edu.UsersQuestion{
			QuestionID: q.ID,
			UserID:     u.TGUserID,
			IsEdu:      true,
			TimeRepeat: time.Now().Add(time.Minute * 5),
		}
		if err = uq.Insert(GetContext(ctx), boil.GetContextDB(), boil.Infer()); err != nil {
			delete(drafts, userID)
			return ctx.Send("Ошибка при привязке вопроса к пользователю.")
		}

		delete(drafts, userID)
		return ctx.Send("Ваш вопрос был добавлен.")
	}
}
