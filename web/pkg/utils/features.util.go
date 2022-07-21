package utils

import (
	"math/rand"

	"github.com/go-playground/validator/v10"
)

func RandInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func RandomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(RandInt(65, 90))
	}
	return string(bytes)
}

func ValidateRpcRequest[T any](structForm T) error {
	validate := validator.New()

	err := validate.Struct(structForm)

	if err != nil {
		return err
	}

	return nil
}

func ValidateRpcMethod(method string) bool {
	methods := []string{"boo.create", "boo.update", "foo.create", "foo.update"}

	for _, m := range methods {
		if m == method {
			return true
		}
	}

	return false
}
