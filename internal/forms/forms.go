package forms

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"net/http"
	"net/url"
	"strings"
)

type Form struct {
	url.Values
	Errors errors
}

func (f *Form) IsValid() bool {
	return len(f.Errors) == 0
}

func NewForm(data url.Values) *Form {
	return &Form{data, make(errors)} // Initialize Errors map
}

func (f *Form) RequirementChecking(fields ...string) bool { // "...string" ka matla
	// b hai bhot sari strings ek sath pass kar skta hun
	for _, field := range fields {
		x := f.Get(field)
		if strings.TrimSpace(x) == "" {
			f.Errors.AppendError(field, "this field is required")
			return false
		}
	}
	return true
}

func (f *Form) HasError(field string, r *http.Request) bool {
	x := r.Form.Get(field)
	if x == "" {
		f.Errors.AppendError(field, "this field is required")
		return false
	}
	return true
}

func (f *Form) MinLength(field string, length int, r *http.Request) bool {
	x := r.Form.Get(field)
	if len(x) < length {
		f.Errors.AppendError(field, fmt.Sprintf("content of field should be larger than or equal to %d", length))
	}
	return true
}

func (f *Form) IsValidEmail(email string) bool {
	if !govalidator.IsEmail(f.Get(email)) {
		f.Errors.AppendError(email, "enter valid email address")
		return false
	}
	return true
}
