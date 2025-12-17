package object

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryToTransform_WidthValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{"valid width", "width=100", false, ""},
		{"invalid width text", "width=invalid", true, "invalid syntax"},
		{"negative width", "width=-100", true, "width cannot be negative"},
		{"zero width", "width=0", true, "width cannot be zero"},
		{"empty width", "width=", true, "empty parameter value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://example.com/image.jpg?" + tt.query)
			trans, err := queryToTransform(u.Query())
			if tt.wantErr {
				require.NotNil(t, err, "should return error for %s", tt.query)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.Nil(t, err, "should not return error for %s", tt.query)
				assert.True(t, trans.NotEmpty)
			}
		})
	}
}

func TestQueryToTransform_HeightValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{"valid height", "height=200", false, ""},
		{"invalid height text", "height=invalid", true, "invalid syntax"},
		{"negative height", "height=-200", true, "height cannot be negative"},
		{"zero height", "height=0", true, "height cannot be zero"},
		{"empty height", "height=", true, "empty parameter value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://example.com/image.jpg?" + tt.query)
			trans, err := queryToTransform(u.Query())
			if tt.wantErr {
				require.NotNil(t, err, "should return error for %s", tt.query)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.Nil(t, err, "should not return error for %s", tt.query)
				assert.True(t, trans.NotEmpty)
			}
		})
	}
}

func TestQueryToTransform_QualityValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{"valid quality 1", "width=100&quality=1", false, ""},
		{"valid quality 50", "width=100&quality=50", false, ""},
		{"valid quality 100", "width=100&quality=100", false, ""},
		{"invalid quality 0", "width=100&quality=0", true, "quality must be between 1 and 100"},
		{"invalid quality 101", "width=100&quality=101", true, "quality must be between 1 and 100"},
		{"invalid quality -1", "width=100&quality=-1", true, "quality must be between 1 and 100"},
		{"invalid quality text", "width=100&quality=invalid", true, "invalid syntax"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://example.com/image.jpg?" + tt.query)
			_, err := queryToTransform(u.Query())
			if tt.wantErr {
				assert.NotNil(t, err, "should return error for %s", tt.query)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.Nil(t, err, "should not return error for %s", tt.query)
			}
		})
	}
}

func TestQueryToTransform_ResizeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{"valid resize with width", "operation=resize&width=100", false, ""},
		{"valid resize with height", "operation=resize&height=200", false, ""},
		{"valid resize with both", "operation=resize&width=100&height=200", false, ""},
		{"invalid negative width", "operation=resize&width=-100", true, "width and height cannot be negative"},
		{"invalid negative height", "operation=resize&height=-200", true, "width and height cannot be negative"},
		{"invalid both zero", "operation=resize&width=0&height=0", true, "at least one of width or height must be specified"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://example.com/image.jpg?" + tt.query)
			trans, err := queryToTransform(u.Query())
			if tt.wantErr {
				require.NotNil(t, err, "should return error for %s", tt.query)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.Nil(t, err, "should not return error for %s", tt.query)
				assert.True(t, trans.NotEmpty)
			}
		})
	}
}

func TestQueryToTransform_CropValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{"valid crop", "operation=crop&width=100&height=200", false, ""},
		{"invalid negative width", "operation=crop&width=-100&height=200", true, "width and height cannot be negative"},
		{"invalid negative height", "operation=crop&width=100&height=-200", true, "width and height cannot be negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://example.com/image.jpg?" + tt.query)
			trans, err := queryToTransform(u.Query())
			if tt.wantErr {
				require.NotNil(t, err, "should return error for %s", tt.query)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.Nil(t, err, "should not return error for %s", tt.query)
				assert.True(t, trans.NotEmpty)
			}
		})
	}
}

func TestQueryToTransform_ExtractValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{"valid extract", "operation=extract&top=0&left=0&areaWith=100&areaHeight=100", false, ""},
		{"valid extract with offset", "operation=extract&top=10&left=20&areaWith=100&areaHeight=100", false, ""},
		{"invalid negative top", "operation=extract&top=-1&left=0&areaWith=100&areaHeight=100", true, "extract coordinates cannot be negative"},
		{"invalid negative left", "operation=extract&top=0&left=-1&areaWith=100&areaHeight=100", true, "extract coordinates cannot be negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://example.com/image.jpg?" + tt.query)
			trans, err := queryToTransform(u.Query())
			if tt.wantErr {
				require.NotNil(t, err, "should return error for %s", tt.query)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.Nil(t, err, "should not return error for %s", tt.query)
				assert.True(t, trans.NotEmpty)
			}
		})
	}
}

func TestQueryToTransform_WatermarkValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{"valid watermark", "operation=watermark&image=test.jpg&position=top-left&opacity=0.5", false, ""},
		{"valid opacity 0", "operation=watermark&image=test.jpg&position=top-left&opacity=0", false, ""},
		{"valid opacity 1", "operation=watermark&image=test.jpg&position=top-left&opacity=1.0", false, ""},
		{"invalid opacity negative", "operation=watermark&image=test.jpg&position=top-left&opacity=-0.1", true, "opacity must be between 0 and 1"},
		{"invalid opacity > 1", "operation=watermark&image=test.jpg&position=top-left&opacity=1.1", true, "opacity must be between 0 and 1"},
		{"invalid opacity text", "operation=watermark&image=test.jpg&position=top-left&opacity=invalid", true, "invalid opacity value"},
		{"missing opacity", "operation=watermark&image=test.jpg&position=top-left", true, "opacity parameter is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://example.com/image.jpg?" + tt.query)
			trans, err := queryToTransform(u.Query())
			if tt.wantErr {
				require.NotNil(t, err, "should return error for %s", tt.query)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.Nil(t, err, "should not return error for %s", tt.query)
				assert.True(t, trans.NotEmpty)
			}
		})
	}
}

func TestQueryToTransform_BlurValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{"valid blur", "operation=blur&sigma=1.0", false, ""},
		{"valid blur with minAmpl", "operation=blur&sigma=5.0&minAmpl=2.0", false, ""},
		{"invalid sigma 0", "operation=blur&sigma=0", true, "sigma must be positive"},
		{"invalid sigma negative", "operation=blur&sigma=-1.0", true, "sigma must be positive"},
		{"invalid sigma text", "operation=blur&sigma=invalid", true, "invalid sigma value"},
		{"missing sigma", "operation=blur", true, "sigma parameter is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://example.com/image.jpg?" + tt.query)
			trans, err := queryToTransform(u.Query())
			if tt.wantErr {
				require.NotNil(t, err, "should return error for %s", tt.query)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.Nil(t, err, "should not return error for %s", tt.query)
				assert.True(t, trans.NotEmpty)
			}
		})
	}
}

func TestQueryToTransform_RotateValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{"valid rotate 90", "operation=rotate&angle=90", false, ""},
		{"valid rotate 180", "operation=rotate&angle=180", false, ""},
		{"valid rotate 270", "operation=rotate&angle=270", false, ""},
		{"invalid angle 45", "operation=rotate&angle=45", true, "wrong angle"},
		{"invalid angle text", "operation=rotate&angle=invalid", true, "invalid angle value"},
		{"missing angle", "operation=rotate", true, "angle parameter is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://example.com/image.jpg?" + tt.query)
			trans, err := queryToTransform(u.Query())
			if tt.wantErr {
				require.NotNil(t, err, "should return error for %s", tt.query)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.Nil(t, err, "should not return error for %s", tt.query)
				assert.True(t, trans.NotEmpty)
			}
		})
	}
}

func TestQueryToTransform_FormatValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{"valid format jpeg", "width=100&format=jpeg", false, ""},
		{"valid format webp", "width=100&format=webp", false, ""},
		{"valid format png", "width=100&format=png", false, ""},
		{"invalid format", "width=100&format=invalid_format", true, "Unknown format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://example.com/image.jpg?" + tt.query)
			_, err := queryToTransform(u.Query())
			if tt.wantErr {
				assert.NotNil(t, err, "should return error for %s", tt.query)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.Nil(t, err, "should not return error for %s", tt.query)
			}
		})
	}
}

func TestValidatePositiveInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     int
		paramName string
		wantErr   bool
	}{
		{"valid positive", 100, "width", false},
		{"valid 1", 1, "width", false},
		{"invalid zero", 0, "width", true},
		{"invalid negative", -100, "width", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePositiveInt(tt.value, tt.paramName)
			if tt.wantErr {
				assert.NotNil(t, err, "validatePositiveInt(%d) should return error", tt.value)
			} else {
				assert.Nil(t, err, "validatePositiveInt(%d) should not return error", tt.value)
			}
		})
	}
}
