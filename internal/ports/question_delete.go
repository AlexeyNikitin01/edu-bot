package ports

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

const (
	MSG_SUCESS_DELETE_QUESTION = "🤫Вопрос удален👁"
)

// deleteQuestion Обрабатывает нажатие на кнопку удаления с учетом пагинации
func deleteQuestion(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		parts := strings.Split(ctx.Data(), "_")
		if len(parts) < 3 {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Ошибка формата данных"})
		}

		questionID, err := strconv.Atoi(parts[0])
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		page, err := strconv.Atoi(parts[1])
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		tag := strings.Join(parts[2:], "_")

		_, err = edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
		).DeleteAll(GetContext(ctx), boil.GetContextDB(), false)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		btns := getQuestionBtns(ctx, tag, page)

		if len(btns) == 0 {
			tagButtons, err := getButtonsTags(ctx, domain)
			if err != nil {
				return err
			}

			return ctx.Edit(MSG_LIST_TAGS, &telebot.ReplyMarkup{
				InlineKeyboard: tagButtons,
			})
		}

		return ctx.Edit(fmt.Sprintf("%s %s (Стр. %d)", tag, MSG_LIST_QUESTION, page+1), &telebot.ReplyMarkup{
			InlineKeyboard: btns,
		})
	}
}

// deleteQuestionByTag Удаление категории вопросов
func deleteQuestionByTag(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		tag := ctx.Data()

		qs, err := edu.UsersQuestions(
			qm.InnerJoin(
				fmt.Sprintf("%s ON %s = %s",
					edu.TableNames.Questions,
					edu.UsersQuestionTableColumns.QuestionID,
					edu.QuestionTableColumns.ID,
				),
			),
			qm.InnerJoin(
				fmt.Sprintf("%s ON %s = %s",
					edu.TableNames.Tags,
					edu.QuestionTableColumns.TagID,
					edu.TagTableColumns.ID,
				),
			),
			edu.TagWhere.Tag.EQ(tag),
			edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
		).All(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if _, err = qs.DeleteAll(GetContext(ctx), boil.GetContextDB(), false); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		return showRepeatTagList(domain)(ctx)
	}
}

func deleteQuestionAfterPoll(_ app.Apper, dispatcher *QuestionDispatcher) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		_, err = edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
		).DeleteAll(GetContext(ctx), boil.GetContextDB(), false)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = ctx.Delete(); err != nil {
			return ctx.Send(err.Error())
		}

		if err = ctx.Send(MSG_SUCESS_DELETE_QUESTION); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = dispatcher.cache.SetUserWaiting(dispatcher.ctx, GetUserFromContext(ctx).TGUserID, false); err != nil {
			log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", GetUserFromContext(ctx).TGUserID, err)
		}
		return nil
	}
}

func deleteQuestionAfterPollHigh(_ app.Apper, dispatcher *QuestionDispatcher) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		_, err = edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
		).DeleteAll(GetContext(ctx), boil.GetContextDB(), false)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = ctx.Delete(); err != nil {
			return ctx.Send(err.Error())
		}

		if err = ctx.Send(MSG_SUCESS_DELETE_QUESTION); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = dispatcher.cache.SetUserWaiting(dispatcher.ctx, GetUserFromContext(ctx).TGUserID, false); err != nil {
			log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", GetUserFromContext(ctx).TGUserID, err)
		}

		return nil
	}
}
