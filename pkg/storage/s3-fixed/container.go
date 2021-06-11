package s3_fixed

import (
	"io"
	"strings"
	"sync"

	"os"

	"github.com/aldor007/stow"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

// Amazon S3 bucket contains a creation date and a name.
type container struct {
	// name is needed to retrieve items.
	name string
	// client is responsible for performing the requests.
	client *s3.S3
	// region describes the AWS Availability Zone of the S3 Bucket.
	region         string
	customEndpoint string
	lock           sync.Mutex
}

type s3DataType struct {
	contentType        *string
	cacheControl       *string
	contentDisposition *string
	storageClass       *string
	contentMd5         *string
	tags               *string
	cannedAcl          *string
}

// ID returns a string value which represents the name of the container.
func (c *container) ID() string {
	return c.name
}

// Name returns a string value which represents the name of the container.
func (c *container) Name() string {
	return c.name
}

// Item returns a stow.Item instance of a container based on the name of the container and the key representing. The
// retrieved item only contains metadata about the object. This ensures that only the minimum amount of information is
// transferred. Calling item.Open() will actually do a get request and open a stream to read from.
func (c *container) Item(id string) (stow.Item, error) {
	return c.getItem(id)
}

// Items sends a request to retrieve a list of items that are prepended with
// the prefix argument. The 'cursor' variable facilitates pagination.
func (c *container) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	itemLimit := int64(count)

	params := &s3.ListObjectsV2Input{
		Bucket:     aws.String(c.Name()),
		StartAfter: &cursor,
		MaxKeys:    &itemLimit,
		Prefix:     &prefix,
	}

	response, err := c.client.ListObjectsV2(params)
	if err != nil {
		return nil, "", errors.Wrap(err, "Items, listing objects")
	}

	var containerItems []stow.Item

	for _, object := range response.Contents {
		if *object.StorageClass == "GLACIER" {
			continue
		}
		etag := cleanEtag(*object.ETag) // Copy etag value and remove the strings.
		object.ETag = &etag             // Assign the value to the object field representing the item.

		newItem := &item{
			container: c,
			client:    c.client,
			properties: properties{
				ETag:         object.ETag,
				Key:          object.Key,
				LastModified: object.LastModified,
				Owner:        object.Owner,
				Size:         object.Size,
				StorageClass: object.StorageClass,
			},
		}
		containerItems = append(containerItems, newItem)
	}

	// Create a marker and determine if the list of items to retrieve is complete.
	// If not, the last file is the input to the value of after which item to start
	startAfter := ""
	if *response.IsTruncated {
		startAfter = containerItems[len(containerItems)-1].Name()
	}

	return containerItems, startAfter, nil
}

func (c *container) RemoveItem(id string) error {
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(c.Name()),
		Key:    aws.String(id),
	}

	_, err := c.client.DeleteObject(params)
	if err != nil {
		return errors.Wrapf(err, "RemoveItem, deleting object %+v", params)
	}
	return nil
}

// Put sends a request to upload content to the container. The arguments
// received are the name of the item (S3 Object), a reader representing the
// content, and the size of the file. Many more attributes can be given to the
// file, including metadata. Keeping it simple for now.
func (c *container) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	uploader := s3manager.NewUploaderWithClient(c.client)

	// Convert map[string]interface{} to map[string]*string
	mdPrepped, s3Data, err := prepMetadata(metadata)

	// Perform an upload.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:             aws.String(c.name),
		Key:                aws.String(name),
		Body:               r,
		Metadata:           mdPrepped,
		ContentType:        s3Data.contentType,
		CacheControl:       s3Data.cacheControl,
		ContentDisposition: s3Data.contentDisposition,
		ContentMD5:         s3Data.contentMd5,
		StorageClass:       s3Data.storageClass,
		ACL:                s3Data.cannedAcl,
		Tagging:            s3Data.tags,
	})

	if err != nil {
		return nil, errors.Wrap(err, "Put, uploading object")
	}

	newItem := &item{
		container: c,
		client:    c.client,
		properties: properties{
			ETag: &result.UploadID,
			Key:  &name,
			// Owner        *s3.Owner
			// StorageClass *string
		},
	}
	switch file := r.(type) {
	case *os.File:
		if st, err := file.Stat(); err == nil && !st.IsDir() {
			newItem.properties.Size = aws.Int64(st.Size())
			newItem.properties.LastModified = aws.Time(st.ModTime())
		}
	default:
		newItem.properties.Size = aws.Int64(size)
	}
	return newItem, nil
}

// Region returns a string representing the region/availability zone of the container.
func (c *container) Region() string {
	return c.region
}

// A request to retrieve a single item includes information that is more specific than
// a PUT. Instead of doing a request within the PUT, make this method available so that the
// request can be made by the field retrieval methods when necessary. This is the case for
// fields that are left out, such as the object's last modified date. This also needs to be
// done only once since the requested information is retained.
// May be simpler to just stick it in PUT and and do a request every time, please vouch
// for this if so.
func (c *container) getItem(id string) (*item, error) {
	params := &s3.HeadObjectInput{
		Bucket: aws.String(c.name),
		Key:    aws.String(id),
	}
	res, err := c.client.HeadObject(params)
	if err != nil {
		// stow needs ErrNotFound to pass the test but amazon returns an opaque error
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			return nil, stow.ErrNotFound
		}
		return nil, errors.Wrap(err, "getItem, getting the object")
	}

	var etag string

	if res.ETag != nil {
		etag = cleanEtag(*res.ETag) // etag string value contains quotations. Remove them.
	}

	md, err := parseMetadata(res.Metadata)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve Item information, parsing metadata")
	}

	if res.CacheControl != nil {
		md["cache-control"] = *res.CacheControl
	}

	if res.ContentDisposition != nil {
		md["content-disposition"] = *res.ContentDisposition
	}

	if res.ContentEncoding != nil {
		md["content-encoding"] = *res.ContentEncoding
	}

	if res.ContentType != nil {
		md["content-type"] = *res.ContentType
	}

	if res.ContentLanguage != nil {
		md["content-language"] = *res.ContentLanguage
	}

	i := &item{
		container: c,
		client:    c.client,
		properties: properties{
			ETag:         &etag,
			Key:          &id,
			LastModified: res.LastModified,
			Owner:        nil, // not returned in the response.
			Size:         res.ContentLength,
			StorageClass: res.StorageClass,
			Metadata:     md,
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
func prepMetadata(md map[string]interface{}) (map[string]*string, s3DataType, error) {
	m := make(map[string]*string, len(md))
	s3Data := s3DataType{}
	for key, value := range md {
		key = strings.ToLower(key)
		strValue, valid := value.(string)
		if !valid {
			return nil, s3Data, errors.Errorf(`value of key '%s' in metadata must be of type string`, key)
		}
		awsValue := aws.String(strValue)
		switch key {
		case "cache-control":
			s3Data.cacheControl = awsValue
		case "content-type":
			s3Data.contentType = awsValue
		case "content-disposition":
			s3Data.contentDisposition = awsValue
		case "x-amz-storage-class":
			s3Data.storageClass = awsValue
		case "x-amz-tagging":
			s3Data.tags = awsValue
		case "content-md5":
			s3Data.contentMd5 = awsValue
		case "x-amz-acl":
			s3Data.cannedAcl = awsValue
		default:
			m[key] = awsValue
		}

	}
	return m, s3Data, nil
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
