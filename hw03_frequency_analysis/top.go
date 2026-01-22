package hw03frequencyanalysis

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
)

var taskWithAsteriskIsCompleted = true

var ErrNoWords = fmt.Errorf("no words")

var wordTemplate, multipleDashes *regexp.Regexp

// заполнение wordTemplate.
func init() {
	logger := &Applogger{} // создаем пустой экземпляр.
	if taskWithAsteriskIsCompleted {
		wordTemplate = SafeRegexCompile(`^[^\p{L}]+|[^\p{L}]+$`, logger)
	} else {
		wordTemplate = SafeRegexCompile("", logger)
	}
	multipleDashes = SafeRegexCompile(`^--+$`, logger)
}

// интерфейс логгера.
type Logger interface {
	Fatalf(format string, v ...any)
}

// структура, реализует интерфейс.
type Applogger struct{}

func (l *Applogger) Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}

// компилция шаблона без panic.
func SafeRegexCompile(pattern string, logger Logger) *regexp.Regexp {
	defer func() {
		if err := recover(); err != nil {
			logger.Fatalf("regexp compile: %v", err)
		}
	}()
	return regexp.MustCompile(pattern)
}

// извлечение и подсчет слов.
func ExtractWords(text, sep string) (map[string]int, error) {
	var wordSlice []string

	// разделяем на слова.
	if sep == "" {
		wordSlice = strings.Fields(text)
	} else {
		wordSlice = strings.Split(text, sep)
	}

	// мапа для подсчета слов.
	wordMap := map[string]int{}

	// заполнение мапы, подсчет слов.
	for _, word := range wordSlice {
		// не множественный дефис - извлекаем слово по шаблону. множественный дефис - добавляем как есть.
		if !multipleDashes.MatchString(word) {
			word = wordTemplate.ReplaceAllString(word, "")
		}

		// для задания со * одиночный дефис не считается словом/после обработки не осталось ничего.
		if taskWithAsteriskIsCompleted && word == "-" || word == "" {
			continue
		}

		wordMap[word]++
	}

	if len(wordMap) == 0 {
		return nil, ErrNoWords
	}

	return wordMap, nil
}

func Top10(s string) []string {
	if s == "" {
		return []string{}
	}

	// для задания со * убираем чувствительность к регистру.
	if taskWithAsteriskIsCompleted {
		s = strings.ToLower(s)
	}

	// извлечение и подсчет слов.
	wordMap, err := ExtractWords(s, "")
	if err != nil {
		if !errors.Is(err, ErrNoWords) {
			log.Println("extract words: ", err)
		}
		return []string{}
	}

	// подготовим массив для сортировки.
	type words struct {
		key   string
		value int
	}
	wordCount := len(wordMap)
	sortSlice := make([]words, wordCount)

	// заполняем массив для сортировки.
	i := 0
	for k, v := range wordMap {
		sortSlice[i].key = k
		sortSlice[i].value = v
		i++
	}

	// сортировка.
	if wordCount > 1 {
		sort.Slice(sortSlice, func(i, j int) bool {
			a, b := sortSlice[i], sortSlice[j]
			return a.value > b.value || a.value == b.value && a.key < b.key
		})
	}

	// берем первые 10 элементов.
	result := make([]string, 10)

	// заполняем результирующий слайс.
	if wordCount < 10 {
		for i := range wordCount {
			result[i] = sortSlice[i].key
		}
	} else {
		for i := range 10 {
			result[i] = sortSlice[i].key
		}
	}

	return result
}
