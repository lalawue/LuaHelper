package parser

import (
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"strings"
)

type ParserInterface interface {
	BeginAnalyze() (block *ast.Block, commentMap map[int]*lexer.CommentInfo, errList []lexer.ParseError)
	BeginAnalyzeExp() (exp ast.Exp)
	GetErrList() (errList []lexer.ParseError)
}

// CreateParser 创建一个分析对象
func CreateParser(chunk []byte, chunkName string) ParserInterface {
	if ok := strings.HasSuffix(chunkName, ".mooc"); ok {
		return createMoocParser(chunk, chunkName)
	} else {
		return createLuaParser(chunk, chunkName)
	}
}
