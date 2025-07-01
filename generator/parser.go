package generator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateStruct 包含validate标签的结构体信息
type ValidateStruct struct {
	Name   string
	Fields []ValidateField
}

// ValidateField 包含validate标签的字段信息
type ValidateField struct {
	Name         string
	Type         string
	ValidateRule string
	JsonTag      string
}

// Options 插件选项
type Options struct {
	EnableTranslator bool // 是否生成translator
}

// parseAPIFileForValidateStructs 解析API文件获取带有validate标签的结构体（支持import）
func parseAPIFileForValidateStructs(apiFilePath string) ([]ValidateStruct, error) {
	var allValidateStructs []ValidateStruct
	processedFiles := make(map[string]bool)

	// 解析主API文件和所有import的文件
	if err := parseAPIFileRecursively(apiFilePath, &allValidateStructs, processedFiles); err != nil {
		return nil, err
	}

	fmt.Printf("goctl-validate: found %d structures with validate tags across %d files\n",
		len(allValidateStructs), len(processedFiles))
	return allValidateStructs, nil
}

// parseAPIFileRecursively 递归解析API文件及其import的文件
func parseAPIFileRecursively(apiFilePath string, validateStructs *[]ValidateStruct, processedFiles map[string]bool) error {
	// 避免重复处理同一个文件
	if processedFiles[apiFilePath] {
		return nil
	}
	processedFiles[apiFilePath] = true

	fmt.Printf("goctl-validate: parsing API file: %s\n", apiFilePath)

	file, err := os.Open(apiFilePath)
	if err != nil {
		return fmt.Errorf("failed to open API file %s: %v", apiFilePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var currentStruct *ValidateStruct
	var inTypeBlock bool
	var inImportBlock bool
	var inStruct bool
	var braceCount int

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 检查import块开始
		if strings.HasPrefix(line, "import (") {
			inImportBlock = true
			continue
		}

		// 检查import块结束
		if inImportBlock && line == ")" {
			inImportBlock = false
			continue
		}

		// 在import块内部或单行import
		if inImportBlock {
			if importPath := parseImportPathFromLine(line, apiFilePath); importPath != "" {
				// 递归解析import的文件
				if err := parseAPIFileRecursively(importPath, validateStructs, processedFiles); err != nil {
					fmt.Printf("goctl-validate: warning - failed to parse imported file %s: %v\n", importPath, err)
				}
			}
			continue
		} else if importPath := parseImportLine(line, apiFilePath); importPath != "" {
			// 单行import
			if err := parseAPIFileRecursively(importPath, validateStructs, processedFiles); err != nil {
				fmt.Printf("goctl-validate: warning - failed to parse imported file %s: %v\n", importPath, err)
			}
			continue
		}

		// 检查是否进入type块
		if strings.HasPrefix(line, "type (") {
			inTypeBlock = true
			continue
		}

		// 检查是否退出type块
		if inTypeBlock && line == ")" {
			inTypeBlock = false
			continue
		}

		if !inTypeBlock {
			continue
		}

		// 检查结构体定义开始
		if strings.Contains(line, "{") && !inStruct {
			structName := extractStructName(line)
			if structName != "" {
				currentStruct = &ValidateStruct{
					Name:   structName,
					Fields: []ValidateField{},
				}
				inStruct = true
				braceCount = strings.Count(line, "{") - strings.Count(line, "}")
				continue
			}
		}

		// 在结构体内部
		if inStruct && currentStruct != nil {
			// 更新大括号计数
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")

			// 检查结构体结束
			if braceCount <= 0 {
				if len(currentStruct.Fields) > 0 {
					*validateStructs = append(*validateStructs, *currentStruct)
					fmt.Printf("goctl-validate: found struct with validate tags: %s (%d fields) in %s\n",
						currentStruct.Name, len(currentStruct.Fields), apiFilePath)
				}
				currentStruct = nil
				inStruct = false
				continue
			}

			// 解析字段
			if field := parseFieldLine(line); field != nil {
				currentStruct.Fields = append(currentStruct.Fields, *field)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading API file %s: %v", apiFilePath, err)
	}

	return nil
}

// extractStructName 从行中提取结构体名称
func extractStructName(line string) string {
	// 匹配结构体定义: StructName {
	re := regexp.MustCompile(`(\w+)\s*\{`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// parseFieldLine 解析字段行
func parseFieldLine(line string) *ValidateField {
	// 跳过注释行
	if strings.HasPrefix(line, "//") {
		return nil
	}

	// 匹配字段定义: FieldName Type `tags`
	re := regexp.MustCompile(`(\w+)\s+([*\[\]]*\w+)\s*` + "`" + `([^` + "`" + `]*)` + "`")
	matches := re.FindStringSubmatch(line)

	if len(matches) < 4 {
		return nil
	}

	fieldName := matches[1]
	fieldType := matches[2]
	tags := matches[3]

	// 解析标签
	validateRule := extractValidateFromTags(tags)
	if validateRule == "" {
		return nil
	}

	jsonTag := extractJsonFromTags(tags)

	fmt.Printf("goctl-validate: found field with validate: %s (%s) validate='%s'\n",
		fieldName, fieldType, validateRule)

	return &ValidateField{
		Name:         fieldName,
		Type:         fieldType,
		ValidateRule: validateRule,
		JsonTag:      jsonTag,
	}
}

// extractValidateFromTags 从标签字符串中提取validate值
func extractValidateFromTags(tags string) string {
	// 匹配 validate:"value"
	re := regexp.MustCompile(`validate:"([^"]*)"`)
	matches := re.FindStringSubmatch(tags)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractJsonFromTags 从标签字符串中提取json值
func extractJsonFromTags(tags string) string {
	// 匹配 json:"value"
	re := regexp.MustCompile(`json:"([^"]*)"`)
	matches := re.FindStringSubmatch(tags)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// parseImportLine 解析单行import语句
func parseImportLine(line, currentFilePath string) string {
	// 匹配 import "path/to/file.api"
	re := regexp.MustCompile(`import\s+"([^"]+\.api)"`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return resolveImportPath(matches[1], currentFilePath)
	}
	return ""
}

// parseImportPathFromLine 解析import块内的路径行
func parseImportPathFromLine(line, currentFilePath string) string {
	// 匹配 "path/to/file.api"
	re := regexp.MustCompile(`"([^"]+\.api)"`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return resolveImportPath(matches[1], currentFilePath)
	}
	return ""
}

// resolveImportPath 解析并验证import路径
func resolveImportPath(importPath, currentFilePath string) string {
	// 如果是相对路径，转换为绝对路径
	if !filepath.IsAbs(importPath) {
		baseDir := filepath.Dir(currentFilePath)
		importPath = filepath.Join(baseDir, importPath)
	}

	// 检查文件是否存在
	if _, err := os.Stat(importPath); err == nil {
		return importPath
	}

	fmt.Printf("goctl-validate: warning - imported API file not found: %s\n", importPath)
	return ""
}

// dirExists 检查目录是否存在
func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	return !os.IsNotExist(err) && info.IsDir()
}
