package main

import (
	"fmt"
	"github.com/88250/lute/ast"
	"github.com/88250/lute/lex"
	"github.com/88250/lute/parse"
	"github.com/88250/lute/render"
	"github.com/88250/lute/util"
	"github.com/signintech/gopdf"
	"log"
	"math"
	"strconv"
	"strings"
)

// PdfRenderer 描述了 PDF 渲染器。
type PdfRenderer struct {
	*render.BaseRenderer
	needRenderFootnotesDef bool
	headingCnt             int

	pdf          *gopdf.GoPdf // PDF 生成器句柄
	pageSize     *gopdf.Rect  // 页面大小
	factor       float64      // 字体、行高大小倍数
	fontSize     float64      // 字体大小
	lineHeight   float64      // 行高
	heading1Size float64      // 一级标题大小
	heading2Size float64      // 二级标题大小
	heading3Size float64      // 三级标题大小
	heading4Size float64      // 四级标题大小
	heading5Size float64      // 五级标题大小
	heading6Size float64      // 六级标题大小
	margin       float64      // 页边距
	x            []float64    // 当前横坐标栈
}

// NewPdfRenderer 创建一个 HTML 渲染器。
func NewPdfRenderer(tree *parse.Tree) render.Renderer {
	pdf := &gopdf.GoPdf{}

	ret := &PdfRenderer{BaseRenderer: render.NewBaseRenderer(tree), needRenderFootnotesDef: false, headingCnt: 0, pdf: pdf}
	ret.factor = 0.8
	ret.fontSize = 14 * ret.factor
	ret.lineHeight = 24.0 * ret.factor
	ret.heading1Size = 24 * ret.factor
	ret.heading2Size = 22 * ret.factor
	ret.heading3Size = 20 * ret.factor
	ret.heading4Size = 18 * ret.factor
	ret.heading5Size = 16 * ret.factor
	ret.heading6Size = 14 * ret.factor
	ret.margin = 30 * ret.factor
	pdf.SetX(ret.margin)
	pdf.SetY(ret.margin)

	ret.pageSize = gopdf.PageSizeA4
	pdf.Start(gopdf.Config{PageSize: *ret.pageSize})
	pdf.AddPage()

	var err error
	err = pdf.AddTTFFont("msyh", "fonts/msyh.ttf")
	if err != nil {
		log.Fatal(err)
	}

	err = pdf.AddTTFFontWithOption("msyhb", "fonts/msyhb.ttf", gopdf.TtfOption{Style: gopdf.Bold})
	if err != nil {
		log.Fatal(err)
	}

	pdf.SetFontWithStyle("msyh", gopdf.Regular, int(ret.fontSize))
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
	ret.RendererFuncs[ast.NodeFootnotesDef] = ret.renderFootnotesDef
	ret.RendererFuncs[ast.NodeFootnotesRef] = ret.renderFootnotesRef
	ret.RendererFuncs[ast.NodeToC] = ret.renderToC
	ret.RendererFuncs[ast.NodeBackslash] = ret.renderBackslash
	ret.RendererFuncs[ast.NodeBackslashContent] = ret.renderBackslashContent
	return ret
}

func (r *PdfRenderer) renderBackslashContent(node *ast.Node, entering bool) ast.WalkStatus {
	r.Write(util.EscapeHTML(node.Tokens))
	return ast.WalkStop
}

func (r *PdfRenderer) renderBackslash(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderToC(node *ast.Node, entering bool) ast.WalkStatus {
	headings := r.headings()
	length := len(headings)
	if 1 > length {
		return ast.WalkStop
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

	return ast.WalkStop
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
	r.WriteString("<div class=\"footnotes-defs-div\">")
	r.WriteString("<hr class=\"footnotes-defs-hr\" />\n")
	r.WriteString("<ol class=\"footnotes-defs-ol\">")
	for i, def := range context.FootnotesDefs {
		r.WriteString("<li id=\"footnotes-def-" + strconv.Itoa(i+1) + "\">")
		tree := &parse.Tree{Name: "", Context: context}
		tree.Context.Tree = tree
		tree.Root = &ast.Node{Type: ast.NodeDocument}
		tree.Root.AppendChild(def)
		defRenderer := NewPdfRenderer(tree)
		lc := tree.Root.LastDeepestChild()
		for i = len(def.FootnotesRefs) - 1; 0 <= i; i-- {
			ref := def.FootnotesRefs[i]
			gotoRef := " <a href=\"#footnotes-ref-" + ref.FootnotesRefId + "\" class=\"footnotes-goto-ref\">↩</a>"
			link := &ast.Node{Type: ast.NodeInlineHTML, Tokens: util.StrToBytes(gotoRef)}
			lc.InsertAfter(link)
		}
		defRenderer.(*PdfRenderer).needRenderFootnotesDef = true
		defContent, err := defRenderer.Render()
		if nil != err {
			break
		}
		r.Write(defContent)

		r.WriteString("</li>\n")
	}
	r.WriteString("</ol></div>")
	return r.Writer.Bytes()
}

func (r *PdfRenderer) renderFootnotesRef(node *ast.Node, entering bool) ast.WalkStatus {
	idx, _ := r.Tree.Context.FindFootnotesDef(node.Tokens)
	idxStr := strconv.Itoa(idx)
	//r.tag("sup", [][]string{{"class", "footnotes-ref"}, {"id", "footnotes-ref-" + node.FootnotesRefId}}, false)
	//r.tag("a", [][]string{{"href", "#footnotes-def-" + idxStr}}, false)
	r.WriteString(idxStr)
	//r.tag("/a", nil, false)
	//r.tag("/sup", nil, false)
	return ast.WalkStop
}

func (r *PdfRenderer) renderFootnotesDef(node *ast.Node, entering bool) ast.WalkStatus {
	if !r.needRenderFootnotesDef {
		return ast.WalkStop
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderCodeBlockCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderCodeBlockInfoMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderCodeBlockOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderEmojiAlias(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderEmojiImg(node *ast.Node, entering bool) ast.WalkStatus {
	// TODO: r.Write(node.Tokens)
	return ast.WalkStop
}

func (r *PdfRenderer) renderEmojiUnicode(node *ast.Node, entering bool) ast.WalkStatus {
	// TODO: r.Write(node.Tokens)
	return ast.WalkStop
}

func (r *PdfRenderer) renderEmoji(node *ast.Node, entering bool) ast.WalkStatus {
	// TODO: Render Emoji
	return ast.WalkStop
}

func (r *PdfRenderer) renderInlineMathCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	//r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *PdfRenderer) renderInlineMathContent(node *ast.Node, entering bool) ast.WalkStatus {
	r.Write(util.EscapeHTML(node.Tokens))
	return ast.WalkStop
}

func (r *PdfRenderer) renderInlineMathOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	//attrs := [][]string{{"class", "vditor-math"}}
	//r.tag("span", attrs, false)
	return ast.WalkStop
}

func (r *PdfRenderer) renderInlineMath(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *PdfRenderer) renderMathBlockCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	//r.tag("/div", nil, false)
	return ast.WalkStop
}

func (r *PdfRenderer) renderMathBlockContent(node *ast.Node, entering bool) ast.WalkStatus {
	r.Write(util.EscapeHTML(node.Tokens))
	return ast.WalkStop
}

func (r *PdfRenderer) renderMathBlockOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	//attrs := [][]string{{"class", "vditor-math"}}
	//r.tag("div", attrs, false)
	return ast.WalkStop
}

func (r *PdfRenderer) renderMathBlock(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	return ast.WalkContinue
}

func (r *PdfRenderer) renderTableCell(node *ast.Node, entering bool) ast.WalkStatus {
	//tag := "td"
	if ast.NodeTableHead == node.Parent.Parent.Type {
		//tag = "th"
	}
	if entering {
		var attrs [][]string
		switch node.TableCellAlign {
		case 1:
			attrs = append(attrs, []string{"align", "left"})
		case 2:
			attrs = append(attrs, []string{"align", "center"})
		case 3:
			attrs = append(attrs, []string{"align", "right"})
		}
		//r.tag(tag, attrs, false)
	} else {
		//r.tag("/"+tag, nil, false)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderTableRow(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		//r.tag("tr", nil, false)
		r.Newline()
	} else {
		//r.tag("/tr", nil, false)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderTableHead(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		//r.tag("thead", nil, false)
		r.Newline()
	} else {
		//r.tag("/thead", nil, false)
		r.Newline()
		if nil != node.Next {
			//r.tag("tbody", nil, false)
		}
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderTable(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		//r.tag("table", nil, false)
		r.Newline()
	} else {
		if nil != node.FirstChild.Next {
			//r.tag("/tbody", nil, false)
		}
		r.Newline()
		//r.tag("/table", nil, false)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrikethrough(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		// TODO: r.TextAutoSpacePrevious(node)
	} else {
		// TODO: r.TextAutoSpaceNext(node)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrikethrough1OpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.pushX(r.pdf.GetX())
	return ast.WalkStop
}

func (r *PdfRenderer) renderStrikethrough1CloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	x := r.popX()
	r.pdf.Line(x, r.pdf.GetY()+r.fontSize/2, r.pdf.GetX(), r.pdf.GetY()+r.fontSize/2)
	return ast.WalkStop
}

func (r *PdfRenderer) renderStrikethrough2OpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.pushX(r.pdf.GetX())
	return ast.WalkStop
}

func (r *PdfRenderer) renderStrikethrough2CloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	x := r.popX()
	r.pdf.Line(x, r.pdf.GetY()+r.fontSize/2, r.pdf.GetX(), r.pdf.GetY()+r.fontSize/2)
	return ast.WalkStop
}

func (r *PdfRenderer) renderLinkTitle(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderLinkDest(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderLinkSpace(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderLinkText(node *ast.Node, entering bool) ast.WalkStatus {
	if r.Option.AutoSpace {
		r.Space(node)
	}
	r.Write(util.EscapeHTML(node.Tokens))
	return ast.WalkStop
}

func (r *PdfRenderer) renderCloseParen(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderOpenParen(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderCloseBracket(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderOpenBracket(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderBang(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderImage(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		if 0 == r.DisableTags {
			r.WriteString("<img src=\"")
			destTokens := node.ChildByType(ast.NodeLinkDest).Tokens
			destTokens = r.Tree.Context.RelativePath(destTokens)
			r.Write(util.EscapeHTML(destTokens))
			r.WriteString("\" alt=\"")
		}
		r.DisableTags++
		return ast.WalkContinue
	}

	r.DisableTags--
	if 0 == r.DisableTags {
		r.WriteString("\"")
		if title := node.ChildByType(ast.NodeLinkTitle); nil != title && nil != title.Tokens {
			r.WriteString(" title=\"")
			r.Write(util.EscapeHTML(title.Tokens))
			r.WriteString("\"")
		}
		r.WriteString(" />")
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderLink(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		// TODO: r.LinkTextAutoSpacePrevious(node)
		r.pushX(r.pdf.GetX())
		r.pdf.SetTextColor(66, 133, 244)
	} else {
		x := r.popX()
		width := r.pdf.GetX() - x
		dest := node.ChildByType(ast.NodeLinkDest)
		destTokens := dest.Tokens
		destTokens = r.Tree.Context.RelativePath(destTokens)
		r.pdf.AddExternalLink(util.BytesToStr(util.EscapeHTML(destTokens)), x, r.pdf.GetY(), width, r.lineHeight)
		// TODO: r.LinkTextAutoSpaceNext(node)
		r.pdf.SetTextColor(0, 0, 0)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderHTML(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	r.Write(node.Tokens)
	r.Newline()
	return ast.WalkStop
}

func (r *PdfRenderer) renderInlineHTML(node *ast.Node, entering bool) ast.WalkStatus {
	r.Write(node.Tokens)
	return ast.WalkStop
}

func (r *PdfRenderer) renderDocument(node *ast.Node, entering bool) ast.WalkStatus {
	if !entering {
		if err := r.pdf.WritePdf(r.Tree.Name + ".pdf"); nil != err {
			log.Fatal(err)
		}
		if err := r.pdf.Close(); nil != err {
			log.Fatal(err)
		}
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderParagraph(node *ast.Node, entering bool) ast.WalkStatus {
	if grandparent := node.Parent.Parent; nil != grandparent && ast.NodeList == grandparent.Type && grandparent.Tight { // List.ListItem.Paragraph
		return ast.WalkContinue
	}

	if entering {
		r.Newline()
		r.pdf.SetY(r.pdf.GetY() + 6)
	} else {
		r.pdf.SetY(r.pdf.GetY() + 6)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderText(node *ast.Node, entering bool) ast.WalkStatus {
	if r.Option.AutoSpace {
		r.Space(node)
	}
	if r.Option.FixTermTypo {
		r.FixTermTypo(node)
	}
	if r.Option.ChinesePunct {
		r.ChinesePunct(node)
	}

	text := util.BytesToStr(util.EscapeHTML(node.Tokens))
	width := gopdf.PageSizeA4.W - r.margin - r.pdf.GetX()
	if 0 > width {
		width = gopdf.PageSizeA4.W - r.margin
	}
	lines, _ := r.pdf.SplitText(text, width)
	isMultiLine := 1 < len(lines)
	for _, line := range lines {
		r.WriteString(line)
		if isMultiLine {
			r.Newline()
		}
	}
	return ast.WalkStop
}

func (r *PdfRenderer) renderCodeSpan(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		//if r.Option.AutoSpace {
		//	if text := node.PreviousNodeText(); "" != text {
		//		lastc, _ := utf8.DecodeLastRuneInString(text)
		//		if unicode.IsLetter(lastc) || unicode.IsDigit(lastc) {
		//			r.WriteByte(lex.ItemSpace)
		//		}
		//	}
		//}
	} else {
		//if r.Option.AutoSpace {
		//	if text := node.NextNodeText(); "" != text {
		//		firstc, _ := utf8.DecodeRuneInString(text)
		//		if unicode.IsLetter(firstc) || unicode.IsDigit(firstc) {
		//			r.WriteByte(lex.ItemSpace)
		//		}
		//	}
		//}
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderCodeSpanOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderCodeSpanContent(node *ast.Node, entering bool) ast.WalkStatus {
	content := util.BytesToStr(util.EscapeHTML(node.Tokens))
	width, _ := r.pdf.MeasureTextWidth(content)
	r.pdf.SetFillColor(227, 236, 245)
	r.pdf.RectFromUpperLeftWithStyle(r.pdf.GetX(), r.pdf.GetY(), width, r.fontSize, "F")
	r.pdf.SetFillColor(0, 0, 0)
	r.WriteString(content)
	return ast.WalkStop
}

func (r *PdfRenderer) renderCodeSpanCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderEmphasis(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		// TODO: r.TextAutoSpacePrevious(node)
	} else {
		// TODO: r.TextAutoSpaceNext(node)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderEmAsteriskOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.pdf.SetFontWithStyle("msyh", gopdf.Italic, int(r.fontSize))
	return ast.WalkStop
}

func (r *PdfRenderer) renderEmAsteriskCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.pdf.SetFontWithStyle("msyh", gopdf.Regular, int(r.fontSize))
	return ast.WalkStop
}

func (r *PdfRenderer) renderEmUnderscoreOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.pdf.SetFontWithStyle("msyh", gopdf.Italic, int(r.fontSize))
	return ast.WalkStop
}

func (r *PdfRenderer) renderEmUnderscoreCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.pdf.SetFontWithStyle("msyh", gopdf.Regular, int(r.fontSize))
	return ast.WalkStop
}

func (r *PdfRenderer) renderStrong(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		// TODO: r.TextAutoSpacePrevious(node)
	} else {
		// TODO: r.TextAutoSpaceNext(node)
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderStrongA6kOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.pdf.SetFontWithStyle("msyhb", gopdf.Bold, int(r.fontSize))
	return ast.WalkStop
}

func (r *PdfRenderer) renderStrongA6kCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.pdf.SetFontWithStyle("msyh", gopdf.Regular, int(r.fontSize))
	return ast.WalkStop
}

func (r *PdfRenderer) renderStrongU8eOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.pdf.SetFontWithStyle("msyhb", gopdf.Bold, int(r.fontSize))
	return ast.WalkStop
}

func (r *PdfRenderer) renderStrongU8eCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.pdf.SetFontWithStyle("msyh", gopdf.Regular, int(r.fontSize))
	return ast.WalkStop
}

func (r *PdfRenderer) renderBlockquote(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
		r.pdf.SetTextColor(106, 115, 125)
		r.pushX(r.pdf.GetX())
	} else {
		x := r.popX()
		r.pdf.SetX(r.pdf.GetX() - x + r.margin)
		r.pdf.SetTextColor(0, 0, 0)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderBlockquoteMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *PdfRenderer) renderHeading(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
		r.pdf.SetY(r.pdf.GetY() + 6)
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
			headingSize = r.fontSize
		}

		r.pdf.SetFontWithStyle("msyhb", gopdf.Bold, int(math.Round(headingSize)))

		// TODO: ToC
		//if r.Option.ToC {
		//	r.WriteString(" id=\"toc_h" + fmt.Sprint(node.HeadingLevel) + "_" + strconv.Itoa(r.headingCnt) + "\"")
		//	r.headingCnt++
		//}
		// TODO: HeadingAnchor
		//if r.Option.HeadingAnchor {
		//	anchor := node.Text()
		//	anchor = strings.ReplaceAll(anchor, " ", "-")
		//	anchor = strings.ReplaceAll(anchor, ".", "")
		//	r.tag("a", [][]string{{"id", "vditorAnchor-" + anchor}, {"class", "vditor-anchor"}, {"href", "#" + anchor}}, false)
		//	r.WriteString(`<svg viewBox="0 0 16 16" version="1.1" width="16" height="16"><path fill-rule="evenodd" d="M4 9h1v1H4c-1.5 0-3-1.69-3-3.5S2.55 3 4 3h4c1.45 0 3 1.69 3 3.5 0 1.41-.91 2.72-2 3.25V8.59c.58-.45 1-1.27 1-2.09C10 5.22 8.98 4 8 4H4c-.98 0-2 1.22-2 2.5S3 9 4 9zm9-3h-1v1h1c1 0 2 1.22 2 2.5S13.98 12 13 12H9c-.98 0-2-1.22-2-2.5 0-.83.42-1.64 1-2.09V6.25c-1.09.53-2 1.84-2 3.25C6 11.31 7.55 13 9 13h4c1.45 0 3-1.69 3-3.5S14.5 6 13 6z"></path></svg>`)
		//	r.tag("/a", nil, false)
		//}
	} else {
		r.pdf.SetFontWithStyle("msyh", gopdf.Regular, int(r.fontSize))
		r.pdf.SetY(r.pdf.GetY() + 6)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *PdfRenderer) renderHeadingC8hMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
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

		if 3 == node.ListData.Typ && "" != r.Option.GFMTaskListItemClass &&
			nil != node.FirstChild && nil != node.FirstChild.FirstChild && ast.NodeTaskListItemMarker == node.FirstChild.FirstChild.Type {
			r.WriteString(fmt.Sprintf("%s", node.ListData.Marker))
		} else {
			if 0 != node.BulletChar {
				r.WriteString("● ")
			} else {
				r.WriteString(fmt.Sprint(node.Num) + ". ")
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
	r.Newline()
	//r.tag("hr", nil, true)
	r.Newline()
	return ast.WalkStop
}

func (r *PdfRenderer) renderHardBreak(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	return ast.WalkStop
}

func (r *PdfRenderer) renderSoftBreak(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	return ast.WalkStop
}

func (r *PdfRenderer) pushX(x float64) {
	r.x = append(r.x, x)
}

func (r *PdfRenderer) popX() float64 {
	ret := r.x[len(r.x)-1]
	r.x = r.x[:len(r.x)]
	return ret
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
	if length := len(content); 0 < length {
		if r.pdf.GetY() > r.pageSize.H-r.margin*2 {
			r.pdf.AddPage()
		}
		r.pdf.Cell(nil, content)
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
