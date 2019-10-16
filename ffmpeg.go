package split

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

//const sliceM3u8FFmpegTemplate = `-y -i %s -strict -2 -ss %s -to %s -c:v %s -c:a %s -bsf:v h264_mp4toannexb -vsync 0 -f hls -hls_list_size 0 -hls_time %d -hls_segment_filename %s %s`
const sliceM3u8FFmpegTemplate = `-y -i %s -strict -2 -c:v %s -c:a %s -bsf:v h264_mp4toannexb -f hls -hls_list_size 0 -hls_time %d -hls_segment_filename %s %s`
const sliceM3u8ScaleTemplate = `-y -i %s -strict -2 -c:v %s -c:a %s -bsf:v h264_mp4toannexb %s -f hls -hls_list_size 0 -hls_time %d -hls_segment_filename %s %s`
const scaleOutputTemplate = "-vf scale=-2:%d"
const bitRateOutputTemplate = "-b:v %dK"
const frameRateOutputTemplate = "-r %3.2f"

type Argument struct {
	StreamFormat    *StreamFormat
	Auto            bool
	Scale           int64
	Start           string
	End             string
	Output          string
	Video           string
	Audio           string
	M3U8            string
	SegmentFileName string
	HLSTime         int
	probe           func(string) (*StreamFormat, error)
	BitRate         int64
	FrameRate       float64
	ShowLog         bool
}

// FFmpegContext ...
type ffmpegContext struct {
	once   sync.Once
	mu     sync.RWMutex
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	done   chan bool
}

// Context ...
func (c *ffmpegContext) Context() context.Context {
	return c.ctx
}

// Add ...
func (c *ffmpegContext) Add(i int) {
	c.wg.Add(i)
}

// Wait ...
func (c *ffmpegContext) Wait() {
	select {
	case <-c.Waiting():
		return
	}
}

// Waiting ...
func (c *ffmpegContext) Waiting() <-chan bool {
	c.once.Do(func() {
		go func() {
			c.wg.Wait()
			c.done <- true
		}()
	})

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.done == nil {
		c.done = make(chan bool)
	}
	d := c.done
	return d
}

// Done ...
func (c *ffmpegContext) Done() {
	c.wg.Done()
}

// Cancel ...
func (c *ffmpegContext) Cancel() {
	if c.cancel != nil {
		c.cancel()
	}
}

// Context ...
type Context interface {
	Cancel()
	Add(int)
	Waiting() <-chan bool
	Wait()
	Done()
	Context() context.Context
}

// FFmpegContext ...
func FFmpegContext() Context {
	ctx, cancel := context.WithCancel(context.Background())
	return &ffmpegContext{
		wg:     &sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
	}
}

// ArgumentOptions ...
type ArgumentOptions func(args *Argument)

// HLSTimeOption ...
func HLSTimeOption(i int) ArgumentOptions {
	return func(args *Argument) {
		args.HLSTime = i
	}
}

func ShowLogOption(b bool) ArgumentOptions {
	return func(args *Argument) {
		args.ShowLog = b
	}
}

// ScaleOption ...
func ScaleOption(s int64, v ...string) ArgumentOptions {
	return func(args *Argument) {
		args.Video = "libx264"
		for _, value := range v {
			args.Video = value
		}
		args.Scale = s
	}
}

// OutputOption ...
func OutputOption(s string) ArgumentOptions {
	return func(args *Argument) {
		args.Output = s
	}
}

// AutoOption ...
func AutoOption(s bool) ArgumentOptions {
	return func(args *Argument) {
		args.Auto = s
	}
}

// VideoOption ...
func VideoOption(s string) ArgumentOptions {
	return func(args *Argument) {
		args.Video = s
	}
}

// AudioOption ...
func AudioOption(s string) ArgumentOptions {
	return func(args *Argument) {
		args.Video = s
	}
}

// StreamFormatOption ...
func StreamFormatOption(s *StreamFormat) ArgumentOptions {
	return func(args *Argument) {
		args.StreamFormat = s
	}
}

// BitRateOption ...
func BitRateOption(b int64) ArgumentOptions {
	return func(args *Argument) {
		args.BitRate = b
	}
}

// ProbeInfoOption ...
func ProbeInfoOption(f func(string) (*StreamFormat, error)) ArgumentOptions {
	return func(args *Argument) {
		args.probe = f
	}
}

// FFMpegSplitToM3U8WithProbe ...
func FFMpegSplitToM3U8WithProbe(ctx context.Context, file string, args ...ArgumentOptions) (sa *Argument, e error) {
	args = append(args, ProbeInfoOption(FFProbeStreamFormat))
	return FFMpegSplitToM3U8(ctx, file, args...)
}

// Scale ...
const (
	Scale480P  = 0
	Scale720P  = 1
	Scale1080P = 2
)

var bitRateList = []int64{
	//Scale480P:  1000 * 1024,
	//Scale720P:  2000 * 1024,
	//Scale1080P: 4000 * 1024,
	Scale480P:  500 * 1024,
	Scale720P:  1000 * 1024,
	Scale1080P: 2000 * 1024,
}

var frameRateList = []float64{
	Scale480P:  float64(24000)/1001 - 0.005,
	Scale720P:  float64(24000)/1001 - 0.005,
	Scale1080P: float64(30000)/1001 - 0.005,
}

func scaleIndex(scale int64) int {
	if scale == 480 {
		return Scale480P
	} else if scale == 1080 {
		return Scale1080P
	}
	return Scale720P
}

func outputScale(sa *Argument) string {
	outputs := []string{fmt.Sprintf(scaleOutputTemplate, sa.Scale)}

	if sa.BitRate != 0 {
		outputs = append(outputs, fmt.Sprintf(bitRateOutputTemplate, sa.BitRate/1024))
	}
	if sa.FrameRate > 0 {
		outputs = append(outputs, fmt.Sprintf(frameRateOutputTemplate, sa.FrameRate))
	}
	log.With("args", strings.Join(outputs, " ")).Info("info")
	return strings.Join(outputs, " ")
}

// optimizeScale ...
func optimizeScale(sa *Argument, video *Stream) {
	if sa.Scale != 0 {
		if video.Height != nil && *video.Height < sa.Scale {
			//pass when video is smaller then input
			sa.Scale = 0
			return
		}

		idx := scaleIndex(sa.Scale)
		i, e := strconv.ParseInt(video.BitRate, 10, 64)
		if e != nil {
			log.Error(e)
			i = math.MaxInt64
		}

		if sa.BitRate == 0 {
			sa.BitRate = bitRateList[idx]
		}
		if sa.BitRate != 0 {
			if sa.BitRate > i {
				sa.BitRate = 0
			}
		}
		fr := strings.Split(video.RFrameRate, "/")
		il := 1
		ir := 1
		if len(fr) == 2 {
			il, e = strconv.Atoi(fr[0])
			if e != nil {
				il = 1
				log.Error(e)
			}
			ir, e = strconv.Atoi(fr[1])
			if e != nil {
				ir = 1
				log.Error(e)
			}
		}
		if sa.FrameRate == 0 {
			sa.FrameRate = frameRateList[idx]
		}
		log.With("framerate", sa.FrameRate, "rateLeft", il, "rateRight", ir, "left/right", il/ir).Info("info")
		if sa.FrameRate > 0 {
			if sa.FrameRate > float64(il)/float64(ir) {
				sa.FrameRate = 0
			}
		}
	}
}

// FFMpegSplitToM3U8WithOptimize ...
func FFMpegSplitToM3U8WithOptimize(ctx context.Context, file string, args ...ArgumentOptions) (sa *Argument, e error) {
	args = append(args, ProbeInfoOption(FFProbeStreamFormat))
	return FFMpegSplitToM3U8(ctx, file, args...)
}

// FFMpegSplitToM3U8 ...
func FFMpegSplitToM3U8(ctx context.Context, file string, args ...ArgumentOptions) (sa *Argument, e error) {
	if strings.Index(file, " ") != -1 {
		return nil, errors.New("file name cannot have spaces")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	sa = &Argument{
		StreamFormat:    nil,
		Auto:            true,
		Scale:           0,
		Start:           "",
		End:             "",
		Output:          "",
		Video:           "libx264",
		Audio:           "aac",
		M3U8:            "media.m3u8",
		SegmentFileName: "media-%05d.ts",
		HLSTime:         10,
		probe:           nil,
		BitRate:         0,
		FrameRate:       0,
		ShowLog:         true,
	}
	for _, o := range args {
		o(sa)
	}

	if sa.probe != nil {
		sa.StreamFormat, e = sa.probe(file)
		if e != nil {
			return nil, fmt.Errorf("%w", e)
		}
	}
	if sa.StreamFormat != nil {
		video := sa.StreamFormat.Video()
		audio := sa.StreamFormat.Audio()
		if !sa.StreamFormat.IsVideo() || audio == nil || video == nil {
			return nil, errors.New("open file failed with ffprobe")
		}

		//check scale before codec check
		optimizeScale(sa, video)

		if video.CodecName == "h264" && sa.Scale == 0 {
			sa.Video = "copy"
		}

		if audio.CodecName == "aac" {
			sa.Audio = "copy"
		}
	}

	sa.Output, e = filepath.Abs(sa.Output)
	if e != nil {
		return nil, fmt.Errorf("%w", e)
	}
	log.With("path", sa.Output).Info("info")
	if sa.Auto {
		sa.Output = filepath.Join(sa.Output, uuid.New().String())
		_ = os.MkdirAll(sa.Output, os.ModePerm)
	}

	sfn := filepath.Join(sa.Output, sa.SegmentFileName)
	m3u8 := filepath.Join(sa.Output, sa.M3U8)

	tpl := fmt.Sprintf(sliceM3u8FFmpegTemplate, file, sa.Video, sa.Audio, sa.HLSTime, sfn, m3u8)
	if sa.Scale != 0 {
		tpl = fmt.Sprintf(sliceM3u8ScaleTemplate, file, sa.Video, sa.Audio, outputScale(sa), sa.HLSTime, sfn, m3u8)
	}

	var runLog chan []byte
	if sa.ShowLog {
		runLog = make(chan []byte)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := FFMpegRun(ctx, runLog, tpl); err != nil {
			log.With("error", err).Info("end")
			e = err
		}
	}()
	for msg := range runLog {
		m := string(msg)
		if m != "" {
			log.With("message", m).Info("process")
		}
	}

	wg.Wait()
	return sa, e
}

// FFMpegRun ...
func FFMpegRun(ctx context.Context, output chan<- []byte, args string) (e error) {
	f := NewFFMpeg()
	f.SetArgs(args)
	if err := f.RunContext(ctx, output); err != nil {

		return err
	}
	return nil
}
