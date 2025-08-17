package ports

import (
	"strconv"
	"strings"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

const (
	INLINE_BTN_TAGS                            = "tags"
	INLINE_BTN_REPEAT_QUESTION                 = "toggle_repeat"
	INLINE_BTN_DELETE_QUESTION                 = "delete_question"
	INLINE_BTN_DELETE_QUESTIONS_BY_TAG         = "delete_tag"
	INLINE_BTN_DELETE_QUESTION_AFTER_POLL      = "delete_question_after_poll"
	INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH = "delete_question_after_poll_high"
	INLINE_BTN_REPEAT_QUESTION_AFTER_POLL      = "repeat_question_after_poll"
	INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH = "repeat_question_after_poll_high"
	INLINE_BTN_QUESTION_BY_TAG                 = "question_by_tag"
	INLINE_FORGOT_HIGH_QUESTION                = "forgot_high_question"
	INLINE_REMEMBER_HIGH_QUESTION              = "remember_high_question"
	INLINE_NEXT_QUESTION                       = "next_question"
	INLINE_EDIT_TAG                            = "edit_tag"
	INLINE_EDIT_QUESTION                       = "edit_question"
	INLINE_EDIT_NAME_QUESTION                  = "inline_edit_name_question"
	INLINE_EDIT_ANSWER_QUESTION                = "inline_edit_answer_question"
	INLINE_EDIT_NAME_TAG_QUESTION              = "inline_edit_name_tag_question"
	INLINE_BACK_TAGS                           = "back_to_tags"
	INLINE_PAUSE_TAG                           = "pause_tag"
	INLINE_BTN_QUESTION_PAGE                   = "inline_btn_page"

	INLINE_NAME_DELETE_AFTER_POLL = "🗑️"
	INLINE_NAME_REPEAT_AFTER_POLL = "️ПОВТОРЕНИЕ"
	INLINE_NAME_DELETE            = "🗑️"

	BTN_ADD_QUESTION       = "➕ Вопрос"
	BTN_MANAGMENT_QUESTION = "📚 Управление"
	BTN_ADD_CSV            = "➕ Вопросы CSV"
	BTN_NEXT_QUESTION      = "🌀 Дальше"

	MSG_WRONG_BTN = "⚠️ Неизвестная команда. Используйте меню ниже."

	MSG_GRETING = "Добро пожаловать!\n\n" +
		"🤖 Этот бот предназначен для интервального повторения собственной базы вопросов. Бот сам отправляет периодически вопросы!\n\n" +
		"✨ Выберите действие с помощью кнопок ниже:\n\n" +
		"🔹 \"➕ Вопрос\" — Создать новый вопрос вручную.\n\n" +
		"🔹 \"➕ Вопросы CSV\" — Массовая загрузка вопросов из файла CSV.\n\n" +
		"🔹 \"📚 Управление\" — Просмотр и редактирование существующих вопросов.\n\n" +
		"🔹 \"🌀 Дальше\" — Получить случайный вопрос для проверки знаний!"

	MSG_CSV = `📤 Отправьте CSV данные (файл или текст):

	Формат:
	Вопрос;Тег;Правильный ответ[;Другие ответы...]
	
	Если в вопросе/ответе есть ";", заключите его в кавычки:
	"Вопрос с ; внутри";Тег;"Ответ с ;"
	
	Примеры:
	1. Простой: Что такое GPT?;AI;Generative Pre-trained Transformer
	2. Сложный: "Что выведет: x++; y--;?";Программирование;"1; 2; 3"`

	CMD_START         = "/start"
	CMD_CANCEL string = "/cancel"
)

func routers(b *telebot.Bot, domain *app.App, dispatcher *QuestionDispatcher) {
	b.Handle(CMD_START, func(ctx telebot.Context) error {
		return ctx.Send(MSG_GRETING, mainMenu())
	})

	// INLINES BUTTONS
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_REPEAT_QUESTION}, handleToggleRepeat(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTION}, deleteQuestion(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTIONS_BY_TAG}, deleteQuestionByTag(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_TAGS}, func(c telebot.Context) error {
		return add(domain)(c)
	})
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_QUESTION_BY_TAG}, func(ctx telebot.Context) error {
		return questionByTag(ctx.Data())(ctx)
	})
	b.Handle(&telebot.InlineButton{Unique: INLINE_BACK_TAGS}, func(ctx telebot.Context) error {
		return backTags(domain)(ctx)
	})
	b.Handle(&telebot.InlineButton{Unique: INLINE_PAUSE_TAG}, func(ctx telebot.Context) error {
		return pauseTag(domain)(ctx)
	})
	b.Handle(&telebot.InlineButton{Unique: INLINE_FORGOT_HIGH_QUESTION}, forgotQuestion(domain, dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_REMEMBER_HIGH_QUESTION}, rememberQuestion(domain, dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL}, repeatQuestionAfterPoll(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH}, repeatQuestionAfterPollHigh(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL}, deleteQuestionAfterPoll(domain, dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH}, deleteQuestionAfterPollHigh(domain, dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_NEXT_QUESTION}, nextQuestion(dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_EDIT_TAG}, setEdit(edu.TableNames.Tags, domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_EDIT_QUESTION}, getForUpdate(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_EDIT_NAME_QUESTION}, setEdit(edu.QuestionTableColumns.Question, domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_EDIT_NAME_TAG_QUESTION}, setEdit(edu.QuestionTableColumns.TagID, domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_EDIT_ANSWER_QUESTION}, setEdit(edu.AnswerTableColumns.Answer, domain))

	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_QUESTION_PAGE + "_prev"}, func(ctx telebot.Context) error {
		dataParts := strings.Split(ctx.Data(), "_")
		if len(dataParts) != 2 {
			return ctx.Respond(&telebot.CallbackResponse{
				Text: "Ошибка: неверный формат данных",
			})
		}

		page, err := strconv.Atoi(dataParts[0])
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{
				Text: "Ошибка: неверный номер страницы",
			})
		}

		tag := dataParts[1]
		if tag == "" {
			return ctx.Respond(&telebot.CallbackResponse{
				Text: "Ошибка: не указан тег",
			})
		}

		return showQuestionsPage(ctx, tag, page-1)
	})

	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_QUESTION_PAGE + "_next"}, func(ctx telebot.Context) error {
		dataParts := strings.Split(ctx.Data(), "_")
		if len(dataParts) != 2 {
			return ctx.Respond(&telebot.CallbackResponse{
				Text: "Ошибка: неверный формат данных",
			})
		}

		page, err := strconv.Atoi(dataParts[0])
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{
				Text: "Ошибка: неверный номер страницы",
			})
		}

		tag := dataParts[1]
		if tag == "" {
			return ctx.Respond(&telebot.CallbackResponse{
				Text: "Ошибка: не указан тег",
			})
		}

		return showQuestionsPage(ctx, tag, page+1)
	})

	// ADD CSV
	b.Handle(telebot.OnDocument, setQuestionsByCSV(domain))

	// WORK WITH MENU
	b.Handle(telebot.OnText, func(ctx telebot.Context) error {
		// Если пользователь в процессе добавления вопроса
		if draft, ok := drafts[GetUserFromContext(ctx).TGUserID]; ok && draft.Step > 0 {
			return add(domain)(ctx)
		}

		text := ctx.Text()

		// Проверяем, может ли текст быть CSV (содержит хотя бы один разделитель)
		if strings.Contains(text, ";") && len(strings.Split(text, ";")) >= 3 {
			return setQuestionsByCSV(domain)(ctx)
		}

		switch ctx.Text() {
		case BTN_ADD_QUESTION:
			return add(domain)(ctx)
		case BTN_MANAGMENT_QUESTION:
			return showRepeatTagList(domain)(ctx)
		case BTN_ADD_CSV:
			return ctx.Send(MSG_CSV, telebot.ModeHTML)
		case BTN_NEXT_QUESTION:
			return nextQuestion(dispatcher)(ctx)
		default:
			return ctx.Send(MSG_WRONG_BTN, mainMenu())
		}
	})

	b.Handle(telebot.OnPollAnswer, checkPollAnswer(domain, dispatcher))

	// Воркер для каждого пользователя, каждые 2 секунды рассылка вопросов для пользователей
	dispatcher.StartPollingLoop()
}

func mainMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btnAdd := menu.Text(BTN_ADD_QUESTION)
	btnMark := menu.Text(BTN_MANAGMENT_QUESTION)
	btnCSV := menu.Text(BTN_ADD_CSV)
	btnNext := menu.Text(BTN_NEXT_QUESTION)

	menu.Reply(
		menu.Row(btnAdd, btnCSV),
		menu.Row(btnMark, btnNext),
	)

	return menu
}
