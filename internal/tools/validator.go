package tools

import (
	"errors"
	"github.com/stickpro/go-store/internal/tools/apierror"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var defaultStructValidator *StructValidator

type StructValidator struct {
	validate *validator.Validate
	trans    ut.Translator
}
type Validatable interface {
	Validate() error
}

func (s *StructValidator) Engine() any {
	return s.validate
}

func (s *StructValidator) Validate(out any) error {
	err := s.validate.Struct(out)
	if err == nil {
		return nil
	}
	var validateErrors validator.ValidationErrors
	if !errors.As(err, &validateErrors) || len(validateErrors) == 0 {
		return createAPIError("Struct parameter error", "", fiber.StatusBadRequest)
	}

	apiErrors := make([]apierror.Error, 0, len(validateErrors))
	for _, validateErr := range validateErrors {
		apiErrors = append(apiErrors, apierror.Error{
			Message: validateErr.Translate(s.trans),
			Field:   validateErr.Namespace(),
		})
	}

	apiErr := apierror.New(apiErrors...)
	_ = apiErr.SetHttpCode(fiber.StatusUnprocessableEntity)
	res, _ := json.Marshal(apiErr)
	return fiber.NewError(fiber.StatusUnprocessableEntity, string(res))
}

func init() {
	defaultStructValidator = newStruckValidator()
}

func createAPIError(message, field string, code int) error {
	apiErr := apierror.New(apierror.Error{
		Message: message,
		Field:   field,
	})
	_ = apiErr.SetHttpCode(code)
	res, _ := json.Marshal(apiErr)
	return fiber.NewError(code, string(res))
}

func newStruckValidator() *StructValidator {
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New()

	_ = enTranslations.RegisterDefaultTranslations(validate, trans)
	return &StructValidator{
		validate: validate,
		trans:    trans,
	}
}

func DefaultStructValidator() *StructValidator {
	return defaultStructValidator
}

func ValidateUUID(id string) (uuid.UUID, error) {
	if len(id) != 36 {
		return uuid.Nil, apierror.New().AddError(errors.New("invalid UUID length")).SetHttpCode(fiber.StatusBadRequest)
	}
	uuidParsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return uuidParsed, nil
}
