# audiobook-chapter-splitter

`audiobook-chapter-splitter` is a batch CLI to split an M4B Audiobook
into one mp3 per chapter using `ffprobe`, `ffmpeg` and `lame`. It
should work with any `m4b` audiobook compatible with the `ffmpeg`
suite, but the use-case was specifically the audio version of
[Reinventing Organizations](https://www.reinventingorganizations.com/)
by Frederic Laloux.

## Build

```console
$ go build -o splitter .
```

or...

```console
$ go install github.com/sa6mwa/audiobook-chapter-splitter@latest
```

You should find the binary in the `bin` directory under `$(go env GOPATH)`.

## Usage

```console
$ audiobook-chapter-splitter -h
usage: audiobook-chapter-splitter input.m4b output_directory

  -t string
        Title. If empty, title tag from metadata or basename of input file will be used
```

`mp3` files will be created under `output_directory` (directory at its
parents will be created if it does not exist). The output filename
will be parsed into `FormatTitle - ChapterTitle.mp3` For example,
Reinventing Organizations looks like this...

```
Reinventing Organizations - 001 - Introduction.mp3
Reinventing Organizations - 002 - 1.1 Changing Paradigms.mp3
Reinventing Organizations - 003 - 1.1 Infrared through Red paradigms.mp3
Reinventing Organizations - 004 - 1.1 Conformist-Amber paradigm.mp3
Reinventing Organizations - 005 - 1.1 Amber Organizations.mp3
Reinventing Organizations - 006 - 1.1 Achievement-Orange paradigm.mp3
Reinventing Organizations - 007 - 1.1 Orange Organizations.mp3
Reinventing Organizations - 008 - 1.1 Organizations as Machines.mp3
Reinventing Organizations - 009 - 1.1 Pluralistic-Green paradigm.mp3
Reinventing Organizations - 010 - 1.1 Green Organizations.mp3
Reinventing Organizations - 011 - 1.1 From Red to Green.mp3
Reinventing Organizations - 012 - 1.2 About Stages of Development.mp3
Reinventing Organizations - 013 - 1.3 Evolutionary-Teal.mp3
```

The ID3 tag(s) set by `lame` looks like this...

```console
$ id3info Reinventing\ Organizations\ -\ 030\ -\ 2.2\ Trust\ vs.\ Control.mp3 

*** Tag information for Reinventing Organizations - 030 - 2.2 Trust vs. Control.mp3
=== TSSE (Software/Hardware and settings used for encoding): LAME 64bits version 3.100 (http://lame.sf.net)
=== TIT2 (Title/songname/content description): 030 - 2.2 Trust vs. Control
=== TPE1 (Lead performer(s)/Soloist(s)): Frederic Laloux
=== TALB (Album/Movie/Show title): Reinventing Organizations
=== COMM (Comments): ()[eng]: The way we manage organizations seems increasingly out of date. "Reinventing Organizations" describes how a new management paradigm is currently emerging, and discusses in practical detail how organizations large and small can operate in fundamentally new
=== TYER (Year): 2016
=== TRCK (Track number/Position in set): 29
=== TCON (Content type): Audiobooks
=== TLEN (Length): 4294967295
=== COMM (Comments): (ID3v1 Comment)[XXX]: The way we manage organizati
*** mp3 info
MPEG2/layer III
Bitrate: 128KBps
Frequency: 22KHz
```

## fflame package

The two functions in
`github.com/sa6mwa/audiobook-chapter-splitter/fflame` does the lifting
in the main function and can be imported elsewhere, see `main.go` for usage example.

```go
// GetMeta uses ffprobe to retrieve the chapter and format metadata
// from inputFile (intended to be an m4b, but if ffmpeg/ffprobe
// supports other formats with compatible output, the function should
// work the same). Return a GetMetaOutput structure or error on
// failure.
GetMeta(inputFile string) (*GetMetaOutput, error)

// Encode encodes inputFile (m4b or similar if supported) into one mp3
// per chapter in meta.Chapters using lame(1). If title is non-empty
// it will be used as the MP3 album tag and prefix of the output
// file. If title is empty, meta.Format.Tags.Title will be used as MP3
// album and filename prefix. Will loop over all chapters and return
// an error immediately if something fails.
Encode(inputFile, outputDir, title string, meta *GetMetaOutput) error
```
