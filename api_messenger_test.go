package feishu

import (
	"strings"
	"testing"
)

func Test_app_SendMessage(t *testing.T) {
	fsApp := testNewCustomApp()

	// _, err := fsApp.getAppAccessTokenWithContext(context.Background())
	// requireNil(t, err)

	fsApp.opt.debug = true

	var (
		receiver = MessageReceiver{
			IDType: ChatID,
			ID:     "oc_5f2c7c2066c6be483bb0302f2fa0c04f",
		}

		content = `first line
second line
` + StrMentionByOpenID("ou_invalid", "不存在的人") + "\n" + StrMentionAll()
		mdZhCn = `**title**
~~DEL~~
`
	)

	msg := NewMessageText(content)

	receiver.ID = "oc_7e63c85739629f0c0b4a868ca76f4d41"
	msg = NewMessagePost(
		WithPost(LangChinese, "🇨🇳我是一个标题",
			WithPostElementText("🇨🇳第一行："),
			WithPostElementLink("超链接", "https://www.feishu.cn"),
			WithPostElementImage("img_7ea74629-9191-4176-998c-2e603c9c5e8g"),
			WithPostElementText("\n"),
			WithPostElementMentionAll(),
			WithPostElementText("+"),
			WithPostElementMentionByOpenID("ou_c99c5f35d542efc7ee492afe11af19ef"),
			WithPostElementImage("img_ecffc3b9-8f14-400f-a014-05eca1a4310g"),
		),
		WithPost(LangEnglish, "🇺🇸🇬🇧 title",
			WithPostElementText("🇺🇸🇬🇧 first line："),
			WithPostElementLink("link", "https://www.feishu.cn"),
			WithPostElementImage("img_7ea74629-9191-4176-998c-2e603c9c5e8g"),
			WithPostElementText("\n"),
			WithPostElementMentionAll(),
			WithPostElementText("+"),
			WithPostElementMentionByOpenID("ou_c99c5f35d542efc7ee492afe11af19ef"),
			WithPostElementImage("img_ecffc3b9-8f14-400f-a014-05eca1a4310g"),
		),
		WithPost(LangJapanese, "🇯🇵 見出し",
			WithPostElementText("🇯🇵 最初の行："),
			WithPostElementLink("リンク", "https://www.feishu.cn"),
			WithPostElementImage("img_7ea74629-9191-4176-998c-2e603c9c5e8g"),
			WithPostElementText("\n"),
			WithPostElementMentionAll(),
			WithPostElementText("+"),
			WithPostElementMentionByOpenID("ou_c99c5f35d542efc7ee492afe11af19ef"),
			WithPostElementImage("img_ecffc3b9-8f14-400f-a014-05eca1a4310g"),
		),
	)

	receiver.ID = "oc_78b297d4a002835dd4eeafe6f83d7b69"
	msg = NewMessageImage("img_7ea74629-9191-4176-998c-2e603c9c5e8g")

	receiver.ID = "oc_67153cf1cbea58e0936e4ec72c18a268"
	msg = NewMessageShareChat("oc_78b297d4a002835dd4eeafe6f83d7b69")

	msg = NewMessageCard(BgColorOrange, WithCardConfig(WithCardConfigCardLink(
		"https://www.feishu.cn",
		"https://zlink.toutiao.com/kG12?apk=1",
		"https://zlink.toutiao.com/h2Sw",
		"https://www.feishu.cn/download",
	)),
		WithCard(LangChinese, "标题",
			WithCardElementPlainText("文本内容"),
			WithCardElementHorizontalRule(),
			WithCardElementPlainText(strings.Repeat("文本内容2", 20), 2),
			WithCardElementHorizontalRule(),
			WithCardElementMarkdown(mdZhCn),
			WithCardElementHorizontalRule(),
			WithCardElementImage("img_7ea74629-9191-4176-998c-2e603c9c5e8g",
				WithCardElementImageTitle("    *图片标题*", true),
				WithCardElementImageHover("被发现了"),
			),
			WithCardElementNote(
				WithCardElementPlainText("**普通文本**"),
				WithCardElementImage("img_7ea74629-9191-4176-998c-2e603c9c5e8g",
					WithCardElementImageTitle("    *图片标题*", true),
					WithCardElementImageHover("被发现了"),
				),
				WithCardElementMarkdown("*test*"),
			),
			WithCardElementFields(
				WithCardElementField(WithCardElementPlainText("列1\nv1"), true),
				WithCardElementField(WithCardElementMarkdown("**列2**\nv2"), true),
				WithCardElementField(WithCardElementMarkdown("~无效的信息~"), false),
			),
			WithCardElementActions(
				WithCardElementAction(WithCardElementPlainText("入门必读"), "https://www.feishu.cn/hc/zh-CN/articles/360024881814", WithCardElementActionButton(ButtonDefault)),
				WithCardElementAction(WithCardElementPlainText("快速习惯飞书️"), "https://www.feishu.cn/hc/zh-CN/categories-detail?category-id=7018450035717259265", WithCardElementActionButton(ButtonPrimary)),
				WithCardElementAction(
					WithCardElementMarkdown("**多端跳转下载**"), "", WithCardElementActionButton(ButtonDanger), WithCardElementActionMultiURL(
						"https://www.feishu.cn",
						"https://zlink.toutiao.com/kG12?apk=1",
						"https://zlink.toutiao.com/h2Sw",
						"https://www.feishu.cn/download",
					),
				),
			),

			WithCardElementMarkdown("*TEST*", WithCardExtraElementImage("img_7ea74629-9191-4176-998c-2e603c9c5e8g",
				WithCardElementImageTitle("    *图片标题*", true),
				WithCardElementImageHover("被发现了"),
			)),
		),

		WithCard(LangEnglish, "title",
			WithCardElementMarkdown("~~empty~~"),
		),
	)

	msg = NewMessageAudio("75235e0c-4f92-430a-a99b-8446610223cg")
	msg = NewMessageMedia("75235e0c-4f92-430a-a99b-8446610223cg", "img_7ea74629-9191-4176-998c-2e603c9c5e8g")
	msg = NewMessageFile("75235e0c-4f92-430a-a99b-8446610223cg")
	msg = NewMessageSticker("img_7ea74629-9191-4176-998c-2e603c9c5e8g")

	msgDetail, err := fsApp.SendMessage(receiver, msg)
	requireNil(t, err)

	logIndent(t, msgDetail)
}

func Test_app_ReplyMessage(t *testing.T) {
	fsApp := testNewCustomApp()
	fsApp.opt.debug = true

	msgDetail, err := fsApp.ReplyMessage("om_d85801e64eb4135954cd657d4e11c8df", NewMessageText("ok"))
	requireNil(t, err)

	logIndent(t, msgDetail)
}
