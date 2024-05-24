package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797
// https://en.wikipedia.org/wiki/ANSI_escape_code

const ESC rune = 0x1b

func main() {
	fmt.Println(`<link rel="stylesheet" href="default.css" />`)
	fmt.Print(`<pre><code class="ansi">`)

	var lastStyle *Style = nil

	r := bufio.NewReader(os.Stdin)

	for {
		c, n, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if n == 0 {
			break
		}

		if c == ESC {
			// check next rune
			c, n, err := r.ReadRune()
			if err != nil {
				panic(err)
			}
			if n == 0 {
				return
			}
			if c == '[' {
				// scan until we get to a terminator (just [m for now)
				sequence := readUntilTerminator(r)
				style := escapeSequenceToStyle(sequence, lastStyle)

				if lastStyle != nil && *lastStyle != style {
					fmt.Printf("</span>")
				}

				if style.IsEmpty() {
					lastStyle = nil
				} else {
					fmt.Printf("<span class=\"%s\">", style)
					lastStyle = &style
				}

				continue
			}
		}

		// todo: better/faster conversion here?
		os.Stdout.Write([]byte(string(c)))
	}
	fmt.Println("</code></pre>")
}

func readUntilTerminator(r *bufio.Reader) []rune {
	sequence := []rune{}

	for {
		c, n, err := r.ReadRune()
		if err != nil {
			panic(err)
		}
		if n == 0 {
			return sequence // no terminator means not a valid sequence
		}
		if c == 'm' {
			break
		}
		// todo: check allowed characters
		sequence = append(sequence, c)
	}
	return sequence
}

type Style struct {
	Bold          bool
	Dim           bool
	Italic        bool
	Underline     bool
	Blink         bool
	Inverse       bool
	Hidden        bool
	Strikethrough bool
	Foreground    string
	Background    string
}

func (s Style) IsEmpty() bool {
	return !s.Bold && !s.Dim && !s.Italic && !s.Underline && !s.Blink && !s.Inverse && !s.Hidden && !s.Strikethrough && s.Foreground == "" && s.Background == ""
}

func (s Style) String() string {
	attrs := []string{"ansi"}
	if s.Bold {
		attrs = append(attrs, "ansi-bold")
	}
	if s.Dim {
		attrs = append(attrs, "ansi-dim")
	}
	if s.Italic {
		attrs = append(attrs, "ansi-italic")
	}
	if s.Underline {
		attrs = append(attrs, "ansi-underline")
	}
	if s.Blink {
		attrs = append(attrs, "ansi-blink")
	}
	if s.Inverse {
		attrs = append(attrs, "ansi-inverse")
	}
	if s.Hidden {
		attrs = append(attrs, "ansi-hidden")
	}
	if s.Strikethrough {
		attrs = append(attrs, "ansi-strikethrough")
	}
	if s.Foreground != "" {
		attrs = append(attrs, "ansi-fg-"+s.Foreground)
	}
	if s.Background != "" {
		attrs = append(attrs, "ansi-bg-"+s.Background)
	}
	return strings.Join(attrs, " ")
}

func escapeSequenceToStyle(sequence []rune, lastStyle *Style) Style {
	var style Style
	if lastStyle != nil {
		style = *lastStyle
	}

	segments := strings.Split(string(sequence), ";")
	for i, segment := range segments {
		switch segment {
		case "0":
			style = Style{}
		case "1":
			style.Dim = false
			style.Bold = true
		case "2":
			style.Bold = false
			style.Dim = true
		case "22":
			style.Bold = false
			style.Dim = false
		case "3":
			style.Italic = true
		case "23":
			style.Italic = false
		case "4":
			style.Underline = true
		case "24":
			style.Underline = false
		case "5":
			style.Blink = true
		case "25":
			style.Blink = false
		case "7":
			style.Inverse = true
		case "27":
			style.Inverse = false
		case "8":
			style.Hidden = true
		case "28":
			style.Hidden = false
		case "9":
			style.Strikethrough = true
		case "29":
			style.Strikethrough = false
		}

		// 3/4-bit colors
		if color, ok := namedColorsFG[segment]; ok {
			style.Foreground = color
		}
		if color, ok := namedColorsBG[segment]; ok {
			style.Background = color
		}

		// 8/24-bit colors
		if segment == "38" {
			if segments[i+1] == "5" {
				style.Foreground = eightBitColors[segments[i+2]]
			} else if segments[i+1] == "2" {
				style.Foreground = fmt.Sprintf("rgb(%s,%s,%s)", segments[i+2], segments[i+3], segments[i+4])
			}
		}

		if segment == "48" {
			if segments[i+1] == "5" {
				style.Foreground = eightBitColors[segments[i+2]]
			} else if segments[i+1] == "2" {
				style.Foreground = fmt.Sprintf("rgb(%s,%s,%s)", segments[i+2], segments[i+3], segments[i+4])
			}
		}
	}

	return style
}
