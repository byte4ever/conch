package conch

import (
	"fmt"
)

type PanicError struct {
	err   error
	stack string
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("error:%v\nstack:\n%s", e.err, e.stack)
}
