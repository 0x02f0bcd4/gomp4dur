package trakutil

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type TKHDInfo struct {
	Is_64_bit    bool
	Track_in_use bool
	Track_id     uint32
	Duration     uint64
	Width        uint32
	Height       uint32
}

func (info *TKHDInfo) PrintInfo() {
	fmt.Printf("Printing TKHD Information -\n")
	fmt.Printf("Is track in use? %t\n", info.Track_in_use)
	fmt.Printf("Is track's version 64 bit? %t\n", info.Is_64_bit)
	fmt.Printf("Track's duration: %d\n", info.Duration)
	fmt.Printf("Track's ID: %d\n", info.Track_id)
	fmt.Printf("Track's width: %d\n", info.Width)
	fmt.Printf("Track's height: %d\n", info.Height)
	fmt.Print("\n")
}

func ParseTKHDAtom(file *os.File, atom_loc int64, atom_size uint64) (info TKHDInfo, err error) {
	_, err = file.Seek(atom_loc+8, io.SeekStart)

	if err != nil {
		return
	}

	var _4_byte_content []byte = make([]byte, 4)

	total_read, err := file.Read(_4_byte_content)

	if err != nil {
		return
	} else if total_read != 4 {
		err = fmt.Errorf("Failed to read the version-flag of TKHD content to its fullest, expected %d, total_read: %d", 4, total_read)
		return
	}

	info.Is_64_bit = _4_byte_content[0] == 0x01

	_4_byte_content[0] = 0x00

	val := binary.BigEndian.Uint32(_4_byte_content)

	info.Track_in_use = ((val & 0x01) != 0) && ((val & 0x02) != 0)

	var skip_size int32 = 0

	//currently we are skipping the Creation and Modification Time
	//(CMTIME)

	skip_size = 8

	if info.Is_64_bit {
		skip_size += 8
	}

	_, err = file.Seek(int64(skip_size), io.SeekCurrent)

	if err != nil {
		return
	}

	total_read, err = file.Read(_4_byte_content)

	if err != nil {
		return
	} else if total_read != 4 {
		err = fmt.Errorf("Failed to read the track_id of TKHD content to its fullest, expected %d, total_read: %d", 4, total_read)
		return
	}

	info.Track_id = binary.BigEndian.Uint32(_4_byte_content)

	//get the duration of the track

	var _8_byte_content []byte = make([]byte, 8)

	_, err = file.Seek(4, io.SeekCurrent)

	if err != nil {
		return
	}

	if info.Is_64_bit {
		total_read, err = file.Read(_8_byte_content)

		if err != nil {
			return
		} else if total_read != 8 {
			err = fmt.Errorf("Failed to read the duration of TKHD content to its fullest, expected %d, total_read: %d", 8, total_read)
			return
		}

		info.Duration = binary.BigEndian.Uint64(_8_byte_content)
	} else {
		total_read, err = file.Read(_4_byte_content)

		if err != nil {
			return
		} else if total_read != 4 {
			err = fmt.Errorf("Failed to read the duration of TKHD content to its fullest, expected %d, total_read: %d", 4, total_read)
		}

		info.Duration = uint64(binary.BigEndian.Uint32(_4_byte_content))
	}

	//now, skip to the width and height

	// Reserved + Layer + Alt Group + Volume + Reserved + Marix
	skip_size = 8 + 2 + 2 + 2 + 2 + 36

	_, err = file.Seek(int64(skip_size), io.SeekCurrent)

	if err != nil {
		return
	}

	//size will contain either width (in first read) or height (in second read)
	var size []byte = make([]byte, 4)

	//reading width
	total_read, err = file.Read(size)

	if err != nil {
		return
	} else if total_read != 4 {
		err = fmt.Errorf("Failed to read the width of TKHD atom content to its fullest, expected: %d, total_read: %d", 4, total_read)
		return
	}

	info.Width = binary.BigEndian.Uint32(size)

	//reading height
	total_read, err = file.Read(size)

	if err != nil {
		return
	} else if total_read != 4 {
		err = fmt.Errorf("Failed to read the height of TKHD atom content to its fullest, expected: %d, total_read: %d", 4, total_read)
		return
	}

	info.Height = binary.BigEndian.Uint32(size)
	return
}
