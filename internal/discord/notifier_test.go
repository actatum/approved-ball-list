// Package discord provides an implementation of the Notifier interface using discord as the notification medium.
package discord

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_batchSlice(t *testing.T) {
	t.Run("batch some integers", func(t *testing.T) {
		list := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		batchSize := 3
		expected := [][]int{
			{1, 2, 3},
			{4, 5, 6},
			{7, 8, 9},
			{10},
		}

		got := batchSlice(list, batchSize)
		assert.Equal(t, expected, got)
	})
}
