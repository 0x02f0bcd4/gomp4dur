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

	loc, err = file.Seek(0, io.SeekCurrent)
	if err != nil {
		loc = -1
		name = ""
		size = 0
		return
	}

	//we return to the start of the atom
	loc -= 8

	_, err = file.Seek(loc, io.SeekStart)

	if err != nil {
		loc = -1
		name = ""
		size = 0
		return
	}

	return
}
