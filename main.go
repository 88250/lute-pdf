package main

import (
	"log"
	"math"

	"github.com/88250/lute/parse"

	"github.com/signintech/gopdf"
)

func main() {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4}) // 595.28, 841.89 = A4
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

	const factor = 0.8
	fontSize := 16 * factor
	x := 16.0 * factor
	y := 24.0 * factor

	const lineHeight = 24.0 * factor
	const heading2Size = 24 * factor

	if err = pdf.SetFontWithStyle("msyhb", gopdf.Bold, int(math.Round(heading2Size))); nil != err {
		log.Fatal(err)
	}
	pdf.SetX(x)
	pdf.SetY(y)
	_ = pdf.Cell(nil, "Guide")
	y += heading2Size + lineHeight
	pdf.SetX(x)
	pdf.SetY(y)

	pdf.SetFontWithStyle("msyh", gopdf.Regular, int(fontSize))
	pdf.SetFont("msyh", "", int(fontSize))

	markdown := []byte("这是一篇讲解如何正确使用 *Markdown* 的排版示例，学会这个很有必要，能让你的文章有更佳清晰的排版。")
	tree, err := parse.Parse("", markdown, &parse.Options{})
	if nil != err {
		log.Fatal(err)
	}
	renderer := NewPdfRenderer(tree, pdf)
	output, err := renderer.Render()
	if nil != err {
		log.Fatal(err)
	}
	text := string(output)

	rect := &gopdf.Rect{W: gopdf.PageSizeA4.W - x*2, H: gopdf.PageSizeA4.H}
	pdf.MultiCell(rect, text)

	pdf.WritePdf("sample.pdf")

}
