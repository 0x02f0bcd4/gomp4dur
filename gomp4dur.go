package gomp4dur

import (
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/0x02f0bcd4/gomp4dur/internal/atomutil"
	"github.com/0x02f0bcd4/gomp4dur/internal/moofutil"
	"github.com/0x02f0bcd4/gomp4dur/internal/moovutil"
	"github.com/0x02f0bcd4/gomp4dur/internal/trakutil"
)

// This function will try to get the duration of the given MP4.
// User of this function should note that the file should be opened
// with the permission of reading it. There should be no assumption as of
// where the file seeking position will be after the operation of the
// file.
func Get(file *os.File) (duration float64, err error) {
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return
	}

	_, atom_name, _, err := atomutil.ReadAtomNameAndSize(file)

	if err != nil {
		return
	} else if atom_name != "ftyp" {
		err = fmt.Errorf("Failed to parse the file, the starting atom isn't named FTYP. File could be malformed or not an MP4 file at all")
		return
	}

	moov_info, obtained_timescale, obtained_duration, err := moovutil.ReadMVHDTimescaleAndDuration(file)
	if err != nil {
		return
	}

	//else, check if the obtained_duration is 0 or not
	if obtained_duration == 0 {
		//Jarvis, deploy the MOOF-util
		if len(moov_info.TRAC_locs) == 0 {
			err = fmt.Errorf("Failed to parse the MP4 duration, no track was recorded from the MP4 file.")
			return
		}

		//get the track related information
		var track_info []trakutil.TrakInfo
		track_info, err = trakutil.GetVideoTrackInfo(file, moov_info.TRAC_locs)
		if err != nil {
			return
		}
		//else, go over the tracks and find the type of the track
		var video_track_set map[uint32]bool = make(map[uint32]bool)

		for _, track := range track_info {
			if track.Is_video_track {
				video_track_set[track.Track_header_info.Track_id] = true
			}
		}

		if len(video_track_set) == 0 {
			err = fmt.Errorf("Failed to parse the duration of the video, the ")
			return
		}

		var video_duration_map map[uint32]uint64 = make(map[uint32]uint64)
		err = moofutil.ParseDurationFromMOOF(file, video_track_set, video_duration_map)
		if err != nil {
			return
		}

		var video_practical_duration []float64 = make([]float64, len(video_track_set))

		var non_timescaled_duration uint64
		var ok bool
		for _, track := range track_info {
			non_timescaled_duration, ok = video_duration_map[track.Track_header_info.Track_id]
			if !ok {
				err = fmt.Errorf("Failed to obtain the duration of one or more video tracks, file could be damaged or malformed.")
				return
			}

			video_practical_duration = append(video_practical_duration, float64(non_timescaled_duration)/float64(track.MDIA_timescale))
		}

		//get the highest duration value
		duration = slices.Max(video_practical_duration)
		return

	} else if obtained_timescale == 0 {
		err = fmt.Errorf("Timescale was found to be 0, invalid value or malformed MP4")
		return
	}
	//else, the duration can be obtained from the file.

	duration = float64(obtained_duration) / float64(obtained_timescale)
	return
}

// Convert your obtained duration from the Get function
// to a suitable string in the format -> hour:minute(0 padded):second(0 padded)
func Stringify(duration float64) string {
	hour := uint64(duration) / 3600
	practical_duration := uint64(duration) % 3600
	min := practical_duration / 60
	//will contain the second of the duration after this operation
	practical_duration %= 60

	var hour_string string

	if hour > 0 {
		hour_string = fmt.Sprintf("%d:", hour)
	}

	return fmt.Sprintf("%s%02d:%02d", hour_string, min, practical_duration)
}
