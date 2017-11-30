# List of image operations

## Originals

<a href="https://mort.mkaciuba.com/demo/img.jpg">Image</a>

## Rotate

Rotate the picture clockwise

Parameters
* angle - rotation angle

### Preset

<a href="https://mort.mkaciuba.com/demo/rotate/img.jpg">
<figure>
<img src="https://mort.mkaciuba.com/demo/rotate/img.jpg">
<figcaption>Rotate 90 degree</figcaption>
</figure>
</a>

### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?operation=rotate&angle=90">
<figure>
<img src="https://mort.mkaciuba.com/demo/img.jpg?operation=rotate&angle=90">
<figcaption>Rotate 90 degree</figcaption>
</figure>
</a>

<a href="https://mort.mkaciuba.com/demo/img.jpg?operation=rotate&angle=90">
<figure>
<img src="https://mort.mkaciuba.com/demo/img.jpg?operation=rotate&angle=180">
<figcaption>Rotate 180 degree</figcaption>
</figure>
</a>


## Blur

Blur the picture using Gaussian operator

Parameters:
* sigma - strength of operation

### Preset

<a href="https://mort.mkaciuba.com/demo/blur/img.jpg">
<figure>
<img src="https://mort.mkaciuba.com/demo/blur/img.jpg">
<figcaption>Blur image with sigma 5</figcaption>
</figure>
</a>


#### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpd?operation=blur&sigma=10">
<figure>
<img src="https://mort.mkaciuba.com/demo/img.jpg?operation=blur&sigma=10">
<figcaption>Blur image sigma 10</figcaption>
</figure>
</a>


## Grayscale

Converts image to grayscale

Parameters: none

### Preset

<a href="https://mort.mkaciuba.com/demo/grayscale/img.jpg">
<figure>
<img src="https://mort.mkaciuba.com/demo/grayscale/img.jpg">
<figcaption>Change image colors to grayscale</figcaption>
</figure>
</a>

## Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?grayscale=1">
<figure>
<img src="https://mort.mkaciuba.com/demo/img.jpg?grayscale=1">
<figcaption>Change image colors to grayscale</figcaption>
</figure>
</a>


## Resize

Change the size of an image without clipping.
Parameters:
* width - choose width for the image. If not given, it will be calculated to preserve the aspect ratio.
* height - choose height for the image. If not given, it will be calculated to preserve the aspect ratio.

### Preset

<a href="https://mort.mkaciuba.com/demo/medium/img.jpg">
<figure>
<img src="https://mort.mkaciuba.com/demo/medium/img.jpg">
<figcaption>resize with width 500 </figcaption>
</figure>
</a>

### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?width=500">
<figure>
<img src="https://mort.mkaciuba.com/demo/img.jpg?width=500">
<figcaption>resize with width 500 </figcaption>
</figure>
</a>


## Crop

Crop  smart the image.
Parameters:
* width - width of the cropped area.
* height - height of the cropped area.

### Preset 

<a href="https://mort.mkaciuba.com/demo/crop/img.jpg">
<figure>
<img src="https://mort.mkaciuba.com/demo/crop/img.jpg">
<figcaption>crop image with width 500 </figcaption>
</figure>
</a>

### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?operation=crop&width=200&height=200">
<figure>
<img src="https://mort.mkaciuba.com/demo/img.jpg?opetation=crop&width=200&height=200">
<figcaption>crop image with width 200 and height 200 </figcaption>
</figure>
</a>

## Watermark

Add watermark to image

Paramters:
* image: url or path to image for adding
* postion:  anchor point of image to combine with. See section TODO
* opacity: choose transparency of image

### Preset 

<a href="https://mort.mkaciuba.com/demo/watermark/img.jpg">
<figure>
<img src="https://mort.mkaciuba.com/demo/watermark/img.jpg">
<figcaption>Add gradient to image</figcaption>
</figure>
</a>

### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?operation=watermark&image=https://i.imgur.com/uomkVIL.png&position=top-left&opacity=0.5&width=500&operation=resize">
<figure>
<img src="https://mort.mkaciuba.com/demo/crop/img.jpg">
<figcaption>Add gradient to iamge</figcaption>
</figure>
</a>

## Image format

Change image format

Paramters:

format: image format

Formats:
* jpeg
* webp
* png
* bmp

### Preset

<a href="https://mort.mkaciuba.com/demo/webp/img.jpg">
<figure>
<img src="https://mort.mkaciuba.com/demo/webp/img.jpg">
<figcaption>Change image format to webp</figcaption>
</figure>
</a>

### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?format=webp">
<figure>
<img src="https://mort.mkaciuba.com/demo/img.jpg?format=webp">
<figcaption>Change image format to webp</figcaption>
</figure>
</a>
