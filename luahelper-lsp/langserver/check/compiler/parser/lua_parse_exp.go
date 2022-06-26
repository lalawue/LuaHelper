package parser

import (
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// explist ::= exp {‘,’ exp}
func (p *luaParser) parseExpList() []ast.Exp {
	l := p.l
	exps := make([]ast.Exp, 0, 1)
	exps = append(exps, p.parseExp())
	for l.LookAheadKind() == lexer.TkSepComma {
		l.NextToken()
		exps = append(exps, p.parseExp())
	}
	return exps
}

/*
exp ::=  nil | false | true | Numeral | LiteralString | ‘...’ | functiondef |
	 prefixexp | tableconstructor | exp binop exp | unop exp
*/
/*
exp   ::= exp12
exp12 ::= exp11 {or exp11}
exp11 ::= exp10 {and exp10}
exp10 ::= exp9 {(‘<’ | ‘>’ | ‘<=’ | ‘>=’ | ‘~=’ | ‘==’) exp9}
exp9  ::= exp8 {‘|’ exp8}
exp8  ::= exp7 {‘~’ exp7}
exp7  ::= exp6 {‘&’ exp6}
exp6  ::= exp5 {(‘<<’ | ‘>>’) exp5}
exp5  ::= exp4 {‘..’ exp4}
exp4  ::= exp3 {(‘+’ | ‘-’) exp3}
exp3  ::= exp2 {(‘*’ | ‘/’ | ‘//’ | ‘%’) exp2}
exp2  ::= {(‘not’ | ‘#’ | ‘-’ | ‘~’)} exp1
exp1  ::= exp0 {‘^’ exp2}
exp0  ::= nil | false | true | Numeral | LiteralString
		| ‘...’ | functiondef | prefixexp | tableconstructor
*/
func (p *luaParser) parseExp() ast.Exp {
	//return parseExp12(l)
	return p.parseSubExp(0)
}

func (p *luaParser) parseSubExp(limit int) ast.Exp {
	l := p.l
	tokenKind := l.LookAheadKind()
	beginBinoLoc := l.GetHeardTokenLoc()

	var exp ast.Exp
	// # - ~ not
	if tokenKind == lexer.TkOpNen || tokenKind == lexer.TkOpUnm || tokenKind == lexer.TkOpBnot ||
		tokenKind == lexer.TkOpNot {
		_, op, _ := l.NextToken()

		beginLoc := l.GetNowTokenLoc()
		argExp := p.parseSubExp(10)
		endLoc := l.GetNowTokenLoc()
		loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
		exp = &ast.UnopExp{
			Op:  op,
			Exp: argExp,
			Loc: loc,
		}
	} else {
		exp = p.parseExp0()
	}

	tokenKind = l.LookAheadKind()

	nowPriority := getPriority(tokenKind)
	for nowPriority > 0 {
		if nowPriority <= limit {
			break
		}

		beforeTokenKind := tokenKind
		if tokenKind == lexer.TkOpPow || tokenKind == lexer.TkOpConcat {
			nowPriority--
		}

		l.NextToken()
		subExp := p.parseSubExp(nowPriority)
		tokenKind = l.LookAheadKind()
		endLoc := l.GetNowTokenLoc()
		loc := lexer.GetRangeLoc(&beginBinoLoc, &endLoc)
		exp = &ast.BinopExp{
			Op:   beforeTokenKind,
			Exp1: exp,
			Exp2: subExp,
			Loc:  loc,
		}
		nowPriority = getPriority(tokenKind)
	}

	return exp
}

func (p *luaParser) parseExp0() ast.Exp {
	l := p.l
	switch l.LookAheadKind() {
	case lexer.TkVararg: // ...
		l.NextToken()
		return &ast.VarargExp{
			Loc: l.GetNowTokenLoc(),
		}
	case lexer.TkKwNil: // nil
		l.NextToken()
		return &ast.NilExp{
			Loc: l.GetNowTokenLoc(),
		}
	case lexer.TkKwTrue: // true
		l.NextToken()
		return &ast.TrueExp{
			Loc: l.GetNowTokenLoc(),
		}
	case lexer.TkKwFalse: // false
		l.NextToken()
		return &ast.FalseExp{
			Loc: l.GetNowTokenLoc(),
		}
	case lexer.TkString: // LiteralString
		_, _, token := l.NextToken()
		// 这里的位置，包含了前后分号
		loc := l.GetNowTokenLoc()
		return &ast.StringExp{
			Str: token,
			Loc: loc,
		}
	case lexer.TkNumber: // Numeral
		return p.parseNumberExp()
	case lexer.TkSepLcurly: // tableconstructor
		return p.parseTableConstructorExp()
	case lexer.TkKwFunction: // functiondef
		l.NextToken()
		beginLoc := l.GetNowTokenLoc()
		return p.parseFuncDefExp(&beginLoc)
	default: // prefixexp
		return p.parsePrefixExp()
	}
}

func (p *luaParser) parseNumberExp() ast.Exp {
	l := p.l
	_, _, token := l.NextToken()
	if i, ok := parseInteger(token); ok {
		return &ast.IntegerExp{
			Val: i,
			Loc: l.GetNowTokenLoc(),
		}
	} else if f, ok := parseFloat(token); ok {
		return &ast.FloatExp{
			Val: f,
			Loc: l.GetNowTokenLoc(),
		}
	} else if n, ok := parseLuajitNum(token); ok {
		return &ast.IntegerExp{
			Val: n,
			Loc: l.GetNowTokenLoc(),
		}
	} else { // todo
		p.insertParserErr(l.GetPreTokenLoc(), "not a number: "+token)
		return &ast.FloatExp{
			Val: 0,
			//Loc: l.GetNowTokenLoc(),
		}
	}
}

// functiondef ::= function funcbody
// funcbody ::= ‘(’ [parlist] ‘)’ block end
func (p *luaParser) parseFuncDefExp(beginLoc *lexer.Location) *ast.FuncDefExp {
	l := p.l
	l.NextTokenKind(lexer.TkSepLparen)                // (
	parList, parLocList, isVararg := p.parseParList() // [parlist]
	l.NextTokenKind(lexer.TkSepRparen)                // )

	blockBeginLoc := l.GetHeardTokenLoc()
	block := p.parseBlock() // block
	blockEndLoc := l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	l.NextTokenKind(lexer.TkKwEnd) // end

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(beginLoc, &endLoc)

	return &ast.FuncDefExp{
		ParList:    parList,
		ParLocList: parLocList,
		IsVararg:   isVararg,
		Block:      block,
		Loc:        loc,
		IsColon:    false,
	}
}

// [parlist]
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
func (p *luaParser) parseParList() (names []string, locVec []lexer.Location, isVararg bool) {
	l := p.l
	switch l.LookAheadKind() {
	case lexer.TkSepRparen:
		return nil, nil, false
	case lexer.TkVararg:
		l.NextToken()
		return nil, nil, true
	}

	_, name := l.NextIdentifier()
	names = append(names, name)
	locVec = append(locVec, l.GetNowTokenLoc())

	for l.LookAheadKind() == lexer.TkSepComma {
		l.NextToken()
		if l.LookAheadKind() == lexer.TkIdentifier {
			_, name := l.NextIdentifier()
			names = append(names, name)
			locVec = append(locVec, l.GetNowTokenLoc())
		} else {
			l.NextTokenKind(lexer.TkVararg)
			isVararg = true
			break
		}
	}
	return
}

// tableconstructor ::= ‘{’ [fieldlist] ‘}’
func (p *luaParser) parseTableConstructorExp() *ast.TableConstructorExp {
	l := p.l
	l.NextTokenKind(lexer.TkSepLcurly) // {
	beginLoc := l.GetNowTokenLoc()
	keyExps, valExps := p.parseFieldList() // [fieldlist]
	l.NextTokenKind(lexer.TkSepRcurly)     // }
	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)

	// 当table的元素过多时，暂时先截断
	if len(keyExps) == len(valExps) && len(keyExps) > 1000 {
		keyExps = keyExps[:1000]
		valExps = valExps[:1000]
	}

	return &ast.TableConstructorExp{
		KeyExps: keyExps,
		ValExps: valExps,
		Loc:     loc,
	}
}

// fieldlist ::= field {fieldsep field} [fieldsep]
func (p *luaParser) parseFieldList() (ks, vs []ast.Exp) {
	l := p.l
	if l.LookAheadKind() != lexer.TkSepRcurly {
		k, v := p.parseField()
		ks = append(ks, k)
		vs = append(vs, v)

		for _isFieldSep(l.LookAheadKind()) {
			l.NextToken()
			if l.LookAheadKind() != lexer.TkSepRcurly {
				k, v := p.parseField()
				ks = append(ks, k)
				vs = append(vs, v)
			} else {
				break
			}
		}
	}
	return
}

// field ::= ‘[’ exp ‘]’ ‘=’ exp | Name ‘=’ exp | exp
func (p *luaParser) parseField() (k, v ast.Exp) {
	l := p.l
	if l.LookAheadKind() == lexer.TkSepLbrack {
		l.NextToken()                      // [
		k = p.parseExp()                   // exp
		l.NextTokenKind(lexer.TkSepRbrack) // ]
		l.NextTokenKind(lexer.TkOpAssign)  // =
		v = p.parseExp()                   // exp
		return
	}

	exp := p.parseExp()
	if nameExp, ok := exp.(*ast.NameExp); ok {
		//loc := l.GetHeardTokenLoc()
		if l.LookAheadKind() == lexer.TkOpAssign {
			// Name ‘=’ exp => ‘[’ LiteralString ‘]’ = exp
			l.NextToken()

			k = &ast.StringExp{
				Str: nameExp.Name,
				Loc: nameExp.Loc,
			}
			v = p.parseExp()
			return
		}
	}

	return nil, exp
}
