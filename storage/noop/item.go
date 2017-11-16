package noop

import (
	"bytes"
	"io"
	"net/url"
	"time"

	"io/ioutil"
)

// The item struct contains an id (also the name of the file/S3 Object/Item),
// a container which it belongs to (s3 Bucket), a client, and a URL. The last
// field, properties, contains information about the item, including the ETag,
// file name/id, size, owner, last modified date, and storage class.
// see Object type at http://docs.aws.amazon.com/sdk-for-go/api/service/s3/
// for more info.
// All fields are unexported because methods exist to facilitate retrieval.
type item struct {
	url  string
	name string
}

// ID returns a string value that represents the name of a file.
func (i *item) ID() string {
	return ""
}

// Name returns a string value that represents the name of the file.
func (i *item) Name() string {
	return ""
}

// Size returns the size of an item in bytes.
func (i *item) Size() (int64, error) {
	return 0, nil
}

// URL returns a formatted string which follows the predefined format
// that every S3 asset is given.
func (i *item) URL() *url.URL {
	u, _ := url.Parse(i.url)
	return u
}

// Open retrieves specic information about an item based on the container name
// and path of the file within the container. This response includes the body of
// resource which is returned along with an error.
func (i *item) Open() (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader([]byte(""))), nil
}

// LastMod returns the last modified date of the item. The response of an item that is PUT
// does not contain this field. Solution? Detect when the LastModified field (a *time.Time)
// is nil, then do a manual request for it via the Item() method of the container which
// does return the specified field. This more detailed information is kept so that we
// won't have to do it again.
func (i *item) LastMod() (time.Time, error) {
	return time.Time{}, nil
}

// ETag returns the ETag value from the properies field of an item.
func (i *item) ETag() (string, error) {
	return time.Time{}.String(), nil
}

func (i *item) Metadata() (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}
