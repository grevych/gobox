package orerr_test

import (
	"errors"
	"fmt"

	"github.com/grevych/gobox/pkg/orerr"
)

const ErrUsernameTaken orerr.SentinelError = "username already taken"

func createUser(_ string) error {
	return ErrUsernameTaken
}

func ExampleSentinelError() {
	if err := createUser("joe"); err != nil {
		if errors.Is(err, ErrUsernameTaken) {
			fmt.Println("User 'joe' already exists")
			return
		}

		panic(err)
	}

	// Output: User 'joe' already exists
}
