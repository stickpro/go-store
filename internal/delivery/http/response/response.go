package response

import (
	"github.com/gofiber/fiber/v3"
)

type Result[T any] struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data"`
} // @name JSONResponse

func OkByMessage(message string) *Result[any] {
	return &Result[any]{
		Code:    fiber.StatusOK,
		Message: message,
	}
}

func OkByData[T any](data T) *Result[T] {
	return &Result[T]{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    data,
	}
}
