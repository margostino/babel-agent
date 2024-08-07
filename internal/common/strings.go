package common

import (
	"fmt"
	"strings"
)

type String struct {
	value string
}

func NewString(value string) *String {
	return &String{
		value: value,
	}
}

func (s *String) ToLower() *String {
	s.value = strings.ToLower(s.value)
	return s
}

func (s *String) HasPrefix(prefix string) bool {
	return strings.HasPrefix(s.value, prefix)
}

func (s *String) ReplaceAll(old string, new string) *String {
	s.value = strings.ReplaceAll(s.value, old, new)
	return s
}

func (s *String) TrimPrefix(prefix string) *String {
	s.value = strings.TrimPrefix(s.value, prefix)
	return s
}

func (s *String) TrimSpace() *String {
	s.value = strings.TrimSpace(s.value)
	return s
}

func (s *String) GetPrefixBy(separator string) *string {

	values := s.Split(separator).Values()

	if len(values) > 0 {
		return &values[0]
	}

	return nil
}

func (s *String) TrimIndex(index int) *String {
	var prefix string

	for i := 0; i < index; i++ {
		if prefix == "" {
			prefix += strings.Split(s.value, " ")[i]
		} else {
			prefix += fmt.Sprintf(" %s", strings.Split(s.value, " ")[i])
		}
	}

	s.value = strings.TrimPrefix(s.value, prefix)
	s.value = strings.TrimPrefix(s.value, " ")
	return s
}

func (s *String) Trim(prefix string) *String {
	s.value = strings.Trim(s.value, prefix)
	return s
}

func (s *String) Split(separator string) *StringSlice {
	values := &StringSlice{
		values: strings.Split(s.value, separator),
	}

	return values
}

func (s *String) Value() string {
	return s.value
}
