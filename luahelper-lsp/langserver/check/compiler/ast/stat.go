package ast

import (
	"luahelper-lsp/langserver/check/compiler/lexer"
)

/*
stat ::=  ‘;’ |
	 varlist ‘=’ explist |
	 functioncall |
	 label |
	 break |
	 goto Name |
	 do block end |
	 while exp do block end |
	 repeat block until exp |
	 if exp then block {elseif exp then block} [else block] end |
	 for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end |
	 for namelist in explist do block end |
	 function funcname funcbody |
	 local function Name funcbody |
	 local namelist [‘=’ explist]
*/

// Stat 空接口
type Stat interface{}

// EmptyStat 空的
type EmptyStat struct{} // ‘;’

// BreakStat break语句
// break
type BreakStat struct {
	//Loc lexer.LocInfo
}

// LabelStat goto对应的标识符
// ‘::’ Name ‘::’
type LabelStat struct {
	Name string
	Loc  lexer.Location
}

// GotoStat goto语句
// goto Name
type GotoStat struct {
	Name string
	Loc  lexer.Location
}

// DoStat do代码块
// do block end
type DoStat struct {
	Stype lexer.TkKind
	Block *Block
	Loc   lexer.Location
}

// FuncCallStat 函数调用
type FuncCallStat = FuncCallExp // functioncall

// IfStat if代码块
// if exp then block {elseif exp then block} [else block] end
type IfStat struct {
	Exps   []Exp
	Blocks []*Block
	Loc    lexer.Location
}

// WhileStat while代码块
// while exp do block end
type WhileStat struct {
	Exp   Exp
	Block *Block
	Loc   lexer.Location
}

// RepeatStat repeat 代码块
// repeat block until exp
type RepeatStat struct {
	Block *Block
	Exp   Exp
	Loc   lexer.Location
}

// ForNumStat for 整数遍历
// for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
type ForNumStat struct {
	VarName  string
	VarLoc   lexer.Location
	InitExp  Exp
	LimitExp Exp
	StepExp  Exp
	Block    *Block
	Loc      lexer.Location
}

// ForInStat for语句
// for namelist in explist do block end
// namelist ::= Name {‘,’ Name}
// explist ::= exp {‘,’ exp}
type ForInStat struct {
	NameList    []string
	NameLocList []lexer.Location // 所有变量的位置信息
	ExpList     []Exp
	Block       *Block
	Loc         lexer.Location
}

// AssignStat 赋值语句
// varlist ‘=’ explist
// varlist ::= var {‘,’ var}
// var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
type AssignStat struct {
	VarList []Exp
	ExpList []Exp
	Loc     lexer.Location
	Attr    LocalAttr
}

// LocalVarDeclStat 局部变量定义
// local namelist [‘=’ explist]
// namelist ::= Name {‘,’ Name}
// explist ::= exp {‘,’ exp}
type LocalVarDeclStat struct {
	NameList   []string
	VarLocList []lexer.Location // 所有变量的位置信息
	AttrList   []LocalAttr      // 变量的属性
	ExpList    []Exp
	Loc        lexer.Location
}

// LocalFuncDefStat local function Name funcbody
type LocalFuncDefStat struct {
	Name    string
	NameLoc lexer.Location // 函数名的位置信息
	Exp     *FuncDefExp
	Loc     lexer.Location // 整体函数的位置信息
}

// IllegalStat Illegal stat token
type IllegalStat struct {
	Name string
	Loc  lexer.Location
}

/** stat below for moocscript only
 */

// class clsname {
//		fn fnName () {
//		}
// }
type ClassDefStat struct {
	SType lexer.TkKind        // 可以是 class，struct 或者 extension
	Class *AssignStat         // 可能是 export 的
	Super Exp                 // super class name
	Vars  []*LocalVarDeclStat // Self，Super
	List  []*AssignStat       // 变量，函数
	Loc   lexer.Location      // 整体类的位置信息
}

// import "lpeg"
// import lpeg from "lpeg"
// import P, R, S from "lpeg" {}
// import p, r, s from "lpeg" { P, R, S }
// import insert, remove from table {}
type ImportDefStat struct {
	Lib  *FuncCallStat     // require("lpeg")
	Name *LocalVarDeclStat // local R, P, S = require("lpeg").R, require("lpeg").P, require("lpeg").S
}

// export *
type ExportAllStat struct {
	Loc lexer.Location
}

type SwitchStat struct {
	Name *LocalVarDeclStat // 包装为一个 local table
	Case *IfStat
	Loc  lexer.Location
}
