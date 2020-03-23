# 数据验证

- [ozzo-validation](https://github.com/go-ozzo/ozzo-validation)

## 简介
- 主要用于检测数据是否符合预设的数据规范

## 使用
### 验证简单值
```go
package main

import (
	"fmt"

	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func main() {
	data := "example"
	err := validation.Validate(data,
		validation.Required,       // not empty
		validation.Length(5, 100), // length between 5 and 100
		is.URL,                    // is a valid URL
	)
	fmt.Println(err)
	// Output:
	// must be a valid URL
}
```

### 验证结构体
- 请注意，在调用validation.ValidateStruct以验证结构时，应将指向结构的指针（而不是结构本身）传递给方法。
- 同样，在调用validation.Field以指定结构字段的规则时，应使用指向结构字段的指针。
- 内部错误的存在
```go
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
	a := Address{
		Street: "123",
		City:   "Unknown",
		State:  "Virginia",
		Zip:    "12345",
	}

	err := a.Validate()
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
```

### 结构体嵌套
- 结构体嵌套的时候，如果某个结构体中的某个字段实现了 `validation.Validatable` 接口，则回递归验证
```go
package main

import (
	"fmt"
	"regexp"
	"github.com/go-ozzo/ozzo-validation/v4/is"
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
```


### Maps/Slices/Arrays 可验证
- 验证其元素类型实现validation.Validatable接口的可迭代（映射，切片或数组）时，该validation.Validate方法将调用Validate每个非null元素的方法。
- 将返回元素的验证错误，因为validation.Errors它将无效元素的键映射到其相应的验证错误。例如，
```go
addresses := []Address{
    Address{State: "MD", Zip: "12345"},
    Address{Street: "123 Main St", City: "Vienna", State: "VA", Zip: "12345"},
    Address{City: "Unknown", State: "NC", Zip: "123"},
}
err := validation.Validate(addresses)
fmt.Println(err)
// Output:
// 0: (City: cannot be blank; Street: cannot be blank.); 2: (Street: cannot be blank; Zip: must be in a valid format.).
```
