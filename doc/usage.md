# List of image operations

## Originals

## Rotate

Rotate the picture clockwise

Parameters
* rotate - rotation angle

<figure>
<img src="https://mort.mkaciuba.com/demo/rotate/img.jpg">
<figcaption>Rotate 90 degree</figcaption>
</figure>


<figure>
<img src="https://mort.mkaciuba.com/demo-query/img.jpg?operation=rotate&rotate=90">
<figcaption>Rotate 90 degree</figcaption>
</figure>


<figure>
<img src="https://mort.mkaciuba.com/demo-query/img.jpg?operation=rotate&rotate=180">
<figcaption>Rotate 180 degree</figcaption>
</figure>


## Blur

Blur the picture using Gaussian operator

Parameters:
* sigma - strength of operation
* minAmpl - TODO

<figure>
<img src="https://mort.mkaciuba.com/demo/blur/img.jpg">
<figcaption>Rotate 90 degree</figcaption>
</figure>


<figure>
<img src="https://mort.mkaciuba.com/demo-query/img.jpg?operation=blur&sigma=5">
<figcaption>Rotate 90 degree</figcaption>
</figure>


## Grayscale

Converts image to grayscale
Parameters: none


<figure>
<img src="https://mort.mkaciuba.com/demo/grayscale/img.jpg">
<figcaption>Rotate 90 degree</figcaption>
</figure>

<figure>
<img src="https://mort.mkaciuba.com/demo-query/img.jpg?grayscale=1">
<figcaption>Rotate 90 degree</figcaption>
</figure>


## Resize

Change the size of an image without clipping.
Parameters:
* width - choose width for the image. If not given, it will be calculated to preserve the aspect ratio.
* height - choose height for the image. If not given, it will be calculated to preserve the aspect ratio.

<figure>
<img src="https://mort.mkaciuba.com/demo/medium/img.jpg">
<figcaption>resize with width 500 </figcaption>
</figure>

## Crop

Crop the image.
Parameters:
* x - x offset for starting point (optional)
* y - y offset for starting point (optional)
* width - width of the cropped area.
* height - height of the cropped area.

## Watermark

Add watermark to image

Paramters:
* iamge: url or path to image for adding
* postion:  anchor point of image to combine with. See sectio
* opacity: choose transparency of image


## Image format

Change image format

Paramters:

format: iamge format

Formats:
* jpeg
* webp
* png
* bmp
ūū
