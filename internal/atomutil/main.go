package atomutil

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// The function assumes we are at the start of the header of the atom
// Please note, after each reading, the location of the file will be
// restored to its previous position, giving you the peace of mind
// that whatever operation may have been done, the location of the
// file is still at the start of the atom, unless an error has occurred.
// In such case, there's no guarantee that the file location is at the
// start of the atom.
func ReadAtomNameAndSize(file *os.File) (loc int64, name string, size uint64, err error) {
	var content []byte = make([]byte, 8)

	total_read, err := file.Read(content)

	if err != nil {
		loc = -1
		name = ""
		size = 0
		return
	} else if total_read != len(content) {
		loc = -1
		name = ""
		size = 0
		err = fmt.Errorf("Failed to read the entirety of the content, total read: %d, expected to read: %d", total_read, len(content))
		return
	}

	name = string(content[4:])
	size = uint64(binary.BigEndian.Uint32(content[:4]))
	var reverse_length int64 = 8

	if size == 0 {
		// size of the atom reaches the end of the file
		var curr_pos int64
		curr_pos, err = file.Seek(8, io.SeekCurrent)
		if err != nil {
			loc = -1
			name = ""
			size = 0
			return
		}

		//Get the size of the file
		var file_stat os.FileInfo

		file_stat, err = file.Stat()
		if err != nil {
			loc = -1
			name = ""
			size = 0
			return
		}

		//size of the atom will be
		size = uint64(file_stat.Size() - curr_pos)
	} else if size == 1 {
		reverse_length += 8
		//there's an extended 64-bit original size of the atom after that
		total_read, err = file.Read(content)
		if err != nil {
			loc = -1
			name = ""
			size = 0
			return
		}
		size = binary.BigEndian.Uint64(content)
	}

	loc, err = file.Seek(0, io.SeekCurrent)
	if err != nil {
		loc = -1
		name = ""
		size = 0
		return
	}

	//we return to the start of the atom
	loc -= reverse_length

	_, err = file.Seek(loc, io.SeekStart)

	if err != nil {
		loc = -1
		name = ""
		size = 0
		return
	}

	return
}
