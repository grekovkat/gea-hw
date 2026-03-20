package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"testing"
)

type UserRole string

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

// ожидаемый результат.
type expectType int

const (
	expectNil            expectType = iota // ошибок нет.
	expectValidationErrs                   // ValidationErrors с хотя бы одной записью.
	expectProgramErr                       // программная ошибка (не ValidationErrors).
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		in     interface{}
		expect expectType
	}{
		{
			name: "валидный User",
			in: User{
				ID:     "123456789123456789123456789123456789",
				Name:   "Катя",
				Age:    37,
				Email:  "kate@mail.ru",
				Role:   "admin",
				Phones: []string{"79991234567", "79997654321"},
			},
			expect: expectNil,
		},
		{
			name:   "ID слишком короткий",
			in:     User{ID: "short-id", Age: 37, Email: "kate@mail.ru", Role: "admin"},
			expect: expectValidationErrs,
		},
		{
			name:   "Age меньше минимума",
			in:     User{ID: "123456789123456789123456789123456789", Age: 16, Email: "kate@mail.ru", Role: "admin"},
			expect: expectValidationErrs,
		},
		{
			name:   "Age больше максимума",
			in:     User{ID: "123456789123456789123456789123456789", Age: 99, Email: "kate@mail.ru", Role: "admin"},
			expect: expectValidationErrs,
		},
		{
			name:   "Email не соответствует regexp",
			in:     User{ID: "123456789123456789123456789123456789", Age: 37, Email: "not-an-email", Role: "admin"},
			expect: expectValidationErrs,
		},
		{
			name:   "Role не входит в множество",
			in:     User{ID: "123456789123456789123456789123456789", Age: 37, Email: "kate@mail.ru", Role: "superuser"},
			expect: expectValidationErrs,
		},
		{
			name: "телефон неверной длины",
			in: User{
				ID: "123456789123456789123456789123456789", Age: 37,
				Email: "kate@mail.ru", Role: "admin",
				Phones: []string{"123"},
			},
			expect: expectValidationErrs,
		},
		{
			name:   "валидный App",
			in:     App{Version: "1.0.0"},
			expect: expectNil,
		},
		{
			name:   "App неверная длина Version",
			in:     App{Version: "1.0"},
			expect: expectValidationErrs,
		},
		{
			name:   "Token без тегов validate",
			in:     Token{Header: []byte("h"), Payload: []byte("p"), Signature: []byte("s")},
			expect: expectNil,
		},
		{
			name:   "валидный Response 200",
			in:     Response{Code: 200, Body: "OK"},
			expect: expectNil,
		},
		{
			name:   "Response невалидный код",
			in:     Response{Code: 301, Body: "Moved"},
			expect: expectValidationErrs,
		},
		{
			name:   "валидный Response 404",
			in:     Response{Code: 404},
			expect: expectNil,
		},
		{
			name:   "валидный Response 500",
			in:     Response{Code: 500},
			expect: expectNil,
		},
		{
			name:   "несколько ошибок ID и Age",
			in:     User{ID: "bad", Age: 5, Email: "kate@mail.ru", Role: "admin"},
			expect: expectValidationErrs,
		},
		{
			name:   "не структура",
			in:     "not a struct",
			expect: expectProgramErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()
			assertValidateResult(t, Validate(tt.in), tt.expect)
		})
	}
}

// assertValidateResult проверяет результат Validate согласно ожидаемому типу.
func assertValidateResult(t *testing.T, err error, expect expectType) {
	t.Helper()

	switch expect {
	case expectNil:
		if err != nil {
			t.Errorf("ожидали nil, получили: %v", err)
		}

	case expectProgramErr:
		if err == nil {
			t.Errorf("ожидали ошибку, получили nil")
			return
		}
		var ve ValidationErrors
		if errors.As(err, &ve) {
			t.Errorf("ожидали программную ошибку, получили ValidationErrors")
		}

	case expectValidationErrs:
		assertValidationErrors(t, err)
	}
}

// assertValidationErrors проверяет, что err является непустым ValidationErrors
// с корректно заполненными полями.
func assertValidationErrors(t *testing.T, err error) {
	t.Helper()

	var ve ValidationErrors
	if !errors.As(err, &ve) {
		t.Errorf("ожидали ValidationErrors, получили: %T: %v", err, err)
		return
	}
	if len(ve) == 0 {
		t.Error("ValidationErrors пуст, но ожидали хотя бы одну ошибку")
		return
	}
	for _, e := range ve {
		if e.Field == "" {
			t.Error("поле Field не должно быть пустым")
		}
		if e.Err == nil {
			t.Error("поле Err не должно быть nil")
		}
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
