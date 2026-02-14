package trakutil

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/0x02f0bcd4/gomp4dur/internal/atomutil"
)

// This function assumes that the atom_loc is at the start of the
// MDHD atom.
func ParseTimescaleFromMDHD(file *os.File, atom_loc int64, atom_size uint64) (timescale uint32, err error) {
	_, err = file.Seek(atom_loc+8, io.SeekStart)

	if err != nil {
		return
	}

	var _4_byte_content []byte = make([]byte, 4)

	total_read, err := file.Read(_4_byte_content)

	if err != nil {
		return
	} else if total_read != 4 {
		err = fmt.Errorf("Failed to read the MDHD version-flag content to its fullest, total_read: %d, expected: %d", total_read, 4)
		return
	}

	is_64_bit := _4_byte_content[0] == 0x01

	skip_cmtime := 8
	if is_64_bit {
		skip_cmtime += 8
	}

	_, err = file.Seek(int64(skip_cmtime), io.SeekCurrent)

	if err != nil {
		return
	}

	total_read, err = file.Read(_4_byte_content)

	if err != nil {
		return
	} else if total_read != 4 {
		err = fmt.Errorf("Failed to read the MDHD timescale content to its fullest, total_read: %d, expected: %d", total_read, 4)
		return
	}

	timescale = binary.BigEndian.Uint32(_4_byte_content)
	return
}

// This function assumes that HDLR atom type has been identified correctly and
// the file is at the start of the atom. Use ReadAtomNameAndSize to learn about
// the atom and its size.
func ParseTrackTypeFromHDLR(file *os.File, atom_loc int64, atom_size uint64) (ttype string, err error) {
	// Size + Name + Version & Flags ( 1 + 3 ) + Pre-Defined
	_, err = file.Seek(atom_loc+4+4+4+4, io.SeekStart)

	if err != nil {
		return
	}

	var type_byte []byte = make([]byte, 4)
	total_read, err := file.Read(type_byte)

	if err != nil {
		return
	} else if total_read != 4 {
		err = fmt.Errorf("Failed to read the HDLR type content to its fullest, total_read:%d, expected: %d", total_read, 4)
		return
	}

	ttype = string(type_byte)
	return
}

// this function assumes that you have successfully identified the atom
// to be MDIA. To learn how to identify the atom, check out ReadAtomNameAndSize()
func ParseMDIAAtom(file *os.File, atom_loc int64, atom_size uint64) (timescale uint32, ttype string, err error) {
	//Walk into the MDIA atom
	_, err = file.Seek(atom_loc+8, io.SeekStart)

	if err != nil {
		return
	}

	//there are many subatoms under the MDIA atom
	//what we are interested in is MDHD, which contains
	//the Timescale value.

	var atom_name string

	var mdia_atom_loc = atom_loc
	var mdia_atom_size = atom_size

	var ts_type_flag byte = 0x0
	for mdia_curr_loc, mdia_end := mdia_atom_loc+8, mdia_atom_loc+int64(mdia_atom_size); mdia_curr_loc < mdia_end; {
		atom_loc, atom_name, atom_size, err = atomutil.ReadAtomNameAndSize(file)

		if err != nil {
			return
		}

		switch atom_name {
		case "mdhd":
			{
				//that's the one we are interested in
				timescale, err = ParseTimescaleFromMDHD(file, atom_loc, atom_size)
				ts_type_flag |= 0x0F
			}
		case "hdlr":
			{
				ttype, err = ParseTrackTypeFromHDLR(file, atom_loc, atom_size)
				ts_type_flag |= 0xF0
			}
		}

		if (ts_type_flag & 0xFF) == 0xFF {
			return
		}

		mdia_curr_loc += int64(atom_size)
		_, err = file.Seek(mdia_curr_loc, io.SeekStart)
		if err != nil {
			return
		}
	}

	//failed to find the MVHD atom
	err = fmt.Errorf("Failed to find the MDHD subatom inside the MDIA atom")

	return
}
