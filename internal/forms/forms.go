package forms

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Form holds URL values which essentially are Values maps a string key to a list of values.
// It is typically used for query parameters and form values. Unlike in the http.Header map,
// the keys in a Values map are case-sensitive and errors associated with form fields.
type Form struct {
	url.Values
	Errors errors
}

// New creates a new form based on query parameter, or form value data (i.e. url.Values) and
// empty error map.
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Required checks whether or not all specified fields have a value. For those fields that do NOT
// have any value, a new error masseg is added to the Form object errors map.
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "Ovo polje je obavezno!")
		}
	}
}

// MinLength checks if a specific field has at least 'length' (a function parameter) characters.
func (f *Form) MinLength(field string, length int) bool {
	x := f.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("Minimalna dužina ovog polja je %d", length))
		return false
	}
	return true
}

// MinValueInt64 checks if a specific field is greater or equal then specified value.
func (f *Form) MinValueInt64(field string, value int64) bool {
	x := f.Get(field)
	x_value, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		f.Errors.Add(field, "Could not convert form value to int64")
		return false
	}
	if x_value < value {
		f.Errors.Add(field, fmt.Sprintf("Minimalna vrednost ovog polja je %d", value))
		return false
	}
	return true
}

// MinValueFloat64 checks if a specific field is greater or equal then specified value.
func (f *Form) MinValueFloat64(field string, value float64) bool {
	x := f.Get(field)
	x_value, err := strconv.ParseFloat(x, 64)
	if err != nil {
		f.Errors.Add(field, "Could not convert form value to float64")
		return false
	}
	if x_value < value {
		f.Errors.Add(field, fmt.Sprintf("Minimalna vrednost ovog polja je %d", int(value)))
		return false
	}
	return true
}

// Has checks whether or not a specific form field has a value or not.
func (f *Form) Has(field string) bool {
	x := f.Get(field)
	return x != ""
}

// Valid checks if the length of errors map of the specific Form object is empty or not. If it is, form is valid, otherwise, form is invalid.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// IsEmail checks for a valid email address
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
	}
}
