package base

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Data_struct struct {
	url    string
	ip     string
	title  string
	icp    string
	domain string
}

func Isdomain(qstr string) bool {
	domainPattern := `^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`
	domainRegex, _ := regexp.Compile(domainPattern)
	if domainRegex.MatchString(qstr) {
		return true
	}
	return false
}
func Contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
func Get_url(data []interface{}) string {
	var single_data Data_struct
	//host,ip,port,protocol,title,icp

	host := data[0].(string)
	//ip := data[1].(string)
	port := data[2].(string)
	protocol := data[3].(string)
	//title := data[4].(string)
	//icp := data[5].(string)

	//url
	if strings.Contains(host, "http") {
		single_data.url = host
	} else {
		if protocol == "http" {
			if strings.Contains(host, ":") {
				single_data.url = "http://" + host
			} else {
				single_data.url = "http://" + host + ":" + port
			}
		} else if protocol == "https" {
			if strings.Contains(host, ":") {
				single_data.url = "https://" + host
			} else {
				single_data.url = "https://" + host + ":" + port
			}
		}
	}
	return single_data.url
}
func Get_res(url string) (string, string, int) {
	client := http.Client{
		Timeout: time.Second * 1, // 设置超时时间为10秒
	}
	response, err := client.Get(url)
	if err != nil {
		return "", "", 0
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", "", 0
	}
	str_body := string(body)
	header := fmt.Sprintf("%s", response.Header)
	code := response.StatusCode

	//fmt.Println("code:", code)
	return str_body, header, code
}
func ExtractTitle(html string) string {
	pattern := `(?s)<title>(.*?)<\/title>`

	// 编译正则表达式
	reg := regexp.MustCompile(pattern)

	// 查找匹配的结果
	matches := reg.FindStringSubmatch(html)
	//fmt.Println(matches)
	if len(matches) > 1 {
		title1 := strings.Replace(matches[1], "\n", "", -1)
		title2 := strings.Replace(title1, "\r", "", -1)
		return title2
	} else {
		return ""
	}
}
func IsIP(ipStr string) bool {
	parsedIP := net.ParseIP(ipStr)
	return parsedIP != nil
}
