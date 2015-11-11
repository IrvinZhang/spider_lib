package spider_lib

// 基础包
import (
	"github.com/PuerkitoBio/goquery"                        //DOM解析
	"github.com/henrylee2cn/pholcus/app/downloader/context" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	. "github.com/henrylee2cn/pholcus/app/spider/common"    //选用
	"github.com/henrylee2cn/pholcus/logs"                   //信息输出

	// net包
	"net/http" //设置http.Header
	// "net/url"

	// 编码包
	// "encoding/xml"
	// "encoding/json"

	// 字符串处理包
	// "regexp"
	"strconv"
	"strings"

	// 其他包
	// "fmt"
	// "math"
	// "time"
	// "io/ioutil"
)

func init() {
	WeiboFans.Register()
}

var WeiboFans = &Spider{
	Name:         "微博粉丝列表",
	Description:  `新浪微博粉丝 [自定义输入格式 "ID"::"Cookie"][最多支持250页，内设定时1~2s]`,
	Pausetime:    [2]uint{uint(1000), uint(1000)},
	Keyword:      USE,
	EnableCookie: true,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			param := strings.Split(ctx.GetKeyword(), "::")
			if len(param) != 2 {
				logs.Log.Error("自定义输入的参数不正确！")
				return
			}
			id := strings.Trim(param[0], " ")
			cookie := strings.Trim(param[1], " ")

			var count1 = 250
			var count2 = 50
			if ctx.GetMaxPage() < count1 {
				count1 = ctx.GetMaxPage()
			}
			if ctx.GetMaxPage() < count2 {
				count2 = ctx.GetMaxPage()
			}
			for i := count1; i > 0; i-- {
				ctx.AddQueue(&context.Request{
					Url:          "http://weibo.com/" + id + "/fans?cfs=600&relate=fans&t=1&f=1&type=&Pl_Official_RelationFans__68_page=" + strconv.Itoa(i) + "#Pl_Official_RelationFans__68",
					Rule:         "好友列表",
					Header:       http.Header{"Cookie": []string{cookie}},
					DownloaderID: 0,
				})
			}
			for i := 1; i <= count2; i++ {
				ctx.AddQueue(&context.Request{
					Url:          "http://www.weibo.com/" + id + "/fans?cfs=&relate=fans&t=5&f=1&type=&Pl_Official_RelationFans__68_page=" + strconv.Itoa(i) + "#Pl_Official_RelationFans__68",
					Rule:         "好友列表",
					Header:       http.Header{"Cookie": []string{cookie}},
					DownloaderID: 0,
				})
			}
		},

		Trunk: map[string]*Rule{
			"好友列表": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					query.Find(".follow_list .mod_info").Each(func(i int, s *goquery.Selection) {
						name, _ := s.Find(".info_name a").Attr("title")
						url, _ := s.Find(".info_name a").Attr("href")
						uid := strings.Replace(url, "/u", "", -1)
						uid = strings.Replace(uid, "/", "", -1)
						url = "http://weibo.com/p/100505" + uid + "/info?mod=pedit_more"
						var 认证 string = ""
						if _, isExist := s.Find(".info_name i").Attr("title"); isExist {
							认证 = "认证"
						}
						关注 := s.Find(".info_connect em a").Eq(0).Text()
						粉丝 := s.Find(".info_connect em a").Eq(1).Text()
						微博 := s.Find(".info_connect em a").Eq(2).Text()
						x := &context.Request{
							Url:          url,
							Rule:         "好友资料",
							DownloaderID: 0,
							Temp: map[string]interface{}{
								"好友名":  name,
								"好友ID": uid,
								"认证":   认证,
								"关注":   关注,
								"粉丝":   粉丝,
								"微博":   微博,
							},
						}
						ctx.AddQueue(x)
					})
				},
			},
			"好友资料": {
				OutFeild: []string{
					"好友名",
					"好友ID",
					"认证",
					"关注",
					"粉丝",
					"微博",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					var 属性 map[string]string
					var title string
					var detail string
					query.Find(".li_1").Each(func(i int, s *goquery.Selection) {
						if 属性 == nil {
							属性 = map[string]string{}
						}
						title = s.Find(".pt_title").Text()
						title = Deprive2(title)
						detail = s.Find(".pt_detail").Text()
						detail = Deprive2(detail)
						属性[title] = detail
					})
					临时 := map[int]interface{}{
						0: ctx.GetTemp("好友名"),
						1: ctx.GetTemp("好友ID"),
						2: ctx.GetTemp("认证"),
						3: ctx.GetTemp("关注"),
						4: ctx.GetTemp("粉丝"),
						5: ctx.GetTemp("微博"),
					}
					for k, v := range 属性 {
						idx := ctx.AddOutFeild(k)
						临时[idx] = v
					}

					// 结果存入Response中转
					ctx.Output(临时)
				},
			},
		},
	},
}
