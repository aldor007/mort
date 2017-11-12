package response

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestResponse_Copy(t *testing.T) {
	buf := make([]byte, 1000)
	res := NewBuf(200, buf)
	resCpy, err := res.Copy()
	assert.Nil(t, err, "Should not return error when copying")

	assert.Equal(t, res.StatusCode, resCpy.StatusCode, "status code should be equal")
	assert.Equal(t, res.ContentLength, resCpy.ContentLength, "content type code should be equal")

	buf1, err := res.ReadBody()
	assert.Nil(t, err, "Should not return error when reading body")
	buf2, err := resCpy.ReadBody()
	assert.Nil(t, err, "Should not return error when reading body")
	assert.Equal(t, len(buf1), len(buf2), "buffors from response should have equal length")
	assert.Equal(t, len(buf1), 10000, "buffors from response should have equal length")
}