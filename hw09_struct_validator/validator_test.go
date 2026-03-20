package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		in          interface{}
		expectedErr error
	}{
		// 0: валидный User
		{
			name: "Valid User",
			in: User{
				ID:     "123456789123456789123456789123456789", // 36 символов
				Name:   "Катя",
				Age:    37,
				Email:  "kate@mal.ru",
				Role:   "admin",
				Phones: []string{"89183107252", "89183107251"},
			},
			expectedErr: nil,
		},
		// 1: ID слишком короткий
		{
			name: "Short ID",
			in: User{
				ID:    "123456789",
				Age:   37,
				Email: "kate@mal.ru",
				Role:  "admin",
			},
			expectedErr: ValidationErrors{},
		},
		// 2: Age меньше минимума
		{
			name: "Age < Min",
			in: User{
				ID:    "123456789123456789123456789123456789",
				Age:   10,
				Email: "kate@mal.ru",
				Role:  "admin",
			},
			expectedErr: ValidationErrors{},
		},
		// 3: Age больше максимума
		{
			name: "Age > Max",
			in: User{
				ID:    "123456789123456789123456789123456789",
				Age:   99,
				Email: "kate@mal.ru",
				Role:  "admin",
			},
			expectedErr: ValidationErrors{},
		},
		// 4: Email не соответствует regexp
		{
			name: "Email not compile regexp",
			in: User{
				ID:    "123456789123456789123456789123456789",
				Age:   37,
				Email: "not-an-email",
				Role:  "admin",
			},
			expectedErr: ValidationErrors{},
		},
		// 5: Role не входит в множество
		{
			name: "Role not in list",
			in: User{
				ID:    "123456789123456789123456789123456789",
				Age:   37,
				Email: "kate@mal.ru",
				Role:  "superuser",
			},
			expectedErr: ValidationErrors{},
		},
		// 6: один из телефонов неверной длины
		{
			name: "one phone incorrect",
			in: User{
				ID:     "123456789123456789123456789123456789",
				Age:    37,
				Email:  "kate@mal.ru",
				Role:   "admin",
				Phones: []string{"123, 89183107251"},
			},
			expectedErr: ValidationErrors{},
		},
		// 7: валидный App
		{
			name:        "valid App",
			in:          App{Version: "1.0.0"},
			expectedErr: nil,
		},
		// 8: App с неверной длиной Version
		{
			name:        "App Version len incorrect",
			in:          App{Version: "1.0"},
			expectedErr: ValidationErrors{},
		},
		// 9: Token — нет тегов validate, всегда nil
		{
			name: "no tag validation",
			in: Token{
				Header:    []byte("header"),
				Payload:   []byte("payload"),
				Signature: []byte("sig"),
			},
			expectedErr: nil,
		},
		// 10: валидный Response
		{
			name:        "valid Response",
			in:          Response{Code: 200, Body: "OK"},
			expectedErr: nil,
		},
		// 11: Response с невалидным кодом
		{
			name:        "invalid code Response",
			in:          Response{Code: 301, Body: "Moved"},
			expectedErr: ValidationErrors{},
		},
		// 12: Response с валидным кодом 404
		{
			name:        "valid code Response 400",
			in:          Response{Code: 404},
			expectedErr: nil,
		},
		// 13: Response с валидным кодом 500
		{
			name:        "valid code Response 500",
			in:          Response{Code: 500},
			expectedErr: nil,
		},
		// 14: несколько ошибок сразу (ID короткий + Age вне диапазона)
		{
			name: "multi errors",
			in: User{
				ID:    "123",
				Age:   5,
				Email: "katemal.ru",
				Role:  "admincheg",
			},
			expectedErr: ValidationErrors{},
		},
		// 15: передаём не структуру — ожидаем программную ошибку (не ValidationErrors)
		{
			name:        "not a struct",
			in:          "not a struct",
			expectedErr: fmt.Errorf("not a struct"),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d %s", i, tt.name), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)

			// Случай 15: ожидаем программную ошибку, не ValidationErrors
			if i == 15 {
				if err == nil {
					t.Errorf("ожидали ошибку, получили nil")
				}
				var ve ValidationErrors
				if errors.As(err, &ve) {
					t.Errorf("ожидали программную ошибку, получили ValidationErrors")
				}
				return
			}

			// Ожидаем nil — ошибок быть не должно
			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("ожидали nil, получили: %v", err)
				}
				return
			}

			// Ожидаем ValidationErrors
			var ve ValidationErrors
			if !errors.As(err, &ve) {
				t.Errorf("ожидали ValidationErrors, получили: %T: %v", err, err)
				return
			}

			if len(ve) == 0 {
				t.Errorf("ValidationErrors пуст, но ожидали хотя бы одну ошибку")
			}

			// Проверяем что ошибки содержат sentinel errors из пакета
			for _, validErr := range ve {
				if validErr.Field == "" {
					t.Errorf("поле Field не должно быть пустым")
				}
				if validErr.Err == nil {
					t.Errorf("поле Err не должно быть nil")
				}
			}
		})
	}
}

// Проверка sentinel errors через errors.Is.
func TestValidateSentinelErrors(t *testing.T) {
	t.Run("ErrMin срабатывает", func(t *testing.T) {
		err := Validate(User{
			ID:    "123456789123456789123456789123456789",
			Age:   10, // < 18
			Email: "kate@mail.ru",
			Role:  "admin",
		})
		assertContainsSentinel(t, err, ErrMin)
	})

	t.Run("ErrMax срабатывает", func(t *testing.T) {
		err := Validate(User{
			ID:    "123456789123456789123456789123456789",
			Age:   100, // > 50
			Email: "kate@mail.ru",
			Role:  "admin",
		})
		assertContainsSentinel(t, err, ErrMax)
	})

	t.Run("ErrLen срабатывает", func(t *testing.T) {
		err := Validate(App{Version: "1.0"}) // не 5 символов
		assertContainsSentinel(t, err, ErrLen)
	})

	t.Run("ErrRegexp срабатывает", func(t *testing.T) {
		err := Validate(User{
			ID:    "123456789123456789123456789123456789",
			Age:   37,
			Email: "bad-email",
			Role:  "admin",
		})
		assertContainsSentinel(t, err, ErrRegexp)
	})

	t.Run("ErrIn срабатывает для string", func(t *testing.T) {
		err := Validate(User{
			ID:    "123456789123456789123456789123456789",
			Age:   37,
			Email: "kate@mail.ru",
			Role:  "hacker",
		})
		assertContainsSentinel(t, err, ErrIn)
	})

	t.Run("ErrIn срабатывает для int", func(t *testing.T) {
		err := Validate(Response{Code: 302})
		assertContainsSentinel(t, err, ErrIn)
	})
}

// Проверка метода Error() у ValidationErrors.
func TestValidationErrorsError(t *testing.T) {
	t.Run("пустой слайс", func(t *testing.T) {
		ve := ValidationErrors{}
		if ve.Error() != "" {
			t.Errorf("пустой ValidationErrors должен возвращать пустую строку, получили: %q", ve.Error())
		}
	})

	t.Run("одна ошибка", func(t *testing.T) {
		ve := ValidationErrors{{Field: "Age", Err: ErrMin}}
		s := ve.Error()
		if s == "" {
			t.Error("ожидали непустую строку")
		}
	})

	t.Run("несколько ошибок", func(t *testing.T) {
		ve := ValidationErrors{
			{Field: "Age", Err: ErrMin},
			{Field: "Email", Err: ErrRegexp},
		}
		s := ve.Error()
		if s == "" {
			t.Error("ожидали непустую строку")
		}
	})
}

// assertContainsSentinel проверяет, что среди ValidationErrors есть хотя бы одна с нужным sentinel.
func assertContainsSentinel(t *testing.T, err error, target error) {
	t.Helper()

	var ve ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("ожидали ValidationErrors, получили: %T: %v", err, err)
	}

	for _, e := range ve {
		if errors.Is(e.Err, target) {
			return
		}
	}

	t.Errorf("среди ValidationErrors не нашли sentinel %v; ошибки: %v", target, ve)
}
