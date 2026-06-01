package exitcode

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrom(t *testing.T) {
	assert.Equal(t, OK, From(nil))
	assert.Equal(t, General, From(errors.New("boom")))
	assert.Equal(t, Validation, From(New(Validation, "bad")))
	assert.Equal(t, Auth, From(&Error{Code: Auth, Message: "no key"}))
}

func TestErrorMessageAndCode(t *testing.T) {
	e := New(NotFound, "missing %s", "thing")
	assert.Equal(t, "missing thing", e.Error())
	assert.Equal(t, NotFound, e.ExitCode())
}
