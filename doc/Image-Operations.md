# List of image operations

Table of content

- [List of image operations](#list-of-image-operations)
  * [Originals](#originals)
  * [Rotate](#rotate)
    + [Preset](#preset)
    + [Query string](#query-string)
  * [Blur](#blur)
    + [Preset](#preset-1)
    + [Query string](#query-string-1)
  * [Grayscale](#grayscale)
    + [Preset](#preset-2)
    + [Query string](#query-string-2)
  * [Resize](#resize)
    + [Preset](#preset-3)
    + [Query string](#query-string-3)
  * [Crop](#crop)
    + [Preset](#preset-4)
    + [Query string](#query-string-4)
  * [Watermark](#watermark)
    + [Preset](#preset-5)
    + [Query string](#query-string-5)
  * [Image format](#image-format)
    + [Preset](#preset-6)
    + [Query string](#query-string-6)

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
<figcaption><br/>Rotate 90 degree</figcaption>
</figure>
</a>

### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?operation=rotate&angle=90">
<figure>
<img  align="center" src="https://mort.mkaciuba.com/demo/img.jpg?operation=rotate&angle=90" alt="angle=90">
<figcaption><br/>Rotate 90 degree</figcaption>
</figure>
</a>

<a href="https://mort.mkaciuba.com/demo/img.jpg?operation=rotate&angle=90">
<figure>
<img aling="center" src="https://mort.mkaciuba.com/demo/img.jpg?operation=rotate&angle=180">
<figcaption><br/>Rotate 180 degree</figcaption>
</figure>
</a>


## Blur

Blur the picture using Gaussian operator

Parameters:
* sigma - strength of operation

### Preset

<a href="https://mort.mkaciuba.com/demo/blur/img.jpg">
<figure>
<img align="center" src="https://mort.mkaciuba.com/demo/blur/img.jpg">
<figcaption><br/>Blur image with sigma 5</figcaption>
</figure>
</a>


### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?operation=blur&sigma=10">
<figure>
<img align="center" src="https://mort.mkaciuba.com/demo/img.jpg?operation=blur&sigma=10">
<figcaption><br/>Blur image sigma 10</figcaption>
</figure>
</a>


## Grayscale

Converts image to grayscale

Parameters: none

### Preset

<a href="https://mort.mkaciuba.com/demo/grayscale/img.jpg">
<figure>
<img src="https://mort.mkaciuba.com/demo/grayscale/img.jpg">
 <figcaption></br>Change image colors to grayscale</figcaption>
</figure>
</a>

### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?grayscale=1">
<figure>
<img src="https://mort.mkaciuba.com/demo/img.jpg?grayscale=1">
<figcaption><br/>Change image colors to grayscale</figcaption>
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
<figcaption><br/>resize with width 500 </figcaption>
</figure>
</a>

### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?width=500">
<figure>
<img src="https://mort.mkaciuba.com/demo/img.jpg?width=500">
<figcaption><br/>resize with width 500 </figcaption>
</figure>
</a>


## Crop

Crop  smart the image.
Parameters:
* width - width of the cropped area.
* height - height of the cropped area.
* gravity - position of crop (optional)
Position can be one of:
  + center
  + north
  + west
  + east
  + south
  + smart

### Preset 

<a href="https://mort.mkaciuba.com/demo/crop/img.jpg">
<figure>
<img src="https://mort.mkaciuba.com/demo/crop/img.jpg">
<figcaption><br/>crop image with width 500 </figcaption>
</figure>
</a>

### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?operation=crop&width=200&height=200">
<figure>
<img src="https://mort.mkaciuba.com/demo/img.jpg?opetation=crop&width=200&height=200">
<figcaption><br/>crop image with width 200 and height 200 </figcaption>
</figure>
</a>

</br>

<a href="https://mort.mkaciuba.com/demo/img.jpg?operation=crop&width=200&height=200">
<figure>
<img src="https://mort.mkaciuba.com/demo/img.jpg?opetation=crop&width=200&height=200&gravity=north">
<figcaption><br/>crop image with width 200 and height 200 </figcaption>
</figure>
</a>

## Watermark

Add watermark to image

Paramters:
* image: url or path to image for adding
* opacity: choose transparency of image
* position:  anchor point of image to combine with. Can be one of:
 + top-left
 + top-center
 + top-right
 + center-left
 + center-center
 + center-right
 + bottom-left
 + bottom-center
 + bottom-right

### Preset 

<a href="https://mort.mkaciuba.com/demo/watermark/img.jpg">
<figure>
<img src="https://mort.mkaciuba.com/demo/watermark/img.jpg">
<figcaption><br/>Add gradient to image</figcaption>
</figure>
</a>

### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?operation=watermark&image=https://i.imgur.com/uomkVIL.png&position=top-left&opacity=0.5&width=500&operation=resize">
<figure>
<img src="https://mort.mkaciuba.com/demo/crop/img.jpg">
<figcaption><br/>Add gradient to iamge</figcaption>
</figure>
</a>

## Image format

Change image format

Parameters:

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
<figcaption><br/>Change image format to webp</figcaption>
</figure>
</a>

### Query string

<a href="https://mort.mkaciuba.com/demo/img.jpg?format=webp">
<figure>
<img src="https://mort.mkaciuba.com/demo/img.jpg?format=webp">
<figcaption><br/>Change image format to webp</figcaption>
</figure>
</a>
