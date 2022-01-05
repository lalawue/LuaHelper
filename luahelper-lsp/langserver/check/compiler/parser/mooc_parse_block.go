package parser

import (
	"fmt"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// block ::= {stat} [retstat]
func (p *moocParser) parseBlock() *ast.Block {
	return &ast.Block{
		Stats:   p.parseStats(),
		RetExps: p.parseRetExps(),
	}
}

func (p *moocParser) parseStats() []ast.Stat {
	stats := make([]ast.Stat, 0, 1)
	for !p.isReturnOrBlockEnd(p.l.LookAheadKind()) {
		stat := p.parseStat()
		if _, ok := stat.(*ast.EmptyStat); !ok {
			stats = append(stats, stat)
		}
	}
	current := p.scopes.current()
	if current != nil && current.scope == pscope_lo && current.count > 0 {
		stats = append(stats, &ast.LabelStat{
			Name: fmt.Sprintf("__continue%d", p.scopes.checkStackIndex(0).count),
			Loc:  p.l.GetNowTokenLoc(),
		})
	}
	return stats
}

// retstat ::= return [explist] [‘;’]
// explist ::= exp {‘,’ exp}
func (p *moocParser) parseRetExps() []ast.Exp {
	l := p.l
	if l.LookAheadKind() != lexer.TkKwReturn {
		return nil
	}

	current := p.scopes.current()
	if current != nil && (current.scope == pscope_do || current.scope == pscope_gu) {
		current.count = 1
	}

	l.NextToken()
	switch l.LookAheadKind() {
	case lexer.TkEOF, lexer.TkSepRcurly, lexer.TkKwCase, lexer.TkKwDefault:
		return []ast.Exp{}
	case lexer.TkSepSemi:
		l.NextToken()
		return []ast.Exp{}
	default:
		exps := p.parseExpList()
		if l.LookAheadKind() == lexer.TkSepSemi {
			l.NextToken()
		}
		return exps
	}
}

func (p *moocParser) isReturnOrBlockEnd(tokenKind lexer.TkKind) bool {
	switch tokenKind {
	case lexer.TkKwReturn, lexer.TkEOF, lexer.TkSepRcurly, lexer.TkKwCase, lexer.TkKwDefault:
		return true
	}
	return false
}
