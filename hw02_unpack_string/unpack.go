package hw02unpackstring

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

// проверка на цифру.
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// проверка на слэш.
func isSlash(r rune) bool {
	return string(r) == `\`
}

// проверям допустимость комбинации текущего и следующего рун.
func IsCorrect(current, next rune) bool {
	switch {
	case isDigit(current) && isDigit(next):
		return false
	case isSlash(current) && !isDigit(next) && !isSlash(next):
		return false
	default:
		return true
	}
}

func ValidateRunes(runes []rune) error {
	runeCount := len(runes)

	// проверка первых и последних символов.
	if isDigit(runes[0]) || // первый символ число.
		isSlash(runes[runeCount-1]) { // последний символ слэш.
		return ErrInvalidString
	}

	for i := 0; i < runeCount; i++ {
		if isSlash(runes[i]) && IsCorrect(runes[i], runes[i+1]) {
			i++
			continue
		}

		if i+1 < runeCount && !IsCorrect(runes[i], runes[i+1]) {
			return ErrInvalidString
		}
	}
	return nil
}

// распаковка строки.
func Unpack(s string) (string, error) {
	var builder strings.Builder
	runes := []rune(s)
	runeCount := len(runes)

	if runeCount == 0 {
		return "", nil
	}

	// валидация.
	err := ValidateRunes(runes)
	if err != nil {
		return "", fmt.Errorf("validate runes: %w", err)
	}

	strSlice := []string{}
	sIdx := 0

	// распаковка.
	for i := 0; i < runeCount; i++ {
		switch {
		case isSlash(runes[i]):
			strSlice = append(strSlice, string(runes[i+1]))
			i++
			sIdx++
		case isDigit(runes[i]):
			digit, err := strconv.Atoi(string(runes[i]))
			if err != nil {
				return "", fmt.Errorf("convert to digit: %w", err)
			}
			strSlice[sIdx-1] = strings.Repeat(string(runes[i-1]), digit)
		default:
			strSlice = append(strSlice, string(runes[i]))
			sIdx++
		}
	}

	// посчитаем сколько байт понадобится для будущей строки.
	strSize := 0
	for _, s := range strSlice {
		strSize += len(s)
	}

	// заполнение строки.
	builder.Grow(strSize)
	for _, v := range strSlice {
		builder.WriteString(v)
	}

	return builder.String(), nil
}
