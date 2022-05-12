package forms

// errors is a map of strings that represent error messages associated
// with specific form fields.
type errors map[string][]string

// Add adds new error message to a specific form field.
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get retrieves error message from a specific form field.
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}
