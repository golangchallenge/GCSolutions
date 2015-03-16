/*
Package drum is supposed to implement the decoding of .splice drum machine files.
See golang-challenge.com/go-challenge1/ for more information

The package contains functions to read and decode binary files representing
drum beat patterns.

.splice files describe a collection of tracks with playback instructions for
each. The information encoded therein contains:
	* version
	* tempo
	* tracks with each track containing
		* id
		* name
		* 16 steps
*/
package drum
