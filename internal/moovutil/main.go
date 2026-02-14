package moovutil

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/0x02f0bcd4/gomp4dur/internal/atomutil"
)

type TrakLocInfo struct {
	TRAK_loc  int64
	TRAK_size int64
}

type MoovHeader struct {
	MOOV_loc int64
	//start location of trac atoms
	TRAC_locs []TrakLocInfo
}

// The function doesn't assume anything regarding the position
// of the file. It will start from the beginning, and it will
// read until an MOOV atom is found. If not found, it errors out
// with the position of the file being left most probably at the
// EOF. IF an error occurs even before that, the file position
// is left where the error occurred. The user shouldn't make any
// assumption as of where the file's position is located.
func ReadMVHDTimescaleAndDuration(file *os.File) (moov_info *MoovHeader, timescale uint32, duration uint64, err error) {
	_, err = file.Seek(0, io.SeekStart)

	if err != nil {
		return
	}

	var loc int64
	var name string
	var size uint64

	found_moov := false

	for {
		loc, name, size, err = atomutil.ReadAtomNameAndSize(file)

		if err != nil {
			return
		}

		//else, check if the name is MOOV

		if name == "moov" {
			found_moov = true
			break
		}

		loc += int64(size)

		_, err = file.Seek(loc, io.SeekStart)

		if err != nil {
			return
		}
	}

	if found_moov {
		MOOV_loc := loc
		moov_size := size
		moov_info = &MoovHeader{
			MOOV_loc:  MOOV_loc,
			TRAC_locs: make([]TrakLocInfo, 0, 2),
		}

		//try to find the MVHD atom inside the MOOV atom
		_, err = file.Seek(8, io.SeekCurrent)

		found_mvhd := false
		var mvhd_loc int64 = -1

		// moov_current_loc describes the current position where the reading of the MOOV atom
		// is being taken place.
		for moov_current_loc, moov_end := MOOV_loc+8, MOOV_loc+int64(moov_size); moov_current_loc < moov_end; {
			loc, name, size, err = atomutil.ReadAtomNameAndSize(file)
			if err != nil {
				return
			}

			switch name {

			case "mvhd":
				{
					found_mvhd = true
					mvhd_loc = loc
				}
			case "trak":
				{
					moov_info.TRAC_locs = append(moov_info.TRAC_locs, TrakLocInfo{
						TRAK_loc:  loc,
						TRAK_size: int64(size),
					})
				}
			}

			//else, jump to the next sibling header
			_, err = file.Seek(int64(size), io.SeekCurrent)

			if err != nil {
				return
			}

			moov_current_loc = loc + int64(size)
		}

		if found_mvhd {
			//read the MVHD content to get the timescale and duration
			//first, read the version and flags
			_, err = file.Seek(mvhd_loc+8, io.SeekStart)

			if err != nil {
				return
			}

			var vf []byte = make([]byte, 4)

			var total_read int
			total_read, err = file.Read(vf)

			if err != nil {
				return
			} else if total_read != 4 {
				err = fmt.Errorf("Failed to read the MVHD version and flags to it's entirety, total read: %d, expected: %d", total_read, 4)
			}

			is_64_bit := vf[0] == 0x01

			skip_dist_cmtime := 8

			if is_64_bit {
				skip_dist_cmtime += 8
			}

			tsd_length := 8
			if is_64_bit {
				tsd_length += 4 // the duration is now 8 bytes
			}

			_, err = file.Seek(int64(skip_dist_cmtime), io.SeekCurrent)

			if err != nil {
				return
			}

			var tsd []byte = make([]byte, tsd_length)

			total_read, err = file.Read(tsd)

			if err != nil {
				return
			} else if total_read != tsd_length {
				err = fmt.Errorf("Failed to read MVHD timescale and duration to its entirety, total read: %d, expected to read: %d\n", total_read, tsd_length)
				return
			}

			//else, set the timescale and duration

			timescale = binary.BigEndian.Uint32(tsd[:4])

			if is_64_bit {
				duration = binary.BigEndian.Uint64(tsd[4:])
			} else {
				duration = uint64(binary.BigEndian.Uint32(tsd[4:]))
			}

			return
		} else {
			err = fmt.Errorf("Failed to find a MVHD atom under the MOOV atom in the file, most probably a malformed MOOV atom.")
		}

	}

	err = fmt.Errorf("Failed to find the MOOV atom anywhere in the file, most probably a malformed file.")

	return
}
