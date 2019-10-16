package split

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

// TestFFProbeStreamFormat ...
func TestFFProbeStreamFormat(t *testing.T) {
	format, _ := FFProbeStreamFormat("D:\\video\\周杰伦唱歌贼难听.mp4")
	v, _ := json.Marshal(format)
	ioutil.WriteFile("d:\\test.json", v, os.ModePerm)
	t.Log(string(v), format.Resolution())
	format1, _ := FFProbeStreamFormat("D:\\video\\[BT天堂btbttt.com]我的女友.My.Girlfriend.2018.HD720P.X264.AAC.Korean.中文字幕.mp4")
	v1, _ := json.Marshal(format1)
	ioutil.WriteFile("d:\\test1.json", v, os.ModePerm)
	t.Log(string(v1), format1.Resolution())
}
