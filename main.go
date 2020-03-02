package main

import (
	"bytes"
	"io/ioutil"
	"log"

	"github.com/88250/lute/parse"
)

func main() {
	options := &parse.Options{
		GFMTable:            true,
		GFMTaskListItem:     true,
		GFMStrikethrough:    true,
		GFMAutoLink:         true,
		SoftBreak2HardBreak: true,
		Emoji:               true,
	}
	options.AliasEmoji, options.EmojiAlias = parse.NewEmojis()

	markdown, err := ioutil.ReadFile("sample.md")
	if nil != err {
		log.Fatal(err)
	}

	markdown = bytes.ReplaceAll(markdown, []byte("\t"), []byte("    "))
	for emojiUnicode, emojiAlias := range options.EmojiAlias {
		markdown = bytes.ReplaceAll(markdown, []byte(emojiUnicode), []byte(":"+emojiAlias+":"))
	}

	tree, err := parse.Parse("sample", markdown, options)
	if nil != err {
		log.Fatal(err)
	}
	renderer := NewPdfRenderer(tree)
	_, err = renderer.Render()
	if nil != err {
		log.Fatal(err)
	}
}
