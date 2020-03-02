// Lute PDF - 一款通过 Markdown 生成 PDF 的小工具
// Copyright (c) 2020-present, b3log.org
//
// LianDi is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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

	renderer.Cover = &PdfCover{
		Title:         "Lute PDF - Markdown 生成 PDF",
		AuthorLabel:   "　　作者：",
		Author:        "88250",
		AuthorLink:    "https://hacpai.com/member/88250",
		LinkLabel:     "原文链接：",
		Link:          "https://github.com/88250/lute-pdf",
		SourceLabel:   "来源网站：",
		Source:        "GitHub",
		SourceLink:    "https://github.com",
		LicenseLabel:  "许可协议：",
		License:       "署名-相同方式共享 4.0 国际 (CC BY-SA 4.0)",
		LicenseLink:   "https://creativecommons.org/licenses/by-sa/4.0/",
		LogoImgPath:   "https://static.b3log.org/images/brand/b3log-128.png",
		LogoTitle:     "B3log 开源",
		LogoTitleLink: "https://b3log.org",
	}
	renderer.RenderCover()

	_, err = renderer.Render()
	if nil != err {
		log.Fatal(err)
	}
	log.Println("completed")
}
