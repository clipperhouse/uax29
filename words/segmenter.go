package words

import "github.com/clipperhouse/uax29/segmenter"

func NewSegmenter(data []byte) *segmenter.Segmenter {
	seg := segmenter.New(SplitFunc)
	seg.SetText(data)
	return seg
}
