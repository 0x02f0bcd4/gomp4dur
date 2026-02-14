package trakutil

import (
	"io"
	"os"

	"github.com/0x02f0bcd4/gomp4dur/internal/atomutil"
	"github.com/0x02f0bcd4/gomp4dur/internal/moovutil"
)

type TrakInfo struct {
	Track_header_info TKHDInfo
	MDIA_timescale    uint32
	Is_video_track    bool
}

func GetVideoTrackInfo(file *os.File, tracks []moovutil.TrakLocInfo) ([]TrakInfo, error) {
	var err error
	var atom_loc int64
	var atom_name string
	var atom_size uint64

	var track_info []TrakInfo = make([]TrakInfo, 0, len(tracks))
	var timescale uint32 = 0
	var header_info TKHDInfo
	var info_timescale_finder_flag byte = 0
	var ttype string

	for _, loc_info := range tracks {
		//walk into the atom and read the content
		_, err = file.Seek(loc_info.TRAK_loc+8, io.SeekStart)
		if err != nil {
			return nil, err
		}
		// each track atom has a TKHD atom and a MDIA atom
		// the TKHD atom has the track_id, width and height
		// and the MDIA has MDHD which contains the timescale
		info_timescale_finder_flag = 0x0

		for trak_curr_loc, trak_end_loc := loc_info.TRAK_loc+8, loc_info.TRAK_loc+loc_info.TRAK_size; trak_curr_loc < trak_end_loc; {
			atom_loc, atom_name, atom_size, err = atomutil.ReadAtomNameAndSize(file)

			if err != nil {
				return nil, err
			}

			switch atom_name {
			case "tkhd":
				{
					//track header atom found, parse it.
					header_info, err = ParseTKHDAtom(file, atom_loc, atom_size)

					if err != nil {
						return nil, err
					}

					info_timescale_finder_flag |= 0x0F
				}
			case "mdia":
				{
					timescale, ttype, err = ParseMDIAAtom(file, atom_loc, atom_size)
					if err != nil {
						return nil, err
					}

					info_timescale_finder_flag |= 0xF0
				}
			}

			trak_curr_loc += int64(atom_size)
			_, err = file.Seek(trak_curr_loc, io.SeekStart)
			if err != nil {
				return nil, err
			}

			//meaning, we found both of our flags
			if (info_timescale_finder_flag & 0xFF) == 0xFF {
				//good shit
				track_info = append(track_info, TrakInfo{
					Track_header_info: header_info,
					MDIA_timescale:    timescale,
					Is_video_track:    ttype == "vide",
				})
				break
			}
		}
	}
	return track_info, nil
}
