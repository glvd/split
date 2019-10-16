package split

import (
	"encoding/json"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ExtIdx ...
const (
	CNameIdx = iota
	ExtIdx
	CaptionIdx
	LanguageIdx
	AudioIdx
	VideoIdx
	SharpnessIdx
	DataIdx
	ENameIdx
	MaxSizeIdx
)

// StreamFormat ...
type StreamFormat struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

// Format ...
type Format struct {
	Filename       string     `json:"filename"`
	NbStreams      int64      `json:"nb_streams"`
	NbPrograms     int64      `json:"nb_programs"`
	FormatName     string     `json:"format_name"`
	FormatLongName string     `json:"format_long_name"`
	StartTime      string     `json:"start_time"`
	Duration       string     `json:"duration"`
	Size           string     `json:"size"`
	BitRate        string     `json:"bit_rate"`
	ProbeScore     int64      `json:"probe_score"`
	Tags           FormatTags `json:"tags"`
}

// FormatTags ...
type FormatTags struct {
	MajorBrand       string `json:"major_brand"`
	MinorVersion     string `json:"minor_version"`
	CompatibleBrands string `json:"compatible_brands"`
	Encoder          string `json:"encoder"`
}

// Stream ...
type Stream struct {
	Index              int64            `json:"index"`
	CodecName          string           `json:"codec_name"`
	CodecLongName      string           `json:"codec_long_name"`
	Profile            string           `json:"profile"`
	CodecType          string           `json:"codec_type"`
	CodecTimeBase      string           `json:"codec_time_base"`
	CodecTagString     string           `json:"codec_tag_string"`
	CodecTag           string           `json:"codec_tag"`
	Width              *int64           `json:"width,omitempty"`
	Height             *int64           `json:"height,omitempty"`
	CodedWidth         *int64           `json:"coded_width,omitempty"`
	CodedHeight        *int64           `json:"coded_height,omitempty"`
	HasBFrames         *int64           `json:"has_b_frames,omitempty"`
	SampleAspectRatio  *string          `json:"sample_aspect_ratio,omitempty"`
	DisplayAspectRatio *string          `json:"display_aspect_ratio,omitempty"`
	PixFmt             *string          `json:"pix_fmt,omitempty"`
	Level              *int64           `json:"level,omitempty"`
	ColorRange         *string          `json:"color_range,omitempty"`
	ColorSpace         *string          `json:"color_space,omitempty"`
	ColorTransfer      *string          `json:"color_transfer,omitempty"`
	ColorPrimaries     *string          `json:"color_primaries,omitempty"`
	ChromaLocation     *string          `json:"chroma_location,omitempty"`
	Refs               *int64           `json:"refs,omitempty"`
	IsAVC              *string          `json:"is_avc,omitempty"`
	NalLengthSize      *string          `json:"nal_length_size,omitempty"`
	RFrameRate         string           `json:"r_frame_rate"`
	AvgFrameRate       string           `json:"avg_frame_rate"`
	TimeBase           string           `json:"time_base"`
	StartPts           int64            `json:"start_pts"`
	StartTime          string           `json:"start_time"`
	DurationTs         int64            `json:"duration_ts"`
	Duration           string           `json:"duration"`
	BitRate            string           `json:"bit_rate"`
	BitsPerRawSample   *string          `json:"bits_per_raw_sample,omitempty"`
	NbFrames           string           `json:"nb_frames"`
	Disposition        map[string]int64 `json:"disposition"`
	Tags               StreamTags       `json:"tags"`
	SampleFmt          *string          `json:"sample_fmt,omitempty"`
	SampleRate         *string          `json:"sample_rate,omitempty"`
	Channels           *int64           `json:"channels,omitempty"`
	ChannelLayout      *string          `json:"channel_layout,omitempty"`
	BitsPerSample      *int64           `json:"bits_per_sample,omitempty"`
	MaxBitRate         *string          `json:"max_bit_rate,omitempty"`
}

// StreamTags ...
type StreamTags struct {
	Language    string `json:"language"`
	HandlerName string `json:"handler_name"`
}

//var resolution = []int{120, 144, 160, 200, 240, 320, 360, 480, 540, 576, 600, 640, 720, 768, 800, 864, 900, 960, 1024, 1050, 1080, 1152, 1200, 1280, 1440, 1536, 1600, 1620, 1800, 1824, 1920, 2048, 2160, 2400, 2560, 2880, 3072, 3200, 4096, 4320, 4800}
var resolution = []int{240, 360, 480, 720, 1080, 1920, 2560, 4096, 4800}

func getResolutionIndex(n int64, sta, end int) int {
	//log.Infof("%d,%d,%d", n, sta, end)
	//if int64(resolution[sta]) == n {
	//	return sta
	//}
	if end == -1 {
		end = len(resolution)
	}

	if idx := (sta + end) / 2; idx > sta {
		if int64(resolution[idx]) > n {
			return getResolutionIndex(n, sta, idx)
		}
		return getResolutionIndex(n, idx, end)
	}
	if int64(resolution[sta]) != n && sta < len(resolution)-1 {
		return sta + 1
	}
	return sta
}

// FFProbeStreamFormat ...
func FFProbeStreamFormat(filename string) (*StreamFormat, error) {
	probe := NewFFProbe()
	probe.SetArgs("-v quiet -print_format json -show_format -show_streams")
	probe.AddArgs(filename)
	s, e := probe.Run()
	if e != nil {
		return nil, e
	}
	sf := StreamFormat{}
	e = json.Unmarshal([]byte(s), &sf)
	if e != nil {
		return nil, e
	}
	return &sf, nil
}

// ResolutionInt ...
func (f *StreamFormat) ResolutionInt() int {
	idx := 0
	for _, s := range f.Streams {
		if s.CodecType == "video" {
			if s.Height != nil {
				idx = getResolutionIndex(*s.Height, 0, -1)
				break
			}

		}
	}
	return resolution[idx]
}

// Resolution ...
func (f *StreamFormat) Resolution() string {
	idx := 0
	for _, s := range f.Streams {
		if s.CodecType == "video" {
			if s.Height != nil {
				idx = getResolutionIndex(*s.Height, 0, -1)
				break
			}

		}
	}
	return strconv.FormatInt(int64(resolution[idx]), 10)
}

// Video ...
func (f *StreamFormat) Video() *Stream {
	for _, s := range f.Streams {
		if s.CodecType == "video" {
			return &s
		}
	}
	return nil
}

// IsVideo ...
func (f *StreamFormat) IsVideo() bool {
	return isVideo(f.Format.Filename)
}

// Audio ...
func (f *StreamFormat) Audio() *Stream {
	for _, s := range f.Streams {
		if s.CodecType == "audio" {
			return &s
		}
	}
	return nil
}

// NameAnalyze 解析
func (f *StreamFormat) NameAnalyze() *FileInfo {
	_, name := filepath.Split(f.Format.Filename)
	ext := filepath.Ext(f.Format.Filename)
	name = strings.Replace(name, ext, "", -1)
	audio := f.Audio()
	audioName := ""
	if audio != nil {
		audioName = audio.CodecName
	}
	video := f.Video()
	videoName := ""
	if video != nil {
		videoName = video.CodecName
	}
	info := &FileInfo{
		Ext:       ext,
		Caption:   "None",
		Language:  "Japanese",
		Audio:     audioName,
		Video:     videoName,
		Sharpness: f.Resolution(),
		Date:      strconv.FormatInt(int64(time.Now().Year()), 10),
		CName:     name,
		EName:     "",
		Prefix:    "",
	}

	na := f.Format.NameAnalyze()

	if na != nil {
		_, err := strconv.ParseInt(info.Date, 10, 32)
		if err != nil {
			return info
		}
		info.Date = na.Date
		info.EName = na.EName
		info.CName = na.CName
		info.Caption = na.Caption
		info.Language = na.Language
	}
	return info
}

// FileInfo ...
type FileInfo struct {
	Ext       string //扩展名
	Caption   string //字幕
	Language  string //语种
	Audio     string //音频
	Video     string //视频
	Sharpness string //清晰度
	Date      string //年份
	CName     string //中文名
	EName     string //英文名
	Prefix    string //前缀(广告信息)
}

// ToString ...
func (info *FileInfo) ToString() string {
	var infos []string
	if info.Prefix != "" {
		infos = append(infos, info.Prefix)
	}
	infos = append(infos, info.CName)
	if info.EName != "" {
		infos = append(infos, info.EName)
	}
	infos = append(infos, info.Date)
	infos = append(infos, info.Sharpness)
	infos = append(infos, info.Video)
	infos = append(infos, info.Audio)
	infos = append(infos, info.Language)
	infos = append(infos, info.Caption)
	return strings.Join(infos, ".") + info.Ext
}

// NameAnalyze ...
func (f *Format) NameAnalyze() *FileInfo {
	return NameAnalyze(f.Filename)
}

// NameAnalyze ...
func NameAnalyze(filename string) *FileInfo {
	_, name := filepath.Split(filename)
	compile, e := regexp.Compile("^\\[(.)+\\]")
	if e != nil {
		return nil
	}
	prefix := compile.FindString(name)
	name = compile.ReplaceAllString(name, "")
	n := strings.Split(name, ".")
	size := len(n)

	if isVideo(name) || size < MaxSizeIdx-1 {
		return nil
	}

	cname := n[CNameIdx]
	ename := ""
	if size-ENameIdx > CNameIdx {
		ename = strings.Join(n[CNameIdx+1:size-DataIdx], ".")
	}

	return &FileInfo{
		Ext:       n[size-ExtIdx],
		Caption:   n[size-CaptionIdx],
		Language:  n[size-LanguageIdx],
		Audio:     n[size-AudioIdx],
		Video:     n[size-VideoIdx],
		Sharpness: n[size-SharpnessIdx],
		Date:      n[size-DataIdx],
		CName:     cname,
		EName:     ename,
		Prefix:    prefix,
	}
}

func isVideo(filename string) bool {
	vlist := []string{
		".swf", "flv", ".3gp", "ogm", ".vob", ".m4v", ".mkv", ".mp4", ".mpg", ".mpeg",
		".avi", ".rm", ".rmvb", ".mov", ".wmv", ".asf", ".dat", ".asx", ".wvx", ".mpe", ".mpa",
	}
	ext := path.Ext(filename)
	for _, v := range vlist {
		if ext == v {
			return true
		}
	}
	return false
}
