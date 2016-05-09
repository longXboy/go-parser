package main

import (
	"fmt"
	"strconv"
	"strings"
)

type SymbolType int

const (
	OneNode SymbolType = 0
	Number             = 1
	Var                = 2
	VarInt             = 3
	VarFunc            = 4
)

type Symbol struct {
	name  string
	value int64
	stype SymbolType
}

type AST struct {
	symbol Symbol
	op     byte
	left   *AST
	right  *AST
}

type ASTNode struct {
	node *AST
	next *ASTNode
}

type Variable struct {
	name        string
	vtype       SymbolType
	value       int64
	FuncStr     string
	FuncVarName string
	FuncVarType SymbolType
}

type Env struct {
	variable Variable
	next     *Env
}

func main() {
	program := "var s int =  8 + ( 8 + 2 ) \n"
	program += "var double func =  ( var a int )  \n {  \n  return a * 2 \n } \n"
	program += "var f  func = ( var a int ) \n {  \n var x int = double ( a ) \n  return x + a \n } \n"
	program += "f ( s ) \n"
	var env *Env
	parseCalculate(program, env)
}

var times = 0

func parseCalculate(program string, env *Env) int64 {
	times++
	var start int = 0
	var shouldreturn bool
	for {
		end := strings.Index(program[start:], "\n")
		if end < 0 {
			break
		}
		line := program[start : start+end]
		start = start + end + 1
		if line == "" || line == "\n" || line == " " {
			continue
		}
		root := new(AST)
		tokens := strings.Split(line, " ")
		idx := 0
		var parse func(root *AST)
		parse = func(root *AST) {
			for {
				if idx == len(tokens) {
					return
				}
				var token string = tokens[idx]
				idx++
				if token == "" || token == " " {
					continue
				} else if token == "int" {
					if root.op != 'v' {
						fmt.Println("error! int must follow var")
						return
					}
					root.left.symbol.stype = VarInt
				} else if token == "func" {
					if root.op != 'v' {
						fmt.Println("error! func must follow var")
						return
					}
					root.left.symbol.stype = VarFunc
				} else if token == "=" {
					if root.op != 'v' {
						root.op = token[0]
					}
					root.right = new(AST)
					root = root.right
				} else if token == "var" {
					root.op = token[0]
				} else if token == "return" {
					root.op = token[0]
					root.left = new(AST)
					parse(root.left)
				} else if (token[0] >= 'a' && token[0] <= 'z') || (token[0] >= 'A' && token[0] <= 'Z') {
					root.left = &AST{
						symbol: Symbol{
							name:  token,
							stype: Var,
						},
					}
				} else if token == "(" {
					if root.left == nil {
						root.left = new(AST)
						parse(root.left)
					} else if root.left.symbol.name != "" {
						root.left.op = 'c'
						root.left.left = new(AST)
						parse(root.left.left)
					}
				} else if token == "+" || token == "-" || token == "*" || token == "/" {
					root.op = token[0]
					root.right = new(AST)
					root = root.right
				} else if token[0] >= '0' && token[0] <= '9' {
					number, err := strconv.ParseInt(token, 10, 64)
					if err != nil {
						fmt.Println("token(%s) is not number", token)
					}
					root.left = &AST{
						symbol: Symbol{
							value: number,
							stype: Number,
						},
					}
				} else if token == ")" {
					return
				}
			}
		}
		parse(root)

		var calculate func(root *AST) int64
		calculate = func(root *AST) int64 {
			if root == nil {
				return 0
			} else if root.symbol.stype == OneNode {
				if root.op == 0 {
					return calculate(root.left)
				}
				if root.op == 'v' {
					if env == nil {
						env = new(Env)
					} else {
						newenv := new(Env)
						newenv.next = env
						env = newenv
					}
					if root.left.symbol.stype == VarInt {
						env.variable.name = root.left.symbol.name
						env.variable.vtype = VarInt
						env.variable.value = calculate(root.right)
					} else if root.left.symbol.stype == VarFunc {
						env.variable.name = root.left.symbol.name
						env.variable.vtype = VarFunc
						if root.right.left.op == 'v' {
							env.variable.FuncVarName = root.right.left.left.symbol.name
							env.variable.FuncVarType = root.right.left.left.symbol.stype
						}
						before := strings.Index(program[start:], "{")
						after := strings.Index(program[start:], "}")
						env.variable.FuncStr = program[start+before+1 : start+after]
						start = start + after + 1
						return 0
					}
					return env.variable.value
				} else if root.op == '+' {
					return (calculate(root.left) + calculate(root.right))
				} else if root.op == '-' {
					return (calculate(root.left) - calculate(root.right))
				} else if root.op == '*' {
					return (calculate(root.left) * calculate(root.right))
				} else if root.op == '/' {
					return (calculate(root.left) / calculate(root.right))
				} else if root.op == 'r' {
					shouldreturn = true
					return calculate(root.left)
				}
			} else if root.symbol.stype == Var {
				if root.op == 'c' {
					variable := lookupEnv(root.symbol.name, env)
					if variable.name != root.symbol.name {
						fmt.Println("can not find %s ,undefined!", root.symbol.name)
						return 0
					}
					if variable.FuncVarType == VarInt {
						if env == nil {
							env = new(Env)
						} else {
							newenv := new(Env)
							newenv.next = env
							env = newenv
						}
						env.variable.value = calculate(root.left)
						env.variable.name = variable.FuncVarName
						env.variable.vtype = variable.FuncVarType
					}
					result := parseCalculate(variable.FuncStr, env)
					if variable.FuncVarType == VarInt {
						env = env.next
					}
					return result
				}
				varres := lookupEnv(root.symbol.name, env).value
				return varres
			}
			return root.symbol.value
		}
		rv := calculate(root)
		if shouldreturn {
			fmt.Println(rv)
			return rv
		}
	}
	return 0
}

func lookupEnv(name string, env *Env) Variable {
	for {
		if env == nil {
			return Variable{}
		}
		if env.variable.name == name {
			return env.variable
		}
		if env.next == nil {
			return Variable{}
		}
		env = env.next
	}
}

func printNode(ptNode *AST) {
	if ptNode.left != nil {
		printNode(ptNode.left)
	}
	if ptNode.symbol.stype == Number {
		fmt.Println(ptNode.symbol.value)
		return
	} else if ptNode.symbol.stype == VarInt || ptNode.symbol.stype == VarFunc {
		fmt.Println(ptNode.symbol.name)
		return
	} else {
		if ptNode.right != nil {
			fmt.Println(string(ptNode.op))
		}
	}
	if ptNode.right != nil {
		printNode(ptNode.right)
	}

}
