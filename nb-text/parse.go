package main

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Env struct {
	Vars     map[string]string
	VarTypes map[string]string
	D0       Operand
}

type ExprParser struct {
	ts  *TokStream
	D0  Operand
	Env *Env
}

func NewExprParser(f io.Reader) *ExprParser {
	ep := &ExprParser{
		ts: tokenize(f),
	}
	return ep
}

func (ep *ExprParser) Reset() {
	ep.ts.iPeekTok = 0
}

func (ep *ExprParser) Run(env *Env) (Operand, error) {
	ep.Env = env
	err := ep.compoundCondition()
	if err != nil {
		return Operand{}, err
	}
	return ep.Env.D0, nil
}

// Sample:
// (date >= "2018-08-01" and date < "2018-09-01") or (cat = 'grocery' or
// cat = 'household') and (amt >= 100.0 or (cat = "dineout" and amt > 75.0))
//
// body =~ "todo"
// cat <> ""

// Token list:
// >, >=, <, <=, =, =~, <>
// string literal, num literal
//

func parseErr(s string) error {
	return fmt.Errorf("%s", s)
}

func expectedErr(s string) error {
	return fmt.Errorf("%s expected", s)
}

func toNum(s string) float64 {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return n
}
func toStr(n float64) string {
	return fmt.Sprintf("%f", n)
}

func toBool(s string) bool {
	if s == "0" {
		return false
	}
	return true
}

func (ep *ExprParser) match(tokName string) (Operand, error) {
	tok := ep.ts.PeekTok()
	if tok == nil || tok.Typ != tokName {
		return Operand{}, expectedErr(tokName)
	}
	ep.ts.NextTok() // advance read pos

	typ := "STR"
	if tok.Typ == "NUM" {
		typ = "NUM"
	}
	return Operand{typ, tok.Lit}, nil
}

// D0 <- Vars[ident]
func (ep *ExprParser) field() error {
	opr, err := ep.match("IDENT")
	if err != nil {
		return err
	}
	varName := opr.Val
	varType := ep.Env.VarTypes[varName]

	typ := "STR"
	if varType == "f" || varType == "d" {
		typ = "NUM"
	}
	ep.Env.D0 = Operand{typ, ep.Env.Vars[varName]}

	return nil
}

// D0 <- num | str | field
// D0 <- (expr1)
func (ep *ExprParser) atom() error {
	tok := ep.ts.PeekTok()
	if tok == nil {
		return expectedErr("field, number or string")
	}

	var err error
	if tok.Typ == "IDENT" {
		err = ep.field()
	} else if tok.Typ == "NUM" {
		ep.Env.D0, err = ep.match("NUM")
	} else if tok.Typ == "STR" {
		ep.Env.D0, err = ep.match("STR")
	} else if tok.Typ == "LPAREN" {
		_, err = ep.match("LPAREN")
		ep.expr()
		_, err = ep.match("RPAREN")
	} else {
		err = expectedErr("field, number or string")
	}

	if err != nil {
		return err
	}
	return nil
}

// D0 <- atom
// D0 <- atom * atom ...
// D0 <- atom / atom ...
func (ep *ExprParser) exprTerm() error {
	err := ep.atom()
	if err != nil {
		return err
	}

	tok := ep.ts.PeekTok()
	for tok != nil {
		if tok.Typ == "MULT" {
			ep.ts.NextTok()
			err = ep.mult()
			if err != nil {
				return err
			}
		} else if tok.Typ == "DIV" {
			ep.ts.NextTok()
			err = ep.div()
			if err != nil {
				return err
			}
		} else {
			break
		}
		tok = ep.ts.PeekTok()
	}
	return nil
}

// D0 <- D0 * atom
func (ep *ExprParser) mult() error {
	leftD0 := ep.Env.D0
	err := ep.atom()
	if err != nil {
		return err
	}

	if ep.Env.D0.Typ != "NUM" {
		return parseErr("can't multiply by non-number")
	}

	if leftD0.Typ == "STR" {
		ep.Env.D0.Val = strings.Repeat(leftD0.Val, int(toNum(ep.Env.D0.Val)))
		return nil
	}
	ep.Env.D0.Typ = "NUM"
	ep.Env.D0.Val = toStr(toNum(leftD0.Val) * toNum(ep.Env.D0.Val))
	return nil
}

// D0 <- D0 / atom
func (ep *ExprParser) div() error {
	leftD0 := ep.Env.D0
	err := ep.atom()
	if err != nil {
		return err
	}

	if ep.Env.D0.Typ != "NUM" {
		return parseErr("can't divide by non-number")
	}
	if leftD0.Typ == "STR" {
		return parseErr("can't divide string")
	}
	if toNum(ep.Env.D0.Val) == 0.0 {
		return parseErr("can't divide by zero")
	}
	ep.Env.D0.Typ = "NUM"
	ep.Env.D0.Val = toStr(toNum(leftD0.Val) / toNum(ep.Env.D0.Val))
	return nil
}
func (ep *ExprParser) add() error {
	leftD0 := ep.Env.D0
	err := ep.exprTerm()
	if err != nil {
		return err
	}

	if leftD0.Typ == "STR" || ep.Env.D0.Typ == "STR" {
		ep.Env.D0.Val = leftD0.Val + ep.Env.D0.Val
		return nil
	}
	ep.Env.D0.Val = toStr(toNum(leftD0.Val) + toNum(ep.Env.D0.Val))
	return nil
}
func (ep *ExprParser) minus() error {
	leftD0 := ep.Env.D0
	err := ep.exprTerm()
	if err != nil {
		return err
	}

	if leftD0.Typ == "STR" || ep.Env.D0.Typ == "STR" {
		return expectedErr("number")
	}
	ep.Env.D0.Val = toStr(toNum(leftD0.Val) - toNum(ep.Env.D0.Val))
	return nil
}

// D0 <- exprTerm
// D0 <- exprTerm + exprTerm ...
// D0 <- exprTerm - exprTerm ...
func (ep *ExprParser) expr() error {
	// Check if unary + or -
	unaryMinus := false
	tok := ep.ts.PeekTok()
	if tok.Typ == "PLUS" {
		ep.ts.NextTok()
	} else if tok.Typ == "MINUS" {
		ep.ts.NextTok()
		unaryMinus = true
	}

	err := ep.exprTerm()
	if err != nil {
		return err
	}

	// If unary minus, negate the D0 value
	if unaryMinus {
		if ep.Env.D0.Typ != "NUM" {
			return expectedErr("number after unary minus (-)")
		}
		ep.Env.D0.Val = toStr(0.0 - toNum(ep.Env.D0.Val))
	}

	tok = ep.ts.PeekTok()
	for tok != nil {
		if tok.Typ == "PLUS" {
			ep.ts.NextTok()
			err = ep.add()
			if err != nil {
				return err
			}
		} else if tok.Typ == "MINUS" {
			ep.ts.NextTok()
			err = ep.minus()
			if err != nil {
				return err
			}
		} else {
			break
		}
		tok = ep.ts.PeekTok()
	}

	return nil
}

// D0 <- expr <cmpOp> expr
// Ex.
// title =~ "text within title"
// amt >= 100.0
// debit >= credit - amt + 100.0
// amt / 5 <= credit * 2
func (ep *ExprParser) comparison() error {
	err := ep.expr()
	if err != nil {
		return err
	}
	leftD0 := ep.Env.D0

	tok := ep.ts.PeekTok()
	if tok == nil {
		return parseErr("expression needs a condition")
	}
	if !inSlc([]string{"GT", "GTE", "LT", "LTE", "EQ", "REG_EQ", "NE"}, tok.Typ) {
		return parseErr("expression needs a condition")
	}
	cmpOp := tok.Typ
	ep.ts.NextTok()

	err = ep.expr()
	if err != nil {
		return err
	}
	rightD0 := ep.Env.D0

	ep.Env.D0.Typ = "NUM"
	ep.Env.D0.Val = "0"
	fCmp, err := doCmp(leftD0, cmpOp, rightD0)
	if err != nil {
		return err
	}
	if fCmp {
		ep.Env.D0.Val = "1"
	}
	return nil
}
func doCmp(l Operand, op string, r Operand) (bool, error) {
	if l.Typ != r.Typ {
		return false, parseErr("operand mismatch")
	}

	if l.Typ == "STR" {
		// string comparison
		switch op {
		case "GT":
			return l.Val > r.Val, nil
		case "GTE":
			return l.Val >= r.Val, nil
		case "LT":
			return l.Val < r.Val, nil
		case "LTE":
			return l.Val <= r.Val, nil
		case "EQ":
			return l.Val == r.Val, nil
		case "NE":
			return l.Val != r.Val, nil
		case "REG_EQ":
			matched, _ := regexp.MatchString("(?i)"+r.Val, l.Val)
			return matched, nil
		default:
			return false, parseErr("unknown comparison operator")
		}
	} else {
		// num comparison
		lNum := toNum(l.Val)
		rNum := toNum(r.Val)

		switch op {
		case "GT":
			return lNum > rNum, nil
		case "GTE":
			return lNum >= rNum, nil
		case "LT":
			return lNum < rNum, nil
		case "LTE":
			return lNum <= rNum, nil
		case "EQ":
			return lNum == rNum, nil
		case "NE":
			return lNum != rNum, nil
		case "REG_EQ":
			matched, _ := regexp.MatchString("(?i)"+r.Val, l.Val)
			return matched, nil
		default:
			return false, parseErr("unknown comparison operator")
		}
	}

	return false, parseErr("unknown error")
}

// D0 <- comparison [and|or comparison]...
func (ep *ExprParser) condition() error {
	var err error

	fLParen := false
	tok := ep.ts.PeekTok()
	if tok.Typ == "LPAREN" {
		_, err = ep.match("LPAREN")
		if err != nil {
			return err
		}
		fLParen = true

		err = ep.compoundCondition()
		if err != nil {
			return err
		}

		if fLParen {
			_, err = ep.match("RPAREN")
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = ep.comparison()
	if err != nil {
		return err
	}

	tok = ep.ts.PeekTok()
	for tok != nil {
		if tok.Typ == "AND" {
			ep.ts.NextTok()
			err = ep.conditionAnd()
			if err != nil {
				return err
			}
		} else if tok.Typ == "OR" {
			ep.ts.NextTok()
			err = ep.conditionOr()
			if err != nil {
				return err
			}
		} else {
			break
		}
		tok = ep.ts.PeekTok()
	}

	return nil
}
func (ep *ExprParser) conditionAnd() error {
	leftD0 := ep.Env.D0
	err := ep.comparison()
	if err != nil {
		return err
	}

	if toBool(leftD0.Val) && toBool(ep.Env.D0.Val) {
		ep.Env.D0.Val = "1"
	} else {
		ep.Env.D0.Val = "0"
	}
	return nil
}
func (ep *ExprParser) conditionOr() error {
	leftD0 := ep.Env.D0
	err := ep.comparison()
	if err != nil {
		return err
	}

	if toBool(leftD0.Val) || toBool(ep.Env.D0.Val) {
		ep.Env.D0.Val = "1"
	} else {
		ep.Env.D0.Val = "0"
	}
	return nil
}

func (ep *ExprParser) compoundCondition() error {
	err := ep.condition()
	if err != nil {
		return err
	}

	tok := ep.ts.PeekTok()
	for tok != nil {
		if tok.Typ == "AND" {
			ep.ts.NextTok()
			err = ep.compoundConditionAnd()
			if err != nil {
				return err
			}
		} else if tok.Typ == "OR" {
			ep.ts.NextTok()
			err = ep.compoundConditionOr()
			if err != nil {
				return err
			}
		} else {
			break
		}
		tok = ep.ts.PeekTok()
	}
	return nil
}
func (ep *ExprParser) compoundConditionAnd() error {
	leftD0 := ep.Env.D0
	err := ep.condition()
	if err != nil {
		return err
	}

	if toBool(leftD0.Val) && toBool(ep.Env.D0.Val) {
		ep.Env.D0.Val = "1"
	} else {
		ep.Env.D0.Val = "0"
	}
	return nil
}
func (ep *ExprParser) compoundConditionOr() error {
	leftD0 := ep.Env.D0
	err := ep.condition()
	if err != nil {
		return err
	}

	if toBool(leftD0.Val) || toBool(ep.Env.D0.Val) {
		ep.Env.D0.Val = "1"
	} else {
		ep.Env.D0.Val = "0"
	}
	return nil
}
