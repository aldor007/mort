package helpers

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestIsRangeOrCondition(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://url", nil)

	req.Header.Add("range", "0-1")

	assert.True(t, IsRangeOrCondition(req))

	req, _ = http.NewRequest("GET", "http://url", nil)

	req.Header.Add("if-match", "a")

	assert.True(t, IsRangeOrCondition(req))

	req, _ = http.NewRequest("GET", "http://url", nil)

	req.Header.Add("If-Unmodified-Since", "date")

	assert.True(t, IsRangeOrCondition(req))

	req, _ = http.NewRequest("GET", "http://url", nil)
	req.Header.Add("accept-encoding", "gzip")

	assert.False(t, IsRangeOrCondition(req))
}
