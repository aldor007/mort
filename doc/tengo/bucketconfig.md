# bucket config object
This object is injected into script and available via `bucketConfig` value

## properties

* `transform` - returns [Transform object](#transform) description for image manipulation
* `keys` - array of map with S3 style `accessKey` and `secretAccessKey`
* `headers` - map[string]string with header configured for bucket
* `name` - string name of bucket

Example usage

```go
[...]
parse := func(reqUrl, bucketConfig, obj) {
	trans := bucketConfig.transform
	if ok := trans.presets[presetName]; ok == undefined {
		return ["", error("unknown preset " + presetName)]
	}

```

# transform

Object describing what more should do with image. It is based on mort config


## properties

* `path` - string, mort regexp describing parent and transforms
* `parentStorage` - string, override parent storage
* `parentBucket` - string, overrride parent storage
* `pathRegexp` - compiled regexp from `path` string. can be used as a function
* `kind` - string, type of transformation
* `presets` - map, map of object which contains description of transformations. More about [Preset](#preset)

Example usage

```go
    presetName := "mypreset"
	if ok := bucketConfig.transform.presets[presetName]; ok == undefined {
		return ["", error("unknown preset " + presetName)]
    }
```

# preset

It is tengo representation of mort preset configuration. Each property is lowercase and in camelCase

## properties

* `quality` - int, quality of image
* `format` - string, format of image
* `filters` - filters used for image transformation. More about it [here](/doc/Image-Operations.md) please take a look on "preset" type

Example usage

```go
presetToTransform := func(obj, preset) {
	filters := preset.filters

	if filters.thumbnial {
		err := obj.transforms.resize(filters.thumbnial.width, filters.thumbnial.height, filters.thumbnial.mode == "outbound", filters.thumbnial.preserveAspectRatio, filters.thumbnial.fill)
		if err != undefined {
			return err
		}
	}

	if filters.crop {
		err := obj.transforms.crop(filters.crop.width, filters.crop.height, filters.crop.gravity, filters.crop.mode == "outbound", filters.crop.embed)
		if err != undefined {
			return err
		}
	}

```




