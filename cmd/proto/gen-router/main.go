package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// 可配置常量，解耦硬编码
const (
	projectModule = "cake" // 项目模块名
	handlerSubDir = "internal/game/handler"
	routerOutFile = "internal/gate/router/routergen.go"
	handlerSuffix = "_handler.go"
)

// getRoot 智能获取项目根目录（优先当前工作目录，兜底可执行文件路径）
func getRoot() (string, error) {
	// 方案1：优先使用执行命令的工作目录（开发环境最稳）
	wd, err := os.Getwd()
	if err == nil {
		return wd, nil
	}

	// 方案2：兜底，从可执行文件向上回溯
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 逐级向上查找（不再写死3层，可按需调整）
	root := filepath.Dir(exePath)
	for i := 0; i < 5; i++ {
		// 简单标记：识别项目根（可根据自己项目改为 go.mod 判断）
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			return root, nil
		}
		root = filepath.Dir(root)
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(exePath))), nil
}

// toPascalCase 替代废弃的 strings.Title，首字母大写（帕斯卡命名）
func toPascalCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// GenerateRoutes 自动生成路由注册代码
func GenerateRoutes() error {
	root, err := getRoot()
	if err != nil {
		return err
	}

	handlerDir := filepath.Join(root, handlerSubDir)
	outPath := filepath.Join(root, routerOutFile)

	// 遍历 handler 目录下所有 .go 文件
	files, err := filepath.Glob(filepath.Join(handlerDir, "*.go"))
	if err != nil {
		return fmt.Errorf("扫描目录失败: %w", err)
	}

	var routeItems []string
	for _, file := range files {
		base := filepath.Base(file)

		// 过滤测试文件、非 handler 文件
		if strings.HasSuffix(base, "_test.go") || !strings.HasSuffix(base, handlerSuffix) {
			continue
		}

		// 截取文件名，生成路由结构体名
		baseName := strings.TrimSuffix(base, handlerSuffix)
		structName := toPascalCase(baseName) + "Route"
		routeItems = append(routeItems, fmt.Sprintf("&handler.%s{}", structName))
	}

	// 模板代码
	tpl := `package router

import (
	"%s/internal/game/handler"
	"%s/internal/gate/router/irouter"
)

// Routes 自动生成路由注册列表，请勿手动修改
func Routes() []irouter.IRoute {
	return []irouter.IRoute{
%s
	}
}
`

	// 高效字符串构建
	var builder strings.Builder
	for _, item := range routeItems {
		builder.WriteString(fmt.Sprintf("		%s,\n", item))
	}

	// 填充模板
	content := fmt.Sprintf(tpl, projectModule, projectModule, builder.String())

	// 写入文件 0644 常规读写权限
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入路由文件失败: %w", err)
	}

	//fmt.Printf("✅ 路由生成成功 | 输出文件: %s\n", outPath)
	fmt.Printf("✅ 路由生成成功\n")
	return nil
}

func main() {
	if err := GenerateRoutes(); err != nil {
		fmt.Printf("❌ 路由生成失败: %v\n", err)
		os.Exit(1)
	}
}
