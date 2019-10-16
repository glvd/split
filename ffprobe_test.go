package split

import (
	"testing"
)

// TestFormat_NameAnalyze ...
func TestFormat_NameAnalyze(t *testing.T) {
	sf1, e := FFProbeStreamFormat("d:\\video\\极乐女忍者.LADY.NINJA.2018.HD1080P.X264.AAC.Japanese.CHT.mp4")

	if e != nil {
		log.Error(e)
		return
	}
	t.Logf("%+v", sf1.Video())
}
