package main

import (
	"alidns-edit/type"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

var editQuery _type.EditQuery
var config _type.Config
var listQuery _type.ListQuery
var r *rand.Rand

func getIP() (string, error) {
	res, err := http.Get("https://4.ipw.cn/")
	if err != nil || res.StatusCode != 200 {
		return "", errors.New("获取IP错误!")
	}
	ip, _ := ioutil.ReadAll(res.Body)

	return string(ip), nil
}

func getConfig() (_type.Config, error) {
	filePath := "./config.json"
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return _type.Config{}, err
	}
	var config _type.Config
	if err = json.Unmarshal(file, &config); err != nil {
		return _type.Config{}, errors.New("config.json解析失败,请确保文件格式正确")
	}
	err = config.Check()
	return config, err
}
func getTime() string { //获取当前时间
	now := time.Now().UTC()
	year, mon, day := now.Date()
	hour, min, sec := now.Clock()
	return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ", year, mon, day, hour, min, sec)
}
func throwError() {
	if err := recover(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
func init() {
	defer throwError()
	ip, err := getIP()
	if err != nil {
		panic(err)
	}
	fmt.Println("您的IP为：", ip)
	if config, err = getConfig(); err != nil {
		panic(err)
	}
	r = rand.New(rand.NewSource(time.Now().Unix()))
	listQuery = _type.ListQuery{
		Action:     "DescribeDomainRecords",
		DomainName: config.DomainName,
		PageSize:   50,
		RRKeyWord:  config.RR,
		PublicQuery: _type.PublicQuery{
			AccessKeyId:      config.AccessKeyId,
			Format:           "JSON",
			SignatureMethod:  "HMAC-SHA1",
			SignatureVersion: "1.0",
			Timestamp:        getTime(),
			Version:          "2015-01-09",
		},
	}
	editQuery = _type.EditQuery{
		Action:     "UpdateDomainRecord",
		DomainName: config.DomainName,
		RR:         config.RR,
		Type:       "A",
		Value:      ip,
		PublicQuery: _type.PublicQuery{
			AccessKeyId:      config.AccessKeyId,
			Format:           "JSON",
			SignatureMethod:  "HMAC-SHA1",
			SignatureVersion: "1.0",
			Timestamp:        getTime(),
			Version:          "2015-01-09",
		},
	}
}
func getKeys(json interface{}) []string {
	var keys []string
	_type := reflect.TypeOf(json).Kind().String()
	if _type != "struct" {
		return keys
	}
	types := reflect.TypeOf(json)
	values := reflect.ValueOf(json)
	for i := 0; i < types.NumField(); i++ {
		if types.Field(i).Type.Kind().String() != "struct" {
			keys = append(keys, types.Field(i).Name)
		} else {
			_keys := getKeys(values.Field(i).Interface().(interface{}))
			keys = append(keys, _keys...)
		}
	}
	return keys
}
func getQuery(json interface{}) string {
	keys := getKeys(json)
	sort.Strings(keys)
	var query string = ""
	values := reflect.ValueOf(json)
	for i := 0; i < len(keys); i++ {
		k := values.FieldByName(keys[i]).Kind().String()
		switch k {
		case "string":
			query += keys[i] + "=" + values.FieldByName(keys[i]).String()
		case "int8":
			value := values.FieldByName(keys[i]).Interface().(int8)
			query += keys[i] + "=" + strconv.Itoa(int(value))
		case "int32":
			value := values.FieldByName(keys[i]).Interface().(int32)
			query += keys[i] + "=" + strconv.Itoa(int(value))
		}
		if len(keys)-1 != i {
			query += "&"
		}
	}
	return query
}
func createUrl(query string) string {
	stringToSign := "GET" + "&" + url.QueryEscape("/") + "&" + url.QueryEscape(query) //.replace(/%3A/g, "%253A");
	stringToSign = strings.Replace(stringToSign, "%3A", "%253A", -1)

	_hash := hmac.New(sha1.New, []byte(config.Signature+"&"))

	_hash.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(_hash.Sum(nil))
	test := url.QueryEscape(query + "&Signature=" + signature)
	test = strings.Replace(test, "%3D", "=", -1)
	test = strings.Replace(test, "%26", "&", -1)
	return "https://alidns.aliyuncs.com/?" + test
}
func getRecordId() (string, error) {
	var page int8 = 1
	var count int8
	for ; ; page++ {
		if page > 1 && page > count {
			break
		}
		listQuery.PageNumber = page
		listQuery.SignatureNonce = r.Int31()
		query := getQuery(listQuery)
		url := createUrl(query)
		res, err := http.Get(url)
		if err != nil {
			return "", errors.New("获取域名RecordId失败")
		}
		body, _ := ioutil.ReadAll(res.Body)
		if res.StatusCode != 200 {
			failedRes := &_type.FailedResponse{}
			if err := json.Unmarshal(body, failedRes); err != nil {
				return "", errors.New("数据解析错误获")
			}
			return "", errors.New(fmt.Sprintf(`获取域名RecordId失败:
					Recommend: %s
					Message:   %s
					Code:      %s`,
				failedRes.Recommend, failedRes.Message, failedRes.Code))
		}
		successedRes := &_type.SuccessedResponse{}
		if err := json.Unmarshal(body, successedRes); err != nil {
			return "", errors.New("数据解析错误获")
		}
		for _, v := range successedRes.DomainRecords.Record {
			if v.RR == config.RR {
				return v.RecordId, nil
			}
		}
		count = int8(math.Ceil(successedRes.TotalCount / successedRes.PageSize))
	}
	return "", errors.New("找不到指定的二级域名，请先在阿里云控制面板添加相应解析")
}
func main() {
	defer throwError()
	recordId, err := getRecordId()
	if err != nil {
		panic(err)
	}
	editQuery.RecordId = recordId
	editQuery.SignatureNonce = r.Int31()
	query := getQuery(editQuery)
	url := createUrl(query)
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if res.StatusCode != 200 {
		failedRes := &_type.FailedResponse{}
		if err := json.Unmarshal(body, failedRes); err != nil {
			panic("数据解析错误获")
		}
		if failedRes.Code == "DomainRecordDuplicate" {
			fmt.Println("设置成功")
		} else {
			panic(fmt.Sprintf(`修改域名解析失败:
					Recommend: %s
					Message:   %s
					Code:      %s`,
				failedRes.Recommend, failedRes.Message, failedRes.Code))
		}
	} else {
		fmt.Println("设置成功")
	}
}
