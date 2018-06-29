package helpers

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"testing"

	"github.com/pkg/errors"
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

func TestFetchUrlObject(t *testing.T) {
	defer gock.Off()

	gock.New("http://image.om").
		Get("/bar.jpg").
		Reply(200).
		BodyString("foo foo")

	gock.InterceptClient(client)
	buf, err := FetchObject("http://image.om/bar.jpg")

	assert.Nil(t, err)
	assert.Equal(t, string(buf), "foo foo")
}

func TestFetchUrlObjectErr(t *testing.T) {
	defer gock.Off()

	gock.New("http://image.om").
		Get("/bar.jpg").
		ReplyError(errors.New("error"))

	gock.InterceptClient(client)
	_, err := FetchObject("http://image.om/bar.jpg")

	assert.NotNil(t, err)
}

func TestFetchObjectErr(t *testing.T) {
	_, err := FetchObject("bar.jpg")

	assert.NotNil(t, err)
}

func TestFetchObject(t *testing.T) {
	_, err := FetchObject("./helpers.go")

	assert.Nil(t, err)
}
