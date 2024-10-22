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
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/88250/lute/ast"
	"github.com/88250/lute/lex"
	"github.com/88250/lute/parse"
	"github.com/88250/lute/render"
	"github.com/88250/lute/util"
	"github.com/signintech/gopdf"
)

// PdfRenderer 描述了 PDF 渲染器。
type PdfRenderer struct {
	*render.BaseRenderer

	Cover       *PdfCover // 封面
	RegularFont string    // 正常字体文件路径
	BoldFont    string    // 粗体字体文件路径
	ItalicFont  string    // 斜体字体文件路径

	pdf          *gopdf.GoPdf // PDF 生成器句柄
	pageSize     *gopdf.Rect  // 页面大小
	zoom         float64      // 字体、行高大小倍数
	fontSize     int          // 字体大小
	lineHeight   float64      // 行高
	heading1Size float64      // 一级标题大小
	heading2Size float64      // 二级标题大小
	heading3Size float64      // 三级标题大小
	heading4Size float64      // 四级标题大小
	heading5Size float64      // 五级标题大小
	heading6Size float64      // 六级标题大小
	margin       float64      // 页边距
	x            []float64    // 当前横坐标栈
	fonts        []*Font      // 当前字体栈
	textColors   []*RGB       // 当前文本颜色栈
}

// PdfCover 描述了 PDF 封面。
type PdfCover struct {
	Title         string // 标题
	AuthorLabel   string // 作者：
	Author        string // 作者
	AuthorLink    string // 作者链接
	LinkLabel     string // 原文链接：
	Link          string // 原文链接
	SourceLabel   string // 来源网站：
	Source        string // 来源网站
	SourceLink    string // 来源网站链接
	LicenseLabel  string // 许可协议：
	License       string // 许可协议
	LicenseLink   string // 许可协议链接
	LogoLink      string // 图标链接
	LogoTitle     string // 图片标题
	LogoTitleLink string // 图标标题链接
}

func (r *PdfRenderer) RenderCover() {
	r.pdf.AddPage()

	logoImgPath, ok, isTemp := r.downloadImg(r.Cover.LogoLink)
	if ok {
		imgW, imgH := r.getImgSize(logoImgPath)
		x := (r.pageSize.W)/2 - imgW/2
		y := r.pageSize.H/2 - r.margin - 128
		r.pdf.Image(logoImgPath, x, y, nil)
		r.pdf.SetY(y)
		r.pdf.Br(imgH + 10)
		r.pdf.SetFontWithStyle("regular", gopdf.Regular, 20)
		width, _ := r.pdf.MeasureTextWidth(r.Cover.LogoTitle)
		x = (r.pageSize.W)/2 - width/2
		r.pdf.SetX(x)
		y = r.pdf.GetY()
		r.pdf.Cell(nil, r.Cover.LogoTitle)
		r.pdf.AddExternalLink(r.Cover.LogoTitleLink, x, y, width, 20)
		r.pdf.Br(48)
		if isTemp {
			os.Remove(logoImgPath)
		}
	}

	r.pdf.SetFontWithStyle("regular", gopdf.Regular, 28)
	lines, _ := r.pdf.SplitText(r.Cover.Title, r.pageSize.W-r.margin)
	for _, line := range lines {
		width, _ := r.pdf.MeasureTextWidth(line)
		x := (r.pageSize.W)/2 - width/2
		r.pdf.SetX(x)
		r.pdf.Cell(nil, line)
		r.pdf.Br(30)
	}

	fontSize := 12
	r.pdf.Br(45)
	r.pdf.SetX(r.margin)
	r.pdf.SetFontWithStyle("regular", gopdf.Regular, fontSize)
	r.pdf.Cell(nil, r.Cover.AuthorLabel)
	x := r.pdf.GetX()
	width, _ := r.pdf.MeasureTextWidth(r.Cover.Author)
	r.pdf.SetTextColor(66, 133, 244)
	r.pdf.Cell(nil, r.Cover.Author)
	r.pdf.AddExternalLink(r.Cover.AuthorLink, x, r.pdf.GetY(), width, float64(fontSize))
	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.Br(22)

	r.pdf.Cell(nil, r.Cover.LinkLabel)
	x = r.pdf.GetX()
	width, _ = r.pdf.MeasureTextWidth(r.Cover.Link)
	r.pdf.SetTextColor(66, 133, 244)
	r.pdf.Cell(nil, r.Cover.Link)
	r.pdf.AddExternalLink(r.Cover.Link, x, r.pdf.GetY(), width, float64(fontSize))
	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.Br(22)

	r.pdf.Cell(nil, r.Cover.SourceLabel)
	x = r.pdf.GetX()
	width, _ = r.pdf.MeasureTextWidth(r.Cover.Source)
	r.pdf.SetTextColor(66, 133, 244)
	r.pdf.Cell(nil, r.Cover.Source)
	r.pdf.AddExternalLink(r.Cover.SourceLink, x, r.pdf.GetY(), width, float64(fontSize))
	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.Br(22)

	r.pdf.Cell(nil, r.Cover.LicenseLabel)
	x = r.pdf.GetX()
	width, _ = r.pdf.MeasureTextWidth(r.Cover.License)
	r.pdf.SetTextColor(66, 133, 244)
	r.pdf.Cell(nil, r.Cover.License)
	r.pdf.AddExternalLink(r.Cover.LicenseLink, x, r.pdf.GetY(), width, float64(fontSize))
	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.Br(20)

	r.pdf.AddPage()
}

// NewPdfRenderer 创建一个 HTML 渲染器。
func NewPdfRenderer(tree *parse.Tree, options *render.Options, regularFont, boldFont, italicFont string) *PdfRenderer {
	pdf := &gopdf.GoPdf{}

	ret := &PdfRenderer{BaseRenderer: render.NewBaseRenderer(tree, options), pdf: pdf}
	ret.zoom = 0.8
	ret.fontSize = int(math.Floor(14 * ret.zoom))
	ret.lineHeight = 24.0 * ret.zoom
	ret.heading1Size = 24 * ret.zoom
	ret.heading2Size = 22 * ret.zoom
	ret.heading3Size = 20 * ret.zoom
	ret.heading4Size = 18 * ret.zoom
	ret.heading5Size = 16 * ret.zoom
	ret.heading6Size = 14 * ret.zoom
	ret.margin = 60 * ret.zoom

	ret.RegularFont = regularFont
	ret.BoldFont = boldFont
	ret.ItalicFont = italicFont

	ret.pageSize = gopdf.PageSizeA4
	pdf.Start(gopdf.Config{PageSize: *ret.pageSize})

	var err error
	err = pdf.AddTTFFont("regular", ret.RegularFont)
	if err != nil {
		logger.Fatal(err)
	}

	err = pdf.AddTTFFontWithOption("bold", ret.BoldFont, gopdf.TtfOption{Style: gopdf.Bold})
	if err != nil {
		logger.Fatal(err)
	}

	err = pdf.AddTTFFontWithOption("italic", ret.ItalicFont, gopdf.TtfOption{Style: gopdf.Italic})
	if err != nil {
		logger.Fatal(err)
	}

	//err = pdf.AddTTFFont("emoji", "fonts/seguiemj.ttf")
	//if err != nil {
	//	logger.Fatal(err)
	//}

	ret.pushFont(&Font{"regular", "R", ret.fontSize})
	ret.pushTextColor(&RGB{0, 0, 0})
	pdf.SetMargins(ret.margin, ret.margin, ret.margin, ret.margin)

	ret.RendererFuncs[ast.NodeDocument] = ret.renderDocument
	ret.RendererFuncs[ast.NodeParagraph] = ret.renderParagraph
	ret.RendererFuncs[ast.NodeText] = ret.renderText
	ret.RendererFuncs[ast.NodeCodeSpan] = ret.renderCodeSpan
	ret.RendererFuncs[ast.NodeCodeSpanOpenMarker] = ret.renderCodeSpanOpenMarker
	ret.RendererFuncs[ast.NodeCodeSpanContent] = ret.renderCodeSpanContent
	ret.RendererFuncs[ast.NodeCodeSpanCloseMarker] = ret.renderCodeSpanCloseMarker
	ret.RendererFuncs[ast.NodeCodeBlock] = ret.renderCodeBlock
	ret.RendererFuncs[ast.NodeCodeBlockFenceOpenMarker] = ret.renderCodeBlockOpenMarker
	ret.RendererFuncs[ast.NodeCodeBlockFenceInfoMarker] = ret.renderCodeBlockInfoMarker
	ret.RendererFuncs[ast.NodeCodeBlockCode] = ret.renderCodeBlockCode
	ret.RendererFuncs[ast.NodeCodeBlockFenceCloseMarker] = ret.renderCodeBlockCloseMarker
	ret.RendererFuncs[ast.NodeMathBlock] = ret.renderMathBlock
	ret.RendererFuncs[ast.NodeMathBlockOpenMarker] = ret.renderMathBlockOpenMarker
	ret.RendererFuncs[ast.NodeMathBlockContent] = ret.renderMathBlockContent
	ret.RendererFuncs[ast.NodeMathBlockCloseMarker] = ret.renderMathBlockCloseMarker
	ret.RendererFuncs[ast.NodeInlineMath] = ret.renderInlineMath
	ret.RendererFuncs[ast.NodeInlineMathOpenMarker] = ret.renderInlineMathOpenMarker
	ret.RendererFuncs[ast.NodeInlineMathContent] = ret.renderInlineMathContent
	ret.RendererFuncs[ast.NodeInlineMathCloseMarker] = ret.renderInlineMathCloseMarker
	ret.RendererFuncs[ast.NodeEmphasis] = ret.renderEmphasis
	ret.RendererFuncs[ast.NodeEmA6kOpenMarker] = ret.renderEmAsteriskOpenMarker
	ret.RendererFuncs[ast.NodeEmA6kCloseMarker] = ret.renderEmAsteriskCloseMarker
	ret.RendererFuncs[ast.NodeEmU8eOpenMarker] = ret.renderEmUnderscoreOpenMarker
	ret.RendererFuncs[ast.NodeEmU8eCloseMarker] = ret.renderEmUnderscoreCloseMarker
	ret.RendererFuncs[ast.NodeStrong] = ret.renderStrong
	ret.RendererFuncs[ast.NodeStrongA6kOpenMarker] = ret.renderStrongA6kOpenMarker
	ret.RendererFuncs[ast.NodeStrongA6kCloseMarker] = ret.renderStrongA6kCloseMarker
	ret.RendererFuncs[ast.NodeStrongU8eOpenMarker] = ret.renderStrongU8eOpenMarker
	ret.RendererFuncs[ast.NodeStrongU8eCloseMarker] = ret.renderStrongU8eCloseMarker
	ret.RendererFuncs[ast.NodeBlockquote] = ret.renderBlockquote
	ret.RendererFuncs[ast.NodeBlockquoteMarker] = ret.renderBlockquoteMarker
	ret.RendererFuncs[ast.NodeHeading] = ret.renderHeading
	ret.RendererFuncs[ast.NodeHeadingC8hMarker] = ret.renderHeadingC8hMarker
	ret.RendererFuncs[ast.NodeList] = ret.renderList
	ret.RendererFuncs[ast.NodeListItem] = ret.renderListItem
	ret.RendererFuncs[ast.NodeThematicBreak] = ret.renderThematicBreak
	ret.RendererFuncs[ast.NodeHardBreak] = ret.renderHardBreak
	ret.RendererFuncs[ast.NodeSoftBreak] = ret.renderSoftBreak
	ret.RendererFuncs[ast.NodeHTMLBlock] = ret.renderHTML
	ret.RendererFuncs[ast.NodeInlineHTML] = ret.renderInlineHTML
	ret.RendererFuncs[ast.NodeLink] = ret.renderLink
	ret.RendererFuncs[ast.NodeImage] = ret.renderImage
	ret.RendererFuncs[ast.NodeBang] = ret.renderBang
	ret.RendererFuncs[ast.NodeOpenBracket] = ret.renderOpenBracket
	ret.RendererFuncs[ast.NodeCloseBracket] = ret.renderCloseBracket
	ret.RendererFuncs[ast.NodeOpenParen] = ret.renderOpenParen
	ret.RendererFuncs[ast.NodeCloseParen] = ret.renderCloseParen
	ret.RendererFuncs[ast.NodeLinkText] = ret.renderLinkText
	ret.RendererFuncs[ast.NodeLinkSpace] = ret.renderLinkSpace
	ret.RendererFuncs[ast.NodeLinkDest] = ret.renderLinkDest
	ret.RendererFuncs[ast.NodeLinkTitle] = ret.renderLinkTitle
	ret.RendererFuncs[ast.NodeStrikethrough] = ret.renderStrikethrough
	ret.RendererFuncs[ast.NodeStrikethrough1OpenMarker] = ret.renderStrikethrough1OpenMarker
	ret.RendererFuncs[ast.NodeStrikethrough1CloseMarker] = ret.renderStrikethrough1CloseMarker
	ret.RendererFuncs[ast.NodeStrikethrough2OpenMarker] = ret.renderStrikethrough2OpenMarker
	ret.RendererFuncs[ast.NodeStrikethrough2CloseMarker] = ret.renderStrikethrough2CloseMarker
	ret.RendererFuncs[ast.NodeTaskListItemMarker] = ret.renderTaskListItemMarker
	ret.RendererFuncs[ast.NodeTable] = ret.renderTable
	ret.RendererFuncs[ast.NodeTableHead] = ret.renderTableHead
	ret.RendererFuncs[ast.NodeTableRow] = ret.renderTableRow
	ret.RendererFuncs[ast.NodeTableCell] = ret.renderTableCell
	ret.RendererFuncs[ast.NodeEmoji] = ret.renderEmoji
	ret.RendererFuncs[ast.NodeEmojiUnicode] = ret.renderEmojiUnicode
	ret.RendererFuncs[ast.NodeEmojiImg] = ret.renderEmojiImg
	ret.RendererFuncs[ast.NodeEmojiAlias] = ret.renderEmojiAlias
	ret.RendererFuncs[ast.NodeFootnotesDefBlock] = ret.renderFootnotesDefBlock
	ret.RendererFuncs[ast.NodeFootnotesDef] = ret.renderFootnotesDef
	ret.RendererFuncs[ast.NodeFootnotesRef] = ret.renderFootnotesRef
	ret.RendererFuncs[ast.NodeToC] = ret.renderToC
	ret.RendererFuncs[ast.NodeBackslash] = ret.renderBackslash
	ret.RendererFuncs[ast.NodeBackslashContent] = ret.renderBackslashContent
	ret.RendererFuncs[ast.NodeHTMLEntity] = ret.renderHtmlEntity
	ret.RendererFuncs[ast.NodeYamlFrontMatter] = ret.renderYamlFrontMatter
	return ret
}

func (r *PdfRenderer) Render() (output []byte) {
	r.LastOut = lex.ItemNewline

	ast.Walk(r.Tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		extRender := r.ExtRendererFuncs[n.Type]
		if nil != extRender {
			output, status := extRender(n, entering)
			r.WriteString(output)
			return status
		}

		render := r.RendererFuncs[n.Type]
		if nil == render {
			if nil != r.DefaultRendererFunc {
				return r.DefaultRendererFunc(n, entering)
			} else {
				return r.renderDefault(n, entering)
			}
		}
		return render(n, entering)
	})

	if 0 < len(r.FootnotesDefs) {
		output = r.RenderFootnotesDefs(r.Tree.Context)
	}
	return
}

func (r *PdfRenderer) renderDefault(n *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderYamlFrontMatter(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderCodeBlockLike(node.Tokens)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderHtmlEntity(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Write(node.HtmlEntityTokens)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderBackslashContent(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Write(node.Tokens)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderBackslash(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderToC(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		headings := r.headings()
		length := len(headings)
		if 1 > length {
			return ast.WalkContinue
		}
		r.WriteString("<div class=\"toc-div\">")
		for i, heading := range headings {
			level := strconv.Itoa(heading.HeadingLevel)
			spaces := (heading.HeadingLevel - 1) * 2
			r.WriteString(strings.Repeat("&emsp;", spaces))
			r.WriteString("<span class=\"toc-h" + level + "\">")
			r.WriteString("<a class=\"toc-a\" href=\"#toc_h" + level + "_" + strconv.Itoa(i) + "\">" + heading.Text() + "</a></span><br>")
		}
		r.WriteString("</div>\n\n")
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) headings() (ret []*ast.Node) {
	for n := r.Tree.Root.FirstChild; nil != n; n = n.Next {
		r.headings0(n, &ret)
	}
	return
}

func (r *PdfRenderer) headings0(n *ast.Node, headings *[]*ast.Node) {
	if ast.NodeHeading == n.Type {
		*headings = append(*headings, n)
		return
	}
	if ast.NodeList == n.Type || ast.NodeListItem == n.Type || ast.NodeBlockquote == n.Type {
		for c := n.FirstChild; nil != c; c = c.Next {
			r.headings0(c, headings)
		}
	}
}

func (r *PdfRenderer) RenderFootnotesDefs(context *parse.Context) []byte {
	if 1 > len(r.FootnotesDefs) || r.RenderingFootnotes {
		return nil
	}

	r.RenderingFootnotes = true
	r.addPage()
	r.renderThematicBreak(nil, false)
	for i, def := range r.FootnotesDefs {
		r.pdf.SetAnchor(string(def.Tokens))
		r.WriteString(fmt.Sprint(i+1) + ". ")
		tree := &parse.Tree{Name: "", Context: context}
		tree.Context.Tree = tree
		tree.Root = &ast.Node{Type: ast.NodeDocument}
		tree.Root.AppendChild(def)
		r.Tree = tree
		r.Render()
		r.Newline()
	}
	r.renderFooter()
	return nil
}

func (r *PdfRenderer) renderFootnotesRef(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		x := r.pdf.GetX() + 1
		r.pdf.SetX(x)
		y := r.pdf.GetY()
		r.pdf.SetFont("regular", "R", 8)
		r.pdf.SetTextColor(66, 133, 244)

		idx := string(node.Tokens)
		width, _ := r.pdf.MeasureTextWidth(idx[1:])
		r.pdf.SetY(y - 4)
		r.pdf.Cell(nil, idx[1:])
		r.pdf.AddInternalLink(idx, x-3, y-9, width+4, r.lineHeight)

		x += width
		r.pdf.SetX(x)
		r.pdf.SetY(y)
		font := r.peekFont()
		r.pdf.SetFont(font.family, font.style, font.size)
		textColor := r.peekTextColor()
		r.pdf.SetTextColor(textColor.R, textColor.G, textColor.B)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderFootnotesDefBlock(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderFootnotesDef(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		if !r.RenderingFootnotes {
			var found bool
			for _, n := range r.FootnotesDefs {
				if bytes.EqualFold(node.Tokens, n.Tokens) {
					found = true
					break
				}
			}
			if !found {
				r.FootnotesDefs = append(r.FootnotesDefs, node)
			}
			return ast.WalkSkipChildren
		}
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderCodeBlock(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		if !node.IsFencedCodeBlock {
			// 缩进代码块处理
			r.renderCodeBlockLike(node.Tokens)
			return ast.WalkContinue
		}
	}
	return ast.WalkContinue
}

// renderCodeBlockCode 进行代码块 HTML 渲染，实现语法高亮。
func (r *PdfRenderer) renderCodeBlockCode(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderCodeBlockLike(node.Tokens)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderCodeBlockLike(content []byte) {
	r.Newline()
	r.pdf.SetY(r.pdf.GetY() + 6)
	r.pushTextColor(&RGB{86, 158, 61})
	r.WriteString(util.BytesToStr(content))
	r.popTextColor()
	r.pdf.SetY(r.pdf.GetY() + 6)
	r.Newline()
}

func (r *PdfRenderer) renderCodeSpanLike(content []byte) {
	r.pushTextColor(&RGB{255, 153, 51})
	r.WriteString(util.BytesToStr(content))
	r.popTextColor()
}

func (r *PdfRenderer) renderCodeBlockCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderCodeBlockInfoMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderCodeBlockOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderEmojiAlias(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.pushFont(&Font{"emoji", "R", r.fontSize})
		alias := node.Tokens[1 : len(node.Tokens)-1]
		r.Write(alias)
		r.popFont()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderEmojiImg(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderEmojiUnicode(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.pushFont(&Font{"emoji", "R", r.fontSize})
		r.Write(node.Tokens)
		r.popFont()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderEmoji(node *ast.Node, entering bool) ast.WalkStatus {
	// 暂不渲染 Emoji，字体似乎有问题
	return ast.WalkContinue
}

func (r *PdfRenderer) renderInlineMathCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderInlineMathContent(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderCodeSpanLike(node.Tokens)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderInlineMathOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderInlineMath(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderMathBlockCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderMathBlockContent(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderCodeBlockLike(node.Tokens)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderMathBlockOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderMathBlock(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	return ast.WalkContinue
}

func (r *PdfRenderer) renderTableCell(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		// TODO: table align
		//var attrs [][]string
		//switch node.TableCellAlign {
		//case 1:
		//	attrs = append(attrs, []string{"align", "left"})
		//case 2:
		//	attrs = append(attrs, []string{"align", "center"})
		//case 3:
		//	attrs = append(attrs, []string{"align", "right"})
		//}
		x := r.pdf.GetX()
		cols := float64(r.tableCols(node))
		maxWidth := (r.pageSize.W - r.margin*2) / cols
		if node.Parent.FirstChild != node {
			prevWidth, _ := r.pdf.MeasureTextWidth(strings.Repeat("爱", node.Previous.TableCellContentWidth))
			x += maxWidth - prevWidth
			r.pdf.SetX(x)
		}
		// TODO: table border
		// r.pdf.RectFromUpperLeftWithStyle(x, r.pdf.GetY(), maxWidth, r.lineHeight, "D")
		r.pdf.SetX(r.pdf.GetX() + 4)
		r.pdf.SetY(r.pdf.GetY() + 4)
	} else {
		r.pdf.SetX(r.pdf.GetX() - 4)
		r.pdf.SetY(r.pdf.GetY() - 4)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) tableCols(cell *ast.Node) int {
	for parent := cell.Parent; nil != parent; parent = parent.Parent {
		if nil != parent.TableAligns {
			return len(parent.TableAligns)
		}
	}
	return 0
}

func (r *PdfRenderer) renderTableRow(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	return ast.WalkContinue
}

func (r *PdfRenderer) renderTableHead(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.pushFont(&Font{"bold", "B", r.fontSize})
	} else {
		r.popFont()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderTable(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.pdf.SetY(r.pdf.GetY() + 6)
	} else {
		r.pdf.SetY(r.pdf.GetY() + 6)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrikethrough(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrikethrough1OpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.pushX(r.pdf.GetX())
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrikethrough1CloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		x := r.popX()
		r.pdf.Line(x, r.pdf.GetY()+float64(r.fontSize)/2, r.pdf.GetX(), r.pdf.GetY()+float64(r.fontSize)/2)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrikethrough2OpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.pushX(r.pdf.GetX())
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrikethrough2CloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		x := r.popX()
		r.pdf.Line(x, r.pdf.GetY()+float64(r.fontSize)/2, r.pdf.GetX(), r.pdf.GetY()+float64(r.fontSize)/2)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderLinkTitle(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderLinkDest(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderLinkSpace(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderLinkText(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		if ast.NodeImage != node.Parent.Type {
			r.Write(node.Tokens)
		}
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderCloseParen(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderOpenParen(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderCloseBracket(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderOpenBracket(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderBang(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderImage(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		if 0 == r.DisableTags {
			destTokens := node.ChildByType(ast.NodeLinkDest).Tokens
			src := util.BytesToStr(destTokens)
			src, ok, isTemp := r.downloadImg(src)
			if ok {
				_, height := r.getImgSize(src)
				y := r.pdf.GetY()
				if math.Ceil(y)+height > math.Floor(r.pageSize.H-r.margin) {
					r.addPage()
				}
				r.pdf.Image(src, r.pdf.GetX(), r.pdf.GetY(), nil)
				r.pdf.SetY(r.pdf.GetY() + height)
				if isTemp {
					os.Remove(src)
				}
			}
		}
		r.DisableTags++
		return ast.WalkContinue
	}

	r.DisableTags--
	if 0 == r.DisableTags {
		//r.WriteString("\"")
		//if title := node.ChildByType(ast.NodeLinkTitle); nil != title && nil != title.Tokens {
		//	r.WriteString(" title=\"")
		//	r.Write(title.Tokens)
		//	r.WriteString("\"")
		//}
		//r.WriteString(" />")
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderLink(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.pushX(r.pdf.GetX())
		r.pushTextColor(&RGB{66, 133, 244})
	} else {
		x := r.popX()
		width := r.pdf.GetX() - x
		dest := node.ChildByType(ast.NodeLinkDest)
		destTokens := dest.Tokens
		destTokens = r.RelativePath(destTokens)
		r.pdf.AddExternalLink(util.BytesToStr(destTokens), x, r.pdf.GetY(), width, r.lineHeight)
		r.popTextColor()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderHTML(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderCodeBlockLike(node.Tokens)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderInlineHTML(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderCodeSpanLike(node.Tokens)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderDocument(node *ast.Node, entering bool) ast.WalkStatus {
	if !entering {
		r.renderFooter()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) Save(pdfPath string) {
	if err := r.pdf.WritePdf(pdfPath); nil != err {
		logger.Fatal(err)
	}
	if err := r.pdf.Close(); nil != err {
		logger.Fatal(err)
	}
}

func (r *PdfRenderer) renderParagraph(node *ast.Node, entering bool) ast.WalkStatus {
	inList := false
	grandparent := node.Parent.Parent
	inTightList := false
	if nil != grandparent && ast.NodeList == grandparent.Type {
		inList = true
		inTightList = grandparent.ListData.Tight
	}

	if inTightList { // List.ListItem.Paragraph
		return ast.WalkContinue
	}

	if entering {
		if !inList {
			r.Newline()
			r.pdf.SetY(r.pdf.GetY() + 6)
		}
	} else {
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderText(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		text := util.BytesToStr(node.Tokens)
		r.WriteString(text)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderCodeSpan(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderCodeSpanOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderCodeSpanContent(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderCodeSpanLike(node.Tokens)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderCodeSpanCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderEmphasis(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderEmAsteriskOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.pushFont(&Font{"italic", "I", r.fontSize})
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderEmAsteriskCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.popFont()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderEmUnderscoreOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.pushFont(&Font{"italic", "I", r.fontSize})
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderEmUnderscoreCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.popFont()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrong(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrongA6kOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.pushFont(&Font{"bold", "B", r.fontSize})
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrongA6kCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.popFont()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrongU8eOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.pushFont(&Font{"bold", "B", r.fontSize})
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrongU8eCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.popFont()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderBlockquote(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
		r.pushTextColor(&RGB{106, 115, 125})
		r.pushX(r.pdf.GetX())
	} else {
		x := r.popX()
		r.pdf.SetX(r.pdf.GetX() - x + r.margin)
		r.popTextColor()
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderBlockquoteMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderHeading(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
		r.pdf.SetY(r.pdf.GetY() + 10)
		headingSize := r.heading2Size
		switch node.HeadingLevel {
		case 1:
			headingSize = r.heading1Size
		case 2:
			headingSize = r.heading2Size
		case 3:
			headingSize = r.heading3Size
		case 4:
			headingSize = r.heading4Size
		case 5:
			headingSize = r.heading5Size
		case 6:
			headingSize = r.heading6Size
		default:
			headingSize = float64(r.fontSize)
		}
		r.pushFont(&Font{"bold", "B", int(math.Round(headingSize))})
	} else {
		r.popFont()
		r.pdf.SetY(r.pdf.GetY() + 6)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderHeadingC8hMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderList(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
		r.pdf.SetY(r.pdf.GetY() + 4)
		nestedLevel := r.countParentContainerBlocks(node)
		indent := float64(nestedLevel * 16)
		r.pdf.SetX(r.pdf.GetX() + indent)
	} else {
		r.pdf.SetY(r.pdf.GetY() + 4)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderListItem(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		if node.Parent.FirstChild != node {
			nestedLevel := r.countParentContainerBlocks(node) - 1
			indent := float64(nestedLevel * 16)
			r.pdf.SetX(r.pdf.GetX() + indent)
		}

		if 3 == node.ListData.Typ && "" != r.Options.GFMTaskListItemClass &&
			nil != node.FirstChild && nil != node.FirstChild.FirstChild && ast.NodeTaskListItemMarker == node.FirstChild.FirstChild.Type {
			r.WriteString(fmt.Sprintf("%s", node.ListData.Marker))
		} else {
			if 0 != node.ListData.BulletChar {
				r.WriteString("● ")
			} else {
				r.WriteString(fmt.Sprint(node.ListData.Num) + ". ")
			}
		}
	} else {
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderTaskListItemMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		var attrs [][]string
		if node.TaskListItemChecked {
			attrs = append(attrs, []string{"checked", ""})
		}
		attrs = append(attrs, []string{"disabled", ""}, []string{"type", "checkbox"})
		//r.tag("input", attrs, true)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderThematicBreak(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
		r.pdf.SetY(r.pdf.GetY() + 14)
		r.pdf.SetStrokeColor(106, 115, 125)
		r.pdf.Line(r.pdf.GetX()+float64(r.fontSize), r.pdf.GetY(), r.pageSize.W-r.margin-float64(r.fontSize), r.pdf.GetY())
		r.pdf.SetY(r.pdf.GetY() + 12)
		r.pdf.SetStrokeColor(0, 0, 0)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderHardBreak(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderSoftBreak(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) pushX(x float64) {
	r.x = append(r.x, x)
}

func (r *PdfRenderer) popX() float64 {
	ret := r.x[len(r.x)-1]
	r.x = r.x[:len(r.x)-1]
	return ret
}

func (r *PdfRenderer) pushFont(font *Font) {
	r.fonts = append(r.fonts, font)
}

func (r *PdfRenderer) popFont() *Font {
	ret := r.fonts[len(r.fonts)-1]
	r.fonts = r.fonts[:len(r.fonts)-1]
	return ret
}

func (r *PdfRenderer) peekFont() *Font {
	return r.fonts[len(r.fonts)-1]
}

func (r *PdfRenderer) pushTextColor(textColor *RGB) {
	r.textColors = append(r.textColors, textColor)
}

func (r *PdfRenderer) popTextColor() *RGB {
	ret := r.textColors[len(r.textColors)-1]
	r.textColors = r.textColors[:len(r.textColors)-1]
	return ret
}

func (r *PdfRenderer) peekTextColor() *RGB {
	return r.textColors[len(r.textColors)-1]
}

func (r *PdfRenderer) countParentContainerBlocks(n *ast.Node) (ret int) {
	for parent := n.Parent; nil != parent; parent = parent.Parent {
		if ast.NodeBlockquote == parent.Type || ast.NodeList == parent.Type {
			ret++
		}
	}
	return
}

// WriteByte 输出一个字节 c。
func (r *PdfRenderer) WriteByte(c byte) {
	r.WriteString(string(c))
}

// Write 输出指定的字节数组 content。
func (r *PdfRenderer) Write(content []byte) {
	r.WriteString(util.BytesToStr(content))
}

// WriteString 输出指定的字符串 content。
func (r *PdfRenderer) WriteString(content string) {
	// PDF 引擎处理不了不换行空格（NBSP）
	content = strings.ReplaceAll(content, " ", " ")

	if length := len(content); 0 < length {
		buf := bytes.Buffer{}
		x := r.pdf.GetX()
		startX := x
		runes := []rune(content)
		pageRight := r.pageSize.W - r.margin
		font := r.peekFont()
		if nil != font {
			r.pdf.SetFont(font.family, font.style, font.size)
		}
		textColor := r.peekTextColor()
		if nil != textColor {
			r.pdf.SetTextColor(textColor.R, textColor.G, textColor.B)
		}

		for i, c := range runes {
			if r.pdf.GetY() > r.pageSize.H-r.margin*2 {
				r.addPage()
			}

			if '\n' == c {
				if 0 < buf.Len() {
					r.pdf.Cell(nil, buf.String())
					buf.Reset()
				}

				r.pdf.Br(float64(r.fontSize) + 2)
				r.pdf.SetX(startX)
				x = startX
				continue
			}

			width, _ := r.pdf.MeasureTextWidth(string(c))
			if i < len(runes)-1 {
				nextC := runes[i+1]
				nextWidth, _ := r.pdf.MeasureTextWidth(string(nextC))
				if x+width+nextWidth > pageRight {
					r.pdf.Cell(nil, buf.String())
					buf.Reset()
					r.pdf.Br(float64(r.fontSize) + 2)
					x = r.pdf.GetX()
					continue
				}
			}

			buf.WriteRune(c)
			x += width
		}
		if 0 < buf.Len() {
			if r.pdf.GetY() > r.pageSize.H-r.margin*2 {
				r.addPage()
			}
			r.pdf.Cell(nil, buf.String())
		}

		r.LastOut = content[length-1]
	}
}

// Newline 会在最新内容不是换行符 \n 时输出一个换行符。
func (r *PdfRenderer) Newline() {
	if lex.ItemNewline != r.LastOut {
		r.pdf.Br(r.lineHeight)
		r.LastOut = lex.ItemNewline
	}
}

func (r *PdfRenderer) downloadImg(src string) (localPath string, ok, isTemp bool) {
	if strings.HasPrefix(src, "//") {
		src = "https:" + src
	}

	u, err := url.Parse(src)
	if nil != err {
		logger.Infof("image src [%s] is not an valid URL, treat it as local path", src)
		return src, true, false
	}

	if !strings.HasPrefix(u.Scheme, "http") {
		logger.Infof("image src [%s] scheme is not [http] or [https], treat it as local path", src)
		return src, true, false
	}

	src = r.qiniuImgProcessing(src)
	u, _ = url.Parse(src)

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req := &http.Request{
		Header: http.Header{
			"User-Agent": []string{"Lute-PDF; +https://github.com/88250/lute-pdf"},
		},
		URL: u,
	}
	resp, err := client.Do(req)
	if nil != err {
		logger.Warnf("download image [%s] failed: %s", src, err)
		return src, false, false
	}
	defer resp.Body.Close()
	if 200 != resp.StatusCode {
		logger.Warnf("download image [%s] failed, status code is [%d]", src, resp.StatusCode)
		return src, false, false
	}

	data, err := ioutil.ReadAll(resp.Body)
	file, err := ioutil.TempFile("", "lute-pdf.img.")
	if nil != err {
		logger.Warnf("create temp image [%s] failed: %s", src, err)
		return src, false, false
	}
	_, err = file.Write(data)
	if nil != err {
		logger.Warnf("write temp image [%s] failed: %s", src, err)
		return src, false, false
	}
	file.Close()
	return file.Name(), true, true
}

// qiniuImgProcessing 七牛云图片样式处理。
func (r *PdfRenderer) qiniuImgProcessing(src string) string {
	if !strings.Contains(src, "img.hacpai.com") && !strings.Contains(src, "b3logfile.com") && !strings.Contains(src, "imageView") {
		return src
	}

	if 0 < strings.Index(src, "?") {
		src = src[:strings.Index(src, "?")]
	}

	maxWidth := int(math.Round(r.pageSize.W-r.margin*2) * 128 / 72)
	style := "imageView2/2/w/%d/interlace/1/format/jpg"
	style = fmt.Sprintf(style, maxWidth)
	src += "?" + style
	return src
}

func (r *PdfRenderer) getImgSize(imgPath string) (width, height float64) {
	file, err := os.Open(imgPath)
	if nil != err {
		logger.Fatal(err)
	}
	img, _, err := image.Decode(file)
	if nil != err {
		logger.Fatal(err)
	}
	file.Close()

	imageRect := img.Bounds()
	k := 1
	w := -128
	h := -128
	if w < 0 {
		w = -imageRect.Dx() * 72 / w / k
	}
	if h < 0 {
		h = -imageRect.Dy() * 72 / h / k
	}
	if w == 0 {
		w = h * imageRect.Dx() / imageRect.Dy()
	}
	if h == 0 {
		h = w * imageRect.Dy() / imageRect.Dx()
	}
	return float64(w), float64(h)
}

func (r *PdfRenderer) addPage() {
	r.renderFooter()
	r.pdf.AddPage()
}

func (r *PdfRenderer) renderFooter() {
	if 0 < len(r.FootnotesDefs) {
		return
	}
	footer := r.Cover.LinkLabel + r.Cover.Title
	r.pdf.SetFont("regular", "R", 8)
	r.pdf.SetTextColor(0, 0, 0)
	labelWidth, _ := r.pdf.MeasureTextWidth(r.Cover.LinkLabel)
	width, _ := r.pdf.MeasureTextWidth(footer)
	x := r.pageSize.W - r.margin - width
	r.pdf.SetX(x)
	y := r.pageSize.H - r.margin
	r.pdf.SetY(y)
	r.pdf.Cell(nil, r.Cover.LinkLabel)

	r.pdf.SetTextColor(66, 133, 244)
	r.pdf.Cell(nil, r.Cover.Title)
	r.pdf.AddExternalLink(r.Cover.Link, x+labelWidth, y, width-labelWidth, 8)

	font := r.peekFont()
	r.pdf.SetFont(font.family, font.style, font.size)
	textColor := r.peekTextColor()
	r.pdf.SetTextColor(textColor.R, textColor.G, textColor.B)
}

type Font struct {
	family string
	style  string // R|B|I|U
	size   int
}

type RGB struct {
	R, G, B uint8
}
