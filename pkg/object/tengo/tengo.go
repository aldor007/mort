package tengo

import (
	"errors"
	"net/url"

	"github.com/aldor007/mort/pkg/config"

	"github.com/aldor007/mort/pkg/object"
	tengoLib "github.com/d5/tengo/v2"
)

func init() {
	object.RegisterParser("tengo", decodeUsingTengo)
}

// decodeUsingTengo parse given url by executing tengo script
func decodeUsingTengo(url *url.URL, bucketConfig config.Bucket, obj *object.FileObject) (string, error) {
	t := bucketConfig.Transform.TengoScript.Clone()
	err := t.Set("url", &URL{Value: url})
	if err != nil {
		return "", err
	}
	err = t.Set("bucketConfig", &BucketConfig{Value: bucketConfig})
	if err != nil {
		return "", err
	}
	err = t.Set("obj", &FileObject{Value: obj})
	if err != nil {
		return "", err
	}

	err = t.Run()
	if err != nil {
		return "", err
	}

	parentTengo := t.Get("parent")
	errorTengo := t.Get("err")
	if errorTengo.Object() != tengoLib.UndefinedValue {
		return "", errors.New(errorTengo.String())
	}
	parent := parentTengo.String()

	return parent, err
}
