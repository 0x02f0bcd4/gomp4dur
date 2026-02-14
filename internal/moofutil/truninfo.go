package moofutil

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// this function assumes that the TRUN atom has been identified
func parseTRUNAtom(file *os.File, atom_loc int64, check_sample_duration bool, default_sample_duration uint32) (duration uint64, err error) {
	//Walk into the TRUN atom
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
		err = fmt.Errorf("Failed to read the TRUN atom's version-flag content to its fullest, total_read: %d bytes, expected: %d", total_read, 4)
		return
	}

	//TODO: Re-add to check for the flags if check_sample_duration is true
	var flags uint32 = binary.BigEndian.Uint32(_4_byte_content)
	flags &= 0x00_FF_FF_FF

	//reading the sample_count
	total_read, err = file.Read(_4_byte_content)
	if err != nil {
		return
	} else if total_read != 4 {
		err = fmt.Errorf("Failed to read the TRUN atom's sample_count content to its fullest, total_read: %d bytes, expected: %d", total_read, 4)
		return
	}

	sample_count := binary.BigEndian.Uint32(_4_byte_content)

	if !check_sample_duration {
		if (flags & 0x00_000100) == 0x00_000100 {
			err = fmt.Errorf("Stop the cap - you said not to check the sample duration, but the sample_durations are provided for the track. File is malformed.")
			return
		}
		duration = uint64(sample_count) * uint64(default_sample_duration)
		return
	} else if (flags & 0x00_000100) != 0x000100 {
		err = fmt.Errorf("Stop the cap - you asked to check the sample duration, but the sample_durations aren't provided for the track. File is malformed.")
		return
	}

	fmt.Println("Check Sample Duration is set to true, currently return 0, later we will read the rows of sample_duration and return its value")

	//else, read the sample durations, and add them to get the final duration
	var skip_length int64 = 0

	//Data offset flag is set
	if (flags & 0x00_00_00_01) == 0x00_00_00_01 {
		skip_length += 4
	}

	//First Sample Flag is set
	if (flags & 0x00_00_00_04) == 0x00_00_00_04 {
		skip_length += 4
	}

	if skip_length > 0 {
		_, err = file.Seek(skip_length, io.SeekCurrent)
		if err != nil {
			return
		}
	}

	//duration skip length
	skip_length = 0
	//skip sample_size
	if (flags & 0x00_000200) == 0x00_000200 {
		skip_length += 4
	}

	if (flags & 0x00_000400) == 0x00_000400 {
		skip_length += 4
	}

	if (flags & 0x00_000800) == 0x00_000800 {
		skip_length += 4
	}

	var sample_duration uint32
	for i := sample_count; i > 0; i-- {
		total_read, err = file.Read(_4_byte_content)
		if err != nil {
			return
		} else if total_read != 4 {
			err = fmt.Errorf("Failed to read TRUN atom's sample_duration from the SAMPLE_TABLE at the row: %d(0 based index), expected to read: %d bytes, total_read: %d bytes\n", sample_count-i, 4, total_read)
			return
		}
		sample_duration = binary.BigEndian.Uint32(_4_byte_content)
		duration += uint64(sample_duration)
		//jump to the next sample_duration
		if skip_length > 0 {
			_, err = file.Seek(skip_length, io.SeekCurrent)
			if err != nil {
				return
			}
		}
	}

	return
}
