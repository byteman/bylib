package byhttp

import (
	"bylib/bylog"
	"bylib/byutils"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	_ "net/http/pprof"
	"os"
)
//go tool pprof http://weizhi.com/debug/pprof/
//https://blog.csdn.net/suiban7403/article/details/79144394
func IndexHandler(w http.ResponseWriter, r *http.Request) {

	path,err:=byutil.GetCurrentPath()
	if err!=nil{

	}

	t, err := template.ParseFiles(path+"/web/index.html")
	if err != nil {
		fmt.Println(err)
	}
	t.Execute(w, nil)
}

type MuxerContext struct{
	w http.ResponseWriter
	r *http.Request
	Result gjson.Result
}
//写入json格式的回应数据.
func (c *MuxerContext)Json(status int,v interface{})error{
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(status)

	return json.NewEncoder(c.w).Encode(v)
}
func (c *MuxerContext) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.r.FormFile(name)
	return fh, err
}
func (c *MuxerContext) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	io.Copy(out, src)
	return nil
}
func (c *MuxerContext) GetQuery(key string) (string, bool) {
	if values, ok := c.GetQueryArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *MuxerContext) GetQueryArray(key string) ([]string, bool) {
	if values, ok := c.r.URL.Query()[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}
func (c *MuxerContext)BindJson(v interface{})error{
	b, err := ioutil.ReadAll(c.r.Body)
	if err != nil {
		return err
	}
	defer c.r.Body.Close()

	err = json.Unmarshal(b, v)
	return err
}
func (c *MuxerContext)Bind2Json()(result *gjson.Result,err error){
	b, err := ioutil.ReadAll(c.r.Body)
	if err != nil {
		return nil,err
	}
	defer c.r.Body.Close()
	res :=  gjson.Parse(string(b))
	return &res,nil

}
func DefaultHandler(w http.ResponseWriter, r *http.Request) {

	path:=r.URL.Path

	//Debug("URL=%s",path)
	ctx:=MuxerContext{
		w:w,
		r:r,
	}

	if r.Method == "GET"{
		//Debug("GET path=%s",path)
		if h,ok:=mux.GetHandlers[path];ok{
			//fmt.Println("----------")
			h(&ctx)
		}

	}else if r.Method == "POST"{
		bylog.Debug("POST path=%s",path)
		if h,ok:=mux.PostHandlers[path];ok{
			h(&ctx)
		}

	}

}


type HttpHandler func(ctx *MuxerContext)error
type MuxServer struct {
	AllUrl map[string]int
	GetHandlers  map[string]HttpHandler
	PostHandlers map[string]HttpHandler
}
var mux MuxServer
func GetMuxServer()*MuxServer{
	if mux.GetHandlers == nil{
		mux.GetHandlers = make(map[string]HttpHandler)
	}
	if mux.PostHandlers == nil{
		mux.PostHandlers = make(map[string]HttpHandler)
	}
	if mux.AllUrl == nil{
		mux.AllUrl = make(map[string]int)
	}
	return &mux
}


//注册一个Get请求.
func (m *MuxServer)Get(url string, handler HttpHandler)error{
	m.AllUrl[url]=0
	mux.GetHandlers[url] = handler
	return nil
}
//注册一个Post请求.
func (m *MuxServer)Post(url string, handler HttpHandler)error{
	m.AllUrl[url]=0
	mux.PostHandlers[url] = handler
	return nil
}

var server *http.Server=nil

func StopMuxServer(){
	if server!=nil{
		server.Close()
	}
}
func listenServer(addr string, handler http.Handler)error{

	server = &http.Server{Addr: addr, Handler: handler}
	return http.ListenAndServe(addr, nil)
}
func StartMuxServer(port int) {
	//m := mux.NewRouter()
	path,err:=byutil.GetCurrentPath()
	if err!=nil{
		return
	}
	bylog.Debug("workDir=%s",path)

	http.Handle("/",http.FileServer(http.Dir(path+"/web")))

	for url,_:= range mux.AllUrl{
		bylog.Debug("registered url=%s",url)
		http.HandleFunc(url,DefaultHandler)
	}
	bylog.Debug("http server start @%d",port)
	listenServer(fmt.Sprintf("0.0.0.0:%d",port),nil)

}
func StartMuxServerWithFs(port int,fs http.FileSystem) {

	http.Handle("/", http.FileServer(fs))
	for url,_:= range mux.AllUrl{
		bylog.Debug("registered url=%s",url)
		http.HandleFunc(url,DefaultHandler)
	}
	bylog.Debug("http server start @%d",port)
	listenServer(fmt.Sprintf("0.0.0.0:%d",port),nil)
}