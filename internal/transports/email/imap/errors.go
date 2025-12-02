package imap

import "fmt"

// Err wraps configuration errors.
type Err string

func (e Err) Error() string { return fmt.Sprintf("imap: %s", string(e)) }
