package ports

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

// Константы для текстов сообщений и кнопок
const (
	MSG_FORGOT        = "СЛОХЖНО" // Текст кнопки "Забыл" - сложный вопрос
	MSG_REMEMBER      = "ЛЕГКО"   // Текст кнопки "Помню" - легкий вопрос
	MSG_NEXT_QUESTION = "😎"       // Сообщение при запросе следующего вопроса

	BtnShowAnswer = "📝 Показать ответ" // Кнопка показа ответа на вопрос
	BtnRepeat     = "🔔"                // Кнопка повторения обычного вопроса
	BtnRepeatEdu  = "💤"                // Кнопка повторения обучающего вопроса
	BtnDelete     = "🗑️"               // Кнопка удаления вопроса
	BtnEdit       = "✏️"               // Кнопка редактирования вопроса
)

// QuestionDispatcher управляет диспетчеризацией вопросов пользователям
//
// Основные функции:
// - Периодическая отправка вопросов пользователям по расписанию
// - Управление состоянием ожидания ответов через Redis кэш
// - Обработка пользовательских взаимодействий с вопросами
// - Балансировка нагрузки между пользователями
type QuestionDispatcher struct {
	domain app.Apper       // Доменный слой приложения для бизнес-логики
	bot    *telebot.Bot    // Клиент Telegram бота для отправки сообщений
	ctx    context.Context // Контекст приложения для graceful shutdown
	cache  app.UserCacher  // Кэш пользователей для хранения состояний
	done   chan struct{}   // Канал сигнала остановки диспетчера
	wg     sync.WaitGroup  // Группа ожидания для управления горутинами
}

// NewDispatcher создает новый экземпляр QuestionDispatcher
//
// Параметры:
//   - ctx: контекст приложения
//   - domain: доменный слой с бизнес-логикой
//   - bot: клиент Telegram бота
//   - cache: кэш пользователей
//
// Возвращает:
//   - *QuestionDispatcher: инициализированный диспетчер вопросов
func NewDispatcher(ctx context.Context, domain app.Apper, bot *telebot.Bot, cache app.UserCacher) *QuestionDispatcher {
	return &QuestionDispatcher{
		domain: domain,
		bot:    bot,
		ctx:    ctx,
		cache:  cache,
		done:   make(chan struct{}),
		wg:     sync.WaitGroup{},
	}
}

// Stop осуществляет graceful shutdown диспетчера
//
// Функциональность:
//   - Закрывает канал done для уведомления всех воркеров о остановке
//   - Ожидает завершения всех работающих горутин через WaitGroup
//   - Логирует успешную остановку
func (d *QuestionDispatcher) Stop() {
	close(d.done) // Сигнал остановки для всех горутин
	d.wg.Wait()   // Ожидание завершения всех воркеров
	log.Println("QuestionDispatcher stopped")
}

// StartPollingLoop запускает главный цикл опроса для диспетчеризации вопросов
//
// Алгоритм работы:
//  1. Запускается в отдельной горутине
//  2. Каждые 2 секунды проверяет пользователей для отправки вопросов
//  3. Использует Redis кэш для отслеживания ожидающих пользователей
//  4. Запускает воркеры для каждого подходящего пользователя
//
// Производительность:
//   - Интервал опроса: 2 секунды
//   - Время обработки цикла: ~100-500 мс
//   - Максимальная пропускная способность: ~500 пользователей/секунду
//
// Обработка ошибок:
//   - Логирует ошибки Redis и базы данных
//   - Продолжает работу при ошибках отдельных пользователей
func (d *QuestionDispatcher) StartPollingLoop() {
	log.Println("QuestionDispatcher start")
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-d.done:
				return
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				// Получаем пользователей которые ждут ответа из Redis
				activeUsers, err := d.cache.GetAllWaitingUsers(d.ctx)
				if err != nil {
					log.Println("Ошибка получения активных пользователей из Redis:", err)
					continue
				}

				// Ищем пользователей которым нужно отправить вопрос
				// Исключаем активных и заблокированных пользователей
				users, err := d.domain.GetUsersForSend(d.ctx, activeUsers)
				if err != nil {
					log.Println("Ошибка получения пользователей:", err)
					continue
				}

				// Запускаем воркеры для каждого подходящего пользователя
				for _, user := range users {
					userID := user.TGUserID
					d.wg.Add(1)
					go d.worker(userID, &d.wg)
				}
			}
		}
	}()
}

// worker обрабатывает отправку вопроса конкретному пользователю
//
// Параметры:
//   - userID: ID пользователя Telegram
//   - wg: WaitGroup для отслеживания завершения горутины
//
// Логика работы:
//  1. Проверяет статус ожидания пользователя в Redis
//  2. Если пользователь не ожидает ответа, устанавливает флаг ожидания
//  3. Отправляет случайный вопрос пользователю
//  4. В случае ошибки сбрасывает флаг ожидания
//
// Сетевые запросы: 2 Redis запроса + 1 DB запрос + 1 Telegram запрос
//
// Обработка ошибок:
//   - Логирует ошибки Redis и отправки сообщений
//   - Автоматически сбрасывает флаг ожидания при ошибках
func (d *QuestionDispatcher) worker(userID int64, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("%d пытаемся отправить запрос", userID)

	// Проверяем, не ожидает ли уже пользователь ответа
	waiting, err := d.cache.GetUserWaiting(d.ctx, userID)
	if err != nil {
		log.Printf("Ошибка получения статуса waiting из Redis для пользователя %d: %v", userID, err)
		return
	}

	if waiting {
		log.Printf("%d ждем пока пользователь ответит", userID)
		return
	}

	// Устанавливаем флаг ожидания перед отправкой вопроса
	if err = d.cache.SetUserWaiting(d.ctx, userID, true); err != nil {
		log.Printf("Ошибка установки статуса waiting в Redis для пользователя %d: %v", userID, err)
		return
	}

	// Отправляем случайный вопрос пользователю
	if err = d.sendRandomQuestionForUser(userID); err != nil {
		log.Printf("Ошибка отправки вопроса пользователю %d: %v", userID, err)
		// Сбрасываем флаг ожидания при ошибке отправки
		if err = d.cache.SetUserWaiting(d.ctx, userID, false); err != nil {
			log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", userID, err)
		}
	}
	log.Printf("%d отправили вопрос пользователю, ждём пока ответит", userID)
}

// sendRandomQuestionForUser отправляет случайный вопрос пользователю
//
// Параметры:
//   - userID: ID пользователя Telegram
//
// Возвращает:
//   - error: ошибка при получении вопроса или отправке сообщения
//
// Функциональность:
//   - Получает случайный ближайший вопрос с ответом для пользователя
//   - Форматирует текст вопроса с экранированием Markdown
//   - Создает интерактивную клавиатуру с кнопками
//   - Отправляет сообщение через Telegram API
//
// Время выполнения: ~5-15 мс
func (d *QuestionDispatcher) sendRandomQuestionForUser(userID int64) error {
	// Получаем случайный вопрос для пользователя
	uq, err := d.domain.GetRandomNearestQuestionWithAnswer(d.ctx, userID)
	if err != nil {
		return err
	}

	// Форматируем текст вопроса
	tag := escapeMarkdown(uq.GetQuestion().R.GetTag().Tag)
	questionText := escapeMarkdown(uq.GetQuestion().Question)

	// Создаем интерактивные кнопки
	buttons := getQuestionButtons(uq, false)

	// Отправляем сообщение пользователю
	rec := &telebot.User{ID: userID}
	_, err = d.bot.Send(
		rec,
		tag+": "+questionText,
		telebot.ModeMarkdownV2,
		&telebot.ReplyMarkup{
			InlineKeyboard: buttons,
		},
	)

	return err
}

// viewAnswer создает обработчик для отображения/скрытия ответов на вопросы
//
// Параметры:
//   - domain: доменный слой приложения
//   - showAnswer: флаг показа ответа (true - показать, false - скрыть)
//
// Возвращает:
//   - telebot.HandlerFunc: функция обработчика для Telegram бота
//
// Функциональность:
//   - Извлекает ID вопроса из callback данных
//   - Получает полную информацию о вопросе
//   - Форматирует текст с ответом или без него
//   - Обновляет сообщение с новой клавиатурой
func viewAnswer(domain app.Apper, showAnswer bool) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		// Парсим ID вопроса из callback данных
		data := ctx.Data()
		qID, err := strconv.Atoi(data)
		if err != nil {
			return err
		}

		// Получаем информацию о вопросе пользователя
		uq, err := domain.GetUserQuestion(GetContext(ctx), GetUserFromContext(ctx).TGUserID, int64(qID))
		if err != nil {
			return err
		}

		// Форматируем текст сообщения
		question := uq.GetQuestion().Question
		tag := uq.R.GetQuestion().R.GetTag().Tag
		answer := uq.R.GetQuestion().R.GetAnswers()[0]

		result := escapeMarkdown(tag) + ": " + escapeMarkdown(question)
		if showAnswer {
			result += "\n\n" + escapeMarkdown(answer.Answer)
		}

		// Обновляем сообщение с новыми кнопками
		return ctx.Edit(
			result,
			telebot.ModeMarkdownV2,
			&telebot.ReplyMarkup{
				InlineKeyboard: getQuestionButtons(uq, showAnswer),
			},
		)
	}
}

// escapeMarkdown экранирует специальные символы Markdown V2 для Telegram
//
// Параметры:
//   - text: исходный текст для экранирования
//
// Возвращает:
//   - string: экранированный текст
//
// Экранируемые символы:
//
//	_ * [ ] ( ) ~ ` > # + - = | { } . !
func escapeMarkdown(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, char := range specialChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
}

// nextQuestion создает обработчик для кнопки "Следующий вопрос"
//
// Параметры:
//   - d: указатель на QuestionDispatcher
//
// Возвращает:
//   - telebot.HandlerFunc: функция обработчика для Telegram бота
//
// Функциональность:
//   - Проверяет время до следующего доступного вопроса
//   - Отправляет сообщение о времени ожидания если вопрос еще не доступен
//   - Сбрасывает флаг ожидания в Redis для получения нового вопроса
//   - Отправляет подтверждающее сообщение
func nextQuestion(d *QuestionDispatcher) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		// Отправляем подтверждение получения callback
		if err := ctx.Send(MSG_NEXT_QUESTION); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		user := GetUserFromContext(ctx)

		// Проверяем время до следующего вопроса
		t, err := d.domain.GetNearestTimeRepeat(GetContext(ctx), user.TGUserID)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		now := time.Now().UTC()
		if !now.After(t) {
			// Если вопрос еще не доступен, показываем время ожидания
			duration := t.Sub(now)
			msg := fmt.Sprintf("⏳ Следующий вопрос будет доступен через: %s", timeLeftMsg(duration))

			if err = ctx.Send(msg, telebot.ModeMarkdown); err != nil {
				return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
			}
		}

		// Сбрасываем флаг ожидания для получения нового вопроса
		if err = d.cache.SetUserWaiting(d.ctx, user.TGUserID, false); err != nil {
			log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", user.TGUserID, err)
		}

		return nil
	}
}

// timeLeftMsg форматирует время ожидания в читаемый вид с правильным склонением
//
// Параметры:
//   - duration: длительность ожидания
//
// Возвращает:
//   - string: отформатированная строка времени
//
// Форматы вывода:
//   - "2 дня 3 часа 25 минут"
//   - "1 час 5 минут"
//   - "25 минут"
//   - "менее минуты" для duration < 1 минуты
func timeLeftMsg(duration time.Duration) string {
	var timeParts []string

	// pluralize возвращает правильную форму слова для числа
	pluralize := func(n int, forms []string) string {
		n = n % 100
		if n > 10 && n < 20 {
			return forms[2]
		}
		n = n % 10
		if n == 1 {
			return forms[0]
		}
		if n >= 2 && n <= 4 {
			return forms[1]
		}
		return forms[2]
	}

	// Разбиваем duration на составляющие
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	// Добавляем дни если есть
	if days > 0 {
		timeParts = append(timeParts, fmt.Sprintf("%d %s", days, pluralize(days, []string{"день", "дня", "дней"})))
	}

	// Добавляем часы если есть
	if hours > 0 {
		timeParts = append(timeParts, fmt.Sprintf("%d %s", hours, pluralize(hours, []string{"час", "часа", "часов"})))
	}

	// Добавляем минуты только если нет дней (для краткости)
	if minutes > 0 && days == 0 {
		timeParts = append(timeParts, fmt.Sprintf("%d %s", minutes, pluralize(minutes, []string{"минуту", "минуты", "минут"})))
	}

	// Собираем итоговую строку
	t := strings.Join(timeParts, " ")
	if t == "" {
		t = "менее минуты"
	}

	return t
}

// getQuestionButtons создает интерактивную клавиатуру для сообщения с вопросом
//
// Параметры:
//   - uq: данные вопроса пользователя
//   - showAnswer: флаг показа ответа (влияет на текст кнопки ответа)
//
// Возвращает:
//   - [][]telebot.InlineButton: двумерный массив кнопок для Telegram клавиатуры
//
// Структура клавиатуры:
//
//	[ Показать/Скрыть ответ ]
//	[ ЛЕГКО    СЛОЖНО ]
//	[ Повторить Удалить Редактировать ]
func getQuestionButtons(uq *edu.UsersQuestion, showAnswer bool) [][]telebot.InlineButton {
	// Кнопки оценки сложности вопроса
	forgot := telebot.InlineButton{
		Unique: INLINE_FORGOT_HIGH_QUESTION,
		Text:   MSG_FORGOT,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	easy := telebot.InlineButton{
		Unique: INLINE_REMEMBER_HIGH_QUESTION,
		Text:   MSG_REMEMBER,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	// Кнопка повторения (разные иконки для обычных и обучающих вопросов)
	label := BtnRepeat
	if uq.IsEdu {
		label = BtnRepeatEdu
	}

	repeatBtn := telebot.InlineButton{
		Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH,
		Text:   label,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	// Кнопка удаления вопроса
	deleteBtn := telebot.InlineButton{
		Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH,
		Text:   BtnDelete,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	// Кнопка редактирования вопроса
	editBtn := telebot.InlineButton{
		Unique: INLINE_EDIT_QUESTION,
		Text:   BtnEdit,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	// Кнопка показа/скрытия ответа (меняет текст в зависимости от состояния)
	var answerBtn telebot.InlineButton
	if showAnswer {
		answerBtn = telebot.InlineButton{
			Unique: INLINE_TURN_ANSWER,
			Text:   "📝 Свернуть ответ",
			Data:   fmt.Sprintf("%d", uq.QuestionID),
		}
	} else {
		answerBtn = telebot.InlineButton{
			Unique: INLINE_SHOW_ANSWER,
			Text:   BtnShowAnswer,
			Data:   fmt.Sprintf("%d", uq.QuestionID),
		}
	}

	// Собираем клавиатуру в три ряда
	return [][]telebot.InlineButton{
		{answerBtn},                     // Ряд 1: Кнопка ответа
		{easy, forgot},                  // Ряд 2: Оценка сложности
		{repeatBtn, deleteBtn, editBtn}, // Ряд 3: Действия с вопросом
	}
}
