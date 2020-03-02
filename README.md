# Lute PDF

Lute PDF 是一款通过 Markdown 生成 PDF 的小工具。通过 [Lute](https://github.com/88250/lute) 解析 Markdown 然后再通过 [gopdf](github.com/signintech/gopdf) 生成 PDF。

## 特性

* 几乎支持所有 Markdown 语法元素
* 图片会通过地址自动拉取并渲染
* 支持封面配置

## 已知问题

* 没有代码高亮，代码块统一使用绿色渲染
* 没有渲染 Emoji
* 表格没有边框
* 表格单元格折行计算有问题
* 粗体、斜体需要字体本身支持

## 