package balls

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
		diff := cmp.Diff(got, expected)
		if diff != "" {
			t.Fatalf("(-got, +want):\n%s", diff)
		}
	})
}
