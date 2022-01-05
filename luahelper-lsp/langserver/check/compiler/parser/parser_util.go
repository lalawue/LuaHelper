package parser

import (
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

var _statEmpty = &ast.EmptyStat{}

// 通过tokenKind 获取优先级
func getPriority(tokenKind lexer.TkKind) int {
	switch tokenKind {
	case lexer.TkOpPow: // ^
		return 12
	case lexer.TkOpMul, lexer.TkOpMod, lexer.TkOpDiv, lexer.TkOpIdiv: // *, %, /, //
		return 10
	case lexer.TkOpAdd, lexer.TkOpSub: // +, -
		return 9
	case lexer.TkOpConcat: // ..
		return 8
	case lexer.TkOpShl, lexer.TkOpShr: // shift,  <<  >>
		return 7
	case lexer.TkOpBand: // &
		return 6
	case lexer.TkOpBxor: // x ~ y
		return 5
	case lexer.TkOpBor: // x | y
		return 4
	case lexer.TkOpLt, lexer.TkOpGt, lexer.TkOpNe,
		lexer.TkOpLe, lexer.TkOpGe, lexer.TkOpEq: // (‘<’ | ‘>’ | ‘<=’ | ‘>=’ | ‘~=’ | ‘==’)
		return 3
	case lexer.TkOpAnd: // x and y
		return 2
	case lexer.TkOpOr: // x or y
		return 1
	}

	return 0
}

// fieldsep ::= ‘,’ | ‘;’
func _isFieldSep(tokenKind lexer.TkKind) bool {
	return tokenKind == lexer.TkSepComma || tokenKind == lexer.TkSepSemi
}
