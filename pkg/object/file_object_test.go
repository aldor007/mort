package object

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/transforms"
	"github.com/h2non/bimg"
	"net/url"
)

var imageInfo = transforms.NewImageInfo(bimg.ImageMetadata{Size: bimg.ImageSize{Width: 100, Height: 100}}, "jpg")

func pathToURL(urlPath string) *url.URL {
	u, _ := url.Parse(urlPath)
	return u
}

func TestNewFileObjectWhenUnknowBucket(t *testing.T) {
	mortConfig := config.GetInstance()
	_, err := NewFileObject(pathToURL("/bucket/path"), mortConfig)

	assert.NotNil(t, err)
}

func TestNewFileObjectNoTransform(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-no-transform.yml")
	obj, err := NewFileObject(pathToURL("/bucket/path"), mortConfig)

	assert.Nil(t, err)

	assert.NotNil(t, obj)

	assert.False(t, obj.HasParent(), "obj shouldn't have parent")

	assert.Equal(t, "local", obj.Storage.Kind, "obj should have storage with kind of local")

}

func TestNewFileObjectTransform(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform.yml")
	obj, err := NewFileObject(pathToURL("/bucket/blog_small/bucket/parent.jpg"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")

	parent := obj.Parent

	assert.Equal(t, "/parent.jpg", parent.Key, "invalid parent key")

	assert.False(t, parent.HasParent(), "parent should't have parent")

	assert.True(t, obj.HasTransform(), "obj should have transform")

	transCfgArr, err := obj.Transforms.BimgOptions(imageInfo)
	transCfg := transCfgArr[0]

	assert.Nil(t, err, "Unexpected to have error when getting transforms")

	assert.Equal(t, 100, transCfg.Width, "invalid width for transform")

	assert.Equal(t, 100, transCfg.Height, "invalid height for transform")

}

func TestNewFileObjectTransformParentBucket(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform-parent-bucket.yml")
	obj, err := NewFileObject(pathToURL("/bucket/blog_small/thumb_2334.jpg"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")

	parent := obj.Parent

	assert.Equal(t, "/2334.jpg", parent.Key, "invalid parent key")
	assert.Equal(t, "bucket", parent.Bucket, "invalid parent key")

	assert.False(t, parent.HasParent(), "parent should't have parent")

	assert.True(t, obj.HasTransform(), "obj should have transform")

	transCfgArr, err := obj.Transforms.BimgOptions(imageInfo)
	assert.Nil(t, err, "Unexpected to have error when getting transforms")
	transCfg := transCfgArr[0]

	assert.Equal(t, 100, transCfg.Width, "invalid width for transform")

	assert.Equal(t, 100, transCfg.Height, "invalid height for transform")
}

func TestNewFileObjectTransformParentStorage(t *testing.T) {
	mortConfig := config.GetInstance()
	err := mortConfig.Load("testdata/bucket-transform-parent-storage.yml")
	assert.Nil(t, err, "Unexpected to have error when parsing config")
	obj, err := NewFileObject(pathToURL("/bucket/blog_small/thumb_2334.jpg"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")

	parent := obj.Parent

	assert.Equal(t, "http", parent.Storage.Kind, "invalid parent storage")

}

func TestNewFileObjectTransformOnlyWitdh(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform.yml")
	obj, err := NewFileObject(pathToURL("/bucket/width/bucket/parent.jpg"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")

	parent := obj.Parent

	assert.Equal(t, "/parent.jpg", parent.Key, "invalid parent key")

	assert.False(t, parent.HasParent(), "parent should't have parent")

	assert.True(t, obj.HasTransform(), "obj should have transform")

	transCfgArr, err := obj.Transforms.BimgOptions(imageInfo)
	assert.Nil(t, err, "Unexpected to have error when getting transforms")
	transCfg := transCfgArr[0]

	assert.Equal(t, 100, transCfg.Width, "invalid width for transform")

	assert.Equal(t, 0, transCfg.Height, "invalid height for transform")
}

func TestNewFileObjecWithNestedParent(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform.yml")
	obj, err := NewFileObject(pathToURL("/bucket/width/bucket/height/bucket/parent.jpg"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")

	parent := obj.Parent

	assert.True(t, parent.HasParent(), "parent should have parent")

	assert.Equal(t, "/parent.jpg", parent.Parent.Key, "parent of parent should have correct path")
}

func TestNewFileObjecWithNestedParentHash(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform-hash.yml")
	obj, err := NewFileObject(pathToURL("/bucket/width/bucket/height/bucket/parent.jpg"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")

	parent := obj.Parent

	assert.True(t, parent.HasParent(), "parent should have parent")

	assert.Equal(t, "/6ca/hei/height-bucket-parent.jpg-6ca0dabe9909875a", obj.Key)
}

func TestNewFileObjecWithNestedParentHashParent(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform-hashParent.yml")
	obj, err := NewFileObject(pathToURL("/bucket/width/bucket/height/bucket/parent.jpg"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")

	parent := obj.Parent

	assert.True(t, parent.HasParent(), "parent should have parent")

	assert.Equal(t, "/parent.jpg/2e805241bb54d7f7a200a56572d63805", obj.Key)
}

func TestNewFileObjectQueryResize(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	obj, err := NewFileObject(pathToURL("/bucket/parent.jpg?width=100"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")
	assert.True(t, obj.HasTransform(), "obj should have transforms")

	parent := obj.Parent

	assert.Equal(t, "/parent.jpg", parent.Key, "invalid parent key")

	transCfgArr, err := obj.Transforms.BimgOptions(imageInfo)
	assert.Nil(t, err, "Unexpected to have error when getting transforms")
	transCfg := transCfgArr[0]

	assert.Equal(t, 100, transCfg.Width, "invalid width for transform")
	assert.Equal(t, 0, transCfg.Height, "invalid width for transform")
}

func TestNewFileObjectQueryResizeDef(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	obj, err := NewFileObject(pathToURL("/bucket/parent.jpg?width=101&height=292"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")
	assert.True(t, obj.HasTransform(), "obj should have transforms")

	parent := obj.Parent

	assert.Equal(t, "/parent.jpg", parent.Key, "invalid parent key")

	transCfgArr, err := obj.Transforms.BimgOptions(imageInfo)
	assert.Nil(t, err, "Unexpected to have error when getting transforms")
	transCfg := transCfgArr[0]

	assert.Equal(t, 101, transCfg.Width, "invalid width for transform")
	assert.Equal(t, 292, transCfg.Height, "invalid width for transform")
}

func TestNewFileObjectQueryResize2(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	obj, err := NewFileObject(pathToURL("/bucket/parent.jpg?width=100&operation=resize"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")
	assert.True(t, obj.HasTransform(), "obj should have transforms")

	parent := obj.Parent

	assert.Equal(t, "/parent.jpg", parent.Key, "invalid parent key")

	transCfgArr, err := obj.Transforms.BimgOptions(imageInfo)
	assert.Nil(t, err, "Unexpected to have error when getting transforms")
	transCfg := transCfgArr[0]

	assert.Equal(t, 100, transCfg.Width, "invalid width for transform")
	assert.Equal(t, 0, transCfg.Height, "invalid width for transform")
}

func TestNewFileObjectQueryCrop(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	obj, err := NewFileObject(pathToURL("/bucket/parent.jpg?width=100&operation=crop&quality=20"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")
	assert.True(t, obj.HasTransform(), "obj should have transforms")

	parent := obj.Parent

	assert.Equal(t, "/parent.jpg", parent.Key, "invalid parent key")

	transCfgArr, err := obj.Transforms.BimgOptions(imageInfo)
	assert.Nil(t, err, "Unexpected to have error when getting transforms")
	transCfg := transCfgArr[0]

	assert.Equal(t, 100, transCfg.Width, "invalid width for transform")
	assert.Equal(t, 0, transCfg.Height, "invalid height for transform")
	assert.True(t, transCfg.Crop)
}

func TestNewFileObjectQueryNoTransform(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	obj, err := NewFileObject(pathToURL("/bucket/parent.jpg"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.False(t, obj.HasParent(), "obj shouldn't have parent")
	assert.False(t, obj.HasTransform(), "obj shouldn't have transforms")
}

func TestNewFileObjectPresetQueryBlur(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	obj, err := NewFileObject(pathToURL("/bucket/parent.jpg?width=100&operation=blur&format=jpg&grayscale=true&sigma=1&minAmpl=6&format=tiff"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")
	assert.True(t, obj.HasTransform(), "obj should have transforms")

	parent := obj.Parent

	assert.Equal(t, "/parent.jpg", parent.Key, "invalid parent key")

	transCfgArr, err := obj.Transforms.BimgOptions(imageInfo)
	assert.Nil(t, err, "Unexpected to have error when getting transforms")
	transCfg := transCfgArr[0]

	assert.Equal(t, 0, transCfg.Height, "invalid height for transform")
	assert.Equal(t, transCfg.GaussianBlur.Sigma, 1.)
	assert.Equal(t, transCfg.GaussianBlur.MinAmpl, 6.)
	assert.Equal(t, transCfg.Interpretation, bimg.InterpretationBW)
}

func TestNewFileObjectPresetQueryRotate(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	obj, err := NewFileObject(pathToURL("/bucket/parent.jpg?width=100&operation=rotate&format=jpg&angle=90&format=tiff"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")
	assert.True(t, obj.HasTransform(), "obj should have transforms")

	parent := obj.Parent

	assert.Equal(t, "/parent.jpg", parent.Key, "invalid parent key")

	transCfgArr, err := obj.Transforms.BimgOptions(imageInfo)
	assert.Nil(t, err, "Unexpected to have error when getting transforms")
	transCfg := transCfgArr[0]

	assert.Equal(t, 0, transCfg.Height, "invalid height for transform")
	assert.Equal(t, bimg.D90, transCfg.Rotate)
}

func TestNewFileObjectPresetQueryWatermarkErr(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	_, err := NewFileObject(pathToURL("/bucket/parent.jpg?width=100&operation=watermark&opacity=0.5&minAmpl=0.5&image=http://www"), mortConfig)
	assert.NotNil(t, err)

}

func TestNewFileObjectPresetQueryWatermark(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	obj, err := NewFileObject(pathToURL("/bucket/parent.jpg?width=100&operation=watermark&opacity=0.5&minAmpl=0.5&image=http://www&position=top-left"), mortConfig)
	assert.Nil(t, err)

	assert.True(t, obj.HasParent(), "obj should have parent")
	assert.True(t, obj.HasTransform(), "obj should have transforms")

	parent := obj.Parent

	assert.Equal(t, "/parent.jpg", parent.Key, "invalid parent key")
	assert.Equal(t, "/parent.jpg/90e8c676de60865efa843590826a08e1", obj.Key, "invalid object key")

	_, err = obj.Transforms.BimgOptions(imageInfo)

	assert.NotNil(t, err)
}

func TestNewFileObjectPresetPresetBlur(t *testing.T) {
	mortConfig := &config.Config{}
	err := mortConfig.Load("testdata/bucket-transform-preset-query.yml")
	if err != nil {
		t.Fatal(err)
	}
	obj, err := NewFileObject(pathToURL("/bucket/blog/parent.jpg"), mortConfig)
	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")
	assert.True(t, obj.HasTransform(), "obj should have transforms")

	parent := obj.Parent

	assert.Equal(t, "/parent.jpg", parent.Key, "invalid parent key")

	transCfgArr, err := obj.Transforms.BimgOptions(imageInfo)
	assert.Nil(t, err, "Unexpected to have error when getting transforms")
	transCfg := transCfgArr[0]

	assert.Equal(t, 0, transCfg.Height, "invalid height for transform")
	assert.Equal(t, transCfg.GaussianBlur.Sigma, 2.)
	assert.Equal(t, transCfg.GaussianBlur.MinAmpl, 3.)
	assert.Equal(t, transCfg.Interpretation, bimg.InterpretationBW)
}

func TestNewFileUnknownPreset(t *testing.T) {
	mortConfig := &config.Config{}
	err := mortConfig.Load("testdata/bucket-transform-preset-query.yml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewFileObject(pathToURL("/bucket/blog-unknown/parent.jpg"), mortConfig)
	assert.Nil(t, err, "Unexpected to have error when parsing path")
}

func TestObjectType(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	obj, err := NewFileObject(pathToURL("/bucket/parent.jpg?width=100&operation=extract&top=100&left=1&width=12&height=11"), mortConfig)
	assert.Nil(t, err)

	assert.Equal(t, obj.Type(), "transform")
	assert.Equal(t, obj.Parent.Type(), "parent")
}

func TestObjectCacheKeyQuery(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	obj, err := NewFileObject(pathToURL("/bucket/parent.jpg?width=100&operation=resizeCropAuto&width=100&height=100"), mortConfig)
	assert.Nil(t, err)

	assert.Equal(t, obj.GetResponseCacheKey(), "bucket/parent.jpg/179e44cbdd0d290829b89286562fa8b6")
}

func TestFileObject_FillWithRequest(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	obj, err := NewFileObjectFromPath("/bucket/parent.jpg?width=100&operation=resizeCropAuto&width=100&height=100", mortConfig)
	assert.Nil(t, err)

	req, _ := http.NewRequest("GET", "http://mort", nil)
	req.Header.Set("Range", "bytes=1-199")
	obj.FillWithRequest(req, req.Context())

	assert.Equal(t, obj.Range, "bytes=1-199")
}

func TestFileObject_ErrorObject(t *testing.T) {
	mortConfig := &config.Config{}
	mortConfig.Load("testdata/bucket-transform-query-parent-storage.yml")
	obj, err := NewFileObjectFromPath("/bucket/parent.jpg?width=100&operation=resizeCropAuto&width=100&height=100", mortConfig)
	assert.Nil(t, err)
	erroredObject, err := NewFileErrorObject("/parent", obj)
	assert.Nil(t, err)

	assert.Equal(t, erroredObject.GetResponseCacheKey(), "bucket/parenta692a0f768855173")
}
func TestObjectCacheKeyPreset(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform-hash.yml")
	obj, err := NewFileObject(pathToURL("/bucket/width/bucket/height/bucket/parent.jpg"), mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.Equal(t, "bucket/6ca/hei/height-bucket-parent.jpg-6ca0dabe9909875a", obj.GetResponseCacheKey())

	obj.Range = "bytes=1-10"

	assert.Equal(t, "bucket/6ca/hei/height-bucket-parent.jpg-6ca0dabe9909875abytes=1-10", obj.GetResponseCacheKey())
}

func BenchmarkNewFileObject(b *testing.B) {

	benchmarks := []struct {
		path       string
		configPath string
	}{
		{"/bucket/width/thumb_121332.jpg", "testdata/bucket-transform-parent-storage.yml"},
		{"/bucket/parent.jpg", "testdata/bucket-transform.yml"},
		{"/bucket/parent.jpg?width=100", "testdata/bucket-transform-query-parent-storage.yml"},
	}

	b.ReportAllocs()
	for _, bm := range benchmarks {
		configInstance := config.Config{}
		err := configInstance.Load(bm.configPath)
		if err != nil {
			panic(err)
		}

		b.Run(bm.path, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				NewFileObject(pathToURL(bm.path), &configInstance)
			}
		})
	}

}
