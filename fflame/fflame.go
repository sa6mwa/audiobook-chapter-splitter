// fflame is a package to retrieve metadata with chapters from an m4b
// (or similar) and encode a single m4b into separate mp3s per
// chapter.
package fflame

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type GetMetaOutput struct {
	Chapters []Chapter `json:"chapters"`
	Format   struct {
		Filename   string `json:"filename"`
		NbStreams  int    `json:"nb_streams"`
		NbPrograms int    `json:"nb_programs"`
		FormatName string `json:"format_name"`
		StartTime  string `json:"start_time"`
		Duration   string `json:"duration"`
		Size       string `json:"size"`
		BitRate    string `json:"bit_rate"`
		ProbeScore int    `json:"probe_score"`
		Tags       struct {
			MajorBrand       string    `json:"major_brand"`
			MinorVersion     string    `json:"minor_version"`
			CompatibleBrands string    `json:"compatible_brands"`
			CreationTime     time.Time `json:"creation_time"`
			Title            string    `json:"title"`
			Track            string    `json:"track"`
			Album            string    `json:"album"`
			Genre            string    `json:"genre"`
			Date             string    `json:"date"`
			Copyright        string    `json:"copyright"`
			Artist           string    `json:"artist"`
			AlbumArtist      string    `json:"album_artist"`
			Encoder          string    `json:"encoder"`
			Description      string    `json:"description"`
			Synopsis         string    `json:"synopsis"`
			MediaType        string    `json:"media_type"`
		} `json:"tags"`
	} `json:"format"`
}

type Chapter struct {
	ID        int    `json:"id"`
	TimeBase  string `json:"time_base"`
	Start     int    `json:"start"`
	StartTime string `json:"start_time"`
	End       int    `json:"end"`
	EndTime   string `json:"end_time"`
	Tags      struct {
		Title string `json:"title"`
	} `json:"tags"`
}

func (c *Chapter) Chapter() string {
	return fmt.Sprintf("%.04d", c.ID+1)
}

// GetMeta uses ffprobe to retrieve the chapter and format metadata
// from inputFile (intended to be an m4b, but if ffmpeg/ffprobe
// supports other formats with compatible output, the function should
// work the same). activationBytes is intended for .aax input files
// (m4b encoded audiobooks from Audible). Returns a GetMetaOutput
// structure or error on failure.
func GetMeta(inputFile string, activationBytes ...string) (*GetMetaOutput, error) {
	c := []string{"/usr/bin/env",
		"ffprobe",
	}
	if len(activationBytes) > 0 && activationBytes[0] != "" {
		c = append(c, "-activation_bytes", activationBytes[0])
	}
	c = append(c, "-i", inputFile, "-print_format", "json", "-show_format", "-show_chapters")

	var ffprobe bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &ffprobe)

	cmd := exec.Cmd{
		Path:   c[0],
		Args:   c,
		Stdin:  os.Stdin,
		Stdout: mw,
		Stderr: os.Stderr,
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %w", strings.Join(c, " "), err)
	}

	var output GetMetaOutput

	if err := json.NewDecoder(&ffprobe).Decode(&output); err != nil {
		return nil, fmt.Errorf("%s: %w", strings.Join(c, " "), err)
	}
	return &output, nil
}

// Encode encodes inputFile (m4b or similar if supported) into one mp3
// per chapter in meta.Chapters using lame(1). If title is non-empty
// it will be used as the MP3 album tag and prefix of the output
// file. If title is empty, meta.Format.Tags.Title will be used as MP3
// album and filename prefix. withChapterNumber will append a four
// character long chapter number after the title. activationBytes is
// intended for Audible .aax audiobooks and will append
// -activation_bytes to the ffmpeg command. Will loop over all
// chapters and return an error immediately if something fails.
func Encode(inputFile, outputDir, title string, withChapterNumber bool, meta *GetMetaOutput, activationBytes ...string) error {
	if meta == nil {
		return errors.New("received nil meta")
	}
	if meta.Chapters == nil {
		return errors.New("nil chapters in meta")
	}
	if len(meta.Chapters) == 0 {
		return fmt.Errorf("zero chapters in %s", inputFile)
	}
	inputFile = filepath.Clean(inputFile)
	outputDir = filepath.Clean(outputDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	if title == "" {
		if meta.Format.Tags.Title != "" {
			title = meta.Format.Tags.Title
		} else {
			title = strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
		}
	}

	for _, chapter := range meta.Chapters {
		var newFile string
		if withChapterNumber {
			newFile = title + " - " + chapter.Chapter() + " " + chapter.Tags.Title + ".mp3"
		} else {
			newFile = title + " - " + chapter.Tags.Title + ".mp3"
		}

		ffmpegCMD := []string{"/usr/bin/env",
			"ffmpeg",
		}
		if len(activationBytes) > 0 && activationBytes[0] != "" {
			ffmpegCMD = append(ffmpegCMD, "-activation_bytes", activationBytes[0])
		}
		ffmpegCMD = append(ffmpegCMD,
			"-i", inputFile,
			"-f", "wav",
			"-c:a", "pcm_s16le",
			"-ss", chapter.StartTime,
			"-to", chapter.EndTime,
			"pipe:",
		)

		lameCMD := []string{"/usr/bin/env",
			"lame", "-b", "128",
			"--add-id3v2",
			"--tt", chapter.Tags.Title,
			"--ta", meta.Format.Tags.Artist,
			"--tl", meta.Format.Tags.Title,
			"--tc", meta.Format.Tags.Description,
			"--ty", fmt.Sprintf("%d", meta.Format.Tags.CreationTime.Year()),
			"--tn", fmt.Sprintf("%d", chapter.ID),
			"--tg", meta.Format.Tags.Genre,
			"-",
			filepath.Join(outputDir, newFile),
		}

		ffmpeg := exec.Cmd{
			Path:   ffmpegCMD[0],
			Args:   ffmpegCMD,
			Stdin:  os.Stdin,
			Stderr: os.Stderr,
		}

		outpipe, err := ffmpeg.StdoutPipe()
		if err != nil {
			return err
		}

		lame := exec.Cmd{
			Path:   lameCMD[0],
			Args:   lameCMD,
			Stdin:  outpipe,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := lame.Start(); err != nil {
			return fmt.Errorf("%s: %w", str(lameCMD), err)
		}

		if err := ffmpeg.Start(); err != nil {
			return fmt.Errorf("%s: %w", str(ffmpegCMD), err)
		}

		var ffmpegErr = make(chan error)
		var lameErr = make(chan error)
		var goroutines int = 2
		go func() {
			ffmpegErr <- ffmpeg.Wait()
		}()
		go func() {
			lameErr <- lame.Wait()
		}()

		for i := 0; i < goroutines; i++ {
			select {
			case err := <-ffmpegErr:
				if err != nil {
					lame.Process.Kill()
					return fmt.Errorf("%s: %w", str(ffmpegCMD), err)
				}
			case err := <-lameErr:
				if err != nil {
					ffmpeg.Process.Kill()
					return fmt.Errorf("%s: %w", str(lameCMD), err)
				}
			}
		}
	}

	return nil
}

func str(s []string) string {
	return strings.Join(s, " ")
}
