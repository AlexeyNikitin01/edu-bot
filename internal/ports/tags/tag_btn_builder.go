// tags/tag_btn_builder.go
package tags

import (
	"bot/internal/repo/edu"
	"fmt"
	"gopkg.in/telebot.v3"
)

type TagButtonsBuilder struct {
	tags        []*edu.Tag
	totalTags   int
	currentPage int
	totalPages  int
	pageSize    int
}

func NewTagButtonsBuilder(tags []*edu.Tag, totalTags int) *TagButtonsBuilder {
	builder := &TagButtonsBuilder{
		tags:        tags,
		totalTags:   totalTags,
		pageSize:    DEFAULT_PAGE_SIZE,
		currentPage: 1,
	}

	builder.calculatePagination()

	return builder
}

func (b *TagButtonsBuilder) calculatePagination() {
	if b.pageSize <= 0 {
		b.pageSize = DEFAULT_PAGE_SIZE
	}

	b.totalPages = (b.totalTags + b.pageSize - 1) / b.pageSize
	if b.totalPages == 0 {
		b.totalPages = 1
	}

	// Корректируем текущую страницу
	if b.currentPage < 1 {
		b.currentPage = 1
	}
	if b.currentPage > b.totalPages {
		b.currentPage = b.totalPages
	}
}

func (b *TagButtonsBuilder) WithPageSize(pageSize int) *TagButtonsBuilder {
	if pageSize > 0 && pageSize <= MAX_PAGE_SIZE {
		b.pageSize = pageSize
		b.calculatePagination()
	}
	return b
}

func (b *TagButtonsBuilder) WithCurrentPage(page int) *TagButtonsBuilder {
	if page > 0 {
		b.currentPage = page
		b.calculatePagination()
	}
	return b
}

// BuildPageRows создает строки кнопок для управления тегами (с кнопками удаления/редактирования/паузы)
func (b *TagButtonsBuilder) BuildPageRows() [][]telebot.InlineButton {
	if len(b.tags) == 0 {
		return [][]telebot.InlineButton{
			{b.buildNoTagsButton()},
		}
	}

	var rows [][]telebot.InlineButton

	pageTags := b.getTagsForCurrentPage()
	for _, tag := range pageTags {
		row := b.BuildRowForTag(tag)
		rows = append(rows, row)
	}

	paginationRow := b.buildPaginationRow()
	if len(paginationRow) > 0 {
		rows = append(rows, paginationRow)
	}

	return rows
}

// BuildTextRows создает строки кнопок только для выбора тегов (без кнопок управления)
func (b *TagButtonsBuilder) BuildTextRows() [][]telebot.InlineButton {
	if len(b.tags) == 0 {
		return nil
	}

	var rows [][]telebot.InlineButton

	pageTags := b.getTagsForCurrentPage()
	for _, tag := range pageTags {
		// Только кнопка выбора тега с другим Unique для создания вопроса
		row := []telebot.InlineButton{
			b.BuildSelectTagButton(tag),
		}
		rows = append(rows, row)
	}

	paginationRow := b.buildPaginationRow()
	if len(paginationRow) > 0 {
		rows = append(rows, paginationRow)
	}

	return rows
}

// BuildSelectTagButton создает кнопку для выбора тега при создании вопроса
func (b *TagButtonsBuilder) BuildSelectTagButton(tag *edu.Tag) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: "select_tag", // Уникальный идентификатор для выбора тега
		Text:   tag.Tag,
		Data:   tag.Tag,
	}
}

// BuildTextTags создает текстовое представление тегов для выбора
func (b *TagButtonsBuilder) BuildTextTags() string {
	if len(b.tags) == 0 {
		return "📭 У вас пока нет тегов. Введите название нового тега:"
	}

	pageTags := b.getTagsForCurrentPage()

	var tagList []string
	for i, tag := range pageTags {
		tagList = append(tagList, fmt.Sprintf("%d. %s", i+1, tag.Tag))
	}

	var message string

	// Добавляем пагинацию если нужно
	if b.totalPages > 1 {
		message += fmt.Sprintf("\n\n📄 Страница %d из %d", b.currentPage, b.totalPages)
	}

	return message
}

func (b *TagButtonsBuilder) getTagsForCurrentPage() []*edu.Tag {
	if b.pageSize <= 0 || b.currentPage <= 0 {
		return b.tags
	}

	return b.tags
}

func (b *TagButtonsBuilder) BuildRowForTag(tag *edu.Tag) []telebot.InlineButton {
	return []telebot.InlineButton{
		b.BuildTagButton(tag),
		b.BuildDeleteButton(tag),
		b.BuildEditButton(tag),
		b.BuildPauseButton(tag),
	}
}

func (b *TagButtonsBuilder) BuildTagButton(tag *edu.Tag) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: INLINE_BTN_QUESTION_BY_TAG,
		Text:   tag.Tag,
		Data:   fmt.Sprintf("%s_page_%d", tag.Tag, b.currentPage),
	}
}

func (b *TagButtonsBuilder) BuildDeleteButton(tag *edu.Tag) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: INLINE_BTN_DELETE_QUESTIONS_BY_TAG,
		Text:   INLINE_NAME_DELETE,
		Data:   tag.Tag,
	}
}

func (b *TagButtonsBuilder) BuildEditButton(tag *edu.Tag) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: INLINE_EDIT_TAG,
		Text:   EMOJI_EDIT,
		Data:   fmt.Sprintf("%d", tag.ID),
	}
}

func (b *TagButtonsBuilder) BuildPauseButton(tag *edu.Tag) telebot.InlineButton {
	label := EMOJI_BELL
	if !tag.IsPause {
		label = EMOJI_SLEEP
	}

	return telebot.InlineButton{
		Unique: INLINE_PAUSE_TAG,
		Text:   label,
		Data:   fmt.Sprintf("%d", tag.ID),
	}
}

func (b *TagButtonsBuilder) buildPaginationRow() []telebot.InlineButton {
	if b.totalPages <= 1 {
		return nil
	}

	var paginationButtons []telebot.InlineButton

	// Кнопка "Назад"
	if b.currentPage > 1 {
		paginationButtons = append(paginationButtons, telebot.InlineButton{
			Unique: INLINE_PAGINATION_PREV,
			Text:   PAGINATION_PREV_TEXT,
			Data:   fmt.Sprintf("%d", b.currentPage-1),
		})
	}

	// Информация о странице
	infoBtn := telebot.InlineButton{
		Unique: INLINE_PAGINATION_INFO,
		Text:   fmt.Sprintf(PAGINATION_INFO_FORMAT, b.currentPage, b.totalPages),
		Data:   PAGINATION_INFO_TEXT,
	}
	paginationButtons = append(paginationButtons, infoBtn)

	// Кнопка "Вперед"
	if b.currentPage < b.totalPages {
		paginationButtons = append(paginationButtons, telebot.InlineButton{
			Unique: INLINE_PAGINATION_NEXT,
			Text:   PAGINATION_NEXT_TEXT,
			Data:   fmt.Sprintf("%d", b.currentPage+1),
		})
	}

	return paginationButtons
}

func (b *TagButtonsBuilder) buildNoTagsButton() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: INLINE_NO_TAGS,
		Text:   NO_TAGS_TEXT,
		Data:   PAGINATION_INFO_TEXT,
	}
}

func (b *TagButtonsBuilder) GetPaginationInfo() string {
	if b.totalPages <= 1 {
		return fmt.Sprintf(PAGINATION_INFO_SIMPLE_FORMAT, b.totalTags)
	}

	start := (b.currentPage-1)*b.pageSize + 1
	end := start + len(b.getTagsForCurrentPage()) - 1

	return fmt.Sprintf(PAGINATION_INFO_FULL_FORMAT, start, end, b.totalTags)
}

func (b *TagButtonsBuilder) BuildSingleTagButton(tag *edu.Tag) telebot.InlineButton {
	return b.BuildTagButton(tag)
}

func (b *TagButtonsBuilder) BuildSingleDeleteButton(tag *edu.Tag) telebot.InlineButton {
	return b.BuildDeleteButton(tag)
}

func (b *TagButtonsBuilder) BuildSingleEditButton(tag *edu.Tag) telebot.InlineButton {
	return b.BuildEditButton(tag)
}

func (b *TagButtonsBuilder) BuildSinglePauseButton(tag *edu.Tag) telebot.InlineButton {
	return b.BuildPauseButton(tag)
}
