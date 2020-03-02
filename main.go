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

	renderer.PdfCover = &PdfCover{
		Title:       "Markdown 使用指南 - 基础语法",
		AuthorLabel: "　　作者：",
		Author:      "88250",
		AuthorLink:  "https://hacpai.com/member/88250",
		LinkLabel:   "原文链接：",
		Link:        "https://hacpai.com/article/1583129520165",
		SourceLabel: "来源网站：",
		Source:      "黑客派",
		SourceLink:  "https://hacpai.com",
		LicenseLabel: "许可协议：",
		License:     "署名-相同方式共享 4.0 国际 (CC BY-SA 4.0)",
		LicenseLink: "https://creativecommons.org/licenses/by-sa/4.0/",
	}
	renderer.renderCover()

	_, err = renderer.Render()
	if nil != err {
		log.Fatal(err)
	}
}
