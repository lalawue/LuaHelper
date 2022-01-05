package parser

import (
	"fmt"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

const (
	pscope_gl = 1 // global
	pscope_pj = 2 // project
	pscope_fi = 3 // file
	pscope_cl = 4 // class
	pscope_fn = 5 // fn
	pscope_lo = 6 // loop (for, while, repeat)
	pscope_if = 7 // if
	pscope_do = 8 // do
	pscope_gu = 9 // guard
)

type parserScope struct {
	scope uint8
	name  string
	count uint8 // for fi scope, defer/guard scope
}

type scopeStack struct {
	index  uint8
	scopes []parserScope
}

func (s *scopeStack) push(scope uint8, name string) {
	s.scopes[s.index] = parserScope{scope: scope, name: name}
	s.index += 1
	if scope == pscope_lo {
		// file scope: count loop for continue
		s.scopes[0].count += 1
	}
}

func (s *scopeStack) pop() *parserScope {
	ret := s.scopes[s.index]
	if s.index > 0 {
		s.index -= 1
		if ret.scope == pscope_lo {
			s.scopes[0].count -= 1
		}
		return &ret
	} else {
		return nil
	}
}

func (p *scopeStack) current() *parserScope {
	if p.index <= 0 {
		return nil
	}
	return &p.scopes[p.index-1]
}

// check scope before stop scope
func (p *scopeStack) checkStackWith(scope uint8, stop uint8) *parserScope {
	for index := int(p.index); index >= 0; index-- {
		iscope := p.scopes[index].scope
		if iscope == scope {
			return &p.scopes[index]
		} else if iscope == stop {
			return nil
		}
	}
	return nil
}

func (p *scopeStack) checkStackIndex(index uint8) *parserScope {
	if index >= p.index {
		return nil
	}
	return &p.scopes[index]
}

// Parser 语法分析器，为把Lua源码文件解析成AST抽象语法树
type moocParser struct {
	// 词法分析器对象
	l *lexer.Lexer
	// 作用域栈
	scopes    scopeStack
	parseErrs []lexer.ParseError
}

// CreateParser 创建一个分析对象
func createMoocParser(chunk []byte, chunkName string) *moocParser {
	parser := &moocParser{}
	errHandler := parser.insertErr
	parser.l = lexer.NewLexer(chunk, chunkName)
	parser.l.SetErrHandler(errHandler)

	return parser
}

// BeginAnalyze 开始分析
func (p *moocParser) BeginAnalyze() (block *ast.Block, commentMap map[int]*lexer.CommentInfo, errList []lexer.ParseError) {
	defer func() {
		if err1 := recover(); err1 != nil {
			block = &ast.Block{}
			commentMap = p.l.GetCommentMap()
			errList = p.parseErrs
			return
		}
	}()

	p.l.SkipFirstLineComment()

	blockBeginLoc := p.l.GetHeardTokenLoc()
	block = p.parseBlock() // block
	blockEndLoc := p.l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	p.l.NextTokenKind(lexer.TkEOF)
	p.l.SetEnd()
	return block, p.l.GetCommentMap(), p.parseErrs
}

// BeginAnalyzeExp ParseExp single exp
func (p *moocParser) BeginAnalyzeExp() (exp ast.Exp) {
	defer func() {
		if err2 := recover(); err2 != nil {
			exp = nil
		}
	}()

	exp = p.parseSubExp(0)
	return exp
}

// GetErrList get parse error list
func (p *moocParser) GetErrList() (errList []lexer.ParseError) {
	return p.parseErrs
}

// insert now token info
func (p *moocParser) insertParserErr(loc lexer.Location, f string, a ...interface{}) {
	err := fmt.Sprintf(f, a...)
	paseError := lexer.ParseError{
		ErrStr:      err,
		Loc:         loc,
		ReadFileErr: false,
	}

	p.insertErr(paseError)
}

func (p *moocParser) insertErr(oneErr lexer.ParseError) {
	if len(p.parseErrs) < 30 {
		p.parseErrs = append(p.parseErrs, oneErr)
	} else {
		oneErr.ErrStr = oneErr.ErrStr + "(too many err...)"
		p.parseErrs = append(p.parseErrs, oneErr)
		manyError := &lexer.TooManyErr{
			ErrNum: 30,
		}

		panic(manyError)
	}
}
