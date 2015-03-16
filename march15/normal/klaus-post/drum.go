// Copyright (c) 2015, Klaus Post
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright
// notice, this list of conditions and the following disclaimer in the
// documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its
// contributors may be used to endorse or promote products derived from this
// software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

/*

Package drum implements the decoding of .splice drum machine files.

To decode a file on disk, and retrieve the pattern stored in the file,
use the DecodeFile() function:

	pattern, err := DecodeFile("filename.splice")
	if err == nil {
		fmt.Println(pattern)
	}

For more advanced usage, you can use the DecodeReader() function,
which works in the same way, but takes a reader as argument.

Once you have loaded the pattern, you can inspect it. Here are some examples:

	p, _ := DecodeFile("filename.splice")
	fmt.Println(p)
	fmt.Println("* Longest track is", p.LongestTrack(), "beats")
	fmt.Println("* Duration of pattern is", p.Duration())
	fmt.Println("* One beat takes", p.BeatDuration())
	fmt.Println("* First track instrument is", p.Tracks[0].Instrument)
	fmt.Println("* At first beat we will play", p.PlayAt(0))

The output could then look like this:

 Saved with HW Version: 0.808-alpha
 Tempo: 120
 (0) kick        |x---|x---|x---|x---|
 (1) snare       |----|x---|----|x---|
 (2) clap        |----|x-x-|----|----|
 (3) hh-open     |--x-|--x-|x-x-|--x-|
 (4) hh-close    |x---|x---|----|x--x|
 (5) cowbell     |----|----|--x-|----|

 * Longest track is 16 beats
 * Duration of pattern is 8s
 * One beat takes 500ms
 * First track Instrument is (0) kick
 * At first beat we will play [(0) kick (4) hh-close]

*/
package drum
