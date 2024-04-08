package main

import "strings"

type Truncate struct {
	FieldName string
	NumChars  int
	Substr    string
}

func (t *Truncate) Truncate(value string) string {
	if t.NumChars > -1 {
		return t.truncAtNumChars(value)
	}

	if t.Substr != "" {
		return t.truncAtSubstr(value)
	}

	return value
}

func (t *Truncate) truncAtNumChars(value string) string {
	if t.NumChars == -1 {
		return value
	}

	if len(value) > t.NumChars {
		return value[:t.NumChars]
	}
	return value
}

func (t *Truncate) truncAtSubstr(value string) string {
	if t.Substr == "" {
		return value
	}

	// If flag is --trunc message="\n" (truncate at newline character) then it'll be read as "\\n" at this point. To match it against newline char \n in a text, we must correct it.
	if strings.Contains(t.Substr, "\\n") {
		t.Substr = "\n"
	}

	// If flag is --trunc message="\t" (truncate at tab character) then it'll be read as "\\t" at this point. To match it against a tab char \t in a text, we must correct it.
	if strings.Contains(t.Substr, "\\t") {
		t.Substr = "\t"
	}

	if indexOfSubstr := strings.Index(value, t.Substr); indexOfSubstr > -1 {
		return value[:indexOfSubstr]
	}
	return value
}
