package cnt

import (
	"errors"
	"fmt"
)

type (
	// Tree is the AST
	Tree struct {
		Name string
		Root *ListNode // top-level root of the tree.
	}

	// Parser parses an cnt file
	Parser struct {
		name       string // filename or name of the buffer
		content    string
		l          *lexer
		tok        *item // token saved for lookahead
		openblocks int
	}
)

// NewParser creates a new parser
func NewParser(name, content string) *Parser {
	return &Parser{
		name:    name,
		content: content,
		l:       lex(name, content),
	}
}

// Parse starts the parsing.
func (p *Parser) Parse() (*Tree, error) {
	root, err := p.parseBlock()

	if err != nil {
		return nil, err
	}

	tr := NewTree(p.name)
	tr.Root = root

	return tr, nil
}

// next returns the next item from lookahead buffer if not empty or
// from the lexer
func (p *Parser) next() item {
	if p.tok != nil {
		t := p.tok
		p.tok = nil
		return *t
	}

	return <-p.l.items
}

// backup puts the item into the lookahead buffer
func (p *Parser) backup(it item) error {
	if p.tok != nil {
		return errors.New("only one slot for backup/lookahead")
	}

	p.tok = &it

	return nil
}

// ignores the next item
func (p *Parser) ignore() {
	if p.tok != nil {
		p.tok = nil
	} else {
		<-p.l.items
	}
}

// peek gets but do not discards the next item (lookahead)
func (p *Parser) peek() item {
	i := p.next()
	p.tok = &i
	return i
}

func (p *Parser) parseCommand() (Node, error) {
	it := p.next()

	// paranoia check
	if it.typ != itemCommand {
		return nil, fmt.Errorf("Invalid command: %v", it)
	}

	n := NewCommandNode(it.pos, it.val)

	for {
		it = p.next()

		switch it.typ {
		case itemArg, itemString:
			arg := NewArg(it.pos, it.val)
			n.AddArg(arg)
		case itemEOF:
			return n, nil
		case itemError:
			return nil, fmt.Errorf("Failed to parse document: %s", it)
		default:
			p.backup(it)
			return n, nil
		}
	}

	return nil, errors.New("unreachable")
}

func (p *Parser) parseRfork() (Node, error) {
	it := p.next()

	if it.typ != itemRfork {
		return nil, fmt.Errorf("Invalid command: %v", it)
	}

	n := NewRforkNode(it.pos)

	it = p.next()

	if it.typ != itemRforkFlags {
		return nil, fmt.Errorf("rfork requires one or more of the following flags: %s", rforkFlags)
	}

	n.SetFlags(NewArg(it.pos, it.val))

	it = p.peek()

	if it.typ == itemLeftBlock {
		p.ignore() // ignore lookaheaded symbol
		p.openblocks++

		n.tree = NewTree("rfork block")
		r, err := p.parseBlock()

		if err != nil {
			return nil, err
		}

		n.tree.Root = r
	}

	// TODO: block

	return n, nil
}

func (p *Parser) parseComment() (Node, error) {
	it := p.next()

	if it.typ != itemComment {
		return nil, fmt.Errorf("Invalid comment: %v", it)
	}

	return NewCommentNode(it.pos, it.val), nil
}

func (p *Parser) parseStatement() (Node, error) {
	it := p.peek()

	switch it.typ {
	case itemCommand:
		return p.parseCommand()
	case itemRfork:
		return p.parseRfork()
	case itemComment:
		return p.parseComment()
	}

	return nil, fmt.Errorf("Unexpected token parsing statement '%d'", it.typ)
}

func (p *Parser) parseBlock() (*ListNode, error) {
	ln := NewListNode()

	for {
		it := p.peek()

		switch it.typ {
		case 0:
			return ln, nil
		case itemEOF:
			return ln, nil
		case itemError:
			return nil, errors.New(it.val)
		case itemLeftBlock:
			p.ignore()

			return nil, errors.New("Blocks are only allowed inside rfork")
		case itemRightBlock:
			p.ignore()

			if p.openblocks <= 0 {
				return nil, fmt.Errorf("No block open for close")
			}

			p.openblocks--
			return ln, nil
		default:
			n, err := p.parseStatement()

			if err != nil {
				return nil, err
			}

			ln.Push(n)
		}
	}

	return ln, nil
}

// NewTree creates a new AST tree
func NewTree(name string) *Tree {
	return &Tree{
		Name: name,
	}
}
