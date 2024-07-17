package errorcmp

import (
	"errors"
	"fmt"
)

func Diff(want, got error) string {
	var s string

	switch {
	case want == nil && got == nil:
		s = ""
	case want != nil && got == nil:
		s = "expected error, but none received"
	case want == nil && got != nil:
		s = fmt.Sprintf("unexpected error: %v", got)
	case want != nil && got != nil:
		if !errors.As(want, &got) {
			s = fmt.Sprintf("want error: %T, but got %T", want, got)
		}
		s = ""
	}

	return s
}
