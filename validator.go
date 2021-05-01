package valkyrie

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type ErrorValidator struct {
	Type    string `json:"error_type,omitempty"`
	Tag     string `json:"error_tag,omitempty"`
	Field   string `json:"error_field,omitempty"`
	Value   string `json:"error_value,omitempty"`
	Message string `json:"error_message,omitempty"`
}

func Validate(s interface{}) (errors []ErrorValidator) {
	validate := validator.New()
	_ = validate.RegisterValidation("date", DateValidation)
	_ = validate.RegisterValidation("datetime", DatetimeValidation)
	_ = validate.RegisterValidation("daterange", DateRangeValidation)
	_ = validate.RegisterValidation("enum", ParseTags)
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	if err := validate.Struct(s); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ErrorValidator{
				Tag:     err.Tag(),
				Value:   fmt.Sprintf("%v", err.Value()),
				Field:   err.Field(),
				Type:    err.Type().String(),
				Message: fmt.Sprintf("Invalid Type %v for input %s", err.Value(), err.Field()),
			})
		}
		return errors
	}
	return nil
}

func DateValidation(fl validator.FieldLevel) bool {
	if _, err := time.Parse("2006-01-02", fl.Field().String()); err != nil {
		return false
	}
	return true
}

func DatetimeValidation(fl validator.FieldLevel) bool {
	if _, err := time.Parse(time.RFC3339, fl.Field().String()); err != nil {
		return false
	}
	return true
}

func DateRangeValidation(fl validator.FieldLevel) bool {

	var date = fl.Field().String()
	var minDate = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	var maxDate = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)

	datetime, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	if datetime.Before(minDate) || datetime.After(maxDate) {
		return false
	}
	return true
}

func ParseDate(dtStr string) time.Time {
	date, err := time.Parse("2006-01-02", dtStr)
	if err != nil {
		return time.Time{}
	}
	return date
}

func ParseDatetime(dtStr string) time.Time {
	dateTime, err := time.Parse(time.RFC3339, dtStr)
	if err != nil {
		return time.Time{}
	}
	return dateTime
}
func ParseTags(fl validator.FieldLevel) bool {
	splitParamsRegex := regexp.MustCompile(`'[^']*'|\S+`)
	params := fl.Param()
	s := splitParamsRegex.FindAllString(params, -1)
	values := make([]string, 0)
	for i := 0; i < len(s); i++ {
		p := strings.Replace(s[i], "'", "", -1)
		p = strings.Replace(p, "[", "", -1)
		p = strings.Replace(p, "]", "", -1)
		values = append(values, strings.Split(p, " ")...)
	}
	field := fl.Field()
	var v string
	switch field.Kind() {
	case reflect.String:
		v = strings.ToLower(field.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = strconv.FormatUint(field.Uint(), 10)
	default:
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}

	for i := 0; i < len(values); i++ {
		if values[i] == v {
			return true
		}
	}

	return false
}
