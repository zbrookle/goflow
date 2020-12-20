package stringutils

import "fmt"

// StringSet is a set of strings
type StringSet map[string]struct{}

// NewStringSet returns a new StringSet
func NewStringSet(strings []string) StringSet {
	set := make(StringSet)
	for _, s := range strings {
		set.Add(s)
	}
	return set
}

// Add adds a string to the set
func (s StringSet) Add(str string) {
	s[str] = struct{}{}
}

// Remove removes a string from the set
func (s StringSet) Remove(str string) {
	delete(s, str)
}

// Contains returns true if the string is in the set
func (s StringSet) Contains(str string) bool {
	_, ok := s[str]
	return ok
}

// GetOne returns a string from the set
func (s StringSet) GetOne() (retString string, err error) {
	retString = ""
	err = fmt.Errorf("nothing to get")
	for str := range s {
		retString, err = str, nil
	}
	return
}
