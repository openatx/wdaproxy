package muuid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOsxUUID(t *testing.T) {
	u1 := UUIDFromOS("any")
	if u1 == "" {
		t.Fatal("got empty uuid string")
	}
	u2 := UUIDFromOS("any")
	assert.Equal(t, u1, u2, "The two uuid should be equal")
	RemoveTempUidFile()
}
