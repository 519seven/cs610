package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	B "github.com/519seven/cs610/battleship/pkg/models/sqlite3"
	gpu "github.com/briandowns/GoPasswordUtilities"
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

// FieldsMatch - check to ensure fields match; ex: passwords
func (f *Form) FieldsMatch(f1, f2 string, shouldTheyMatch bool) {
	field1 := f.Get(f1)
	field2 := f.Get(f2)
	if (field1 == field2 && shouldTheyMatch == true) || (field1 != field2 && shouldTheyMatch == false) {
		return
	} else if (field1 == field2 && shouldTheyMatch == false) || (field1 != field2 && shouldTheyMatch == true) {
		f.Errors.Add("password", "Password and password confirmation do not match")
	}
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

// PasswordComplexity - make sure password meets basic complexity requirements
func (f *Form) PasswordComplexity() {
	pass := f.Get("password")
	gpuPass := gpu.New(pass)    
    gpuPass.ProcessPassword()
    if gpuPass.ComplexityRating() == "Horrible" || gpuPass.ComplexityRating() == "Weak" {
		f.Errors.Add("password", "Your password is too weak. Please use alpha-numeric, mixed case, and special characters.")
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

// ShipLength - how many pins should this ship get?
func ShipLength(shipName string) int {
	// the first coordinate is given so we want n-1 more
	switch shipName {
	case "carrier":
		return 5
	case "battleship":
		return 4
	case "cruiser":
		return 3
	case "submarine":
		return 3
	case "destroyer":
		return 2
	}
	return 0
}

// MaxLength - for checking maximum number of characters
func (f *Form) SpacesAbsent(field string) {
	value := f.Get(field)
	if strings.Contains(value, " ") || strings.Contains(value, "\t") || strings.Contains(value, "\n") || strings.Contains(value, "\f") || strings.Contains(value, "\r") {
		f.Errors.Add(field, "This field cannot contain whitespace")
	}
}


// ValidNumberOfItems - make sure that the ship in question has proper pin placement
func (f *Form) ValidNumberOfItems(coordinates []string, shipName string) {
	// once we have a boardID, update coordinates table with each ship's XY
	// First, we have to define a set of coordinates to a ship
	// If our coordinates don't meet our requirements,
	// return to the form with an error message
	// loop through the values, pick a row, find out what is adjacent
	// figure out which ship it is, remember the ship
	// if our ship definitions are violated, fail this routine
	searchDirection := "initialize"
	numberOfFathomsRemaining := ShipLength(shipName) - 1
	for _, rc := range coordinates {
		s := strings.Split(rc, ",")
		//fmt.Println(s)
		row, col := s[0], s[1]
		nxtR, _ := strconv.Atoi(row)
		nextRow := strconv.Itoa(nxtR+1)
		nextCol := string(B.GetNextChar(col, 10))
		if (searchDirection == "initialize" || searchDirection == "row") && numberOfFathomsRemaining > 0 && B.MatchFound(row+","+nextCol, strings.Join(coordinates, " ")) {
			// is the next column in the slice?
			//fmt.Println("match found in the next column: ", row+"|"+nextCol, "searching this row only")
			searchDirection = "row"
			numberOfFathomsRemaining -= 1
		}
		if (searchDirection == "initialize" || searchDirection == "column") && numberOfFathomsRemaining > 0 && B.MatchFound(nextRow+","+col, strings.Join(coordinates, " ")) {
			// is the next row in the slice?
			//fmt.Println("match found in the next row: ", nextRow+"|"+col, "searching this column only")
			searchDirection = "column"
			numberOfFathomsRemaining -= 1
		}
	}
	// after looping through all of the coordinates of a ship, we ought to be at 0 fathoms remaining
	// (a fathom is a unit of measurement based on one's outstretched arms)
	if numberOfFathomsRemaining != 0 {
		// we did not receive enough coordinates to satisfy the requirement for this ship
		//log.Info("numberOfFathomsRemaining is not 0")
		fmt.Println("numberOfFathomsRemaining is not 0!  Sending you back to the form with your data.", numberOfFathomsRemaining)
		fmt.Println("The ship that is in error is:", shipName)
		msg := fmt.Sprintf("Unable to calculate the correct number of coordinates (%d) necessary for a %s", ShipLength(shipName), shipName)
		f.Errors.Add(shipName, msg)
	}
}

// Implement a Valid method which returns true if there are no errors.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}