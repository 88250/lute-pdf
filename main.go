package main

import (
	"log"

	"github.com/88250/lute/parse"
)

func main() {
	markdown := []byte(sample)
	tree, err := parse.Parse("sample", markdown, &parse.Options{
		GFMTable:            true,
		GFMTaskListItem:     true,
		GFMStrikethrough:    true,
		GFMAutoLink:         true,
		SoftBreak2HardBreak: true})
	if nil != err {
		log.Fatal(err)
	}
	renderer := NewPdfRenderer(tree)
	_, err = renderer.Render()
	if nil != err {
		log.Fatal(err)
	}
}

const sample = `Vditor 是一款**所见即所得**编辑器，支持 *Markdown*。

* 不熟悉 Markdown 可使用工具栏或快捷键进行排版
* 熟悉 Markdown 可直接排版，也可切换为分屏预览

更多细节和用法请参考 [Vditor - 浏览器端的 Markdown 编辑器](https://hacpai.com/article/1549638745630)，同时也欢迎向我们提出建议或报告问题，谢谢

## Guide

这是一篇讲解如何正确使用 **Markdown** 的排版示例，学会这个很有必要，能让你的文章有更佳清晰的排版。

> 引用文本：Markdown is a text formatting syntax inspired

## 语法指导

### 普通内容

这段内容展示了在内容里面一些排版格式，比如：

- **加粗** - ` + "`**加粗**`" + `
- *倾斜* - ` + "`*倾斜*`" + `
- ~~删除线~~ - ` + "`~~删除线~~`" + `
- ` + "`Code 标记` - " + "`` `Code 标记` ``" + `
- [超级链接](https://hacpai.com) - ` + "`[超级链接](https://hacpai.com)`" + `
- [username@gmail.com](mailto:username@gmail.com) - ` + "`[username@gmail.com](mailto:username@gmail.com)`" + `

### 提及用户

@Vanessa 通过 ` + "`@User`" + ` 可以在内容中提及用户，被提及的用户将会收到系统通知。
`
