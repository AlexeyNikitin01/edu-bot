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
			return ctx.Send("⚠️ Вы не зарегистрированы.")
		}

		msg := ctx.Message().Text
		draft, exists := drafts[userID]
		if !exists {
			return ctx.Send("ℹ️ Начните с команды /add или выберите «➕ Добавить вопрос» в меню.")
		}

		switch draft.Step {
		case 1:
			draft.Question = msg
			draft.Step++
			return ctx.Send("🏷 Введите тэг вопроса:")
		case 2:
			draft.Tag = msg
			draft.Step++
			return ctx.Send("✅ Введите правильный ответ:")
		case 3:
			if len(draft.Answers) > 100 {
				return ctx.Send("ℹ️ нельзя больше 100 символов в ответе")
			}
			draft.Answers = append(draft.Answers, msg) // правильный
			draft.Step++
			return ctx.Send("❌ Введите неправильный ответ 1 (или /done, чтобы завершить):")
		case 4:
			if len(draft.Answers) > 100 {
				return ctx.Send("ℹ️ нельзя больше 100 символов в ответе")
			}
			if msg == "/done" {
				goto Save
			}
			draft.Answers = append(draft.Answers, msg)
			return ctx.Send("❌ Ещё неправильный ответ (или /done):")
		}

	Save:
		q := &edu.Question{
			Question: draft.Question,
			Tag:      draft.Tag,
		}
		if err := q.Insert(GetContext(ctx), boil.GetContextDB(), boil.Infer()); err != nil {
			delete(drafts, userID)
			return ctx.Send("❗ Ошибка при сохранении вопроса.")
		}

		for i, answer := range draft.Answers {
			a := edu.Answer{
				QuestionID: q.ID,
				Answer:     answer,
				IsCorrect:  i == 0,
			}
			if err := a.Insert(GetContext(ctx), boil.GetContextDB(), boil.Infer()); err != nil {
				delete(drafts, userID)
				return ctx.Send("❗ Ошибка при сохранении ответа.")
			}
		}

		uq := edu.UsersQuestion{
			QuestionID: q.ID,
			UserID:     u.TGUserID,
			IsEdu:      true,
			TimeRepeat: time.Now().Add(time.Minute * 5),
		}
		if err := uq.Insert(GetContext(ctx), boil.GetContextDB(), boil.Infer()); err != nil {
			delete(drafts, userID)
			return ctx.Send("❗ Ошибка при привязке вопроса к пользователю.")
		}

		delete(drafts, userID)
		return ctx.Send("✅ Вопрос успешно добавлен!", mainMenu())
	}
}
