package config

type PresetsYaml struct {
	Quality int `yaml:"quality"`
	Filters struct {
		Thumbnail struct {
			Size []int  `yaml:"size"`
			Mode string `yaml:"mode"`
		} `yaml:"thumbnail"`
		Interlace struct {
			Mode string `yaml:"mode"`
		} `yaml:"interlace"`
		Crop struct {
			Size  []int  `yaml:"size"`
			Start []int  `yaml:"start"`
			Mode  string `yaml:"mode"`
		} `yaml:"crop"`
		SmartCrop struct {
			Size []int  `yaml:"size"`
			Mode string `yaml:"mode"`
		} `yaml:"entropy_crop"`
		AutoRotate interface{} `yaml:"auto_rtate"`
		Strip      interface{} `yaml:"strip"`
	} `yaml:"filters"`
}

type TransformYaml struct {
	Path    string                 `yaml:"path"`
	Kind    string                 `yaml:"kind"`
	Presets map[string]PresetsYaml `yaml:"presets"`
}

type Storage struct {
	RootPath   string `yaml:"rootPath"`
	Kind       string `yaml:"kind"`
}

type StorageTypes struct {
	Transform Storage `yaml:"transform"`
	Basic     Storage `yaml:"basic"`
}

type Bucket struct {
	Transform TransformYaml `yaml:"transform"`
	Storages  StorageTypes `yaml:"storages"`
}
