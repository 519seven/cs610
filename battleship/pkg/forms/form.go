package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Create a custom Form struct, which anonymously embeds a url.Values object
// (to hold the form data) and an Errors field to hold any validation errors
type Form struct {
	url.Values
	Errors errors 
}

// Define a New function to initialize a custom Form struct
//  Takes the Form data as the parameter
func New(data url.Values) *Form {
	return &Form{ data,
	errors(map[string][]string{}), }
}

// Matches Pattern - for verifying a pattern match
func (f *Form) MatchesPattern(field string, pattern *regexp.Regexp) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if !pattern.MatchString(value) {
		f.Errors.Add(field, "This field is invalid")
	}
}

// MaxLength - for checking maximum number of characters
func (f *Form) MaxLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) > d {
		f.Errors.Add(field, fmt.Sprintf("This field is too long (maximum is %d characters)", d))
	}
}

// MinLength - for checking presence of a minimum number of characters
func (f *Form) MinLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) < d {
		f.Errors.Add(field, fmt.Sprintf("This field is too short (minimum is %d characters)", d))
	}
}

// PermittedValues - matches one of a set of specific permitted values
func (f *Form) PermittedValues(field string, opts ...string) {
	value := f.Get(field)
	if value == "" {
		return
	}
	for _, opt := range opts {
		if value == opt {
			return
		}
	}
	f.Errors.Add(field, "This field is invalid")
}

// Required - check for the presence of an item in a field list
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}
	
// RequiredNumberOfItems - require X number of coordinates for the specific ship
func (f *Form) RequiredNumberOfItems(shipType string, requiredNumber int, countedNumber int) {
	if requiredNumber != countedNumber {
		var msg string
		if countedNumber < requiredNumber {
			msg = fmt.Sprintf("The number of coordinates for %s is too low.  We need %d, got %d", shipType, requiredNumber, countedNumber)
		} else if countedNumber > requiredNumber {
			msg = fmt.Sprintf("The number of coordinates for %s is too high.  We need %d, got %d", shipType, requiredNumber, countedNumber)
		}
		f.Errors.Add(shipType, msg)
	}
}
	
// Implement a Valid method which returns true if there are no errors.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}