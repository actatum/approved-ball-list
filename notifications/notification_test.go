package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDuplicateTargetError_Error(t *testing.T) {
	e := DuplicateTargetError{
		targetType:  TargetTypeDiscord,
		destination: "some_channel_id",
	}
	assert.Equal(t, "target already exists for discord: some_channel_id", e.Error())
}
