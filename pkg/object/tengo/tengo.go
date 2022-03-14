package tengo

import (
	"fmt"
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/transforms"
	"net/url"

	"github.com/aldor007/mort/pkg/object"
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

	err = t.Set("transforms", &Transforms{Value: transforms.Transforms{}})
	if err != nil {
		return "", err
	}

	err = t.Run()
	if err != nil {
		return "", err
	}

	parentTengo := t.Get("parent")
	transTengo := t.Get("transforms")
	transTengoObj, ok := transTengo.Object().(*Transforms)
	if !ok {
		return "", fmt.Errorf("unable to convert objet")
	}
	obj.Transforms = transTengoObj.Value
	parent := parentTengo.String()

	return parent, err
}
