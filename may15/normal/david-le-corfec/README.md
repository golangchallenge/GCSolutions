Mosy - a mosaic generator
=========================

Mosy generates a mosaic image representing a target image, using supplied tile images.
The number of tiles and their resolution are adjustable.

### Usage

Usage: ./mosy -target=TARGET -tiles=TILEDIR -out=OUTPUT [OPTION]...

Mandatory parameters:
* target: the path of the image the mosaic will (hopefully) look like
* tiles: the directory containing the tile images (JPEG or PNG only)
* out: the path of the rendered image, ending in .png for PNG, else defaults to JPEG format

Optional parameters:
* xt: number of horizontal tiles in mosaic. Defaults to 120.
* yt: number of vertical tiles in mosaic. Defaults to 80.
* tw: tile width (in pixels). Larger tiles are downscaled to this size. Defaults to 50.
* th: tile height (in pixels). Defaults to 30.
* h: display help

### Exit status

* 0: success
* 1: missing mandatory parameter or help displayed
* 2: a dimension (xt, yt, tw or th) is not greater than zero
* 3: failed to render mosaic
* 4: failed to save result

### Example

> ./mosy -target=example/target.jpg -tiles=example/tiles -out=example/mosaic.png

### Internals

#### Tile setup

Each tile is downscaled (using averaging interpolation) to tw x th pixels.
The original aspect ratio is not preserved.
The average of each RGB channel is computed, giving a tile average color.

#### Tile selection

Given a target color, for each tile, the euclidean distance between the target
and the average tile color is computed. The tile with the smallest distance is selected.

#### Rendering

The target image is downscaled (using averaging interpolation) to xt x yt pixels.
For each target pixel, the closest tile to its color is selected and written to output.

