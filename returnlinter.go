package returnlinter

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "returnlinter",
	Doc:      "checks that w.WriteHeader() calls are followed by return statements in http.Handler middleware",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Filter for function declarations
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl := n.(*ast.FuncDecl)

		// Check if this function matches the middleware pattern:
		// func <name>(handler http.Handler) http.Handler
		if !isMiddlewarePattern(funcDecl) {
			return
		}

		// Find the inner HandlerFunc
		ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
			// Look for the pattern: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ... })
			if callExpr, ok := node.(*ast.CallExpr); ok {
				if isHandlerFuncCall(callExpr) {
					// Get the function literal inside HandlerFunc
					if len(callExpr.Args) > 0 {
						if funcLit, ok := callExpr.Args[0].(*ast.FuncLit); ok {
							checkHandlerBody(pass, funcLit.Body)
						}
					}
				}
			}
			return true
		})
	})

	return nil, nil
}

// isMiddlewarePattern checks if the function signature matches:
// func <name>(handler http.Handler) http.Handler
func isMiddlewarePattern(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) != 1 {
		return false
	}

	// Check return type is http.Handler
	if !isHTTPHandler(funcDecl.Type.Results.List[0].Type) {
		return false
	}

	return true
}

// isHTTPHandler checks if the type is http.Handler
func isHTTPHandler(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "http" && selector.Sel.Name == "Handler"
}

// isHandlerFuncCall checks if the call is http.HandlerFunc(...)
func isHandlerFuncCall(callExpr *ast.CallExpr) bool {
	selector, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "http" && selector.Sel.Name == "HandlerFunc"
}

// checkHandlerBody inspects the handler function body for WriteHeader calls
func checkHandlerBody(pass *analysis.Pass, body *ast.BlockStmt) {
	for i, stmt := range body.List {
		// Look for expression statements that might contain w.WriteHeader()
		if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
			if IsWriteHeaderCall(exprStmt.X) {
				// Check if the next non-comment/non-empty statement is a return
				if !IsFollowedByReturn(body.List, i) {
					pass.Reportf(exprStmt.Pos(), "WriteHeader call not immediately followed by return statement")
				}
			}
		}

		// Also check inside if/else blocks, switch statements, etc.
		checkNestedWriteHeader(pass, stmt)
	}
}

// checkNestedWriteHeader recursively checks for WriteHeader calls in nested structures
func checkNestedWriteHeader(pass *analysis.Pass, stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.IfStmt:
		checkBlockForWriteHeader(pass, s.Body)
		if s.Else != nil {
			checkNestedWriteHeader(pass, s.Else)
		}
	case *ast.BlockStmt:
		checkBlockForWriteHeader(pass, s)
	case *ast.ForStmt:
		checkBlockForWriteHeader(pass, s.Body)
	case *ast.RangeStmt:
		checkBlockForWriteHeader(pass, s.Body)
	case *ast.SwitchStmt:
		checkBlockForWriteHeader(pass, s.Body)
	case *ast.TypeSwitchStmt:
		checkBlockForWriteHeader(pass, s.Body)
	case *ast.SelectStmt:
		checkBlockForWriteHeader(pass, s.Body)
	case *ast.CaseClause:
		for i, caseStmt := range s.Body {
			if exprStmt, ok := caseStmt.(*ast.ExprStmt); ok {
				if IsWriteHeaderCall(exprStmt.X) {
					if !IsFollowedByReturn(s.Body, i) {
						pass.Reportf(exprStmt.Pos(), "WriteHeader call not immediately followed by return statement")
					}
				}
			}
			checkNestedWriteHeader(pass, caseStmt)
		}
	}
}

// checkBlockForWriteHeader checks a block statement for WriteHeader calls
func checkBlockForWriteHeader(pass *analysis.Pass, block *ast.BlockStmt) {
	if block == nil {
		return
	}
	for i, stmt := range block.List {
		if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
			if IsWriteHeaderCall(exprStmt.X) {
				if !IsFollowedByReturn(block.List, i) {
					pass.Reportf(exprStmt.Pos(), "WriteHeader call not immediately followed by return statement")
				}
			}
		}
		checkNestedWriteHeader(pass, stmt)
	}
}

// IsWriteHeaderCall checks if the expression is w.WriteHeader(...)
func IsWriteHeaderCall(expr ast.Expr) bool {
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	selector, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	return selector.Sel.Name == "WriteHeader"
}

// IsFollowedByReturn checks if the next non-whitespace statement is a return
func IsFollowedByReturn(stmts []ast.Stmt, currentIndex int) bool {
	// Check if there's a next statement
	if currentIndex+1 >= len(stmts) {
		return false
	}

	// The next statement should be a return statement
	_, ok := stmts[currentIndex+1].(*ast.ReturnStmt)
	return ok
}

// isEmptyStmt checks if a statement is effectively empty (just whitespace/comments)
func isEmptyStmt(stmt ast.Stmt) bool {
	_, ok := stmt.(*ast.EmptyStmt)
	return ok
}
