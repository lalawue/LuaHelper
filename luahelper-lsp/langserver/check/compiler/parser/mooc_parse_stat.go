package parser

import (
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/log"
)

/*
stat ::=  ‘;’
	| break
	| ‘::’ Name ‘::’
	| goto Name
	| do block end
	| while exp do block end
	| repeat block until exp
	| if exp then block {elseif exp then block} [else block] end
	| for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
	| for namelist in explist do block end
	| function funcname funcbody
	| local function Name funcbody
	| local namelist [‘=’ explist]
	| varlist ‘=’ explist
	| functioncall
*/
func (p *moocParser) parseStat() ast.Stat {
	token := p.l.LookAheadKind()
	switch token {
	case lexer.TkSepSemi:
		return p.parseEmptyStat()
	case lexer.TkKwBreak:
		return p.parseBreakStat()
	case lexer.TkSepLabel:
		return p.parseLabelStat()
	case lexer.TkKwGoto:
		return p.parseGotoStat()
	case lexer.TkKwDo:
		return p.parseDoStat(token)
	case lexer.TkKwWhile:
		return p.parseWhileStat()
	case lexer.TkKwRepeat:
		return p.parseRepeatStat()
	case lexer.TkKwIf:
		return p.parseIfStat()
	case lexer.TkKwFor:
		return p.parseForStat()
	case lexer.TkKwFn:
		return p.parseFuncDefStat(false)
	case lexer.TkKwLocal:
		return p.parseLocalAssignOrFuncDefStat()
	case lexer.TkKwClass, lexer.TkKwStruct, lexer.TkKwExtension:
		return p.parserClassDefStat(token)
	case lexer.TkKwStatic:
		return p.parseFuncDefStat(true)
	case lexer.TkKwImport:
		return p.parseImportStat()
	case lexer.TkKwDefer:
		return p.parseDoStat(token)
	case lexer.TkKwExport:
		p.l.NextTokenKind(lexer.TkKwExport)
		if p.l.LookAheadKind() == lexer.TkOpMul {
			p.l.NextTokenKind(lexer.TkOpMul)
			return &ast.ExportAllStat{Loc: p.l.GetNowTokenLoc()}
		}
		if p.l.LookAheadKind() == lexer.TkKwFn {
			return p.parseFuncDefStat(false)
		} else {
			return p.parseAssignOrFuncCallStat(true)
		}
	case lexer.IKIllegal:
		return p.parseIKIllegalStat()
	default:
		return p.parseAssignOrFuncCallStat(false)
	}
}

// ;
func (p *moocParser) parseEmptyStat() *ast.EmptyStat {
	p.l.NextTokenKind(lexer.TkSepSemi)
	return _statEmpty
}

// break
func (p *moocParser) parseBreakStat() *ast.BreakStat {
	p.l.NextTokenKind(lexer.TkKwBreak)

	return &ast.BreakStat{
		//Loc: l.GetNowTokenLoc(),
	}
}

// ‘::’ Name ‘::’
func (p *moocParser) parseLabelStat() *ast.LabelStat {
	p.l.NextTokenKind(lexer.TkSepLabel) // ::
	_, name := p.l.NextIdentifier()     // name
	loc := p.l.GetNowTokenLoc()
	p.l.NextTokenKind(lexer.TkSepLabel) // ::
	return &ast.LabelStat{
		Name: name,
		Loc:  loc,
	}
}

// goto Name
func (p *moocParser) parseGotoStat() *ast.GotoStat {
	p.l.NextTokenKind(lexer.TkKwGoto) // goto
	_, name := p.l.NextIdentifier()   // name
	return &ast.GotoStat{
		Name: name,
		Loc:  p.l.GetNowTokenLoc(),
	}
}

// do block end
func (p *moocParser) parseDoStat(token lexer.TkKind) *ast.DoStat {
	l := p.l
	l.NextTokenKind(lexer.TkKwDo) // do
	beginLoc := l.GetNowTokenLoc()
	l.NextTokenKind(lexer.TkSepLcurly) // {

	if token == lexer.TkKwDefer {
		if p.scopes.checkStackWith(pscope_fn, pscope_fn) == nil {
			p.insertParserErr(l.GetNowTokenLoc(), "defer should inside function body")
		}
		p.scopes.push(pscope_fn, "defer") // defer implement as fn
	} else {
		p.scopes.push(pscope_do, "do")
	}

	blockBeginLoc := l.GetHeardTokenLoc()
	block := p.parseBlock() // block
	blockEndLoc := l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	p.scopes.pop()

	l.NextTokenKind(lexer.TkSepRcurly) // end
	endLoc := l.GetNowTokenLoc()

	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
	return &ast.DoStat{
		Block: block,
		Loc:   loc,
	}
}

// while exp do block end
func (p *moocParser) parseWhileStat() *ast.WhileStat {
	l := p.l
	l.NextTokenKind(lexer.TkKwWhile) // while
	beginLoc := l.GetNowTokenLoc()
	exp := p.parseExp()           // exp
	l.NextTokenKind(lexer.TkKwDo) // do

	blockBeginLoc := l.GetHeardTokenLoc()
	block := p.parseBlock() // block
	blockEndLoc := l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	l.NextTokenKind(lexer.TkKwEnd) // end
	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)

	return &ast.WhileStat{
		Exp:   exp,
		Block: block,
		Loc:   loc,
	}
}

// repeat block until exp
func (p *moocParser) parseRepeatStat() *ast.RepeatStat {
	l := p.l
	l.NextTokenKind(lexer.TkKwRepeat) // repeat
	beginLoc := l.GetNowTokenLoc()

	blockBeginLoc := l.GetHeardTokenLoc()
	block := p.parseBlock() // block
	blockEndLoc := l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	l.NextTokenKind(lexer.TkKwUntil) // until
	exp := p.parseExp()              // exp
	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)

	return &ast.RepeatStat{
		Block: block,
		Exp:   exp,
		Loc:   loc,
	}
}

// if exp then block {elseif exp then block} [else block] end
func (p *moocParser) parseIfStat() *ast.IfStat {
	l := p.l
	exps := make([]ast.Exp, 0, 1)
	blocks := make([]*ast.Block, 0, 1)

	l.NextTokenKind(lexer.TkKwIf) // if
	beginLoc := l.GetNowTokenLoc()

	exps = append(exps, p.parseExp()) // exp
	l.NextTokenKind(lexer.TkKwThen)   // then

	thenBlockBeginLoc := l.GetHeardTokenLoc()
	thenBlock := p.parseBlock()
	thenBlockEndLoc := l.GetHeardTokenLoc()
	thenBlock.Loc = lexer.GetRangeLocExcludeEnd(&thenBlockBeginLoc, &thenBlockEndLoc)
	blocks = append(blocks, thenBlock) // block

	for l.LookAheadKind() == lexer.TkKwElseif {
		l.NextToken()                     // elseif
		exps = append(exps, p.parseExp()) // exp
		l.NextTokenKind(lexer.TkKwThen)   // then

		elseifBlockBeginLoc := l.GetHeardTokenLoc()
		elseifBlock := p.parseBlock()
		elseifBlockEndLoc := l.GetHeardTokenLoc()
		elseifBlock.Loc = lexer.GetRangeLocExcludeEnd(&elseifBlockBeginLoc, &elseifBlockEndLoc)

		blocks = append(blocks, elseifBlock) // block
	}

	// else block => elseif true then block
	if l.LookAheadKind() == lexer.TkKwElse {
		l.NextToken() // else
		exps = append(exps, &ast.TrueExp{
			Loc: l.GetNowTokenLoc(),
		})

		elseBlockBeginLoc := l.GetHeardTokenLoc()
		elseBlock := p.parseBlock()

		elseBlockEndLoc := l.GetHeardTokenLoc()
		elseBlock.Loc = lexer.GetRangeLocExcludeEnd(&elseBlockBeginLoc, &elseBlockEndLoc)

		blocks = append(blocks, elseBlock) // block
	}

	l.NextTokenKind(lexer.TkKwEnd) // end

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
	return &ast.IfStat{
		Exps:   exps,
		Blocks: blocks,
		Loc:    loc,
	}
}

// for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
// for namelist in explist do block end
func (p *moocParser) parseForStat() ast.Stat {
	l := p.l
	lineOfFor, _ := l.NextTokenKind(lexer.TkKwFor)
	beginLoc := l.GetNowTokenLoc()

	_, name := l.NextIdentifier()
	if l.LookAheadKind() == lexer.TkOpAssign {
		return p.finishForNumStat(lineOfFor, name, &beginLoc)
	}
	return p.finishForInStat(name, &beginLoc)
}

// for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
func (p *moocParser) finishForNumStat(lineOfFor int, varName string, beginLoc *lexer.Location) *ast.ForNumStat {
	l := p.l
	varNameLoc := l.GetNowTokenLoc()
	l.NextTokenKind(lexer.TkOpAssign) // for name =
	initExp := p.parseExp()           // exp
	l.NextTokenKind(lexer.TkSepComma) // ,
	limitExp := p.parseExp()          // exp

	var stepExp ast.Exp
	if l.LookAheadKind() == lexer.TkSepComma {
		l.NextToken()          // ,
		stepExp = p.parseExp() // exp
	} else {
		// 这里的位置值可能不太准确
		stepExp = &ast.IntegerExp{
			Val: 1,
			//Loc: l.GetNowTokenLoc(),
		}
	}

	l.NextTokenKind(lexer.TkKwDo) // do

	blockBeginLoc := l.GetHeardTokenLoc()
	block := p.parseBlock() // block
	blockEndLoc := l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	l.NextTokenKind(lexer.TkKwEnd) // end

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(beginLoc, &endLoc)

	return &ast.ForNumStat{
		VarName:  varName,
		VarLoc:   varNameLoc,
		InitExp:  initExp,
		LimitExp: limitExp,
		StepExp:  stepExp,
		Block:    block,
		Loc:      loc,
	}
}

// for namelist in explist do block end
// namelist ::= Name {‘,’ Name}
// explist ::= exp {‘,’ exp}
func (p *moocParser) finishForInStat(name0 string, beginLoc *lexer.Location) *ast.ForInStat {
	l := p.l
	varLoc0 := l.GetNowTokenLoc()
	nameList, nameLocList := p.finishNameList(name0, varLoc0) // for namelist
	l.NextTokenKind(lexer.TkKwIn)                             // in
	expList := p.parseExpList()                               // explist
	l.NextTokenKind(lexer.TkKwDo)                             // do

	blockBeginLoc := l.GetHeardTokenLoc()
	block := p.parseBlock() // block
	blockEndLoc := l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	l.NextTokenKind(lexer.TkKwEnd) // end

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(beginLoc, &endLoc)

	return &ast.ForInStat{
		NameList:    nameList,
		NameLocList: nameLocList,
		ExpList:     expList,
		Block:       block,
		Loc:         loc,
	}
}

// namelist ::= Name {‘,’ Name}
func (p *moocParser) finishNameList(name0 string, varLoc0 lexer.Location) ([]string, []lexer.Location) {
	l := p.l
	names := []string{name0}
	locs := []lexer.Location{varLoc0}
	for l.LookAheadKind() == lexer.TkSepComma {
		l.NextToken()                 // ,
		_, name := l.NextIdentifier() // Name
		loc := l.GetNowTokenLoc()
		locs = append(locs, loc)
		names = append(names, name)
	}
	return names, locs
}

// get local var attribute, add by guochuliang 2020-08-20
func (p *moocParser) getLocalAttribute() ast.LocalAttr {
	l := p.l
	if l.LookAheadKind() == lexer.TkOpLt {
		l.NextToken()
		_, attr := l.NextIdentifier()

		if attr == "close" {
			l.NextTokenKind(lexer.TkOpGt)
			return ast.RDKTOCLOSE
		} else if attr == "const" {
			l.NextTokenKind(lexer.TkOpGt)
			return ast.RDKCONST
		} else {
			p.insertParserErr(l.GetNowTokenLoc(), "unrecognized local varible attribute '%s' ", attr)
			l.NextTokenKind(lexer.TkOpGt)
		}
	}
	return ast.VDKREG
}

//namelist for loacl var after lua 5.4 add support for local var attribute, added by guochuliang 2020-08-20
//5.3
// namelist ::= Name {‘,’ Name}
// 5.4
// namelist ::= Name attrib {‘,’ Name attrib}
// attrib ::= [ '<' Name '>' ] 5.4
func (p *moocParser) finishLocalNameList(name0 string, varLoc0 lexer.Location, kind ast.LocalAttr) ([]string,
	[]lexer.Location, []ast.LocalAttr) {
	l := p.l
	index := -1
	if kind == ast.RDKTOCLOSE {
		index++
	}
	names := []string{name0}
	kinds := []ast.LocalAttr{kind}
	locs := []lexer.Location{varLoc0}
	for l.LookAheadKind() == lexer.TkSepComma {
		l.NextToken()                 // ,
		_, name := l.NextIdentifier() // Name
		loc := l.GetNowTokenLoc()
		kind := p.getLocalAttribute()
		if kind == ast.RDKTOCLOSE {
			if index != -1 {
				p.insertParserErr(l.GetPreTokenLoc(), "more than one to_be_close variables found in local list")
			} else {
				index++
			}
		}
		locs = append(locs, loc)
		kinds = append(kinds, kind)
		names = append(names, name)
	}
	return names, locs, kinds
}

// local function Name funcbody
// local namelist [‘=’ explist]
func (p *moocParser) parseLocalAssignOrFuncDefStat() ast.Stat {
	l := p.l
	l.NextTokenKind(lexer.TkKwLocal)
	if l.LookAheadKind() == lexer.TkKwFunction {
		return p.finishLocalFuncDefStat()
	}

	return p.finishLocalVarDeclStat()
}

/*
http://www.lua.org/manual/5.3/manual.html#3.4.11

function f() end          =>  f = function() end
function t.a.b.c.f() end  =>  t.a.b.c.f = function() end
function t.a.b.c:f() end  =>  t.a.b.c.f = function(self) end
local function f() end    =>  local f; f = function() end

The statement `local function f () body end`
translates to `local f; f = function () body end`
not to `local f = function () body end`
(This only makes a difference when the body of the function
 contains references to f.)
*/
// local function Name funcbody
func (p *moocParser) finishLocalFuncDefStat() *ast.LocalFuncDefStat {
	l := p.l
	beginLoc := l.GetNowTokenLoc()

	l.NextTokenKind(lexer.TkKwFunction) // local function
	_, name := l.NextIdentifier()       // name
	nameLoc := l.GetNowTokenLoc()
	fdExp := p.parseFuncDefExp(false, &beginLoc) // funcbody

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
	return &ast.LocalFuncDefStat{
		Name:    name,
		NameLoc: nameLoc,
		Exp:     fdExp,
		Loc:     loc,
	}
}

// local namelist [‘=’ explist]
func (p *moocParser) finishLocalVarDeclStat() *ast.LocalVarDeclStat {
	l := p.l
	beginLoc := l.GetNowTokenLoc()
	_, name0 := l.NextIdentifier() // local Name
	varLoc0 := l.GetNowTokenLoc()
	kind0 := p.getLocalAttribute()                                              // added to support lua5.4
	nameList, locList, attrList := p.finishLocalNameList(name0, varLoc0, kind0) // { , Name attrib}
	var expList []ast.Exp
	if l.LookAheadKind() == lexer.TkOpAssign {
		l.NextToken()              // ==
		expList = p.parseExpList() // explist
	}

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
	return &ast.LocalVarDeclStat{
		NameList:   nameList,
		VarLocList: locList,
		AttrList:   attrList,
		ExpList:    expList,
		Loc:        loc,
	}
}

// varlist ‘=’ explist
// functioncall
func (p *moocParser) parseAssignOrFuncCallStat(isExport bool) ast.Stat {
	l := p.l
	beginLoc := l.GetHeardTokenLoc()
	prefixExp := p.parsePrefixExp()
	if _, ok := prefixExp.(*ast.BadExpr); ok {
		return &ast.EmptyStat{}
	}

	if fc, ok := prefixExp.(*ast.FuncCallExp); ok {
		endLoc := l.GetNowTokenLoc()
		fc.Loc = lexer.GetRangeLoc(&beginLoc, &endLoc)
		return fc
	}

	assignStat := p.parseAssignStat(beginLoc, prefixExp)
	switch v := assignStat.(type) {
	case *ast.AssignStat:
		if isExport {
			v.Attr = ast.VDKEXPORT
		}
	}

	return assignStat
}

// varlist ‘=’ explist |
func (p *moocParser) parseAssignStat(preLoc lexer.Location, var0 ast.Exp) ast.Stat {
	l := p.l
	symList := p.finishVarList(var0) // varlist

	aheadKind := l.LookAheadKind()
	if len(symList) == 1 &&
		(aheadKind == lexer.TkOpMul ||
			aheadKind == lexer.TkOpDiv ||
			aheadKind == lexer.TkOpMod ||
			aheadKind == lexer.TkOpAdd ||
			aheadKind == lexer.TkOpMinus ||
			aheadKind == lexer.TkOpConcat ||
			aheadKind == lexer.TkOpOr ||
			aheadKind == lexer.TkOpAnd ||
			aheadKind == lexer.TkOpPow) {
		l.NextToken()

		if l.LookAheadKind() != lexer.TkOpAssign {
			nowLoc := l.GetNowTokenLoc()
			loc := lexer.GetRangeLoc(&preLoc, &nowLoc)
			p.insertParserErr(loc, "expression cannot be used as a statement")
			return &ast.EmptyStat{}
		}

		l.NextTokenKind(lexer.TkOpAssign)
		expList := p.parseExpList() // explist
		endLoc := l.GetNowTokenLoc()
		loc := lexer.GetRangeLoc(&preLoc, &endLoc)
		return &ast.AssignStat{
			VarList: symList,
			ExpList: []ast.Exp{&ast.BinopExp{
				Op:   aheadKind,
				Exp1: symList[0],
				Exp2: expList,
				Loc:  loc,
			}},
			Loc: loc,
		}
	} else {
		if aheadKind != lexer.TkOpAssign {
			nowLoc := l.GetNowTokenLoc()
			loc := lexer.GetRangeLoc(&preLoc, &nowLoc)
			p.insertParserErr(loc, "expression cannot be used as a statement")
			return &ast.EmptyStat{}
		}

		l.NextTokenKind(lexer.TkOpAssign) // =
		expList := p.parseExpList()       // explist
		endLoc := l.GetNowTokenLoc()
		loc := lexer.GetRangeLoc(&preLoc, &endLoc)

		return &ast.AssignStat{
			VarList: symList,
			ExpList: expList,
			Loc:     loc,
		}
	}
}

// varlist ::= var {‘,’ var}
func (p *moocParser) finishVarList(var0 ast.Exp) []ast.Exp {
	l := p.l
	vars := []ast.Exp{p.checkVar(var0)}         // var := p
	for l.LookAheadKind() == lexer.TkSepComma { // {
		l.NextToken()                        // ,
		exp := p.parsePrefixExp()            // var
		vars = append(vars, p.checkVar(exp)) //
	} // }
	return vars
}

// var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
func (p *moocParser) checkVar(exp ast.Exp) ast.Exp {
	l := p.l
	switch exp.(type) {
	case *ast.NameExp, *ast.TableAccessExp, *ast.BadExpr:
		return exp
	}

	loc := l.GetNowTokenLoc()
	return &ast.BadExpr{
		Loc: loc,
	}
	// l.NextTokenKind(-1) // trigger error
	// panic("unreachable!")
}

// function funcname funcbody
// funcname ::= Name {‘.’ Name} [‘:’ Name]
// funcbody ::= ‘(’ [parlist] ‘)’ block end
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
// namelist ::= Name {‘,’ Name}
func (p *moocParser) parseFuncDefStat(isStaticAttr bool) *ast.AssignStat {
	l := p.l
	if isStaticAttr {
		l.NextTokenKind(lexer.TkKwStatic)
	}
	l.NextTokenKind(lexer.TkKwFn) // function
	beginLoc := l.GetNowTokenLoc()
	fnExp, hasColon := p.parseFuncName() // funcname
	selfLoc := l.GetNowTokenLoc()
	fdExp := p.parseFuncDefExp(false, &beginLoc) // funcbody
	if hasColon {                                // insert self
		fdExp.ParList = append(fdExp.ParList, "")
		copy(fdExp.ParList[1:], fdExp.ParList)
		fdExp.ParList[0] = "self"
		fdExp.IsColon = true

		fdExp.ParLocList = append(fdExp.ParLocList, lexer.Location{})
		copy(fdExp.ParLocList[1:], fdExp.ParLocList)
		fdExp.ParLocList[0] = selfLoc
	}

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
	return &ast.AssignStat{
		VarList: []ast.Exp{fnExp},
		ExpList: []ast.Exp{fdExp},
		Loc:     loc,
	}
}

// funcname ::= Name {‘.’ Name} [‘:’ Name]
func (p *moocParser) parseFuncName() (exp ast.Exp, hasColon bool) {
	l := p.l
	_, name := l.NextIdentifier()
	loc := l.GetNowTokenLoc()

	beginTableLoc := l.GetNowTokenLoc()
	exp = &ast.NameExp{
		Name: name,
		Loc:  loc,
	}

	for l.LookAheadKind() == lexer.TkSepDot {
		l.NextToken()
		_, name := l.NextIdentifier()
		loc := l.GetNowTokenLoc()
		idx := &ast.StringExp{
			Str: name,
			Loc: loc,
		}

		endTableLoc := l.GetNowTokenLoc()
		tableLoc := lexer.GetRangeLoc(&beginTableLoc, &endTableLoc)

		exp = &ast.TableAccessExp{
			PrefixExp: exp,
			KeyExp:    idx,
			Loc:       tableLoc,
		}
	}
	if l.LookAheadKind() == lexer.TkSepColon {
		l.NextToken()
		_, name := l.NextIdentifier()
		loc := l.GetNowTokenLoc()
		idx := &ast.StringExp{
			Str: name,
			Loc: loc,
		}

		endTableLoc := l.GetNowTokenLoc()
		tableLoc := lexer.GetRangeLoc(&beginTableLoc, &endTableLoc)

		exp = &ast.TableAccessExp{
			PrefixExp: exp,
			KeyExp:    idx,
			Loc:       tableLoc,
		}
		hasColon = true
	}

	return
}

// func (p *moocParser) parseIKIllegalStat() *ast.IllegalStat{
// 	l := p.l
// 	loc := l.GetNowTokenLoc()
// 	l.NextToken()
// 	return &ast.IllegalStat{
// 		Name: "",
// 		Loc:  loc,
// 	}
// }

func (p *moocParser) parseIKIllegalStat() *ast.EmptyStat {
	p.l.NextToken()
	return _statEmpty
}

// class identifier {
//	fn name() {
//		return b
//	}
// }
func (p *moocParser) parserClassDefStat(token lexer.TkKind) ast.Stat {
	log.Debug("parserClassDefStat begin")
	l := p.l

	beginLoc := l.GetNowTokenLoc()
	l.NextTokenKind(token)

	_, cname := l.NextIdentifier() // name
	cnameLoc := l.GetNowTokenLoc()

	var super *ast.NameExp
	if l.LookAheadKind() == lexer.TkSepColon {
		if token == lexer.TkKwClass || token == lexer.TkKwExtension {
			l.NextTokenKind(lexer.TkSepColon)
		} else {
			p.insertParserErr(l.GetNowTokenLoc(), "struct can not inherit")
		}
		_, sname := l.NextIdentifier()
		super = &ast.NameExp{
			Name: sname,
			Loc:  l.GetNowTokenLoc(),
		}
	}

	p.scopes.push(pscope_cl, cname)

	// 类名
	nameList := []string{cname}
	locList := []lexer.Location{cnameLoc}
	attrList := []ast.LocalAttr{ast.VDKREG}
	expList := []ast.Exp{&ast.TableConstructorExp{
		Loc: cnameLoc,
	}}
	nameStat := &ast.LocalVarDeclStat{
		NameList:   nameList,
		VarLocList: locList,
		AttrList:   attrList,
		ExpList:    expList,
		Loc:        cnameLoc,
	}

	l.NextTokenKind(lexer.TkSepLcurly)

	vfList := []*ast.AssignStat{}
	for {
		token := l.LookAheadKind()
		if token == lexer.TkKwStatic || token == lexer.TkKwFn {
			// 类函数
			vfList = append(vfList, p.parseFuncDefStat(token == lexer.TkKwStatic))
		} else {
			if token == lexer.TkIdentifier {
				// 类变量，模拟 tableAccess 的 assign stat
				cexp := &ast.NameExp{
					Name: cname,
					Loc:  beginLoc,
				}
				_, vname := l.NextIdentifier() // Name
				endLoc := l.GetNowTokenLoc()
				tableLoc := lexer.GetRangeLoc(&beginLoc, &endLoc)
				keyExp := &ast.StringExp{
					Str: vname,
					Loc: endLoc,
				}
				prefixExp := &ast.TableAccessExp{
					PrefixExp: cexp,
					KeyExp:    keyExp,
					Loc:       tableLoc,
				}
				switch v := p.parseAssignStat(tableLoc, prefixExp).(type) {
				case *ast.AssignStat:
					vfList = append(vfList, v)
				}
			} else {
				break
			}
		}
	}

	l.NextTokenKind(lexer.TkSepRcurly)
	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
	log.Debug("parserClassDefStat end")
	p.scopes.pop()
	return &ast.ClassDefStat{
		SType: token,
		Super: super,
		Name:  nameStat,
		List:  vfList,
		Loc:   loc,
	}
}

// import A from "a"
// import A, B from "a" {}
// import A, B from "a" { a, b }
// import concat, remove from table {}
func (p *moocParser) parseImportStat() ast.Stat {
	log.Debug("parse import stat begin")
	l := p.l

	beginLoc := l.GetNowTokenLoc()
	l.NextTokenKind(lexer.TkKwImport)
	importLoc := l.GetNowTokenLoc()

	var libStat *ast.FuncCallStat
	var nameStat *ast.LocalVarDeclStat

	if l.LookAheadKind() == lexer.TkString {
		// import "lpeg" as require "lpeg"
		prefixExp := &ast.ParensExp{
			Exp: ast.NameExp{
				Name: "require",
				Loc:  lexer.GetRangeLoc(&beginLoc, &importLoc),
			},
		}
		args := p.parseArgs()
		endLoc := l.GetNowTokenLoc()
		libStat = &ast.FuncCallExp{
			PrefixExp: prefixExp,
			Args:      args,
			Loc:       lexer.GetRangeLoc(&beginLoc, &endLoc),
		}
	} else {
		//var toList []string

		// 变量列表, local A, B = nil, nil
		beginLoc := l.GetNowTokenLoc()
		_, name0 := l.NextIdentifier() // local Name
		varLoc0 := l.GetNowTokenLoc()
		kind0 := p.getLocalAttribute()                                              // added to support lua5.4
		nameList, locList, attrList := p.finishLocalNameList(name0, varLoc0, kind0) // { , Name attrib}

		// from
		l.NextTokenKind(lexer.TkKwFrom)
		fromLoc := l.GetNowTokenLoc()

		// 获取库名称
		var libStr string
		var libLib string
		var libLoc lexer.Location

		if l.LookAheadKind() == lexer.TkString {
			_, libStr = l.NextTokenKind(lexer.TkString)
		} else {
			_, libLib = l.NextIdentifier()
		}
		libLoc = l.GetNowTokenLoc()

		var expList []ast.Exp

		// 检查是否是子库
		if l.LookAheadKind() == lexer.TkSepLcurly {
			// import A from "lib" {}
			// import concat from table {}
			l.NextTokenKind(lexer.TkSepLcurly)
			expList = []ast.Exp{}
			count := 0
			for {
				if l.LookAheadKind() == lexer.TkIdentifier {
					_, key := l.NextIdentifier()
					if len(libLib) > 0 {
						expList = append(expList, p.importSubLib(libLib, "", key, l.GetNowTokenLoc()))
					} else {
						expList = append(expList, p.importSubLib("require", libStr, key, l.GetNowTokenLoc()))
					}
				} else if count <= 0 {
					for _, name := range nameList {
						if len(libLib) > 0 {
							expList = append(expList, p.importSubLib(libLib, "", name, l.GetNowTokenLoc()))
						} else {
							expList = append(expList, p.importSubLib("require", libStr, name, l.GetNowTokenLoc()))
						}
					}
				}
				count += 1
				if l.LookAheadKind() != lexer.TkSepComma {
					break
				}
				l.NextToken()
			}
			l.NextTokenKind(lexer.TkSepRcurly)
			if len(nameList) != len(expList) {
				panic("name list and var list should be same")
			}
		} else {
			// import A from "a"
			if len(libStr) <= 0 {
				p.insertParserErr(l.GetPreTokenLoc(), "invalid libname, should be string")
			}
			prefixExp := &ast.NameExp{
				Name: "require",
				Loc:  lexer.GetRangeLoc(&fromLoc, &fromLoc),
			}
			fnCallExp := &ast.FuncCallExp{
				PrefixExp: prefixExp,
				Args: []ast.Exp{&ast.StringExp{
					Str: libStr,
					Loc: libLoc,
				}},
				Loc: lexer.GetRangeLoc(&fromLoc, &libLoc),
			}
			expList = []ast.Exp{fnCallExp}
		}

		endLoc := l.GetNowTokenLoc()
		nameStat = &ast.LocalVarDeclStat{
			NameList:   nameList,
			VarLocList: locList,
			AttrList:   attrList,
			ExpList:    expList,
			Loc:        lexer.GetRangeLoc(&beginLoc, &endLoc),
		}
	}

	log.Debug("parse import stat end")
	return &ast.ImportDefStat{
		Lib:  libStat,
		Name: nameStat,
	}
}

// 返回 table.concat 此时 lib 是 ""，或者 require("table").concat，此时 lib 是 "table"
func (p *moocParser) importSubLib(fnName string, argName string, tblKey string, loc lexer.Location) *ast.TableAccessExp {
	prefixExp := &ast.NameExp{
		Name: fnName,
		Loc:  loc,
	}
	idx := &ast.StringExp{
		Str: tblKey,
		Loc: loc,
	}
	if len(argName) <= 0 {
		return &ast.TableAccessExp{
			PrefixExp: prefixExp,
			KeyExp:    idx,
			Loc:       loc,
		}
	} else {
		fnCallExp := &ast.FuncCallExp{
			PrefixExp: prefixExp,
			Args: []ast.Exp{&ast.StringExp{
				Str: argName,
				Loc: loc,
			}},
		}
		return &ast.TableAccessExp{
			PrefixExp: fnCallExp,
			KeyExp:    idx,
			Loc:       loc,
		}
	}
}
