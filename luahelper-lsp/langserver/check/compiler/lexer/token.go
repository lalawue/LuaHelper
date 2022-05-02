package lexer

import "strconv"

//TkKind token kind
type TkKind int

// The list of tokens.
const (
	IKIllegal     TkKind = iota      // illegal
	TkEOF                            // end-of-file
	TkVararg                         // ...
	TkSepSemi                        // ;
	TkSepComma                       // ,
	TkSepDot                         // .
	TkSepColon                       // :
	TkSepLabel                       // ::
	TkSepLparen                      // (
	TkSepRparen                      // )
	TkSepLbrack                      // [
	TkSepRbrack                      // ]
	TkSepLcurly                      // {
	TkSepRcurly                      // }
	TkOpAssign                       // =
	TkOpMinus                        // - (sub or unm)
	TkOpWave                         // ~ (bnot or bxor)
	TkOpAdd                          // +
	TkOpMul                          // *
	TkOpDiv                          // /
	TkOpIdiv                         // //
	TkOpPow                          // ^
	TkOpMod                          // %
	TkOpBand                         // &
	TkOpBor                          // |
	TkOpShr                          // >>
	TkOpShl                          // <<
	TkOpConcat                       // ..
	TkOpLt                           // <
	TkOpLe                           // <=
	TkOpGt                           // >
	TkOpGe                           // >=
	TkOpEq                           // ==
	TkOpNe                           // ~=
	TkOpNen                          // #
	TkOpAnd                          // and
	TkOpOr                           // or
	TkOpNot                          // not
	TkKwBreak                        // break
	TkKwCase                         // case
	TkKwClass                        // class
	TkKwContinue                     // continue
	TkKwDefer                        // defer
	TkKwDefault                      // default
	TkKwDo                           // do
	TkKwElse                         // else
	TkKwElseif                       // elseif
	TkKwEnd                          // end
	TkKwExport                       // export
	TkKwExtension                    // extension
	TkKwFalse                        // false
	TkKwFn                           // fn
	TkKwFor                          // for
	TkKwFunction                     // function
	TkKwFrom                         // from
	TkKwGoto                         // goto
	TkKwGuard                        // guard
	TkKwIf                           // if
	TkKwImport                       // import
	TkKwIn                           // in
	TkKwLocal                        // local
	TkKwNil                          // nil
	TkKwPublic                       // public
	TkKwRepeat                       // repeat
	TkKwReturn                       // return
	TkKwStatic                       // static
	TkKwStruct                       // struct
	TkKwSwitch                       // switch
	TkKwThen                         // then
	TkKwTrue                         // true
	TkKwUntil                        // until
	TkKwWhile                        // while
	TkIdentifier                     // identifier
	TkNumber                         // number literal
	TkString                         // string literal
	TkOpUnm       TkKind = TkOpMinus // unary minus
	TkOpSub       TkKind = TkOpMinus
	TkOpBnot      TkKind = TkOpWave
	TkOpBxor      TkKind = TkOpWave
)

var tokenKinds = [...]string{
	IKIllegal: "ILLEGAL",

	TkEOF:         "EOF",            // end-of-file
	TkVararg:      "...",            // ...
	TkSepSemi:     ";",              // ;
	TkSepComma:    ",",              // ,
	TkSepDot:      ".",              // .
	TkSepColon:    ":",              // :
	TkSepLabel:    "::",             // ::
	TkSepLparen:   "(",              // (
	TkSepRparen:   ")",              // )
	TkSepLbrack:   "[",              // [
	TkSepRbrack:   "]",              // ]
	TkSepLcurly:   "{",              // {
	TkSepRcurly:   "}",              // }
	TkOpAssign:    "=",              // =
	TkOpMinus:     "-",              // - (sub or unm)
	TkOpWave:      "~",              // ~ (bnot or bxor)
	TkOpAdd:       "+",              // +
	TkOpMul:       "*",              // *
	TkOpDiv:       "/",              // /
	TkOpIdiv:      "//",             // //
	TkOpPow:       "^",              // ^
	TkOpMod:       "%",              // %
	TkOpBand:      "&",              // &
	TkOpBor:       "|",              // |
	TkOpShr:       ">>",             // >>
	TkOpShl:       "<<",             // <<
	TkOpConcat:    "..",             // ..
	TkOpLt:        "<",              // <
	TkOpLe:        "<=",             // <=
	TkOpGt:        ">",              // >
	TkOpGe:        ">=",             // >=
	TkOpEq:        "==",             // ==
	TkOpNe:        "~=",             // ~=
	TkOpNen:       "#",              // #
	TkOpAnd:       "and",            // and
	TkOpOr:        "or",             // or
	TkOpNot:       "not",            // not
	TkKwBreak:     "break",          // break
	TkKwCase:      "case",           // case
	TkKwClass:     "class",          // class
	TkKwContinue:  "continue",       // continue
	TkKwDefer:     "defer",          // defer
	TkKwDefault:   "default",        // default
	TkKwDo:        "do",             // do
	TkKwElse:      "else",           // else
	TkKwElseif:    "elseif",         // elseif
	TkKwEnd:       "end",            // end
	TkKwExport:    "export",         // export
	TkKwExtension: "extension",      // extension
	TkKwFalse:     "false",          // false
	TkKwFn:        "fn",             // fn
	TkKwFor:       "for",            // for
	TkKwFrom:      "from",           // from
	TkKwFunction:  "function",       // function
	TkKwGoto:      "goto",           // goto
	TkKwIf:        "if",             // if
	TkKwGuard:     "guard",          // guard
	TkKwImport:    "import",         // import
	TkKwIn:        "in",             // in
	TkKwLocal:     "local",          // local
	TkKwNil:       "nil",            // nil
	TkKwPublic:    "public",         // public
	TkKwRepeat:    "repeat",         // repeat
	TkKwReturn:    "return",         // return
	TkKwStatic:    "static",         // static
	TkKwStruct:    "struct",         // struct
	TkKwThen:      "then",           // then
	TkKwTrue:      "true",           // true
	TkKwUntil:     "until",          // until
	TkKwWhile:     "while",          // while
	TkIdentifier:  "identifier",     // identifier
	TkNumber:      "number literal", // number literal
	TkString:      "string literal", // string literal
}

func (tok TkKind) String() string {
	s := ""
	if 0 <= tok && tok < TkKind(len(tokenKinds)) {
		s = tokenKinds[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

var keywords = map[string]TkKind{
	"and":      TkOpAnd,
	"break":    TkKwBreak,
	"do":       TkKwDo,
	"else":     TkKwElse,
	"elseif":   TkKwElseif,
	"end":      TkKwEnd,
	"false":    TkKwFalse,
	"for":      TkKwFor,
	"function": TkKwFunction,
	"goto":     TkKwGoto,
	"if":       TkKwIf,
	"in":       TkKwIn,
	"local":    TkKwLocal,
	"nil":      TkKwNil,
	"not":      TkOpNot,
	"or":       TkOpOr,
	"repeat":   TkKwRepeat,
	"return":   TkKwReturn,
	"then":     TkKwThen,
	"true":     TkKwTrue,
	"until":    TkKwUntil,
	"while":    TkKwWhile,
}

var keywordsMooc = map[string]TkKind{
	"and":       TkOpAnd,
	"break":     TkKwBreak,
	"case":      TkKwCase,
	"class":     TkKwClass,
	"continue":  TkKwContinue,
	"default":   TkKwDefault,
	"defer":     TkKwDefer,
	"do":        TkKwDo,
	"else":      TkKwElse,
	"elseif":    TkKwElseif,
	"end":       TkKwEnd,
	"export":    TkKwExport,
	"extension": TkKwExtension,
	"false":     TkKwFalse,
	"fn":        TkKwFn,
	"for":       TkKwFor,
	"from":      TkKwFrom,
	"function":  TkKwFunction,
	"goto":      TkKwGoto,
	"guard":     TkKwGuard,
	"if":        TkKwIf,
	"import":    TkKwImport,
	"in":        TkKwIn,
	"local":     TkKwLocal,
	"nil":       TkKwNil,
	"not":       TkOpNot,
	"or":        TkOpOr,
	"public":    TkKwPublic,
	"repeat":    TkKwRepeat,
	"return":    TkKwReturn,
	"static":    TkKwStatic,
	"struct":    TkKwStruct,
	"switch":    TkKwSwitch,
	"then":      TkKwThen,
	"true":      TkKwTrue,
	"until":     TkKwUntil,
	"while":     TkKwWhile,
}
