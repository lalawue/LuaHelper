package langserver

import (
	"context"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/lspcommon"
	"luahelper-lsp/langserver/pathpre"
	lsp "luahelper-lsp/langserver/protocol"
	"strings"
)

// TextDocumentSymbol 提示文件中生成所有的符合 @使用
func (l *LspServer) TextDocumentSymbol(ctx context.Context, vs lsp.DocumentSymbolParams) (itemsResult []lsp.DocumentSymbol, err error) {
	strFile := pathpre.VscodeURIToString(string(vs.TextDocument.URI))
	project := l.getAllProject()
	if !project.IsNeedHandle(strFile) {
		log.Debug("not need to handle strFile=%s", strFile)
		return
	}

	// 文件截断后的名字
	fileSymbolVec := project.FindFileAllSymbol(strFile)

	// 将filesymbols 转换为 lsp document
	itemsResult = transferSymbolVec(strFile, 0, fileSymbolVec)
	return
}

// transferSymbolVec 转换fileSybmols为DocumentSymbol
func transferSymbolVec(strFile string, level int, fileSymbolVec []common.FileSymbolStruct) (items []lsp.DocumentSymbol) {
	vecLen := len(fileSymbolVec)
	items = make([]lsp.DocumentSymbol, 0, vecLen)

	isMooc := strings.HasSuffix(strFile, ".mooc")
	lastFuncSymbol := lsp.DocumentSymbol{}

	for _, oneSymbol := range fileSymbolVec {
		ra := lspcommon.LocToRange(&oneSymbol.Loc)

		fullName := oneSymbol.Name
		if oneSymbol.ContainerName == "" {
			if isMooc {
				// mark 标记，或者丢掉多余的类前缀
				if oneSymbol.Kind == common.IKAnnotateMark {
					fullName = "---"
				} else if level <= 0 {
					fullName = "export " + fullName
				} else {
					idx := strings.Index(fullName, ".")
					if idx > 0 {
						if oneSymbol.Kind == common.IKFunction {
							fullName = fullName[idx:]
						} else {
							fullName = fullName[idx+1:]
						}
					} else {
						idx = strings.Index(fullName, ":")
						if idx > 0 && oneSymbol.Kind == common.IKFunction {
							fullName = fullName[idx:]
						}
					}
				}
			}
		} else {
			if oneSymbol.ContainerName == "local" {
				if !isMooc {
					fullName = oneSymbol.ContainerName + " " + fullName
				}
			} else {
				fullName = oneSymbol.ContainerName + "." + fullName
			}
		}

		symbol := lsp.DocumentSymbol{
			Name:           fullName,
			Kind:           lsp.Variable,
			Range:          ra,
			SelectionRange: ra,
		}

		if oneSymbol.Children != nil {
			symbol.Children = transferSymbolVec(strFile, level+1, oneSymbol.Children)
		}
		if oneSymbol.Kind == common.IKAnnotateAlias {
			symbol.Kind = lsp.Interface
			symbol.Detail = "annotate alias"
		} else if oneSymbol.Kind == common.IKAnnotateClass {
			symbol.Kind = lsp.Interface
			symbol.Detail = "annotate class"
		} else if oneSymbol.Kind == common.IKFunction {
			symbol.Kind = lsp.Function
			symbol.Detail = "function"
			lastFuncSymbol = symbol
		} else if oneSymbol.Kind == common.IKAnnotateMark {
			symbol.Kind = lsp.Field
			symbol.Detail = oneSymbol.Name
		} else if len(oneSymbol.Children) != 0 {
			symbol.Kind = lsp.Class
			symbol.Detail = "table"
		} else {
			symbol.Detail = "variable"
		}

		if symbol.Detail == "variable" &&
			oneSymbol.ContainerName == "local" &&
			lastFuncSymbol.Detail == "function" &&
			symbol.Range.Start.Line > lastFuncSymbol.Range.Start.Line &&
			symbol.Range.End.Line < lastFuncSymbol.Range.End.Line {
			// igonre local variable inside function
		} else {
			items = append(items, symbol)
		}
	}
	return
}

// for sort lsp.DocumentSymbol
type lspSymbolSlice []lsp.DocumentSymbol

func (a lspSymbolSlice) Len() int {
	return len(a)
}

func (a lspSymbolSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a lspSymbolSlice) Less(i, j int) bool {
	return a[i].Range.Start.Line < a[j].Range.Start.Line
}

// check is symbol in function
func isSymbolRangeIn(items []lsp.DocumentSymbol, index *int, startLine uint32) bool {
	for i := *index; i < len(items); i++ {
		it := items[i]
		if startLine >= it.Range.Start.Line && startLine <= it.Range.End.Line {
			*index = i
			return true
		}
	}
	return false
}
