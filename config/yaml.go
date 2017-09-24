package config

type LiipFiltersYAML struct {
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
			Start []int `yaml:"start"`
			Mode string `yaml:"mode"`
		} `yaml:"crop"`
		SmartCrop struct {
			Size []int  `yaml:"size"`
			Mode string `yaml:"mode"`
		} `yaml:"entropy_crop"`
		AutoRotate interface{} `yaml:"auto_rtate"`
		Strip      interface{}       `yaml:"strip"`
	} `yaml:"filters"`
}

type LiipConfigYAML struct {
	LiipImagine struct {
		Resolvers struct {
			Default struct {
				WebPath interface{} `yaml:"web_path"`
			} `yaml:"default"`
		} `yaml:"resolvers"`
		FilterSets map[string] LiipFiltersYAML `yaml:"filter_sets"`
	} `yaml:"liip_imagine"`
}

