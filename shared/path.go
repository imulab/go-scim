package shared

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// interface to represent a single segment in a Path
// i.e. 'emails' in 'emails.value'
// i.e. 'groups[type Eq "direct"]' in 'groups[type Eq "direct"].value'
type Path interface {
	Next() Path                        // next Path, nil means this is the last one
	Value() string                     // text value, unprocessed
	Base() string                      // base Path value, i.e. 'groups' in 'groups[type Eq "direct"]'
	FilterRoot() FilterNode            // root of the filter tree, i.e. 'Eq' in 'type Eq "direct"'
	SeparateAtLast() (Path, Path)      // break up the path chain at the last node
	CollectValue() string              // all path value downstream, separated by period.
	CorrectCase(AttributeSource, bool) // correct the case to defined values, and whether process downstream as well
}

// interface to represent a node in the filter tree
type FilterNode interface {
	Data() interface{}
	Type() FilterNodeType
	Left() FilterNode
	Right() FilterNode
	CorrectCase(guide AttributeSource)
}

// type of the filter node
type FilterNodeType int

const (
	PathOperand = FilterNodeType(iota + 1)
	ConstantOperand
	LogicalOperator
	RelationalOperator
	Parenthesis
)

// create a new Path from text
func NewPath(text string) (Path, error) {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return nil, Error.InvalidPath(text, "empty path")
	}

	var (
		this, next string
		thisPath   *path
	)

	idx := -1
	textMode := false
	for i, r := range []rune(text) {
		switch r {
		case quoteRune:
			textMode = !textMode
		case periodRune:
			if !textMode && (strings.ToLower(text[:i]) != "urn:ietf:params:scim:schemas:core:2" && strings.ToLower(text[:i]) != "urn:ietf:params:scim:schemas:extension:enterprise:2") {
				idx = i
				break
			}
		}
	}

	if idx == -1 {
		this = text
	} else {
		this = text[:idx]
		next = text[idx+1:]
	}

	this = strings.TrimSpace(this)
	if len(this) == 0 {
		return nil, Error.InvalidPath(text, "empty component")
	} else {
		lbIdx := strings.Index(this, "[")
		rbIdx := strings.Index(this, "]")

		switch {
		case lbIdx == -1 && rbIdx == -1:
			thisPath = &path{text: this, base: this, next: nil, filterRoot: nil}

		case lbIdx > 0 && rbIdx > lbIdx+1 && rbIdx == len(this)-1:
			thisBase := this[:lbIdx]
			thisFilter, err := NewFilter(this[lbIdx+1 : rbIdx])
			if err != nil {
				return nil, err
			}
			thisPath = &path{text: this, base: thisBase, next: nil, filterRoot: thisFilter.(*filterNode)}

		default:
			return nil, Error.InvalidPath(text, "invalid placement of filter brackets")
		}
	}

	next = strings.TrimSpace(next)
	if len(next) > 0 {
		if nextPath, err := NewPath(next); err != nil {
			return nil, err
		} else {
			thisPath.next = nextPath.(*path)
		}
	}

	return thisPath, nil
}

// create a new filter from text
func NewFilter(text string) (FilterNode, error) {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return nil, Error.InvalidFilter(text, "empty filter")
	}

	tokenizer := &filterTokenizer{
		textMode:  false,
		remaining: []rune(text),
		buffer:    make([]rune, 0),
		tokens:    make([]*filterNode, 0),
	}
	if err := tokenizer.tokenize(); err != nil {
		return nil, Error.InvalidFilter(text, err.Error())
	}

	sy := &shuntingYard{
		input:    NewQueueWithoutLimit(),
		operator: NewStackWithoutLimit(),
		output:   NewStackWithoutLimit(),
	}
	root, err := sy.run(tokenizer.tokens)
	if err != nil {
		return nil, Error.InvalidFilter(text, err.Error())
	}

	return root, nil
}

// filter tokenizer
const (
	spaceRune        = ' '
	quoteRune        = '"'
	commaRune        = ','
	periodRune       = '.'
	leftBracketRune  = '['
	rightBracketRune = ']'
	leftParenRune    = '('
	rightParenRune   = ')'
)

type filterTokenizer struct {
	textMode   bool          // treat everything as text
	remaining  []rune        // the remaining runes to become token
	buffer     []rune        // buffer for the runes to be converted to the next token
	tokens     []*filterNode // tokens
	parenLevel int           // matching count for parenthesis
}

func (t *filterTokenizer) tokenize() error {
	for len(t.remaining) > 0 {
		r := t.getAndDropTopRune()
		switch r {
		case spaceRune:
			if t.textMode {
				t.addToBuffer(r)
			} else {
				t.addBufferToTokens()
			}

		case quoteRune:
			t.addToBuffer(r)
			t.textMode = !t.textMode

		case leftBracketRune:
			return errors.New("left bracket not allowed here")

		case rightBracketRune:
			return errors.New("right bracket not allowed here")

		case leftParenRune:
			t.addToTokens(r)
			t.parenLevel++

		case rightParenRune:
			t.addToTokens(r)
			t.parenLevel--

		case commaRune:
			t.addToTokens(r)

		default:
			t.addToBuffer(r)
		}
	}
	t.addBufferToTokens()

	switch {
	case t.parenLevel > 0:
		return errors.New("mismatched parenthesis")
	default:
		return nil
	}
}

func (t *filterTokenizer) getAndDropTopRune() rune {
	r := t.remaining[0]
	t.remaining = t.remaining[1:]
	return r
}

func (t *filterTokenizer) addToBuffer(r rune) {
	t.buffer = append(t.buffer, r)
}

func (t *filterTokenizer) addToTokens(r rune) error {
	if len(t.buffer) > 0 {
		t.addBufferToTokens()
	}

	t.tokens = append(t.tokens, tokenCentral.create(fmt.Sprintf("%c", r)))
	return nil
}

func (t *filterTokenizer) addBufferToTokens() error {
	if len(t.buffer) > 0 {
		t.tokens = append(t.tokens, tokenCentral.create(string(t.buffer)))
		t.buffer = make([]rune, 0)
		return nil
	} else {
		return errors.New("unexpected filter content")
	}
}

// token factory
var (
	oneTokenFactory     sync.Once
	tokenCentral        *tokenFactory
	tokenMetadataLookup tokenMetadataMap
)

const (
	And        = "and"
	Or         = "or"
	Not        = "not"
	Eq         = "eq"
	Ne         = "ne"
	Sw         = "sw"
	Ew         = "ew"
	Co         = "co"
	Pr         = "pr"
	Gt         = "gt"
	Ge         = "ge"
	Lt         = "lt"
	Le         = "le"
	leftParen  = "("
	rightParen = ")"
)

type tokenMetadataMap map[interface{}]tokenMetadata

func (m tokenMetadataMap) get(key interface{}) tokenMetadata {
	if v, ok := m[key]; !ok {
		panic(fmt.Errorf("no metadata configured for %v", key))
	} else {
		return v
	}
}

// create a filterNode out of the face value, note that anything that cannot be resolved to
// logical, relational or constant token will be treated as path, which could delay throw
// from being discovered
type tokenFactory struct{}

func (f tokenFactory) create(text string) *filterNode {
	text = strings.TrimSpace(text)
	switch strings.ToLower(text) {
	case And:
		return &filterNode{data: And, typ: LogicalOperator}
	case Or:
		return &filterNode{data: Or, typ: LogicalOperator}
	case Not:
		return &filterNode{data: Not, typ: LogicalOperator}
	case Eq:
		return &filterNode{data: Eq, typ: RelationalOperator}
	case Ne:
		return &filterNode{data: Ne, typ: RelationalOperator}
	case Sw:
		return &filterNode{data: Sw, typ: RelationalOperator}
	case Ew:
		return &filterNode{data: Ew, typ: RelationalOperator}
	case Co:
		return &filterNode{data: Co, typ: RelationalOperator}
	case Pr:
		return &filterNode{data: Pr, typ: RelationalOperator}
	case Gt:
		return &filterNode{data: Gt, typ: RelationalOperator}
	case Ge:
		return &filterNode{data: Ge, typ: RelationalOperator}
	case Lt:
		return &filterNode{data: Lt, typ: RelationalOperator}
	case Le:
		return &filterNode{data: Le, typ: RelationalOperator}
	case leftParen:
		return &filterNode{data: leftParen, typ: Parenthesis}
	case rightParen:
		return &filterNode{data: rightParen, typ: Parenthesis}
	default:
		if strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") {
			return &filterNode{data: text[1 : len(text)-1], typ: ConstantOperand}
		} else if b, err := strconv.ParseBool(text); err == nil {
			return &filterNode{data: b, typ: ConstantOperand}
		} else if i, err := strconv.ParseInt(text, 10, 64); err == nil {
			return &filterNode{data: i, typ: ConstantOperand}
		} else if f, err := strconv.ParseFloat(text, 64); err == nil {
			return &filterNode{data: f, typ: ConstantOperand}
		} else {
			if path, err := NewPath(text); err != nil {
				return nil
			} else {
				return &filterNode{data: path, typ: PathOperand}
			}
		}
	}
}

// token meta data
type tokenAssociativity int
type tokenPrecedence int

const (
	leftAssociative  = tokenAssociativity(1)
	rightAssociative = tokenAssociativity(2)
	highPrecedence   = tokenPrecedence(1000)
	normalPrecedence = tokenPrecedence(100)
	lowPrecedence    = tokenPrecedence(10)
)

type tokenMetadata struct {
	associativity tokenAssociativity
	precedence    tokenPrecedence
	numOfArgs     int
}

// shunting yard algorithm
type shuntingYard struct {
	input    Queue
	operator Stack
	output   Stack
}

func (sy *shuntingYard) run(tokens []*filterNode) (*filterNode, error) {
	for _, tok := range tokens {
		sy.input.Offer(tok)
	}

	for sy.input.Size() > 0 {
		tok := sy.input.Poll().(*filterNode)

		switch tok.Type() {
		case PathOperand, ConstantOperand:
			if err := sy.pushToOutput(tok); err != nil {
				return nil, err
			}

		case RelationalOperator, LogicalOperator:
			for {
				if peek, ok := sy.operator.Peek().(*filterNode); !ok || peek == nil {
					break
				} else if peek.Type() != RelationalOperator && peek.Type() != LogicalOperator {
					break
				} else {
					tokMetadata := tokenMetadataLookup.get(tok.Data())
					peekMetadata := tokenMetadataLookup.get(peek.Data())

					if tokMetadata.associativity == leftAssociative &&
						tokMetadata.precedence <= peekMetadata.precedence {
						if err := sy.pushToOutput(sy.operator.Pop().(*filterNode)); err != nil {
							return nil, err
						}
					} else if tokMetadata.associativity == rightAssociative &&
						tokMetadata.precedence < peekMetadata.precedence {
						if err := sy.pushToOutput(sy.operator.Pop().(*filterNode)); err != nil {
							return nil, err
						}
					} else {
						break
					}
				}
			}
			sy.operator.Push(tok)

		case Parenthesis:
			switch tok.Data().(string) {
			case leftParen:
				sy.operator.Push(tok)
			case rightParen:
				for {
					if peek, ok := sy.operator.Peek().(*filterNode); !ok || peek == nil {
						return nil, errors.New("parenthesis mismatch")
					} else if peek.Type() == Parenthesis && peek.Data().(string) == leftParen {
						sy.operator.Pop()
						break
					} else {
						if err := sy.pushToOutput(sy.operator.Pop().(*filterNode)); err != nil {
							return nil, err
						}
					}
				}
			}

		default:
			return nil, fmt.Errorf("cannot handle token %v, invalid type", tok.Data())
		}
	}

	for sy.operator.Size() > 0 {
		if peek := sy.operator.Peek(); peek != nil && peek.(*filterNode).Type() == Parenthesis {
			return nil, errors.New("parenthesis mismatch")
		} else {
			if err := sy.pushToOutput(sy.operator.Pop().(*filterNode)); err != nil {
				return nil, err
			}
		}
	}

	return sy.output.Pop().(*filterNode), nil
}

func (sy *shuntingYard) pushToOutput(tok *filterNode) error {
	switch tok.Type() {
	case ConstantOperand, PathOperand:
	default:
		metadata := tokenMetadataLookup.get(tok.Data())
		switch metadata.numOfArgs {
		case 1:
			if arg := sy.output.Pop(); arg == nil {
				return fmt.Errorf("cannot handle token %v, insufficient number of arguments", tok.Data())
			} else {
				tok.left = arg.(*filterNode)
			}

		case 2:
			arg2 := sy.output.Pop()
			arg1 := sy.output.Pop()
			if arg1 == nil || arg2 == nil {
				return fmt.Errorf("cannot handle token %v, insufficient number of arguments", tok.Data())
			} else {
				tok.left = arg1.(*filterNode)
				tok.right = arg2.(*filterNode)
			}

		default:
			return fmt.Errorf("cannot handle token %v, invalid number of arguments", tok.Data())
		}
	}
	sy.output.Push(tok)
	return nil
}

// implementation of Path
type path struct {
	next       Path
	text       string
	base       string
	filterRoot FilterNode
}

func (p *path) Next() Path             { return p.next }
func (p *path) Value() string          { return p.text }
func (p *path) Base() string           { return p.base }
func (p *path) FilterRoot() FilterNode { return p.filterRoot }

func (p *path) SeparateAtLast() (Path, Path) {
	if p.Next() == nil {
		return nil, p
	}

	var c Path = p
	for c.Next().Next() != nil {
		c = c.Next()
	}

	var last = c.Next()
	c.(*path).next = nil
	return p, last
}

func (p *path) CollectValue() string {
	v := make([]string, 0)
	var c Path = p
	for c != nil {
		v = append(v, c.Value())
		c = c.Next()
	}
	return strings.Join(v, ".")
}

func (p *path) CorrectCase(guide AttributeSource, recursive bool) {
	attr := guide.GetAttribute(p, false)

	switch strings.ToLower(p.base) {
	case strings.ToLower(attr.Name):
		p.base = attr.Name
	case strings.ToLower(attr.Assist.FullPath):
		p.base = attr.Assist.FullPath
	}

	if p.filterRoot != nil {
		p.filterRoot.CorrectCase(attr)
	}
	if recursive && p.next != nil {
		p.next.CorrectCase(attr, recursive)
	}
}

// implementation of FilterNode
type filterNode struct {
	data  interface{}
	typ   FilterNodeType
	left  *filterNode
	right *filterNode
}

func (n *filterNode) Data() interface{}    { return n.data }
func (n *filterNode) Type() FilterNodeType { return n.typ }
func (n *filterNode) Left() FilterNode     { return n.left }
func (n *filterNode) Right() FilterNode    { return n.right }
func (n *filterNode) CorrectCase(guide AttributeSource) {
	if n.left != nil {
		n.left.CorrectCase(guide)
	}
	if n.typ == PathOperand {
		n.data.(Path).CorrectCase(guide, true)
	}
	if n.right != nil {
		n.right.CorrectCase(guide)
	}
}

func init() {
	oneTokenFactory.Do(func() {
		tokenCentral = &tokenFactory{}
		tokenMetadataLookup = tokenMetadataMap(map[interface{}]tokenMetadata{
			And: {leftAssociative, normalPrecedence, 2},
			Or:  {leftAssociative, normalPrecedence - 1, 2},
			Not: {rightAssociative, normalPrecedence + 1, 1},
			Eq:  {leftAssociative, highPrecedence, 2},
			Ne:  {leftAssociative, highPrecedence, 2},
			Sw:  {leftAssociative, highPrecedence, 2},
			Ew:  {leftAssociative, highPrecedence, 2},
			Co:  {leftAssociative, highPrecedence, 2},
			Pr:  {leftAssociative, highPrecedence, 1},
			Gt:  {leftAssociative, highPrecedence, 2},
			Ge:  {leftAssociative, highPrecedence, 2},
			Lt:  {leftAssociative, highPrecedence, 2},
			Le:  {leftAssociative, highPrecedence, 2},
		})
	})
}
