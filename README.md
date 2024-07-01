# fofa_search_finger

使用fofa快速查找目标资产，并对资产进行存活探测和指纹识别

目前采用fofa接口，请在当前目录下config.ini中配置fofa_email和key

eg:new_search_finger.exe -t target.txt -l 200 -m icp,finger -o res.txt -n 3

-t 此选项用于设置需要探测的目标，在文件中按行放入ip或域名，根据fofa语法，也可以自定义搜索语法。

-l 此选项用于设置是否需要探测目标存活,以及需要的返回码200/403/xxx,eg:-l 200,403,404

-m 此选项设置需要使用的功能，如icp:获取icp备案，finger:识别资产指纹

-n 此选项设置每次查询的条数，最大仅支持10000条

-o 此选项设置需要输出的文件

target文件内容可以是域名、IP、自定义fofa语法。

target文件内容示例:

app="泛微-协同商务系统"

1.1.1.1

xxx.com

a.b.com

title="xxx系统"
