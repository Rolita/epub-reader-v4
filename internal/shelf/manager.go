package shelf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ShelfInfo 书架信息
type ShelfInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ShelfOrder 书架顺序配置
type ShelfOrder struct {
	Order []string `json:"order"` // 书架名称列表，按顺序排列
}

// Scan 扫描所有书架（查找所有 *_library.json 文件）
// 按照 shelf_order.json 中的顺序返回，如果没有顺序文件则按字母顺序
func Scan(shelfDir string) ([]ShelfInfo, error) {
	entries, err := os.ReadDir(shelfDir)
	if err != nil {
		return nil, err
	}

	// 先扫描所有书架
	allShelves := make(map[string]ShelfInfo)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_library.json") {
			name := strings.TrimSuffix(entry.Name(), "_library.json")
			allShelves[name] = ShelfInfo{
				ID:   name,
				Name: name,
			}
		}
	}

	// 尝试读取顺序文件
	order, err := LoadShelfOrder(shelfDir)
	if err != nil {
		// 没有顺序文件，按字母顺序返回
		var shelves []ShelfInfo
		for _, name := range sortedKeys(allShelves) {
			shelves = append(shelves, allShelves[name])
		}
		return shelves, nil
	}

	// 按照顺序文件返回书架
	var shelves []ShelfInfo
	for _, name := range order.Order {
		if shelf, ok := allShelves[name]; ok {
			shelves = append(shelves, shelf)
			delete(allShelves, name)
		}
	}
	// 添加不在顺序文件中的书架（按字母顺序）
	for _, name := range sortedKeys(allShelves) {
		shelves = append(shelves, allShelves[name])
	}

	return shelves, nil
}

// sortedKeys 返回 map 的 key 按字母顺序排序
func sortedKeys(m map[string]ShelfInfo) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// LoadShelfOrder 加载书架顺序
func LoadShelfOrder(shelfDir string) (*ShelfOrder, error) {
	filePath := filepath.Join(shelfDir, "shelf_order.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var order ShelfOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}

	return &order, nil
}

// SaveShelfOrder 保存书架顺序
func SaveShelfOrder(shelfDir string, order []string) error {
	shelfOrder := ShelfOrder{Order: order}

	data, err := json.MarshalIndent(shelfOrder, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(shelfDir, "shelf_order.json")
	return os.WriteFile(filePath, data, 0644)
}

// LoadShelfData 加载指定书架的数据
func LoadShelfData(shelfDir, shelfName string) (string, error) {
	fileName := shelfName + "_library.json"
	filePath := filepath.Join(shelfDir, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return `{"books":[]}`, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveShelfData 保存指定书架的数据
func SaveShelfData(shelfDir, shelfName, data string) error {
	fileName := shelfName + "_library.json"
	filePath := filepath.Join(shelfDir, fileName)

	var raw interface{}
	json.Unmarshal([]byte(data), &raw)
	prettyJSON, err := json.MarshalIndent(raw, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, prettyJSON, 0644)
}

// Create 创建新书架
func Create(shelfDir, shelfName string) error {
	fileName := shelfName + "_library.json"
	filePath := filepath.Join(shelfDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return nil
	}

	prettyJSON, _ := json.MarshalIndent(map[string]interface{}{"books": []interface{}{}}, "", "    ")
	return os.WriteFile(filePath, prettyJSON, 0644)
}

// Delete 删除书架（同时删除对应的书籍目录）
func Delete(shelfDir, booksDir, shelfName string) error {
	fileName := shelfName + "_library.json"
	filePath := filepath.Join(shelfDir, fileName)
	os.Remove(filePath)

	shelfBookDir := filepath.Join(booksDir, shelfName)
	os.RemoveAll(shelfBookDir)

	return nil
}

// Rename 重命名书架（同时重命名文件、目录和更新内部路径）
func Rename(oldName, newName, shelfDir, booksDir string) error {
	if oldName == newName {
		return nil
	}

	newFileName := newName + "_library.json"
	newFilePath := filepath.Join(shelfDir, newFileName)
	if _, err := os.Stat(newFilePath); err == nil {
		return os.ErrExist
	}

	oldFileName := oldName + "_library.json"
	oldFilePath := filepath.Join(shelfDir, oldFileName)

	if _, err := os.Stat(oldFilePath); os.IsNotExist(err) {
		return err
	}

	content, err := os.ReadFile(oldFilePath)
	if err != nil {
		return err
	}

	var shelfData map[string]interface{}
	if err := json.Unmarshal(content, &shelfData); err != nil {
		return err
	}

	books, ok := shelfData["books"].([]interface{})
	if !ok {
		books = []interface{}{}
	}

	for _, bookInterface := range books {
		book, ok := bookInterface.(map[string]interface{})
		if !ok {
			continue
		}

		book["shelfId"] = newName

		fieldsToUpdate := []string{"coverUrl", "filePath"}
		for _, field := range fieldsToUpdate {
			if value, exists := book[field]; exists {
				if strValue, ok := value.(string); ok {
					book[field] = strings.ReplaceAll(strValue, oldName, newName)
				}
			}
		}
	}

	newContent, err := json.MarshalIndent(shelfData, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(newFilePath, newContent, 0644); err != nil {
		return err
	}

	if err := os.Remove(oldFilePath); err != nil {
		os.Remove(newFilePath)
		return err
	}

	oldBookDir := filepath.Join(booksDir, oldName)
	newBookDir := filepath.Join(booksDir, newName)
	if _, err := os.Stat(oldBookDir); err == nil {
		// 检查新目录是否已存在
		if _, err := os.Stat(newBookDir); err == nil {
			// 目标目录已存在，先删除它
			if err := os.RemoveAll(newBookDir); err != nil {
				// 删除失败，尝试恢复
				os.WriteFile(oldFilePath, content, 0644)
				os.Remove(newFilePath)
				return err
			}
		}

		if err := os.Rename(oldBookDir, newBookDir); err != nil {
			os.WriteFile(oldFilePath, content, 0644)
			os.Remove(newFilePath)
			return err
		}
	}

	return nil
}
