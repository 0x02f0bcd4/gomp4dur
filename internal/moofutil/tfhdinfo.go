package moofutil

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type TFHDInfo struct {
	Is_On_Track                    bool
	Track_ID                       uint32
	Is_default_sample_duration_set bool
	Default_sample_duration        uint32
}

// This function assumes that you have identified TFHD atom correctly
func parseTFHDAtom(file *os.File, atom_loc int64, video_track_set map[uint32]bool) (info TFHDInfo, err error) {
	//Walk into the TRAF atom
	_, err = file.Seek(atom_loc+8, io.SeekStart)

	if err != nil {
		return
	}

	var _4_byte_content []byte = make([]byte, 4)

	//reading the version and flag
	total_read, err := file.Read(_4_byte_content)

	if err != nil {
		return
	} else if total_read != 4 {
		err = fmt.Errorf("Failed to read the TFHD atom's version-flag content to its fullest, total_read: %d bytes, expected: %d\n", total_read, 4)
		return
	}

	var flags uint32 = binary.BigEndian.Uint32(_4_byte_content) & 0x00_FF_FF_FF

	info.Is_default_sample_duration_set = (flags & 0x00_00_00_08) == 0x00_00_00_08

	//reading the Track_ID
	total_read, err = file.Read(_4_byte_content)
	if err != nil {
		return
	} else if total_read != 4 {
		err = fmt.Errorf("Failed to read the TFHD atom's Track ID content to its fullest, total_read: %d bytes, expected: %d\n", total_read, 4)
		return
	}

	info.Track_ID = binary.BigEndian.Uint32(_4_byte_content)
	info.Is_On_Track = video_track_set[info.Track_ID]

	if !info.Is_On_Track {
		return
	}

	if !info.Is_default_sample_duration_set {
		return
	}

	var skip_length int64 = 0

	//the data_offset flag was set
	if (flags & 0x00_00_00_01) == 0x00_00_00_01 {
		skip_length += 8
	}

	//the sample_desc_index flag was set
	if (flags & 0x00_00_00_02) == 0x00_00_00_02 {
		skip_length += 4
	}

	if skip_length > 0 {
		_, err = file.Seek(skip_length, io.SeekCurrent)
		if err != nil {
			return
		}
	}

	//read the 4 bytes of Default_Sample_Duration
	total_read, err = file.Read(_4_byte_content)
	if err != nil {
		return
	} else if total_read != 4 {
		err = fmt.Errorf("Failed to read the TFHD atom's Default Sample Duration content to its fullest, total_read: %d bytes, expected: %d\n", total_read, 4)
		return
	}

	info.Default_sample_duration = binary.BigEndian.Uint32(_4_byte_content)
	return
}
