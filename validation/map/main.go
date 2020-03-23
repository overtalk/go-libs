package main

import (
	"fmt"
	"regexp"

	"github.com/go-ozzo/ozzo-validation/v4"
)

type Address struct {
	Street string
	City   string
	State  string
	Zip    string
}

func (a *Address) Validate() error {
	return validation.ValidateStruct(a,
		// Street cannot be empty, and the length must between 5 and 50
		validation.Field(&a.Street, validation.Required, validation.Length(5, 50)),
		// City cannot be empty, and the length must between 5 and 50
		validation.Field(&a.City, validation.Required, validation.Length(5, 50)),
		// State cannot be empty, and must be a string consisting of two letters in upper case
		validation.Field(&a.State, validation.Required, validation.Match(regexp.MustCompile("^[A-Z]{2}$"))),
		// State cannot be empty, and must be a string consisting of five digits
		validation.Field(&a.Zip, validation.Required, validation.Match(regexp.MustCompile("^[0-9]{5}$"))),
	)
}

func main() {
	a := []*Address{
		&Address{
			Street: "123",
			City:   "Unknown",
			State:  "Virginia",
			Zip:    "12345",
		}, &Address{
			Street: "asdfasdfasd",
			City:   "Unknown",
			State:  "Virginia",
			Zip:    "12345",
		}, &Address{
			Street: "sdfgsdfgsdfg",
			City:   "Unknown",
			State:  "Virginia",
			Zip:    "12345",
		}, &Address{
			Street: "sdfsdfgsdf",
			City:   "Unknown",
			State:  "Virginia",
			Zip:    "12345",
		},
	}

	err := validation.Validate(a)
	if err != nil {
		if e, ok := err.(validation.InternalError); ok {
			fmt.Println("internal error", e)
		} else {
			fmt.Println(err)
		}
	}

	// Output:
	// Street: the length must be between 5 and 50; State: must be in a valid format.
}
