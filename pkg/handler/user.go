package handler

import (
	"asset-management-system/pkg/model"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Levenshtein 距离算法，用于模糊搜索
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			if s1[i-1] == s2[j-1] {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				min := matrix[i-1][j] + 1
				if matrix[i][j-1]+1 < min {
					min = matrix[i][j-1] + 1
				}
				if matrix[i-1][j-1]+1 < min {
					min = matrix[i-1][j-1] + 1
				}
				matrix[i][j] = min
			}
		}
	}
	return matrix[len(s1)][len(s2)]
}

// 全局模板变量
var (
	assetEntryFullTemplate *template.Template
	loginTemplate         *template.Template
)

// 缓存所有资产（模拟缓存，实际可使用 Redis 或其他缓存系统）
var assetCache []struct {
	ID                 int    `json:"id"`
	SerialNumber       string `json:"serial_number"`
	Name               string `json:"name"`
	Category           string `json:"category"`
	Brand              string `json:"brand"`
	ApplicationDate    string `json:"application_date"`
	Specification      string `json:"specification"`
	AssetCode          string `json:"asset_code"`
	OrderDate          string `json:"order_date"`
	CreatedAt          string `json:"created_at"`
	Department         string `json:"department"`
	Location           string `json:"location"`
	Supplier           string `json:"supplier"`
	Recipient          string `json:"recipient"`
	RecipientDepartment string `json:"recipient_department"`
	Remarks            string `json:"remarks"`
}
var cacheMutex sync.Mutex

func init() {
	log.Println("初始化资产录入、列表和登录模板...")
	// 解析资产录入和列表模板
	assetEntryTemplatePath := filepath.Join("static", "templates", "asset-entry-full.html")
	log.Printf("尝试解析模板文件: %s", assetEntryTemplatePath)
	var err error
	assetEntryFullTemplate, err = template.ParseFiles(assetEntryTemplatePath)
	if err != nil {
		log.Printf("解析资产录入模板失败: %v", err)
	}

	// 解析登录模板（更新为 login.html）
	loginTemplatePath := filepath.Join("static", "templates", "login.html")
	log.Printf("尝试解析模板文件: %s", loginTemplatePath)
	loginTemplate, err = template.ParseFiles(loginTemplatePath)
	if err != nil {
		log.Printf("解析登录模板失败: %v", err)
	}
	log.Println("模板初始化成功")

	// 初始化缓存
	loadAssetCache()
}

func loadAssetCache() {
	db, err := model.InitDB()
	if err != nil {
		log.Printf("加载资产缓存失败: %v", err)
		return
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, serial_number, name, category, brand, application_date, specification, asset_code, order_date, created_at, department, location, supplier, recipient, recipient_department, remarks 
		FROM assets 
		ORDER BY created_at DESC`)
	if err != nil {
		log.Printf("查询所有资产失败: %v", err)
		return
	}
	defer rows.Close()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	assetCache = []struct {
		ID                 int    `json:"id"`
		SerialNumber       string `json:"serial_number"`
		Name               string `json:"name"`
		Category           string `json:"category"`
		Brand              string `json:"brand"`
		ApplicationDate    string `json:"application_date"`
		Specification      string `json:"specification"`
		AssetCode          string `json:"asset_code"`
		OrderDate          string `json:"order_date"`
		CreatedAt          string `json:"created_at"`
		Department         string `json:"department"`
		Location           string `json:"location"`
		Supplier           string `json:"supplier"`
		Recipient          string `json:"recipient"`
		RecipientDepartment string `json:"recipient_department"`
		Remarks            string `json:"remarks"`
	}{}

	for rows.Next() {
		var asset struct {
			ID                 int    `json:"id"`
			SerialNumber       string `json:"serial_number"`
			Name               string `json:"name"`
			Category           string `json:"category"`
			Brand              string `json:"brand"`
			ApplicationDate    string `json:"application_date"`
			Specification      string `json:"specification"`
			AssetCode          string `json:"asset_code"`
			OrderDate          string `json:"order_date"`
			CreatedAt          string `json:"created_at"`
			Department         string `json:"department"`
			Location           string `json:"location"`
			Supplier           string `json:"supplier"`
			Recipient          string `json:"recipient"`
			RecipientDepartment string `json:"recipient_department"`
			Remarks            string `json:"remarks"`
		}
		err = rows.Scan(&asset.ID, &asset.SerialNumber, &asset.Name, &asset.Category, &asset.Brand, &asset.ApplicationDate, &asset.Specification, &asset.AssetCode, &asset.OrderDate, &asset.CreatedAt, &asset.Department, &asset.Location, &asset.Supplier, &asset.Recipient, &asset.RecipientDepartment, &asset.Remarks)
		if err != nil {
			log.Printf("解析资产数据失败: %v", err)
			continue
		}
		assetCache = append(assetCache, asset)
	}
	log.Println("资产缓存加载成功")
}

// LoginHandler 处理登录页面和登录逻辑
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("处理登录请求: %s %s, 远程地址: %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method == "GET" {
		log.Println("加载登录页面")
		if loginTemplate == nil {
			log.Println("登录模板未初始化")
			http.Error(w, "登录模板未初始化", http.StatusInternalServerError)
			return
		}
		err := loginTemplate.Execute(w, nil)
		if err != nil {
			log.Printf("渲染登录页面失败: %v", err)
			http.Error(w, "渲染登录页面失败", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		log.Println("处理登录 POST 请求")
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		log.Printf("接收到登录数据: username=%s, password=%s", username, password)

		// 验证用户名和密码（示例，实际应从数据库验证）
		if username == "admin" && password == "admin" {
			log.Println("登录成功")
			// 设置会话或 Cookie 表示已登录
			http.SetCookie(w, &http.Cookie{
				Name:  "authenticated",
				Value: "true",
				Path:  "/",
			})
			http.Redirect(w, r, "/asset-entry", http.StatusSeeOther)
			return
		}

		log.Println("登录失败")
		http.Error(w, "用户名或密码错误", http.StatusUnauthorized)
	}
}

// IsAuthenticated 检查用户是否已登录
func IsAuthenticated(r *http.Request) bool {
	// 检查 Cookie 或会话状态
	cookie, err := r.Cookie("authenticated")
	if err != nil || cookie.Value != "true" {
		return false
	}
	return true
}

// AssetEntryHandler 处理资产录入页面
func AssetEntryHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("处理资产录入请求: %s %s, 远程地址: %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method == "GET" {
		log.Println("加载资产录入和列表页面")
		if assetEntryFullTemplate == nil {
			log.Println("模板未初始化")
			http.Error(w, "模板未初始化", http.StatusInternalServerError)
			return
		}

		data := struct {
			CreatedAt string
		}{
			CreatedAt: time.Now().Format("2006-01-02"),
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := assetEntryFullTemplate.Execute(w, data); err != nil {
			log.Printf("渲染模板失败: %v", err)
			http.Error(w, "渲染模板失败", http.StatusInternalServerError)
		} else {
			log.Println("模板渲染成功")
		}
	}

	if r.Method == "POST" {
		log.Println("处理资产录入 POST 请求")
		r.ParseForm()
		serialNumber := r.FormValue("serialNumber")
		name := r.FormValue("name")
		category := r.FormValue("category")
		brand := r.FormValue("brand")
		applicationDateStr := r.FormValue("applicationDate")
		specification := r.FormValue("specification")
		assetCode := r.FormValue("assetCode")
		orderDateStr := r.FormValue("orderDate")
		createdAtStr := r.FormValue("createdAt")
		department := r.FormValue("department")
		location := r.FormValue("location")
		supplier := r.FormValue("supplier")
		recipient := r.FormValue("recipient")
		recipientDepartment := r.FormValue("recipient_department")
		remarks := r.FormValue("remarks")

		log.Printf("接收到表单数据: serial_number=%s, name=%s, ...", serialNumber, name)

		// 表单验证
		if err := validateAssetForm(serialNumber, name, category, brand, applicationDateStr, specification, assetCode, orderDateStr, department, location, supplier, recipient, recipientDepartment, remarks); err != nil {
			log.Printf("表单验证失败: %v", err)
			http.Error(w, fmt.Sprintf("表单验证失败: %v", err), http.StatusBadRequest)
			return
		}

		// 转换为日期（使用字符串格式存储到数据库）
		var applicationDateStrSQL, orderDateStrSQL, createdAtStrSQL string
		if applicationDateStr != "" {
			applicationDate, err := time.Parse("2006-01-02", applicationDateStr)
			if err != nil {
				log.Printf("解析申请时间失败: %v", err)
				http.Error(w, "申请时间格式错误", http.StatusBadRequest)
				return
			}
			applicationDateStrSQL = applicationDate.Format("2006-01-02")
		}
		if orderDateStr != "" {
			orderDate, err := time.Parse("2006-01-02", orderDateStr)
			if err != nil {
				log.Printf("解析订购日期失败: %v", err)
				http.Error(w, "订购日期格式错误", http.StatusBadRequest)
				return
			}
			orderDateStrSQL = orderDate.Format("2006-01-02")
		}
		if createdAtStr != "" {
			createdAt, err := time.Parse("2006-01-02", createdAtStr)
			if err != nil {
				log.Printf("解析创建日期失败: %v", err)
				http.Error(w, "创建日期格式错误", http.StatusBadRequest)
				return
			}
			createdAtStrSQL = createdAt.Format("2006-01-02")
		} else {
			createdAtStrSQL = time.Now().Format("2006-01-02")
		}

		// 初始化数据库连接
		db, err := model.InitDB()
		if err != nil {
			log.Printf("数据库连接失败: %v", err)
			http.Error(w, "数据库连接失败", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// 使用事务确保数据一致性
		log.Println("开始事务插入或更新资产数据")
		tx, err := db.Begin()
		if err != nil {
			log.Printf("开始事务失败: %v", err)
			http.Error(w, "开始事务失败", http.StatusInternalServerError)
			return
		}

		action := r.FormValue("action") // 区分新建或编辑
		if action == "edit" {
			// 编辑现有资产
			idStr := r.FormValue("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				tx.Rollback()
				log.Printf("无效的资产 ID: %v", err)
				http.Error(w, "无效的资产 ID", http.StatusBadRequest)
				return
			}

			_, err = tx.Exec(`
				UPDATE assets 
				SET serial_number = ?, name = ?, category = ?, brand = ?, application_date = ?, specification = ?, asset_code = ?, order_date = ?, created_at = ?, department = ?, location = ?, supplier = ?, recipient = ?, recipient_department = ?, remarks = ?
				WHERE id = ?`,
				serialNumber, name, category, brand, applicationDateStrSQL, specification, assetCode, orderDateStrSQL, createdAtStrSQL, department, location, supplier, recipient, recipientDepartment, remarks, id)
			if err != nil {
				tx.Rollback()
				log.Printf("资产更新失败: %v", err)
				http.Error(w, "资产更新失败", http.StatusInternalServerError)
				return
			}
		} else {
			// 新建资产
			_, err = tx.Exec(`
				INSERT INTO assets (serial_number, name, category, brand, application_date, specification, asset_code, order_date, created_at, department, location, supplier, recipient, recipient_department, remarks) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				serialNumber, name, category, brand, applicationDateStrSQL, specification, assetCode, orderDateStrSQL, createdAtStrSQL, department, location, supplier, recipient, recipientDepartment, remarks)
			if err != nil {
				tx.Rollback()
				log.Printf("资产录入失败: %v", err)
				http.Error(w, "资产录入失败", http.StatusInternalServerError)
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Printf("提交事务失败: %v", err)
			http.Error(w, "提交事务失败", http.StatusInternalServerError)
			return
		}

		// 更新缓存
		loadAssetCache()

		log.Println("资产操作成功，刷新资产列表")
		// 返回 JSON 响应，刷新列表
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "success", "action": "%s"}`, action)
	}

	if r.Method == "DELETE" {
		log.Println("处理资产删除请求")
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("无效的资产 ID: %v", err)
			http.Error(w, "无效的资产 ID", http.StatusBadRequest)
			return
		}

		// 初始化数据库连接
		db, err := model.InitDB()
		if err != nil {
			log.Printf("数据库连接失败: %v", err)
			http.Error(w, "数据库连接失败", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// 使用事务删除资产
		log.Println("开始事务删除资产数据")
		tx, err := db.Begin()
		if err != nil {
			log.Printf("开始事务失败: %v", err)
			http.Error(w, "开始事务失败", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec("DELETE FROM assets WHERE id = ?", id)
		if err != nil {
			tx.Rollback()
			log.Printf("资产删除失败: %v", err)
			http.Error(w, "资产删除失败", http.StatusInternalServerError)
			return
		}

		err = tx.Commit()
		if err != nil {
			log.Printf("提交事务失败: %v", err)
			http.Error(w, "提交事务失败", http.StatusInternalServerError)
			return
		}

		// 更新缓存
		loadAssetCache()

		log.Println("资产删除成功，刷新资产列表")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "success"}`)
	}
}

// AssetListHandler 处理资产列表页面（优化模糊搜索，使用 LIKE 和 Levenshtein 距离）
func AssetListHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("处理资产列表请求: %s %s, 远程地址: %s", r.Method, r.URL.Path, r.RemoteAddr)
	// 初始化数据库连接
	db, err := model.InitDB()
	if err != nil {
		log.Printf("数据库连接失败: %v", err)
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 获取分页和搜索参数
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 20 // 默认每页 20 条
	}
	offset := (page - 1) * pageSize

	query := r.URL.Query().Get("query") // 模糊搜索关键字

	log.Printf("查询资产列表，页码: %d, 每页条数: %d, 搜索关键字: %s", page, pageSize, query)

	// 使用缓存
	cacheMutex.Lock()
	cachedAssets := make([]struct {
		ID                 int    `json:"id"`
		SerialNumber       string `json:"serial_number"`
		Name               string `json:"name"`
		Category           string `json:"category"`
		Brand              string `json:"brand"`
		ApplicationDate    string `json:"application_date"`
		Specification      string `json:"specification"`
		AssetCode          string `json:"asset_code"`
		OrderDate          string `json:"order_date"`
		CreatedAt          string `json:"created_at"`
		Department         string `json:"department"`
		Location           string `json:"location"`
		Supplier           string `json:"supplier"`
		Recipient          string `json:"recipient"`
		RecipientDepartment string `json:"recipient_department"`
		Remarks            string `json:"remarks"`
	}, len(assetCache))
	copy(cachedAssets, assetCache)
	cacheMutex.Unlock()

	// 模糊搜索逻辑（结合 LIKE 和 Levenshtein 距离）
	var filteredAssets []struct {
		ID                 int    `json:"id"`
		SerialNumber       string `json:"serial_number"`
		Name               string `json:"name"`
		Category           string `json:"category"`
		Brand              string `json:"brand"`
		ApplicationDate    string `json:"application_date"`
		Specification      string `json:"specification"`
		AssetCode          string `json:"asset_code"`
		OrderDate          string `json:"order_date"`
		CreatedAt          string `json:"created_at"`
		Department         string `json:"department"`
		Location           string `json:"location"`
		Supplier           string `json:"supplier"`
		Recipient          string `json:"recipient"`
		RecipientDepartment string `json:"recipient_department"`
		Remarks            string `json:"remarks"`
	}

	if query != "" {
		query = strings.ToLower(query)
		// 首先使用 LIKE 进行快速过滤（提高性能）
		likeQuery := "%" + query + "%"
		rows, err := db.Query(`
			SELECT id, serial_number, name, category, brand, application_date, specification, asset_code, order_date, created_at, department, location, supplier, recipient, recipient_department, remarks 
			FROM assets 
			WHERE serial_number LIKE ? OR name LIKE ? OR category LIKE ? OR brand LIKE ? OR department LIKE ? OR location LIKE ? OR supplier LIKE ? OR recipient LIKE ? OR recipient_department LIKE ? OR remarks LIKE ?
			ORDER BY created_at DESC`,
			likeQuery, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery)
		if err != nil {
			log.Printf("LIKE 模糊搜索失败: %v", err)
			http.Error(w, "模糊搜索失败", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var asset struct {
				ID                 int    `json:"id"`
				SerialNumber       string `json:"serial_number"`
				Name               string `json:"name"`
				Category           string `json:"category"`
				Brand              string `json:"brand"`
				ApplicationDate    string `json:"application_date"`
				Specification      string `json:"specification"`
				AssetCode          string `json:"asset_code"`
				OrderDate          string `json:"order_date"`
				CreatedAt          string `json:"created_at"`
				Department         string `json:"department"`
				Location           string `json:"location"`
				Supplier           string `json:"supplier"`
				Recipient          string `json:"recipient"`
				RecipientDepartment string `json:"recipient_department"`
				Remarks            string `json:"remarks"`
			}
			err = rows.Scan(&asset.ID, &asset.SerialNumber, &asset.Name, &asset.Category, &asset.Brand, &asset.ApplicationDate, &asset.Specification, &asset.AssetCode, &asset.OrderDate, &asset.CreatedAt, &asset.Department, &asset.Location, &asset.Supplier, &asset.Recipient, &asset.RecipientDepartment, &asset.Remarks)
			if err != nil {
				log.Printf("解析资产数据失败: %v", err)
				continue
			}
			// 进一步使用 Levenshtein 距离验证相似度
			fields := []string{
				asset.SerialNumber, asset.Name, asset.Category, asset.Brand,
				asset.Department, asset.Location, asset.Supplier, asset.Recipient,
				asset.RecipientDepartment, asset.Remarks,
			}
			for _, field := range fields {
				if field != "" {
					distance := levenshteinDistance(strings.ToLower(field), query)
					maxLength := len(field)
					if maxLength > 0 && float64(distance)/float64(maxLength) < 0.3 { // 相似度阈值 0.3（可调整）
						filteredAssets = append(filteredAssets, asset)
						break
					}
				}
			}
		}
	} else {
		filteredAssets = cachedAssets
	}

	// 应用分页
	total := len(filteredAssets)
	start := offset
	end := start + pageSize
	if end > total {
		end = total
	}
	pagedAssets := filteredAssets[start:end]

	// 返回 JSON 数据（用于 AJAX 刷新）
	log.Println("返回资产列表 JSON 数据")
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Assets   []struct {
			ID                 int    `json:"id"`
			SerialNumber       string `json:"serial_number"`
			Name               string `json:"name"`
			Category           string `json:"category"`
			Brand              string `json:"brand"`
			ApplicationDate    string `json:"application_date"`
			Specification      string `json:"specification"`
			AssetCode          string `json:"asset_code"`
			OrderDate          string `json:"order_date"`
			CreatedAt          string `json:"created_at"`
			Department         string `json:"department"`
			Location           string `json:"location"`
			Supplier           string `json:"supplier"`
			Recipient          string `json:"recipient"`
			RecipientDepartment string `json:"recipient_department"`
			Remarks            string `json:"remarks"`
		} `json:"assets"`
		Total   int `json:"total"`
		Page    int `json:"page"`
		Pages   int `json:"pages"`
		PageSize int `json:"pageSize"`
	}{
		Page:    page,
		Total:   total,
		Pages:   (total + pageSize - 1) / pageSize,
		PageSize: pageSize,
	}

	for _, asset := range pagedAssets {
		response.Assets = append(response.Assets, struct {
			ID                 int    `json:"id"`
			SerialNumber       string `json:"serial_number"`
			Name               string `json:"name"`
			Category           string `json:"category"`
			Brand              string `json:"brand"`
			ApplicationDate    string `json:"application_date"`
			Specification      string `json:"specification"`
			AssetCode          string `json:"asset_code"`
			OrderDate          string `json:"order_date"`
			CreatedAt          string `json:"created_at"`
			Department         string `json:"department"`
			Location           string `json:"location"`
			Supplier           string `json:"supplier"`
			Recipient          string `json:"recipient"`
			RecipientDepartment string `json:"recipient_department"`
			Remarks            string `json:"remarks"`
		}{
			ID:                 asset.ID,
			SerialNumber:       asset.SerialNumber,
			Name:               asset.Name,
			Category:           asset.Category,
			Brand:              asset.Brand,
			ApplicationDate:    asset.ApplicationDate,
			Specification:      asset.Specification,
			AssetCode:          asset.AssetCode,
			OrderDate:          asset.OrderDate,
			CreatedAt:          asset.CreatedAt,
			Department:         asset.Department,
			Location:           asset.Location,
			Supplier:           asset.Supplier,
			Recipient:          asset.Recipient,
			RecipientDepartment: asset.RecipientDepartment,
			Remarks:            asset.Remarks,
		})
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("编码 JSON 失败: %v", err)
		http.Error(w, "编码 JSON 失败", http.StatusInternalServerError)
	} else {
		log.Println("JSON 数据编码成功")
	}
}

// 表单验证函数
func validateAssetForm(serialNumber, name, category, brand, applicationDate, specification, assetCode, orderDate, department, location, supplier, recipient, recipientDepartment, remarks string) error {
	if serialNumber == "" {
		return fmt.Errorf("序列号不能为空")
	}
	if name == "" {
		return fmt.Errorf("资产名称不能为空")
	}
	if category == "" {
		return fmt.Errorf("设备类型不能为空")
	}
	if brand == "" {
		return fmt.Errorf("品牌不能为空")
	}
	if department == "" {
		return fmt.Errorf("所在部门不能为空")
	}
	if location == "" {
		return fmt.Errorf("所在地不能为空")
	}
	if supplier == "" {
		return fmt.Errorf("供应商不能为空")
	}
	if recipient == "" {
		return fmt.Errorf("领用人不能为空")
	}
	if recipientDepartment == "" {
		return fmt.Errorf("领取部门不能为空")
	}

	// 日期格式验证（YYYY-MM-DD）
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if applicationDate != "" && !dateRegex.MatchString(applicationDate) {
		return fmt.Errorf("申请时间格式错误，应为 YYYY-MM-DD")
	}
	if orderDate != "" && !dateRegex.MatchString(orderDate) {
		return fmt.Errorf("订购日期格式错误，应为 YYYY-MM-DD")
	}
	return nil
}