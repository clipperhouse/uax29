package segmenter

import (
	"bufio"
)

// Segmenter is an iterator for byte arrays. See the New() and Next() funcs.
type Segmenter struct {
	split bufio.SplitFunc
	data  []byte
	token []byte
	pos   int
	err   error
}

// New creates a new segmenter given a SplitFunc. To use the new segmenter,
// call SetText() and then iterate while Next() is true
func New(split bufio.SplitFunc) *Segmenter {
	return &Segmenter{
		split: split,
	}
}

// SetText sets the text for the segmenter to operate on, and resets
// all state
func (sc *Segmenter) SetText(data []byte) {
	sc.data = data
	sc.token = nil
	sc.pos = 0
	sc.err = nil
}

// Next advances the Segmenter to the next segment. It returns false when there
// are no remaining segments, or an error occurred.
//	text := []byte("This is an example.")
//
//	seg := segmenter.New(splitFunc)
//  seg.SetText(text)
//	for segment.Next() {
//		fmt.Printf("%q\n", seg.Bytes())
//	}
//	if err := seg.Err(); err != nil {
//		log.Fatal(err)
//	}
func (seg *Segmenter) Next() bool {
	advance, token, err := seg.split(seg.data[seg.pos:], true)

	seg.pos += advance
	seg.token = token
	seg.err = err

	if advance == 0 {
		return false
	}

	if len(token) == 0 {
		return false
	}

	if seg.err != nil {
		return false
	}

	return true
}

// Err indicates an error occured when calling Next() or Previous(). Next and
// Previous will return false when an error occurs.
func (seg *Segmenter) Err() error {
	return seg.err
}

// Bytes returns the current segment
func (seg *Segmenter) Bytes() []byte {
	return seg.token
}

func All(data []byte, split bufio.SplitFunc) [][]byte {
	var result [][]byte
	var pos int

	for {
		advance, token, err := split(data[pos:], true)

		if advance == 0 {
			break
		}

		pos += advance

		if len(token) == 0 {
			break
		}

		if err != nil {
			break
		}

		result = append(result, token)
	}

	return result
}
