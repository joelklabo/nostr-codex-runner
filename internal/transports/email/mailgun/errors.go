package mailgun

import "fmt"

// Err wraps configuration/validation errors for clarity.
type Err string

func (e Err) Error() string {
	return fmt.Sprintf("mailgun: %s", string(e))
}
