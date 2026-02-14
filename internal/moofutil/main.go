package moofutil

import (
	"fmt"
	"io"
	"os"

	"github.com/0x02f0bcd4/gomp4dur/internal/atomutil"
)

// This function assumes that you have found the TRAF atom.
func parseTheTRAF(file *os.File, atom_loc int64, atom_size uint64, video_track_set map[uint32]bool, video_duration_map map[uint32]uint64) (err error) {
	//Walk into the TRAF atom
	_, err = file.Seek(atom_loc+8, io.SeekStart)

	if err != nil {
		return
	}

	var atom_name string
	var tfhd_info TFHDInfo
	checked_tfhd := false
	checked_trun := false
	var duration uint64

	for traf_curr_loc, traf_end_loc := atom_loc+8, atom_loc+int64(atom_size); traf_curr_loc < traf_end_loc; {
		atom_loc, atom_name, atom_size, err = atomutil.ReadAtomNameAndSize(file)

		if err != nil {
			return
		}

		switch atom_name {

		case "tfhd":
			{
				tfhd_info, err = parseTFHDAtom(file, atom_loc, video_track_set)

				if err != nil {
					return
				}

				checked_tfhd = true
			}
		case "trun":
			{
				if !checked_tfhd {
					err = fmt.Errorf("Failed to read the MOOF atom, expected the TFHD atom to be before TRUN atom. The File could be malformed")
					return
				}

				if tfhd_info.Is_On_Track {
					duration, err = parseTRUNAtom(file, atom_loc, !tfhd_info.Is_default_sample_duration_set, tfhd_info.Default_sample_duration)
				}
				if err != nil {
					return
				}
				checked_trun = true
			}
		}

		traf_curr_loc += int64(atom_size)
		_, err = file.Seek(traf_curr_loc, io.SeekStart)
		if err != nil {
			return
		}
	}

	if !(checked_tfhd && checked_trun) {
		err = fmt.Errorf("Failed to get both TFHD and TRUN atom, search result: TFHD - %t, TRUN - %t\n", checked_tfhd, checked_trun)
		return
	}

	video_duration_map[tfhd_info.Track_ID] += duration
	return
}

// This function assumes that you have found the MOOF atom.
func findAndParseTheTRAF(file *os.File, atom_loc int64, atom_size uint64, video_track_set map[uint32]bool, video_duration_map map[uint32]uint64) (err error) {
	//Walk into the MOOF atom
	_, err = file.Seek(atom_loc+8, io.SeekStart)

	if err != nil {
		return
	}

	var atom_name string
	var found_traf = false
	for moof_curr_loc, moof_end_loc := atom_loc+8, atom_loc+int64(atom_size); moof_curr_loc < moof_end_loc; {
		atom_loc, atom_name, atom_size, err = atomutil.ReadAtomNameAndSize(file)

		if err != nil {
			return
		}

		if atom_name == "traf" {
			err = parseTheTRAF(file, atom_loc, atom_size, video_track_set, video_duration_map)
			if err != nil {
				return
			}

			found_traf = true
			break
		}

		moof_curr_loc += int64(atom_size)

		_, err = file.Seek(moof_curr_loc, io.SeekStart)
		if err != nil {
			return
		}
	}

	if !found_traf {
		err = fmt.Errorf("Failed to find the TRAF atom inside the MOOF atom. File could be malformed.")
	}

	return
}

// The function doesn't assume anything regarding
// about the position of the file. Neither should
// the user assume anything about the file position
// after the operation is completed.
func ParseDurationFromMOOF(file *os.File, video_track_set map[uint32]bool, video_duration_map map[uint32]uint64) (err error) {
	var atom_loc int64
	var atom_size uint64
	var atom_name string

	file_size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return
	}

	for atom_loc < file_size {
		atom_loc, atom_name, atom_size, err = atomutil.ReadAtomNameAndSize(file)
		if err != nil {
			return
		}

		if atom_name == "moof" {
			err = findAndParseTheTRAF(file, atom_loc, atom_size, video_track_set, video_duration_map)
			if err != nil {
				return
			}
		}

		atom_loc += int64(atom_size)
		_, err = file.Seek(atom_loc, io.SeekStart)
		if err != nil {
			return
		}
	}

	return
}
