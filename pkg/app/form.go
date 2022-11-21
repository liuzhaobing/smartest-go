package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"smartest-go/pkg/logf"
	validation "smartest-go/pkg/validation"
)

// BindAndValid binds and validates data
func BindAndValid(c *gin.Context, form interface{}) error {
	var err error
	if c.Request.Method == http.MethodGet {
		err = c.ShouldBindQuery(form)
	} else if c.Request.Method == http.MethodPost {
		err = c.ShouldBindJSON(form)
	} else if c.Request.Method == http.MethodPut {
		err = c.ShouldBindJSON(form)
	} else if c.Request.Method == http.MethodDelete {
		err = c.ShouldBindQuery(form)
	}

	var errStr string
	if err != nil {
		switch err.(type) {
		case validator.ValidationErrors:
			errStr = validation.TranslateOneError(err.(validator.ValidationErrors))
		case *json.UnmarshalTypeError:
			unmarshalTypeError := err.(*json.UnmarshalTypeError)
			errStr = fmt.Errorf("%s type error, expect type is %s", unmarshalTypeError.Field, unmarshalTypeError.Type.String()).Error()
		default:
			errStr = err.Error()
		}
		logf.Error(errStr)
		return errors.New(errStr)
	}

	return nil
}
