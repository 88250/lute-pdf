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
	argBoldFontPath := flag.String("boldFontPath", "D:/88250/lute-pdf/fonts/msyhb.ttf", "粗体字体文件路径")
	argItalicFontPath := flag.String("italicFontPath", "D:/88250/lute-pdf/fonts/msyhl.ttf", "斜体字体文件路径")
	argSavePath := flag.String("savePath", "D:/88250/lute-pdf/sample.pdf", "转换后 PDF 的保存路径")

	argCoverTitle := flag.String("coverTitle", "Lute PDF - Markdown 生成 PDF", "封面 - 标题")
	argCoverAuthor := flag.String("coverAuthor", "88250", "封面 - 作者")
	argCoverAuthorLink := flag.String("coverAuthorLink", "https://hacpai.com/member/88250", "封面 - 作者链接")
	argCoverLink := flag.String("coverLink", "https://github.com/88250/lute-pdf", "封面 - 原文链接")
	argCoverSource := flag.String("coverSource", "GitHub", "封面 - 来源网站")
	argCoverSourceLink := flag.String("coverSourceLink", "https://github.com", "封面 - 来源网站链接")
	argCoverLicense := flag.String("coverLicense", "署名-相同方式共享 4.0 国际 (CC BY-SA 4.0)", "封面 - 文档许可协议")
	argCoverLicenseLink := flag.String("coverLicenseLink", "https://creativecommons.org/licenses/by-sa/4.0/", "封面 - 文档许可协议链接")
	argCoverLogoLink := flag.String("coverLogoLink", "https://static.b3log.org/images/brand/b3log-128.png", "封面 - 图标链接")
	argCoverLogoTitle := flag.String("coverLogoTitle", "B3log 开源", "封面 - 图标标题")
	argCoverLogoTitleLink := flag.String("coverLogoTitleLink", "https://b3log.org", "封面 - 图标标题链接")

	flag.Parse()

	mdPath := *argMdPath
	regularFontPath := *argRegularFontPath
	boldFontPath := *argBoldFontPath
	italicFontPath := *argItalicFontPath
	savePath := *argSavePath

	coverTitle := *argCoverTitle
	coverAuthorLabel := "　　作者："
	coverAuthor := *argCoverAuthor
	coverAuthorLink := *argCoverAuthorLink
	coverLinkLabel := "原文链接："
	coverLink := *argCoverLink
	coverSourceLabel := "来源网站："
	coverSource := *argCoverSource
	coverSourceLink := *argCoverSourceLink
	coverLicenseLabel := "许可协议："
	coverLicense := *argCoverLicense
	coverLicenseLink := *argCoverLicenseLink
	coverLogoLink := *argCoverLogoLink
	coverLogoTitle := *argCoverLogoTitle
	coverLogoTitleLink := *argCoverLogoTitleLink

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
		LogoLink:      coverLogoLink,
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
