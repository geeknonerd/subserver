package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const CnfFileName string = "config.yaml" //配置文件名

var cache = sync.Map{}     //缓存
var dataCache = NewCache() //请求数据缓存

// Config 配置结构体
type Config struct {
	Token        string            `yaml:"TOKEN"`
	ClashSubFmt  string            `yaml:"CLASH_SUB_FMT"`
	ClashSubUrls map[string]string `yaml:"CLASH_SUB_URLS"`
}

// getCnf 读取并解析yaml配置文件
func getCnf(path string) Config {
	cnf := Config{}
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("[error] getCnf ReadFile err: %v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &cnf)
	if err != nil {
		log.Fatalf("[error] getCnf Unmarshal err: %v", err)
	}
	return cnf
}

// getCnfPath 获取配置路径
func getCnfPath(path string) string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("[error] getCnfPath Getwd err: %v", err)
	}
	cnfAbsPath := filepath.Join(pwd, path)
	return cnfAbsPath
}

// getCnfByEnv 从环境变量读取配置
func getCnfByEnv() Config {
	cnf := Config{
		Token:       os.Getenv("TOKEN"),
		ClashSubFmt: os.Getenv("CLASH_SUB_FMT"),
	}
	envUrls := os.Getenv("CLASH_SUB_URLS")
	var urlMap map[string]string
	err := json.Unmarshal([]byte(envUrls), &urlMap)
	if err != nil {
		log.Fatalf("[error] getCnfByEnv Unmarshal err: %v", err)
	}
	cnf.ClashSubUrls = urlMap
	return cnf
}

// CacheItem 表示一个缓存项
type CacheItem struct {
	Value      interface{}
	Expiration int64
}

// Cache 表示缓存对象
type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
}

// NewCache 创建一个新的缓存对象
func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}
	// 启动一个 goroutine 定期清除过期项
	//go cache.cleanup()
	return cache
}

// Set 向缓存中添加一个项
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	expiration := time.Now().Add(duration).UnixNano()
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

// Get 从缓存中获取一个项
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[key]
	if !found {
		return nil, false
	}
	// 检查是否过期
	if time.Now().UnixNano() > item.Expiration {
		delete(c.items, key)
		return nil, false
	}
	return item.Value, true
}

// Delete 从缓存中删除一个项
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// cleanup 定期清除过期项
func (c *Cache) cleanup() {
	for {
		time.Sleep(2 * time.Hour) // 每2小时清理一次
		c.mu.Lock()
		now := time.Now().UnixNano()
		for key, item := range c.items {
			if now > item.Expiration {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

type ReqResItem struct {
	txt     string
	status  int
	headers http.Header
	err     error
}

// reqUrl 请求subconverter服务生成配置信息
func reqUrl(url string, method string, body io.Reader) ReqResItem {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return ReqResItem{"", -1, nil, errors.New(fmt.Sprintf("reqUrl NewRequest err: %v", err))}
	}
	req.Header.Add("User-Agent", "ClashforWindows/0.19.23")
	clt := http.Client{}
	resp, err := clt.Do(req)
	if err != nil {
		return ReqResItem{"", -1, nil, errors.New(fmt.Sprintf("reqUrl Do err: %v", err))}
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ReqResItem{"", -1, nil, errors.New(fmt.Sprintf("reqUrl ReadAll err: %v", err))}
	}
	if resp.StatusCode != 200 {
		return ReqResItem{"", -1, nil, errors.New(fmt.Sprintf("reqUrl StatusCode not 200: url=%s, resp=%v", url, resp))}
	}
	return ReqResItem{string(content), resp.StatusCode, resp.Header, nil}
}

func Convert(w http.ResponseWriter, r *http.Request) {
	//读取配置
	//cnfPath := getCnfPath(CnfFileName)
	//cnf := getCnf(cnfPath)
	//读取缓存
	cnf := Config{}
	c, ok := cache.Load("cnf")
	if !ok {
		cnf = getCnfByEnv()
		//设置缓存
		cache.Store("cnf", cnf)
	} else {
		cnf = c.(Config)
	}
	//检查配置
	if cnf.Token == "" || cnf.ClashSubUrls == nil {
		log.Fatalf("[error] Server config err: cnf=%v", cnf)
	}
	//鉴权
	if r.URL.Query().Get("token") != cnf.Token {
		log.Printf("[warn] Req token err: path=%s", r.URL.String())
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	//获取订阅链接
	subType := r.URL.Query().Get("sub_type")
	var subUrl string
	//针对mix类型进行配置合并
	if subType == "mix" {
		items := r.URL.Query().Get("mix_items")
		//检查 mix_items 参数
		if items == "" {
			log.Printf("[warn] Req mix_items err: path=%s", r.URL.String())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		//获取 mix 后的链接
		subTypes := strings.Split(items, ",")
		urls := make([]string, len(subTypes))
		for i, v := range subTypes {
			urls[i] = cnf.ClashSubUrls[v]
		}
		subUrl = strings.Join(urls, "|")
	} else {
		subUrl = cnf.ClashSubUrls[subType]
	}
	//检查 subUrl
	if subUrl == "" {
		log.Printf("[warn] Req sub_type err: path=%s", r.URL.String())
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	//请求subconverter转换配置
	convUrl := fmt.Sprintf("%s&url=%s&filename=Clash_%s.yaml",
		cnf.ClashSubFmt, url.QueryEscape(subUrl), subType)
	//读取数据缓存
	value, found := dataCache.Get(convUrl)
	//log.Printf("[info] dataCache: path=%s, found=%v", r.URL.String(), found)
	var res ReqResItem
	if found {
		//使用缓存数据
		res = value.(ReqResItem)
	} else {
		//发送请求
		log.Printf("[info] Req call reqUrl: path=%s", r.URL.String())
		res = reqUrl(convUrl, "GET", nil)
	}
	if res.err != nil {
		log.Printf("[warn] Req reqUrl err: path=%s, err=%v", r.URL.String(), res.err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	//去除不适合的头信息
	for k, v := range res.headers {
		if k != "Strict-Transport-Security" && k != "Content-Encoding" && k != "Vary" {
			w.Header().Set(k, v[0])
		}
	}
	//设置缓存
	dataCache.Set(convUrl, res, 22*time.Hour)
	//_, found := dataCache.Get(convUrl)
	//log.Printf("[debug] dataCache: found=%v", found)

	//接口返回
	w.WriteHeader(res.status)
	fmt.Fprintf(w, "%s", res.txt)
}
