package sigma

import (
	"fmt"
)

type ErrInvalidSigma struct {
	Desc string
	Raw  string
}

func (e ErrInvalidSigma) Error() string {
	return fmt.Sprintf("Invalid sigma rule: %+v", e)
}

type UsableRule struct {
}

type RawRule struct {
	// Our custom fields
	// Unique identifier that will be attached to positive match
	ID int `yaml:"id" json:"id"`
	// Detection logic type
	// Is it simple string match or more complex correlation

	// https://github.com/Neo23x0/sigma/wiki/Specification
	Title       string `yaml:"title" json:"title"`
	Status      string `yaml:"status" json:"status"`
	Description string `yaml:"description" json:"description"`
	Author      string `yaml:"author" json:"author"`
	// A list of URL-s to external sources
	References []string `yaml:"references" json:"references"`
	Logsource  struct {
		Product    string `yaml:"product" json:"product"`
		Category   string `yaml:"category" json:"category"`
		Service    string `yaml:"service" json:"service"`
		Definition string `yaml:"definition" json:"definition"`
	} `yaml:"logsource" json:"logsource"`

	Detection Detection `yaml:"detection" json:"detection"`

	Fields         interface{} `yaml:"fields" json:"fields"`
	Falsepositives interface{} `yaml:"falsepositives" json:"falsepositives"`
	Level          interface{} `yaml:"level" json:"level"`
	Tags           []string    `yaml:"tags" json:"tags"`
}

type Detection map[string]interface{}

/*
func Parse(d RawRule) error {
	condition, ok := d.Detection["condition"].(string)
	if !ok {
		return fmt.Errorf("missing condition, or unable to assert string type")
	}
	tokens := []*conditionToken{}

	fmt.Println("---------------------------")
	fmt.Println(d.Detection["condition"].(string))

	for len(condition) > 0 {
	inner:
		for _, def := range tokenDefs {
			if loc := def.FindStringIndex(condition); loc != nil {
				tokens = append(tokens, &conditionToken{
					token:    def.token,
					position: loc,
				})
				condition = condition[loc[1]:]
				break inner
			}
		}
	}

	for _, t := range tokens {
		fmt.Printf("%+v %s", t, t.token.String())
	}
	fmt.Println()

	return nil
}
*/
