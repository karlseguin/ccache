package ccache

import (
	"testing"

	"github.com/karlseguin/ccache/v3/assert"
)

func Test_Configuration_BucketsPowerOf2(t *testing.T) {
	for i := uint32(0); i < 31; i++ {
		c := Configure[int]().Buckets(i)
		if i == 1 || i == 2 || i == 4 || i == 8 || i == 16 {
			assert.Equal(t, c.buckets, int(i))
		} else {
			assert.Equal(t, c.buckets, 16)
		}
	}
}

func Test_Configuration_Buffers(t *testing.T) {
	assert.Equal(t, Configure[int]().DeleteBuffer(24).deleteBuffer, 24)
	assert.Equal(t, Configure[int]().PromoteBuffer(95).promoteBuffer, 95)
}
