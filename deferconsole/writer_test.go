package deferconsole

import (
	"testing"
)

func TestAll(t *testing.T) {
	SetMinLevel(0)
	ToError(1, "%s -> %s ", "some", "thing")
	Sync()
}
