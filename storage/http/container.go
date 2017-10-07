package http

import (
	"io"
	"strings"
	"net/http"
	"github.com/aldor007/stow"
	"github.com/pkg/errors"

"fmt"
)

type container struct {
	name          string
	endpoint       string
	client 			*http.Client
	headers        map[string]string
}

// ID returns a string value which represents the name of the container.
func (c *container) ID() string {
	return c.name
}

// Name returns a string value which represents the name of the container.
func (c *container) Name() string {
	return c.name
}

// Item returns a stow.Item instance of a container based on the
// name of the container and the key representing
func (c *container) Item(id string) (stow.Item, error) {
	return c.getItem(id)
}

// Items sends a request to retrieve a list of items that are prepended with
// the prefix argument. The 'cursor' variable facilitates pagination.
func (c *container) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	return nil, "", errors.New("not implemented")
}

func (c *container) RemoveItem(id string) error {
	return errors.New("not implemented")
}

func (c *container) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	return nil, errors.New("not implemented")
}


// A request to retrieve a single item includes information that is more specific than
// a PUT. Instead of doing a request within the PUT, make this method available so that the
// request can be made by the field retrieval methods when necessary. This is the case for
// fields that are left out, such as the object's last modified date. This also needs to be
// done only once since the requested information is retained.
// May be simpler to just stick it in PUT and and do a request every time, please vouch
// for this if so.
func (c *container) getItem(id string) (*item, error) {
	endpoint := strings.Replace(c.endpoint, "<item>", id, 1)
	fmt.Println("cont", endpoint)
	req, err := http.NewRequest("HEAD", endpoint, nil)
	if err != nil {
		return nil, err
	}

	for h, v := range c.headers {
		req.Header.Set(h, v)
	}

	res, err := c.client.Do(req)

	if err != nil {
		// stow needs ErrNotFound to pass the test but amazon returns an opaque error
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, stow.ErrNotFound
		}
		return nil, errors.Wrap(err, "getItem, getting the object")
	}

	if res.StatusCode == 404 {
		return nil, stow.ErrNotFound
	}

	//defer res.Body.Close()

	//etag := cleanEtag(res.Header().Get("ETag")) // etag string value contains quotations. Remove them.
	//md, err := parseMetadata(res.Header())
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to retrieve Item information, parsing metadata")
	//}

	i := &item{
		container: c,
		client:    c.client,
		name: id,
		url: endpoint,
		properties: properties{
			Key:          &id,
			Size:         &res.ContentLength,
		},
	}

	return i, nil
}

// Remove quotation marks from beginning and end. This includes quotations that
// are escaped. Also removes leading `W/` from prefix for weak Etags.
//
// Based on the Etag spec, the full etag value (<FULL ETAG VALUE>) can include:
// - W/"<ETAG VALUE>"
// - "<ETAG VALUE>"
// - ""
// Source: https://tools.ietf.org/html/rfc7232#section-2.3
//
// Based on HTTP spec, forward slash is a separator and must be enclosed in
// quotes to be used as a valid value. Hence, the returned value may include:
// - "<FULL ETAG VALUE>"
// - \"<FULL ETAG VALUE>\"
// Source: https://www.w3.org/Protocols/rfc2616/rfc2616-sec2.html#sec2.2
//
// This function contains a loop to check for the presence of the three possible
// filler characters and strips them, resulting in only the Etag value.
func cleanEtag(etag string) string {
	for {
		// Check if the filler characters are present
		if strings.HasPrefix(etag, `\"`) {
			etag = strings.Trim(etag, `\"`)

		} else if strings.HasPrefix(etag, `"`) {
			etag = strings.Trim(etag, `"`)

		} else if strings.HasPrefix(etag, `W/`) {
			etag = strings.Replace(etag, `W/`, "", 1)

		} else {
			break
		}
	}
	return etag
}

// prepMetadata parses a raw map into the native type required by S3 to set metadata (map[string]*string).
// TODO: validation for key values. This function also assumes that the value of a key value pair is a string.
func prepMetadata(md map[string]interface{}) (map[string]*string, error) {
	m := make(map[string]*string, len(md))
	for key, value := range md {
		/*strValue*/_, valid := value.(string)
		if !valid {
			return nil, errors.Errorf(`value of key '%s' in metadata must be of type string`, key)
		}
		//m[key] = *strValue
	}
	return m, nil
}

// The first letter of a dash separated key value is capitalized, so perform a ToLower on it.
// This Key transformation of returning lowercase is consistent with other locations..
func parseMetadata(md map[string]*string) (map[string]interface{}, error) {
	m := make(map[string]interface{}, len(md))
	for key, value := range md {
		k := strings.ToLower(key)
		m[k] = *value
	}
	return m, nil
}
