package split

import (
	"context"
	"testing"
)

// TestFFMpegRun ...
func TestFFMpegRun(t *testing.T) {

	ctx := context.Background()
	//ctx.Add(1)

	func() {
		//f, e := FFMpegSplitToM3U8WithProbe(ctx, "D:\\workspace\\goproject\\go-ffmpeg-cmd\\周杰伦唱歌贼难听.2019.1080P.h264.aac.Japanese.None..mp4", ScaleOption(720), OutputOption("tmp"))
		f, e := FFMpegSplitToM3U8WithProbe(ctx, "D:\\video\\QmQQbAKeLpLL5cgCtCDgwdoGCssrHHssyz4echBAi57us7.wmv", ScaleOption(720), OutputOption("tmp"))
		log.Info(f)
		log.Error(e)
	}()
	//ctx.Add(1)
	//go FFMpegSplitToM3U8(ctx, "D:\\video\\极乐女忍者.LADY.NINJA.2018.HD1080P.X264.AAC.Japanese.CHT.mp4", OutputOption("tmp"))
	//ctx.Add(1)
	//go FFMpegSplitToM3U8(ctx, "D:\\video\\极乐女忍者.LADY.NINJA.2018.HD1080P.X264.AAC.Japanese.CHT.mp4", OutputOption("tmp"))
	//ctx.Add(1)
	//go FFMpegSplitToM3U8(ctx, "D:\\video\\极乐女忍者.LADY.NINJA.2018.HD1080P.X264.AAC.Japanese.CHT.mp4", OutputOption("tmp"))
	//ctx.Add(1)
	//go FFMpegSplitToM3U8(ctx, "D:\\video\\极乐女忍者.LADY.NINJA.2018.HD1080P.X264.AAC.Japanese.CHT.mp4", OutputOption("tmp"))
	//for {

	//}

}
