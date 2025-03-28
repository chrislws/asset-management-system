package main

import (
"asset-management-system/pkg/handler"
"log"
"net/http"
)

func main() {
log.Println("启动资产管理系统服务器...")

// 路由
http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
log.Printf("路由 /login 触发，方法: %s, URL: %s", r.Method, r.URL.Path)
handler.LoginHandler(w, r)
})

http.HandleFunc("/asset-entry", func(w http.ResponseWriter, r *http.Request) {
log.Printf("路由 /asset-entry 触发，方法: %s, URL: %s", r.Method, r.URL.Path)
handler.AssetEntryHandler(w, r)
})

http.HandleFunc("/assets/list", func(w http.ResponseWriter, r *http.Request) {
log.Printf("路由 /assets/list 触发，方法: %s, URL: %s", r.Method, r.URL.Path)
handler.AssetListHandler(w, r)
})

// 添加根路径 / 路由，检查登录状态并重定向
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
log.Printf("路由 / 触发，方法: %s, URL: %s", r.Method, r.URL.Path)
if !handler.IsAuthenticated(r) { // 假设 handler 包中有 IsAuthenticated 函数
http.Redirect(w, r, "/login", http.StatusSeeOther)
return
}
// 已登录用户渲染资产录入页面
handler.AssetEntryHandler(w, r) // 或其他默认页面
})

// 自定义静态文件服务
fs := http.FileServer(http.Dir("./static"))
http.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
log.Printf("服务静态文件: %s", r.URL.Path)
fs.ServeHTTP(w, r)
})))

// 启动服务器
log.Println("服务器运行在 :8080...")
log.Fatal(http.ListenAndServe(":8080", nil))
}