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
	"flag"
	"io/ioutil"
	"log"

	"github.com/88250/lute/parse"
)

func main() {
	argMdPath := flag.String("mdPath", "D:/88250/lute-pdf/sample.md", "Markdown 文件路径")
	argRegularFontPath := flag.String("regularFontPath", "D:/88250/lute-pdf/fonts/msyh.ttf", "正常字体文件路径")


	flag.Parse()

	mdPath := *argMdPath
	regularFontPath := *argRegularFontPath
	boldFontPath := "D:/88250/lute-pdf/fonts/msyhb.ttf"
	italicFontPath := "D:/88250/lute-pdf/fonts/msyhl.ttf"
	savePath := "D:/88250/lute-pdf/sample.pdf"

	coverTitle := "Lute PDF - Markdown 生成 PDF"
	coverAuthorLabel := "　　作者："
	coverAuthor := "88250"
	coverAuthorLink := "https://hacpai.com/member/88250"
	coverLinkLabel := "原文链接："
	coverLink := "https://github.com/88250/lute-pdf"
	coverSourceLabel := "来源网站："
	coverSource := "GitHub"
	coverSourceLink := "https://github.com"
	coverLicenseLabel := "许可协议："
	coverLicense := "署名-相同方式共享 4.0 国际 (CC BY-SA 4.0)"
	coverLicenseLink := "https://creativecommons.org/licenses/by-sa/4.0/"
	coverLogoImgPath := "https://static.b3log.org/images/brand/b3log-128.png"
	coverLogoTitle := "B3log 开源"
	coverLogoTitleLink := "https://b3log.org"


	options := &parse.Options{
		GFMTable:            true,
		GFMTaskListItem:     true,
		GFMStrikethrough:    true,
		GFMAutoLink:         true,
		SoftBreak2HardBreak: true,
		Emoji:               true,
	}
	options.AliasEmoji, options.EmojiAlias = parse.NewEmojis()

	markdown, err := ioutil.ReadFile(mdPath)
	if nil != err {
		log.Fatal(err)
	}

	markdown = bytes.ReplaceAll(markdown, []byte("\t"), []byte("    "))
	for emojiUnicode, emojiAlias := range options.EmojiAlias {
		markdown = bytes.ReplaceAll(markdown, []byte(emojiUnicode), []byte(":"+emojiAlias+":"))
	}

	tree, err := parse.Parse("", markdown, options)
	if nil != err {
		log.Fatal(err)
	}

	renderer := NewPdfRenderer(tree, regularFontPath, boldFontPath, italicFontPath)
	renderer.Cover = &PdfCover{
		Title:         coverTitle,
		AuthorLabel:   coverAuthorLabel,
		Author:        coverAuthor,
		AuthorLink:    coverAuthorLink,
		LinkLabel:     coverLinkLabel,
		Link:          coverLink,
		SourceLabel:   coverSourceLabel,
		Source:        coverSource,
		SourceLink:    coverSourceLink,
		LicenseLabel:  coverLicenseLabel,
		License:       coverLicense,
		LicenseLink:   coverLicenseLink,
		LogoImgPath:   coverLogoImgPath,
		LogoTitle:     coverLogoTitle,
		LogoTitleLink: coverLogoTitleLink,
	}
	renderer.RenderCover()

	_, err = renderer.Render()
	if nil != err {
		log.Fatal(err)
	}

	renderer.Save(savePath)

	log.Println("completed")
}
