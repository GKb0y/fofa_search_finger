package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/tidwall/gjson"
	"gopkg.in/ini.v1"
	"io"
	"new_search_finger/base"
	"new_search_finger/search"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
1.识别输入参数，返回参数
2。
*/

func test() {
	donefile, err3 := os.OpenFile("done.txt", os.O_CREATE|os.O_RDWR, 0777)
	if err3 != nil {
		color.Red("打开target文件出现错误!-%s\n", err3)
		return
	}
	defer donefile.Close()

	donefile.Seek(0, io.SeekStart)
	scanner2 := bufio.NewScanner(donefile)
	for scanner2.Scan() {
		line := scanner2.Text()
		fmt.Println(line)
	}
}
func main() {
	//test()
	//os.Exit(0)
	var done_target []string
	args := os.Args
	arg_len := len(args)
	if arg_len < 2 {
		color.Red("参数值不够，请阅读使用说明！\n")
		help()
		return
	}
	_ = color.New(color.FgCyan).Add(color.Underline)
	time_now := time.Now()
	format_time := time_now.Format("2006-01-02 15-04-05")
	target_file, code_arr, module_arr, out_file, count := arg_init(args, arg_len, format_time)
	fmt.Println("配置完成，开始查找.....")
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		return
	}

	// 读取默认节（DEFAULT）中的配置
	fofa_email := cfg.Section("").Key("fofa_email").String()
	fofa_key := cfg.Section("").Key("fofa_key").String()

	//打开target文件，逐行读取语句，并调用fofasearch
	targetfile, err := os.Open(target_file)
	if err != nil {
		color.Red("打开target文件出现错误!-%s\n", err)
		return
	}
	defer targetfile.Close()

	//打开输出文件
	resfile, err2 := os.OpenFile(out_file, os.O_CREATE|os.O_RDWR, 0777)
	if err2 != nil {
		color.Red("打开target文件出现错误!-%s\n", err2)
		return
	}
	defer resfile.Close()
	//打开保存已经查询过的目标
	donefile, err3 := os.OpenFile("done.txt", os.O_CREATE|os.O_RDWR, 0777)
	if err3 != nil {
		color.Red("打开target文件出现错误!-%s\n", err3)
		return
	}
	defer donefile.Close()
	//打开保存新获得资产的文件
	newtargetfile, err4 := os.OpenFile(format_time+"_new-target.txt", os.O_CREATE|os.O_RDWR, 0777)
	if err4 != nil {
		color.Red("打开target文件出现错误!-%s\n", err4)
		return
	}
	defer newtargetfile.Close()

	scanner := bufio.NewScanner(targetfile)
	var fofa_res_arr [][]gjson.Result

	// 将文件指针移动到文件开头
	donefile.Seek(0, io.SeekStart)
	scanner2 := bufio.NewScanner(donefile)
	for scanner2.Scan() {
		line := scanner2.Text()
		done_target = append(done_target, line)
		//fmt.Println(line)
	}
	//fmt.Println(done_target)
	for scanner.Scan() {
		line := scanner.Text()
		donefile.WriteString(line + "\n")
		line = strings.Replace(line, "\n", "", -1)
		line = strings.Replace(line, "\r", "", -1)
		//调用search，将结果存入数组
		res_array := search.Search(line, &done_target, fofa_email, fofa_key, count)
		if res_array != nil {
			fofa_res_arr = append(fofa_res_arr, res_array)
		}
	}
	fmt.Println("查找完成，开始处理......")
	for _, results_arr := range fofa_res_arr {
		for _, result := range results_arr {
			res_str := result.String()
			var data []interface{}
			err := json.Unmarshal([]byte(res_str), &data)
			if err != nil {
				panic(err)
			}
			//开始调用功能函数
			//存活探测
			url := base.Get_url(data)
			if code_arr != nil {
				body, header, code, title := search.GetCode(url)
				for _, s := range code_arr {
					int_s, _ := strconv.Atoi(s)
					if int_s == code {
						icp, finger := "", ""
						if module_arr != nil {
							for _, s2 := range module_arr {
								if s2 == "icp" {
									icp = search.Geticp(url)
								} else if s2 == "finger" {
									finger = search.Getfinger(body, header)
								}
							}
						}
						res_str := fmt.Sprintf("%s  --title:%s   --code:%d  --icp:%s  --finger:%s\n", url, title, code, icp, finger)
						color.Green(res_str)
						resfile.WriteString(res_str)
					}
				}
			}
			var new_target []string
			ip := data[1].(string)
			domain := data[6].(string)
			if !base.Contains(done_target, domain) && !base.Contains(new_target, domain) && domain != "" {

				new_target = append(new_target, domain)
				newtargetfile.WriteString(domain + "\n")
			}
			if !base.Contains(done_target, ip) && !base.Contains(new_target, ip) && ip != "" {
				new_target = append(new_target, ip)
				newtargetfile.WriteString(ip + "\n")
			}

		}
	}
}

func arg_init(args []string, arg_len int, format_time string) (string, []string, []string, string, int) {

	var code_arr []string
	var module_arr []string
	//code_arr[0] = "200"
	//fmt.Println(arg_len)
	target_file := "target.txt"
	out_file := format_time + "_" + "result.txt"
	count := 3000
	for i := 1; i < arg_len; i++ {
		//fmt.Println(args[i])
		switch args[i] {
		case "-h":
			help()
			os.Exit(0)
		case "-t":
			target_file = args[i+1]
			i++
		case "-l":
			code_str := args[i+1]
			i++
			code_arr = strings.Split(code_str, ",")
		case "-m":
			module_str := args[i+1]
			i++
			module_arr = strings.Split(module_str, ",")
		case "-o":
			out_file = args[i+1]
			i++
		case "-n":
			count, _ = strconv.Atoi(args[i+1])
			i++
		}
	}

	return target_file, code_arr, module_arr, out_file, count
}
func help() {
	_ = color.New(color.FgCyan).Add(color.Underline)
	color.Green("此程序的目的是方便资产检索和目标发现！采用fofa接口，请在当前目录下config.ini中配置fofa_email和key\n")
	color.Green("eg:new_search_finger.exe -t target.txt -l -m icp,finger -o res.txt")
	color.Green("-t 此选项用于设置需要探测的目标，在文件中按行放入ip或域名，根据fofa语法，也可以自定义搜索语法。")
	color.Green("-l 此选项用于设置是否需要探测目标存活,以及需要的返回码200/403/xxx,eg:-l 200,403,404\n")
	color.Green("-m 此选项设置需要使用的功能，如icp:获取icp备案，finger:识别资产指纹\n")
	color.Green("-n 此选项设置每次查询的条数，最大仅支持10000条\n")
	color.Green("-o 此选项设置需要输出的文件\n")
}
