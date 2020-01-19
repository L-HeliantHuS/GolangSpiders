package main

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

/*
	这个爬虫完全是无聊写着玩的， 之前用Python实现过，不过速度觉得并不快，于是用Go重写了一遍，速度起飞，代码结构以及注释并不标准QAQ
*/

// 初始化WaitGroup
var (
	wg sync.WaitGroup
)

// 用于匹配中文字符
var reg = regexp.MustCompile("^[\u4e00-\u9fa5]$")

// GetHTMLText 用于网络请求
func GetHTMLText(url string) []byte {

	request, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer request.Body.Close()

	htmls, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	return htmls
}

// GetImagePage 获取详细图片页
func GetImagePage(url string) {
	// URL响应的HTML
	response := GetHTMLText(url)

	// 获得每一个小分类的URL
	node, err := htmlquery.Parse(strings.NewReader(string(response)))
	if err != nil {
		panic("序列化Html到node时出错！")
	}

	nextUrl := htmlquery.Find(node, "//ul[@class='g-piclist-container']//li/div/div[@class='m-img-wrap']/a/@href")
	for _, i := range nextUrl {
		url := htmlquery.InnerText(i)
		// 过滤url 保证是有图片的 不是apk链接
		if len(url) > 0 {
			if url[0] != '/' {
				continue
			} else {
				wg.Add(1)
				go GetImageContent(fmt.Sprintf("https://m.woyaogexing.com%s", url))
			}
		}
	}
}

// GetImageContent 获取图片各种信息
func GetImageContent(url string) {
	htmlText := GetHTMLText(url)
	node, err := htmlquery.Parse(strings.NewReader(string(htmlText)))
	if err != nil {
		panic(err)
	}
	// 疯狂过滤title， 这jb网站 title一堆不合法的 都不能创建文件夹
	title := htmlquery.FindOne(node, "//h1[@class='m-page-title']/text()").Data
	StrFilterNonChinese(&title)

	imageUrls := htmlquery.Find(node, "//ul[@class='m-page-txlist wbpCtr f-clear']//li/div/img/@data-src")

	tempID := 0
	for _, i := range imageUrls {
		tempID++
		imageUrl := fmt.Sprintf("http:%s", htmlquery.InnerText(i))
		SaveImageToDisk(title, imageUrl, tempID)
	}

	wg.Done()
}

// SaveImageToDisk 保存图片到硬盘
func SaveImageToDisk(filepath string, url string, id int) {

	path := fmt.Sprintf("images/%s", filepath)
	os.MkdirAll(path, os.ModePerm)
	imagePath := fmt.Sprintf("%s/%d.jpeg", path, id)
	file, err := os.OpenFile(imagePath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	response := GetHTMLText(url)

	size, err := file.Write(response)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s, 保存成功，大小: %d \n", imagePath, size)
}

// Spider 爬虫主流程
func Spider() {
	var url string
	for i := 1; i < 500; i++ {
		if i == 1 {
			url = "https://m.woyaogexing.com/touxiang/index.html"
		} else {
			url = fmt.Sprintf("https://m.woyaogexing.com/touxiang/index_%d.html", i)
		}

		// 获取图片页面的内容 这里本来也想加go关键字 结果.... 8太行
		GetImagePage(url)
	}
}

// StrFilterNonChinese 过滤非中文字符
func StrFilterNonChinese(src *string) {
	strn := ""
	for _, c := range *src {
		if reg.MatchString(string(c)) {
			strn += string(c)
		}
	}

	*src = strn
}

func main() {
	// 执行Spider主程序
	Spider()

	// WaitGroup等待
	wg.Wait()
}
