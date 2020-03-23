package main

import (
	"fmt"
	"github.com/go-ozzo/ozzo-validation/v4/is"
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

type Customer struct {
	Name    string
	Gender  string
	Email   string
	Address *Address
}

func (c *Customer) Validate() error {
	return validation.ValidateStruct(c,
		// Name cannot be empty, and the length must be between 5 and 20.
		validation.Field(&c.Name, validation.Required, validation.Length(5, 20)),
		// Gender is optional, and should be either "Female" or "Male".
		validation.Field(&c.Gender, validation.In("Female", "Male")),
		// Email cannot be empty and should be in a valid email format.
		validation.Field(&c.Email, validation.Required, is.Email),
		// Validate Address using its own validation rules
		validation.Field(&c.Address),
	)
}

func main() {
	c := &Customer{
		Name:  "Qiang Xue",
		Email: "qinhan_shu@163.com",
		Address: &Address{
			Street: "12",
			City:   "Unknown",
			State:  "Virginia",
			Zip:    "12345",
		},
	}

	err := c.Validate()
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
