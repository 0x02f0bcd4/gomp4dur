# gomp4dur

A simple, stupid golang library to find the duration of mp4 files. 

## Usage

Get the library by running the following command

```
go get -u github.com/0x02f0bcd4/gomp4dur
```

## Why?

Because I needed it in my job. Thought about making it public, so here we are.

### Why didn't I just use ffprobe?

I tried to, but then I got some permission issue. Also, I couldn't bother myself to read through thousand different documentation to understand the different flags of ffprobe. I did some **gemini** to get as of how it works, but halfway through I was like, meh, let's just write a library.

## Technologia

Nothing interesting to be honest, the library finds the duration by _leapfrogging_ throughout the file. It checks if the MOOV atom of the file has duration set to 0 or not, if so, the library identifies the video as fMP4 (fragmented MP4) and tries to read the TRAK atom to find out the video tracks. Then it searches the MOOF atoms to get the duration of each fragments.

## Test coverage

Now, I could cover only two test edges -

1. The MOOV atom has the timescale and duration.
2. The MOOF atom's TFHD has `default_sample_duration` flag set.

However, as for the case where the TFHD atom's `default_sample_duration` flag is not set and to find the sample's duration I need to read the `SAMPLE_TABLE` to get the total SAMPLE duration, I couldn't find a video neither could I generate a video with the help of FFMPEG and MP4BOX to test the case. If you can find it, please let me know. Thank you.


## Support

Please, if you have the PDF of the MP4 file structure documentation ([located here](https://www.iso.org/standard/79110.html)) please consider sharing it with me. I've done the job mostly through reverse engineering (thanks to mp4box.js) and some help from Googling. Also, if you know how to generate a VFR video with FFMPEG (which I abysmally failed to after going through several stackoverflow answers, most of which are decades old at this point and many of the ffmpeg options are no longer available/deprecated) or have a no-copyright, no-royalty VFR video   to test if the library can find the duration reliably, do let me know.

As for supporting through code, do open a PR and let me know about it. Don't bother, however, making a "Fix typo" PR. You have been warned.

## License

This library is licensed under MIT License. Read the LICENSE.txt file to know more.

