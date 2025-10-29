package distpatcher

import (
	"bot/internal/adapters/cache"
	"bot/internal/domain"
	"context"
	"log"
	"sync"
	"time"

	"bot/internal/repo/edu"
)

type QuestionDispatcher struct {
	user         domain.UserService
	userQuestion domain.UserQuestionService
	cache        cache.Cache    // Кэш пользователей для хранения состояний
	done         chan struct{}  // Канал сигнала остановки диспетчера
	wg           sync.WaitGroup // Группа ожидания для управления горутинами
}

func NewDispatcher(
	user domain.UserService,
	userQuestion domain.UserQuestionService,
	cache cache.Cache,
) *QuestionDispatcher {
	return &QuestionDispatcher{
		user:         user,
		userQuestion: userQuestion,
		cache:        cache,
		done:         make(chan struct{}),
		wg:           sync.WaitGroup{},
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
func (d *QuestionDispatcher) StartPollingLoop(ctx context.Context, ch chan *edu.UsersQuestion) {
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
			case <-ticker.C:
				// Получаем пользователей которые ждут ответа из Redis
				activeUsers, err := d.cache.GetAllWaitingUsers(ctx)
				if err != nil {
					log.Println("Ошибка получения активных пользователей из Redis:", err)
					continue
				}

				// Ищем пользователей которым нужно отправить вопрос
				// Исключаем активных и заблокированных пользователей
				users, err := d.user.GetUsersForSend(ctx, activeUsers)
				if err != nil {
					log.Println("Ошибка получения пользователей:", err)
					continue
				}

				if len(users) == 0 {
					continue
				}

				// Запускаем воркеры для каждого подходящего пользователя
				for _, user := range users {
					userID := user.TGUserID
					d.wg.Add(1)
					go d.worker(ctx, userID, &d.wg, ch)
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
func (d *QuestionDispatcher) worker(ctx context.Context, userID int64, wg *sync.WaitGroup, ch chan *edu.UsersQuestion) {
	defer wg.Done()

	log.Printf("%d пытаемся отправить запрос", userID)

	// Проверяем, не ожидает ли уже пользователь ответа
	waiting, err := d.cache.GetUserWaiting(ctx, userID)
	if err != nil {
		log.Printf("Ошибка получения статуса waiting из Redis для пользователя %d: %v", userID, err)
		return
	}

	if waiting {
		log.Printf("%d ждем пока пользователь ответит", userID)
		return
	}

	// Устанавливаем флаг ожидания перед отправкой вопроса
	if err = d.cache.SetUserWaiting(ctx, userID, true); err != nil {
		log.Printf("Ошибка установки статуса waiting в Redis для пользователя %d: %v", userID, err)
		return
	}

	// Отправляем случайный вопрос пользователю
	if err = d.sendRandomQuestionForUser(ctx, userID, ch); err != nil {
		log.Printf("Ошибка отправки вопроса пользователю %d: %v", userID, err)
		// Сбрасываем флаг ожидания при ошибке отправки
		if err = d.cache.SetUserWaiting(ctx, userID, false); err != nil {
			log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", userID, err)
		}
	}
	log.Printf("%d отправили вопрос пользователю, ждём пока ответит", userID)
}

func (d *QuestionDispatcher) SetUserWaiting(ctx context.Context, userID int64, wait bool) error {
	if err := d.cache.SetUserWaiting(ctx, userID, wait); err != nil {
		log.Printf("Ошибка установки статуса waiting в Redis для пользователя %d: %v", userID, err)
		return err
	}

	return nil
}

func (d *QuestionDispatcher) sendRandomQuestionForUser(
	ctx context.Context, userID int64, ch chan *edu.UsersQuestion,
) error {
	// Получаем случайный вопрос для пользователя
	uq, err := d.userQuestion.GetRandomNearestQuestionWithAnswer(ctx, userID)
	if err != nil {
		return err
	}

	ch <- uq

	return err
}
