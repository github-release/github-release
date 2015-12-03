package lib

import "fmt"

type NotFound struct {
	Name  string
	Key   string
	Value string
}

func (e NotFound) Error() string {
	return fmt.Sprintf("%s not found with %s: %s", e.Name, e.Key, e.Value)
}
