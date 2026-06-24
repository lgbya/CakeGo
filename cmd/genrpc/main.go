package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type RpcMethodInfo struct {
	PgkName     string //service的包名
	ServiceName string // 结构体名：RoleService / UserService
	RecvVar     string // 接收器变量：rs / us
	FuncName    string // Rpc方法名：RpcRoleCmd、RpcHeartbeat
	FullDecl    string // 完整方法签名
	FuncBody    string // 方法内部代码
}

var globalMapStr = ""
var globalImportStr = ""
var globalRpcIdMap = make(map[string]bool)

func main() {

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	root := wd + "/internal/game/services"
	writeDir := wd + "/internal/gensvc/rpcgen/"
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // 当前文件访问错误，直接返回或忽略
		}
		if info.IsDir() {
			return nil
		}
		// 过滤只扫描go源码文件
		if filepath.Ext(path) != ".go" {
			return nil
		}

		fileName := info.Name()
		ok := strings.HasSuffix(fileName, "_service.go")
		if !ok {
			return nil
		}

		pkgPath, pkgName := GetPkgInfo(path)
		rpcMethods, err := ParseGoFile(pkgName, path)

		if err != nil {
			return err
		}
		if len(rpcMethods) == 0 {
			return nil
		}

		WriteServiceContext(wd, writeDir, pkgName, pkgPath, rpcMethods)
		return nil
	})
	WriteGlobal(writeDir)
	WriteRpcCmd(writeDir)
}

func WriteRpcCmd(writeDir string) {
	writeDir = writeDir + "rpcid/"
	rpcIdStr := ""
	for rpcId := range globalRpcIdMap {
		rpcIdStr += fmt.Sprintf("const %s = \"%s\"\n", rpcId, rpcId)
	}
	context := fmt.Sprintf(`// Code generated rpcgen.go; DO NOT EDIT
package rpcid	

%s


	`, rpcIdStr)
	WriteFile(writeDir+"rpcid.go", context)

}

func WriteGlobal(writeDir string) {

	context := fmt.Sprintf(`// Code generated rpcgen.go; DO NOT EDIT
package rpcgen	

import (
	"cake/internal/gensvc/router"
%s
	"reflect"
)
	
var regRoutes = map[reflect.Type]map[string]router.RpcFn{
%s
}

func Init() {
	router.Gen = NewGenRouter()
}

type GenRouter struct {
}

func NewGenRouter() router.Router {
	return &GenRouter{}
}

func (*GenRouter) GetRoutes(typ reflect.Type) (map[string]router.RpcFn, bool) {
	routes, ok := regRoutes[typ]
	if !ok {
		return nil, false
	}
	return routes, true
}

	`, globalImportStr, globalMapStr)
	WriteFile(writeDir+"rpcgen.go", context)

}

// IsEmbedRpcService 判断字段是否为匿名内嵌 *rpc.Service
func IsEmbedRpcService(field *ast.Field) bool {
	// 匿名内嵌无字段名
	if len(field.Names) != 0 {
		return false
	}
	// 外层是指针 *
	starExpr, ok := field.Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	// 内部 gensvc.Service
	sel, ok := starExpr.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok || pkgIdent.Name != "rpc" {
		return false
	}
	return sel.Sel.Name == "Service"
}

// GetTargetStructName 获取文件内唯一内嵌 *rpc.Service 的结构体名，无则返回空
func GetTargetStructName(fileNode *ast.File) string {
	var structName string

	ast.Inspect(fileNode, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		st, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}
		// 遍历结构体所有字段
		for _, f := range st.Fields.List {
			if IsEmbedRpcService(f) {
				structName = typeSpec.Name.Name
				return false // 找到直接停止遍历（文件只有一个）
			}
		}
		return true
	})
	return structName
}

func ParseGoFile(pkgName, filePath string) ([]RpcMethodInfo, error) {
	fset := token.NewFileSet()
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	node, err := parser.ParseFile(fset, filePath, src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// 获取当前文件唯一的rpc服务结构体名，为空代表本文件无合法rpc服务
	targetStruct := GetTargetStructName(node)

	if targetStruct == "" {
		return nil, nil
	}
	var rpcMethods []RpcMethodInfo
	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Name == nil {
			return true
		}
		funcName := funcDecl.Name.Name

		// 只匹配 Rpc 开头的方法
		if !strings.HasPrefix(funcName, "Rpc") {
			return true
		}
		// 无接收器直接跳过
		if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			return true
		}

		recvField := funcDecl.Recv.List[0]
		recvVar := recvField.Names[0].Name
		var serviceName string
		switch typ := recvField.Type.(type) {
		case *ast.StarExpr:
			if ident, ok := typ.X.(*ast.Ident); ok {
				serviceName = ident.Name
			}
		case *ast.Ident:
			serviceName = typ.Name
		}

		// 接收器结构体和目标结构体不一致则跳过
		if serviceName != targetStruct {
			return true
		}

		// 截取方法完整签名源码
		startPos := fset.Position(funcDecl.Pos()).Offset
		endPos := fset.Position(funcDecl.End()).Offset
		fullCode := string(src[startPos:endPos])
		splitIdx := strings.Index(fullCode, "{")
		fullDecl := fullCode[:splitIdx+1]

		rpcMethods = append(rpcMethods, RpcMethodInfo{
			PgkName:     pkgName,
			ServiceName: serviceName,
			RecvVar:     recvVar,
			FuncName:    funcName,
			FullDecl:    strings.TrimSpace(fullDecl),
		})
		return true
	})
	return rpcMethods, nil
}

func WriteServiceContext(wd, writeDir, pkgName, pkgPath string, rpcMethods []RpcMethodInfo) {

	tag := FirstUpper(pkgName)
	importStr := fmt.Sprintf("\"%s\"", TrimGamePrefix(wd, pkgPath))
	fnStr := ""
	mapStr := ""
	rpcMethod := rpcMethods[0]
	for _, info := range rpcMethods {
		globalRpcIdMap[info.FuncName] = true
		funName := tag + "_" + info.FuncName
		fnStr += fmt.Sprintf(`func %s(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*%s.%s)
		state,ok := rawState.(*%s.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.%s(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

`, funName, pkgName, info.ServiceName, pkgName, info.FuncName)

		mapStr += fmt.Sprintf("\t\"%s\": %s,\n", info.FuncName, funName)

	}
	mapName := tag + "Map"
	mapStr = fmt.Sprintf(`var %s = map[string]router.RpcFn{
%s
}
`, mapName, mapStr)

	context := fmt.Sprintf(`// Code generated %s.go; DO NOT EDIT
package rpcgen
import (
	"cake/internal/gensvc/router"
	%s
	"errors"
)
%s
%s`, pkgName, importStr, mapStr, fnStr)
	WriteFile(writeDir+pkgName+".go", context)

	globalMapStr += fmt.Sprintf("	reflect.TypeOf((*%s.%s)(nil)):%s,\n", pkgName, rpcMethod.ServiceName, mapName)
	globalImportStr += "	" + importStr + "\n"
}

func WriteFile(filePath string, context string) {
	// 创建目录
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}
	// 创建并清空文件
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)

	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString(context)
}

func FirstUpper(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func GetPkgInfo(filePath string) (string, string) {
	dir := filepath.Dir(filePath)
	pkgName := filepath.Base(dir)
	return dir, pkgName
}

func TrimGamePrefix(wd, absPath string) string {
	return "cake/" + strings.TrimPrefix(absPath, wd+"/")
}
