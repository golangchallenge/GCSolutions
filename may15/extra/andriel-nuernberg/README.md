# Bambu Mosaics

Bambu Mosaics is a web application developed for the [third Go Challenge](http://golang-challenge.com/go-challenge3).

![](http://f.cl.ly/items/2f2W2b1t243d1T1s2d3G/Screen%20Shot%202015-05-24%20at%2014.54.08.png)

# Design

The layout and logo was made by myself, mainly in PhotoShop. The cats in the
background I got from the Instagram user [@cats_of_insragram](https://instagram.com/cats_of_instagram).

The PSD file is [available here](http://cl.ly/1M3R2j072t0d).

# Tile set

I've used the Instagram API to download a bunch of thumbnails (~37k). The default size of Instagram
thumbnails is 150x150, but for a better performance and mosaic results, I resized all tiles to 15x15.

# Closest tiles algorithm

- [http://en.wikipedia.org/wiki/Euclidean_distance](http://en.wikipedia.org/wiki/Euclidean_distance)
- [http://stackoverflow.com/questions/1847092/given-an-rgb-value-what-would-be-the-best-way-to-find-the-closest-match-in-the-d](http://stackoverflow.com/questions/1847092/given-an-rgb-value-what-would-be-the-best-way-to-find-the-closest-match-in-the-d)
- [http://stackoverflow.com/questions/1313/followup-finding-an-accurate-distance-between-colors](http://stackoverflow.com/questions/1313/followup-finding-an-accurate-distance-between-colors)

# Deployment

Docker was used to deploy the app on DigitalOcean. You can see the [Dockerfile here](Dockerfile).

Right now the app is working under http://andrielfn.com:4000. Maybe I move it to http://bambu.andrielfn.com, but not sure yet.