package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	ErrIn     = errors.New("проверка на in")
	ErrMax    = errors.New("проверка на max")
	ErrMin    = errors.New("проверка на min")
	ErrRegexp = errors.New("проверка на regexp")
	ErrLen    = errors.New("проверка на len")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

// констрейнт.
type ValueTypes interface {
	~int | ~string
}

type Value[T ValueTypes] struct {
	value T
	rules []func(T) error
}

// реализуем интерфейс errors.
func (validErros ValidationErrors) Error() string {
	if len(validErros) == 0 {
		return ""
	}

	if len(validErros) == 1 {
		return fmt.Sprintf("ошибка валидации: 1\n %s - %s", validErros[0].Field, validErros[0].Err.Error())
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "ошибок валидации: %d\n", len(validErros))

	for i, err := range validErros {
		fmt.Fprintf(&sb, "  %d. %s: %s\n", i+1, err.Field, err.Err.Error())
	}

	return sb.String()
}

// дженерик конструктор создания экземпляра структуры: значение, слайс функций валидаторов с замыканием.
func NewValue[T ValueTypes](value T, rules []func(T) error) *Value[T] {
	return &Value[T]{
		value: value,
		rules: rules,
	}
}

// Запуск функций валидаторов значения, полученных из тега.
func (v *Value[T]) Valid() []error {
	var errors []error

	for _, valid := range v.rules {
		err := valid(v.value)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// Функция валидация с замыканием.
func minRule(ruleValue string) (func(int) error, error) {
	minVal, err := strconv.Atoi(ruleValue)
	if err != nil {
		return nil, fmt.Errorf("minRule конвертация в число значения %s:%w", ruleValue, err)
	}

	return func(val int) error {
		if val < minVal {
			return fmt.Errorf("%w: ожидали значение больше %d, получили значение %d", ErrMin, minVal, val)
		}
		return nil
	}, nil
}

// Функция валидация с замыканием.
func maxRule(ruleValue string) (func(int) error, error) {
	maxVal, err := strconv.Atoi(ruleValue)
	if err != nil {
		return nil, fmt.Errorf("maxRule конвертация в число значения %s:%w", ruleValue, err)
	}

	return func(val int) error {
		if val > maxVal {
			return fmt.Errorf("%w: ожидали значение меньше %d, получили значение %d", ErrMax, maxVal, val)
		}
		return nil
	}, nil
}

// Функция валидация с замыканием.
func lenRule(ruleValue string) (func(string) error, error) {
	length, err := strconv.Atoi(ruleValue)
	if err != nil {
		return nil, fmt.Errorf("lenRule конвертация в число значения %s:%w", ruleValue, err)
	}

	return func(val string) error {
		if utf8.RuneCountInString(val) != length {
			return fmt.Errorf("%w: ожидаемое количество символов %d, получили %d", ErrLen, length, utf8.RuneCountInString(val))
		}
		return nil
	}, nil
}

// Функция валидация с замыканием.
func regexpRule(ruleValue string) (func(string) error, error) {
	re, err := regexp.Compile(ruleValue)
	if err != nil {
		return nil, fmt.Errorf("regexpRule компиляция выражения %s:%w", ruleValue, err)
	}

	return func(val string) error {
		if ok := re.MatchString(val); !ok {
			return fmt.Errorf("%w: значение %s не удовлетворяет шаблону \"%s\"", ErrRegexp, val, ruleValue)
		}
		return nil
	}, nil
}

// Джененрик Функция  валидация с замыканием.
func inRule[T ValueTypes](ruleValue string) func(T) error {
	mapIn := make(map[string]struct{}, (strings.Count(ruleValue, ",") + 1))

	inValues := strings.Split(ruleValue, ",")
	for _, inValue := range inValues {
		mapIn[inValue] = struct{}{}
	}

	return func(val T) error {
		var valStr string

		switch any(val).(type) {
		case int:
			valStr = strconv.Itoa(any(val).(int))
		case string:
			valStr = any(val).(string)
		}

		if _, ok := mapIn[valStr]; !ok {
			return fmt.Errorf("%w: значение %v не входит в множество %s", ErrIn, valStr, ruleValue)
		}
		return nil
	}
}

// Парсинг тега.
func Parse(tag string, ruleSep string, inRuleSep string) map[string]string {
	ruleMap := make(map[string]string, (strings.Count(tag, ruleSep) + 1))

	for _, rule := range strings.Split(tag, ruleSep) {
		ruleName, ruleValue, _ := strings.Cut(rule, inRuleSep)
		ruleMap[ruleName] = ruleValue
	}

	return ruleMap
}

// Список ф-ций валидаторов согласно тегу для int.
func GetIntRules(ruleValues map[string]string) ([]func(int) error, error) {
	var intRules []func(int) error
	var err error

	for ruleName, ruleValue := range ruleValues {
		var intRule func(int) error

		switch ruleName {
		case "min":
			intRule, err = minRule(ruleValue)
		case "max":
			intRule, err = maxRule(ruleValue)
		case "in":
			intRule = inRule[int](ruleValue)
		}
		if err != nil {
			return nil, fmt.Errorf("добавление int валидатора %s:%w", ruleName, err)
		}
		if intRule != nil {
			intRules = append(intRules, intRule)
		}
	}

	return intRules, nil
}

// Список ф-ций валидаторов согласно тегу для string.
func GetStrRules(ruleValues map[string]string) ([]func(string) error, error) {
	var strRules []func(string) error
	var err error

	for ruleName, ruleValue := range ruleValues {
		var strRule func(string) error

		switch ruleName {
		case "len":
			strRule, err = lenRule(ruleValue)
		case "regexp":
			strRule, err = regexpRule(ruleValue)
		case "in":
			strRule = inRule[string](ruleValue)
		}
		if err != nil {
			return nil, fmt.Errorf("добавление string валидатора %s:%w", ruleName, err)
		}
		if strRule != nil {
			strRules = append(strRules, strRule)
		}
	}

	return strRules, nil
}

// Накопление ошибок.
func (validErros *ValidationErrors) addErrors(fieldName string, fieldValidErrs []error) {
	for _, err := range fieldValidErrs {
		*validErros = append(*validErros, ValidationError{
			Field: fieldName,
			Err:   err,
		})
	}
}

func ValidateIntField(
	fieldName string,
	fieldValue reflect.Value,
	ruleMap map[string]string,
	validErrors *ValidationErrors,
) error {
	validators, err := GetIntRules(ruleMap)
	if err != nil {
		return fmt.Errorf("получение списка валидаторов GetIntRules:%w", err)
	}

	if fieldValue.Kind() == reflect.Int {
		intValue := NewValue(int(fieldValue.Int()), validators)
		fieldValidErrs := intValue.Valid()

		validErrors.addErrors(fieldName, fieldValidErrs)
	} else {
		for i := 0; i < fieldValue.Len(); i++ {
			intValue := NewValue(int(fieldValue.Index(i).Int()), validators)
			fieldValidErrs := intValue.Valid()

			validErrors.addErrors(fieldName, fieldValidErrs)
		}
	}
	return nil
}

func ValidateStrField(
	fieldName string,
	fieldValue reflect.Value,
	ruleMap map[string]string,
	validErrors *ValidationErrors,
) error {
	validators, err := GetStrRules(ruleMap)
	if err != nil {
		return fmt.Errorf("получение списка валидаторов GetStrRules:%w", err)
	}

	if fieldValue.Kind() == reflect.String {
		strValue := NewValue(fieldValue.String(), validators)
		fieldValidErrs := strValue.Valid()

		validErrors.addErrors(fieldName, fieldValidErrs)
	} else {
		for i := 0; i < fieldValue.Len(); i++ {
			strValue := NewValue(fieldValue.Index(i).String(), validators)
			fieldValidErrs := strValue.Valid()

			validErrors.addErrors(fieldName, fieldValidErrs)
		}
	}
	return nil
}

// Валидация.
func Validate(v interface{}) error {
	// Переменная для накапливания ошибок.
	validErrors := &ValidationErrors{}

	// Указатель на данные о значении.
	rvPtr := reflect.ValueOf(v)

	if rvPtr.Kind() != reflect.Struct {
		return fmt.Errorf("валидация типа: ожидалась структура, передали %T", v)
	}

	// Указатель на данные о типе.
	rtPtr := rvPtr.Type()

	// Цикл по полям структуры.
	for i := 0; i < rtPtr.NumField(); i++ {
		field := rtPtr.Field(i)

		// Извлечение тега.
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		// Указатель на структуру со значением поля.
		fieldValue := rvPtr.Field(i)

		// Парсим тег.
		ruleMap := Parse(tag, "|", ":")

		// Валидация.
		if fieldValue.Kind() == reflect.Int ||
			fieldValue.Kind() == reflect.Slice && fieldValue.Type().Elem().Kind() == reflect.Int {
			err := ValidateIntField(field.Name, fieldValue, ruleMap, validErrors)
			if err != nil {
				return fmt.Errorf("валидация int значений, поле %s:%w", field.Name, err)
			}
		}

		if fieldValue.Kind() == reflect.String ||
			fieldValue.Kind() == reflect.Slice && fieldValue.Type().Elem().Kind() == reflect.String {
			err := ValidateStrField(field.Name, fieldValue, ruleMap, validErrors)
			if err != nil {
				return fmt.Errorf("валидация string значений, поле %s:%w", field.Name, err)
			}
		}
	}

	if len(*validErrors) == 0 {
		return nil
	}
	return *validErrors
}
