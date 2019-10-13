package forms

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"
)

// Create a custom Form struct, which anonymously embeds a url.Values object
// (to hold the form data) and an Errors field to hold any validation errors
type Form struct {
	url.Values
	Errors errors 
}

// Define a New function to initialize a custom Form struct.
// Notice that this takes the form data as the parameter?
func New(data url.Values) *Form {
	return &Form{ data,
	errors(map[string][]string{}), }
}

// Implement a Required method to check that specific fields in the form
//  data are present and not blank. If any fields fail this check, add the
//  appropriate message to the form errors.
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}
	
// Implement a MaxLength method to check that a specific field in the form
//  contains a maximum number of characters. If the check fails then add the 
//  appropriate message to the form errors.
func (f *Form) MaxLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) > d {
		f.Errors.Add(field, fmt.Sprintf("This field is too long (maximum is %d characters)", d))
	}
}

// Implement a PermittedValues method to check that a specific field in the form 
//  matches one of a set of specific permitted values. If the check fails
//  then add the appropriate message to the form errors.
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

// Implement a RequiredNumberOfItems method to check that there are X number of
//  coordinates for the specific ship
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