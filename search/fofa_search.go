package search

import (
	"encoding/base64"
	"fmt"
	"github.com/fatih/color"
	"github.com/tidwall/gjson"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"net/url"
	"new_search_finger/base"
	"new_search_finger/search/info"
	"regexp"
	"strings"
	"unicode/utf8"
)

func Search(qstr string, done_target *[]string, fofa_email string, fofa_key string, count int) []gjson.Result {
	if base.Isdomain(qstr) {
		if base.Contains(*done_target, qstr) {
			return nil
		}
		qstr = "domain=\"" + qstr + "\""

	} else if base.IsIP(qstr) {
		if base.Contains(*done_target, qstr) {
			return nil
		}
		qstr = "ip=\"" + qstr + "\""
	} else {
		if base.Contains(*done_target, qstr) {
			return nil
		}
	}
	*done_target = append(*done_target, qstr)

	qstr_b_64 := base64.StdEncoding.EncodeToString([]byte(qstr + "&&country=\"CN\""))
	url := fmt.Sprintf("https://fofa.info/api/v1/search/all?email=%s&key=%s&full=true&page=1&size=%d&fields=host,ip,port,protocol,title,icp,domain&qbase64=%s", fofa_email, fofa_key, count, qstr_b_64)
	response, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer response.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil
	}
	str_body := string(body)

	all_count := gjson.Get(str_body, "size")
	all_count_int := all_count.Int()
	if all_count_int > 30000 {
		color.Red("%s结果过多,总条数:%d,跳过\n", qstr, all_count_int)
		return nil
	}

	res_array := gjson.Get(str_body, "results").Array()
	return res_array
}
func GetCode(url string) (string, string, int, string) {
	body, header, code := base.Get_res(url)
	//fmt.Println(body)
	title := base.ExtractTitle(body)
	if !utf8.ValidString(title) {
		title, _ = toUTF8(title, simplifiedchinese.GBK.NewDecoder())
	}
	return body, header, code, title
}
func toUTF8(src string, decoder transform.Transformer) (string, error) {
	reader := transform.NewReader(strings.NewReader(src), decoder)
	decoded, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
func Geticp(url string) string {
	urla := url
	url = strings.Replace(url, "https://", "", -1)
	url = strings.Replace(url, "http://", "", -1)
	url = strings.Split(url, ":")[0]
	rootDomain := ""
	if base.Isdomain(url) {
		rootDomain, _ = getRootDomain(urla)
	} else {
		rootDomain = getdomain(url)
	}
	if rootDomain != "" {
		//fmt.Printf("https://icp.365jz.com/%s\n", rootDomain)
		icp := Geticp_res(rootDomain)
		return icp
	}
	return ""
}

func getRootDomain(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	// Extract the hostname from the parsed URL
	hostname := parsedURL.Hostname()

	// Use the publicsuffix package to get the root domain
	etld1, err := publicsuffix.EffectiveTLDPlusOne(hostname)
	if err != nil {
		return "", err
	}

	return etld1, nil
}
func getdomain(ip string) string {
	ipstr := "ip=\"" + ip + "\""
	qstr_b_64 := base64.StdEncoding.EncodeToString([]byte(ipstr + "&&country=\"CN\""))
	url := fmt.Sprintf("https://fofa.info/api/v1/search/all?email=%s&key=%s&full=true&page=1&size=3000&fields=domain&qbase64=%s", qstr_b_64)
	response, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer response.Body.Close()
	// 读取响应内容
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ""
	}
	str_body := string(body)

	res_array := gjson.Get(str_body, "results").Array()
	for _, result := range res_array {
		res_str := result.String()
		if res_str != "" && base.Isdomain(res_str) {
			//fmt.Println(res_str)
			return res_str
		}
	}

	return ""
}
func Geticp_res(rootDomain string) string {
	url := "https://icp.365jz.com/" + rootDomain + "/" // 替换为你的目标URL
	// 发送HTTP请求
	//fmt.Println("get icp url:", url)
	// 发送GET请求
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return ""
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error reading response body:", err)
		return ""
	}
	//fmt.Printf("\nResponse Body:\n%s\n", body)

	// 将响应体转换为字符串
	bodyStr := string(body)

	// 使用正则表达式提取特定的HTML内容
	re := regexp.MustCompile(`<td>主办单位名称</td><td><div>(.*?)</div></td>`)
	matches := re.FindStringSubmatch(bodyStr)

	if len(matches) > 1 {
		// 匹配成功，输出提取的内容\
		icp := matches[1]
		return icp
	} else {
		return ""
	}
}

func Getfinger(body string, header string) string {
	var matched bool
	for _, rule := range info.RuleDatas {
		if rule.Type == "code" {
			matched, _ = regexp.MatchString(rule.Rule, body)
		} else {
			matched, _ = regexp.MatchString(rule.Rule, header)
		}
		if matched == true {
			return rule.Name
		}
	}
	return ""
}
