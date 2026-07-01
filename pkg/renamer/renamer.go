package renamer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type RenameMode string

const (
	BasicMode = "BasicMode"
	SmartMode = "SmartMode"
)

type ExtractRule struct {

	// 变量名
	Name string `json:"Name"`
	// 对应的正则表达式
	Pattern string `json:"Pattern"`
}

type RenameRule struct {

	// 模式
	Mode RenameMode `json:"Mode"`

	// 模式一   基础模式
	// 前缀
	Prefix string `json:"Prefix"`
	//后缀
	Suffix string `json:"Suffix"`
	// 要替换的旧字符串
	ReplaceOld string `json:"ReplaceOld"`
	// 替换成的新字符串
	ReplaceNew string `json:"ReplaceNew"`

	// 模式二  智能解析与重排
	SmartRules    []ExtractRule `json:"SmartRules"`
	SmartTemplate string        `json:"SmartTemplate"`
	CleanChars    string        `json:"CleanChars"` //允许用户指定要忽略的垃圾字符
}

type RenamePreview struct {

	// 原始路径
	OriginalPath string `json:"originalPath,omitempty"`
	// 原文件名
	OriginalName string `json:"originalName"`
	// 新文件名
	NewName string `json:"newName"`
	// 新的绝对路径
	NewPath string `json:"newPath"`
	//警告错误
	HasConflict bool `json:"hasConflict"`
	// 格式错误警告
	FormatError string `json:"formatError"`
}

type compiledRule struct {
	Name string
	Re   *regexp.Regexp
}

// GeneratePreview 主控，负责统筹调度和文件IO探查
func GeneratePreview(Paths []string, rule RenameRule) ([]RenamePreview, error) {

	// 检查传入的文件列表是否为空
	if len(Paths) == 0 {
		return nil, errors.New("未传入任何文件，请选择文件！")
	}

	var compiledRules []compiledRule
	var err error

	if rule.Mode == SmartMode {
		compiledRules, err = compileSmartRule(rule.SmartRules)

		if err != nil {
			return nil, err
		}
	}

	var previews []RenamePreview

	for _, oldPath := range Paths {

		if _, err := os.Stat(oldPath); os.IsNotExist(err) {
			previews = append(previews, RenamePreview{
				OriginalPath: oldPath,
				OriginalName: filepath.Base(oldPath),
				FormatError:  "文件已在其他地方被删除或修改，请移除",
			})
		}

		//获取当前文件的父目录
		dir := filepath.Dir(oldPath)
		// 获取当前文件名(包含后缀)
		baseName := filepath.Base(oldPath)
		// 获取当前文件的后缀
		ext := filepath.Ext(baseName)
		//剥离后缀，只处理文件名
		nameWithoutExt := strings.TrimSuffix(baseName, ext)

		var newName, formatErr string

		// 路由分发到处理逻辑
		if rule.Mode == SmartMode {
			newName, formatErr = applySmartMode(nameWithoutExt, ext, baseName, compiledRules, rule.SmartTemplate, rule.CleanChars)
		} else {
			newName = applyBasicMode(nameWithoutExt, ext, rule)
		}

		// 拼接新的文件路径
		newPath := filepath.Join(dir, newName)
		hasConflict := false
		if _, err := os.Stat(newPath); err == nil && oldPath != newPath {
			// 如果目录里文件已存在，进行警告
			hasConflict = true
		}

		previews = append(previews, RenamePreview{
			OriginalPath: oldPath,
			OriginalName: baseName,
			NewName:      newName,
			NewPath:      newPath,
			HasConflict:  hasConflict,
			FormatError:  formatErr,
		})
	}

	return previews, nil
}

func ExecuteRename(previews []RenamePreview) error {
	for _, p := range previews {

		// 旧名字与新名字相同时，跳过
		if p.OriginalPath == p.NewPath {
			continue
		}

		if p.HasConflict {
			return fmt.Errorf("文件%s已存在，拒绝覆盖", p.NewName)

		}

		if p.FormatError != "" {
			return fmt.Errorf("文件%s格式错误: %s", p.OriginalName, p.FormatError)
		}

		// 检测跨驱动器移动（Windows 上 os.Rename 不支持跨卷）
		oldVol := filepath.VolumeName(p.OriginalPath)
		newVol := filepath.VolumeName(p.NewPath)
		if oldVol != newVol {
			return fmt.Errorf("不支持跨驱动器移动: %s → %s", p.OriginalPath, p.NewPath)
		}

		err := os.Rename(p.OriginalPath, p.NewPath)
		if err != nil {
			return fmt.Errorf("重命名文件%s时出错: %v", p.OriginalName, err)
		}
	}
	return nil
}

// 负责把用户的字符串编译成正则
func compileSmartRule(rules []ExtractRule) ([]compiledRule, error) {
	var compiled []compiledRule
	for _, r := range rules {
		if r.Name == "" || r.Pattern == "" {
			continue
		}

		// 创建正则表达式
		re, err := regexp.Compile(r.Pattern)

		if err != nil {
			return nil, fmt.Errorf("正则语法%s错误: %v", r.Pattern, err)
		}

		compiled = append(compiled, compiledRule{Name: r.Name, Re: re})
	}
	return compiled, nil
}

// 处理基础前后缀拼接及重命名 (模式一)
func applyBasicMode(nameWithoutExt string, ext string, rule RenameRule) string {
	if rule.ReplaceOld != "" {
		//将字符串中出现的旧字符串替换成新字符串
		nameWithoutExt = strings.ReplaceAll(nameWithoutExt, rule.ReplaceOld, rule.ReplaceNew)
	}
	return rule.Prefix + nameWithoutExt + rule.Suffix + ext
}

// 处理智能模式的逻辑 (模式二）
func applySmartMode(nameWithoutExt string, ext string, originalBaseName string, compliedRules []compiledRule, template string, cleanChars string) (string, string) {
	extractedVars := make(map[string]string)
	var missing []string
	remain := nameWithoutExt

	var cleanRemain string

	for _, cr := range compliedRules {
		cleanRemain = remain
		// 清洗文件名 去除分隔符
		if cleanChars != "" {
			cleanRemain = strings.Map(func(r rune) rune {
				if strings.ContainsRune(cleanChars, r) {
					return -1
				}
				return r
			}, remain)
		}

		match := cr.Re.FindString(cleanRemain)

		// 匹配失败 则存入缺失切片中
		// 若成功， 则存入匹配结果
		if match == "" {
			missing = append(missing, cr.Name)
		} else {
			extractedVars[cr.Name] = strings.TrimSpace(match)
			remain = strings.Replace(remain, match, "", 1)
		}
	}

	// 对缺失的元素 进行警告提示
	if len(missing) > 0 {
		return originalBaseName, "缺失要素：" + strings.Join(missing, ",")
	}

	newName := template

	for key, val := range extractedVars {
		newName = strings.ReplaceAll(newName, "{"+key+"}", val)
	}

	if strings.Contains(newName, "{") && strings.Contains(newName, "}") {
		return originalBaseName, "格式错误(模板与变量名不匹配)：" + newName
	}

	return newName + ext, ""
}

// 专为前端“单文件快捷修改”提供的底层真实修改接口
func QuickRenameOnDisk(oldPath string, newNameWithExt string) (string, error) {
	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newNameWithExt)

	if oldPath == newPath {
		return newPath, nil
	}
	if _, err := os.Stat(newPath); err == nil {
		return "", fmt.Errorf("目录下已存在同名文件")
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return "", err
	}
	// 返回新的绝对路径，让前端更新状态
	return newPath, nil
}
