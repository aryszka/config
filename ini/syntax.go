/*
This file was generated with treerack (https://github.com/aryszka/treerack).

The contents of this file fall under different licenses.

The code between the "// head" and "// eo head" lines falls under the same
license as the source code of treerack (https://github.com/aryszka/treerack),
unless explicitly stated otherwise, if treerack's license allows changing the
license of this source code.

Treerack's license: MIT https://opensource.org/licenses/MIT
where YEAR=2017, COPYRIGHT HOLDER=Arpad Ryszka (arpad.ryszka@gmail.com)

The rest of the content of this file falls under the same license as the one
that the user of treerack generating this file declares for it, or it is
unlicensed.
*/

package ini

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type charParser struct {
	name   string
	id     int
	not    bool
	chars  []rune
	ranges [][]rune
}
type charBuilder struct {
	name string
	id   int
}

func (p *charParser) nodeName() string {
	return p.name
}
func (p *charParser) nodeID() int {
	return p.id
}
func (p *charParser) commitType() CommitType {
	return Alias
}
func matchChar(chars []rune, ranges [][]rune, not bool, char rune) bool {
	for _, ci := range chars {
		if ci == char {
			return !not
		}
	}
	for _, ri := range ranges {
		if char >= ri[0] && char <= ri[1] {
			return !not
		}
	}
	return not
}
func (p *charParser) match(t rune) bool {
	return matchChar(p.chars, p.ranges, p.not, t)
}
func (p *charParser) parse(c *context) {
	if tok, ok := c.token(); !ok || !p.match(tok) {
		if c.offset > c.failOffset {
			c.failOffset = c.offset
			c.failingParser = nil
		}
		c.fail(c.offset)
		return
	}
	c.success(c.offset + 1)
}
func (b *charBuilder) nodeName() string {
	return b.name
}
func (b *charBuilder) nodeID() int {
	return b.id
}
func (b *charBuilder) build(c *context) ([]*Node, bool) {
	return nil, false
}

type sequenceParser struct {
	name            string
	id              int
	commit          CommitType
	items           []parser
	ranges          [][]int
	generalizations []int
	allChars        bool
}
type sequenceBuilder struct {
	name            string
	id              int
	commit          CommitType
	items           []builder
	ranges          [][]int
	generalizations []int
	allChars        bool
}

func (p *sequenceParser) nodeName() string {
	return p.name
}
func (p *sequenceParser) nodeID() int {
	return p.id
}
func (p *sequenceParser) commitType() CommitType {
	return p.commit
}
func (p *sequenceParser) parse(c *context) {
	if !p.allChars {
		if c.results.pending(c.offset, p.id) {
			c.fail(c.offset)
			return
		}
		c.results.markPending(c.offset, p.id)
	}
	var (
		currentCount int
		parsed       bool
	)
	itemIndex := 0
	from := c.offset
	to := c.offset
	for itemIndex < len(p.items) {
		p.items[itemIndex].parse(c)
		if !c.matchLast {
			if currentCount >= p.ranges[itemIndex][0] {
				itemIndex++
				currentCount = 0
				continue
			}
			c.offset = from
			if c.fromResults(p) {
				if to > c.failOffset {
					c.failOffset = -1
					c.failingParser = nil
				}
				if !p.allChars {
					c.results.unmarkPending(from, p.id)
				}
				return
			}
			if c.failingParser == nil && p.commit&userDefined != 0 && p.commit&Whitespace == 0 && p.commit&FailPass == 0 {
				c.failingParser = p
			}
			c.fail(from)
			if !p.allChars {
				c.results.unmarkPending(from, p.id)
			}
			return
		}
		parsed = c.offset > to
		if parsed {
			currentCount++
		}
		to = c.offset
		if !parsed || p.ranges[itemIndex][1] > 0 && currentCount == p.ranges[itemIndex][1] {
			itemIndex++
			currentCount = 0
		}
	}
	if p.commit&NoKeyword != 0 && c.isKeyword(from, to) {
		if c.failingParser == nil && p.commit&userDefined != 0 && p.commit&Whitespace == 0 && p.commit&FailPass == 0 {
			c.failingParser = p
		}
		c.fail(from)
		if !p.allChars {
			c.results.unmarkPending(from, p.id)
		}
		return
	}
	for _, g := range p.generalizations {
		if c.results.pending(from, g) {
			c.results.setMatch(from, g, to)
		}
	}
	if to > c.failOffset {
		c.failOffset = -1
		c.failingParser = nil
	}
	c.results.setMatch(from, p.id, to)
	c.success(to)
	if !p.allChars {
		c.results.unmarkPending(from, p.id)
	}
}
func (b *sequenceBuilder) nodeName() string {
	return b.name
}
func (b *sequenceBuilder) nodeID() int {
	return b.id
}
func (b *sequenceBuilder) build(c *context) ([]*Node, bool) {
	to, ok := c.results.longestMatch(c.offset, b.id)
	if !ok {
		return nil, false
	}
	from := c.offset
	parsed := to > from
	if b.allChars {
		c.offset = to
		if b.commit&Alias != 0 {
			return nil, true
		}
		return []*Node{{Name: b.name, From: from, To: to, tokens: c.tokens}}, true
	} else if parsed {
		c.results.dropMatchTo(c.offset, b.id, to)
		for _, g := range b.generalizations {
			c.results.dropMatchTo(c.offset, g, to)
		}
	} else {
		if c.results.pending(c.offset, b.id) {
			return nil, false
		}
		c.results.markPending(c.offset, b.id)
		for _, g := range b.generalizations {
			c.results.markPending(c.offset, g)
		}
	}
	var (
		itemIndex    int
		currentCount int
		nodes        []*Node
	)
	for itemIndex < len(b.items) {
		itemFrom := c.offset
		n, ok := b.items[itemIndex].build(c)
		if !ok {
			itemIndex++
			currentCount = 0
			continue
		}
		if c.offset > itemFrom {
			nodes = append(nodes, n...)
			currentCount++
			if b.ranges[itemIndex][1] > 0 && currentCount == b.ranges[itemIndex][1] {
				itemIndex++
				currentCount = 0
			}
			continue
		}
		if currentCount < b.ranges[itemIndex][0] {
			for i := 0; i < b.ranges[itemIndex][0]-currentCount; i++ {
				nodes = append(nodes, n...)
			}
		}
		itemIndex++
		currentCount = 0
	}
	if !parsed {
		c.results.unmarkPending(from, b.id)
		for _, g := range b.generalizations {
			c.results.unmarkPending(from, g)
		}
	}
	if b.commit&Alias != 0 {
		return nodes, true
	}
	return []*Node{{Name: b.name, From: from, To: to, Nodes: nodes, tokens: c.tokens}}, true
}

type choiceParser struct {
	name            string
	id              int
	commit          CommitType
	options         []parser
	generalizations []int
}
type choiceBuilder struct {
	name            string
	id              int
	commit          CommitType
	options         []builder
	generalizations []int
}

func (p *choiceParser) nodeName() string {
	return p.name
}
func (p *choiceParser) nodeID() int {
	return p.id
}
func (p *choiceParser) commitType() CommitType {
	return p.commit
}
func (p *choiceParser) parse(c *context) {
	if c.fromResults(p) {
		return
	}
	if c.results.pending(c.offset, p.id) {
		c.fail(c.offset)
		return
	}
	c.results.markPending(c.offset, p.id)
	var (
		match         bool
		optionIndex   int
		foundMatch    bool
		failingParser parser
	)
	from := c.offset
	to := c.offset
	initialFailOffset := c.failOffset
	initialFailingParser := c.failingParser
	failOffset := initialFailOffset
	for {
		foundMatch = false
		optionIndex = 0
		for optionIndex < len(p.options) {
			p.options[optionIndex].parse(c)
			optionIndex++
			if !c.matchLast {
				if c.failOffset > failOffset {
					failOffset = c.failOffset
					failingParser = c.failingParser
				}
			}
			if !c.matchLast || match && c.offset <= to {
				c.offset = from
				continue
			}
			match = true
			foundMatch = true
			to = c.offset
			c.offset = from
			c.results.setMatch(from, p.id, to)
		}
		if !foundMatch {
			break
		}
	}
	if match {
		if p.commit&NoKeyword != 0 && c.isKeyword(from, to) {
			if c.failingParser == nil && p.commit&userDefined != 0 && p.commit&Whitespace == 0 && p.commit&FailPass == 0 {
				c.failingParser = p
			}
			c.fail(from)
			c.results.unmarkPending(from, p.id)
			return
		}
		if failOffset > to {
			c.failOffset = failOffset
			c.failingParser = failingParser
		} else if to > initialFailOffset {
			c.failOffset = -1
			c.failingParser = nil
		} else {
			c.failOffset = initialFailOffset
			c.failingParser = initialFailingParser
		}
		c.success(to)
		c.results.unmarkPending(from, p.id)
		return
	}
	if failOffset > initialFailOffset {
		c.failOffset = failOffset
		c.failingParser = failingParser
		if c.failingParser == nil && p.commitType()&userDefined != 0 && p.commitType()&Whitespace == 0 && p.commitType()&FailPass == 0 {
			c.failingParser = p
		}
	}
	c.results.setNoMatch(from, p.id)
	c.fail(from)
	c.results.unmarkPending(from, p.id)
}
func (b *choiceBuilder) nodeName() string {
	return b.name
}
func (b *choiceBuilder) nodeID() int {
	return b.id
}
func (b *choiceBuilder) build(c *context) ([]*Node, bool) {
	to, ok := c.results.longestMatch(c.offset, b.id)
	if !ok {
		return nil, false
	}
	from := c.offset
	parsed := to > from
	if parsed {
		c.results.dropMatchTo(c.offset, b.id, to)
		for _, g := range b.generalizations {
			c.results.dropMatchTo(c.offset, g, to)
		}
	} else {
		if c.results.pending(c.offset, b.id) {
			return nil, false
		}
		c.results.markPending(c.offset, b.id)
		for _, g := range b.generalizations {
			c.results.markPending(c.offset, g)
		}
	}
	var option builder
	for _, o := range b.options {
		if c.results.hasMatchTo(c.offset, o.nodeID(), to) {
			option = o
			break
		}
	}
	n, _ := option.build(c)
	if !parsed {
		c.results.unmarkPending(from, b.id)
		for _, g := range b.generalizations {
			c.results.unmarkPending(from, g)
		}
	}
	if b.commit&Alias != 0 {
		return n, true
	}
	return []*Node{{Name: b.name, From: from, To: to, Nodes: n, tokens: c.tokens}}, true
}

type idSet struct{ ids []uint }

func divModBits(id int) (int, int) {
	return id / strconv.IntSize, id % strconv.IntSize
}
func (s *idSet) set(id int) {
	d, m := divModBits(id)
	if d >= len(s.ids) {
		if d < cap(s.ids) {
			s.ids = s.ids[:d+1]
		} else {
			s.ids = s.ids[:cap(s.ids)]
			for i := cap(s.ids); i <= d; i++ {
				s.ids = append(s.ids, 0)
			}
		}
	}
	s.ids[d] |= 1 << uint(m)
}
func (s *idSet) unset(id int) {
	d, m := divModBits(id)
	if d >= len(s.ids) {
		return
	}
	s.ids[d] &^= 1 << uint(m)
}
func (s *idSet) has(id int) bool {
	d, m := divModBits(id)
	if d >= len(s.ids) {
		return false
	}
	return s.ids[d]&(1<<uint(m)) != 0
}

type results struct {
	noMatch   []*idSet
	match     [][]int
	isPending [][]int
}

func ensureOffsetInts(ints [][]int, offset int) [][]int {
	if len(ints) > offset {
		return ints
	}
	if cap(ints) > offset {
		ints = ints[:offset+1]
		return ints
	}
	ints = ints[:cap(ints)]
	for i := len(ints); i <= offset; i++ {
		ints = append(ints, nil)
	}
	return ints
}
func ensureOffsetIDs(ids []*idSet, offset int) []*idSet {
	if len(ids) > offset {
		return ids
	}
	if cap(ids) > offset {
		ids = ids[:offset+1]
		return ids
	}
	ids = ids[:cap(ids)]
	for i := len(ids); i <= offset; i++ {
		ids = append(ids, nil)
	}
	return ids
}
func (r *results) setMatch(offset, id, to int) {
	r.match = ensureOffsetInts(r.match, offset)
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id || r.match[offset][i+1] != to {
			continue
		}
		return
	}
	r.match[offset] = append(r.match[offset], id, to)
}
func (r *results) setNoMatch(offset, id int) {
	if len(r.match) > offset {
		for i := 0; i < len(r.match[offset]); i += 2 {
			if r.match[offset][i] != id {
				continue
			}
			return
		}
	}
	r.noMatch = ensureOffsetIDs(r.noMatch, offset)
	if r.noMatch[offset] == nil {
		r.noMatch[offset] = &idSet{}
	}
	r.noMatch[offset].set(id)
}
func (r *results) hasMatchTo(offset, id, to int) bool {
	if len(r.match) <= offset {
		return false
	}
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id {
			continue
		}
		if r.match[offset][i+1] == to {
			return true
		}
	}
	return false
}
func (r *results) longestMatch(offset, id int) (int, bool) {
	if len(r.match) <= offset {
		return 0, false
	}
	var found bool
	to := -1
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id {
			continue
		}
		if r.match[offset][i+1] > to {
			to = r.match[offset][i+1]
		}
		found = true
	}
	return to, found
}
func (r *results) longestResult(offset, id int) (int, bool, bool) {
	if len(r.noMatch) > offset && r.noMatch[offset] != nil && r.noMatch[offset].has(id) {
		return 0, false, true
	}
	to, ok := r.longestMatch(offset, id)
	return to, ok, ok
}
func (r *results) dropMatchTo(offset, id, to int) {
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id {
			continue
		}
		if r.match[offset][i+1] == to {
			r.match[offset][i] = -1
			return
		}
	}
}
func (r *results) resetPending() {
	r.isPending = nil
}
func (r *results) pending(offset, id int) bool {
	if len(r.isPending) <= id {
		return false
	}
	for i := range r.isPending[id] {
		if r.isPending[id][i] == offset {
			return true
		}
	}
	return false
}
func (r *results) markPending(offset, id int) {
	r.isPending = ensureOffsetInts(r.isPending, id)
	for i := range r.isPending[id] {
		if r.isPending[id][i] == -1 {
			r.isPending[id][i] = offset
			return
		}
	}
	r.isPending[id] = append(r.isPending[id], offset)
}
func (r *results) unmarkPending(offset, id int) {
	for i := range r.isPending[id] {
		if r.isPending[id][i] == offset {
			r.isPending[id][i] = -1
			break
		}
	}
}

type context struct {
	reader        io.RuneReader
	keywords      []parser
	offset        int
	readOffset    int
	consumed      int
	offsetLimit   int
	failOffset    int
	failingParser parser
	readErr       error
	eof           bool
	results       *results
	tokens        []rune
	matchLast     bool
}

func newContext(r io.RuneReader, keywords []parser) *context {
	return &context{reader: r, keywords: keywords, results: &results{}, offsetLimit: -1, failOffset: -1}
}
func (c *context) read() bool {
	if c.eof || c.readErr != nil {
		return false
	}
	token, n, err := c.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			if n == 0 {
				c.eof = true
				return false
			}
		} else {
			c.readErr = err
			return false
		}
	}
	c.readOffset++
	if token == unicode.ReplacementChar {
		c.readErr = ErrInvalidUnicodeCharacter
		return false
	}
	c.tokens = append(c.tokens, token)
	return true
}
func (c *context) token() (rune, bool) {
	if c.offset == c.offsetLimit {
		return 0, false
	}
	if c.offset == c.readOffset {
		if !c.read() {
			return 0, false
		}
	}
	return c.tokens[c.offset], true
}
func (c *context) fromResults(p parser) bool {
	to, m, ok := c.results.longestResult(c.offset, p.nodeID())
	if !ok {
		return false
	}
	if m {
		c.success(to)
	} else {
		c.fail(c.offset)
	}
	return true
}
func (c *context) isKeyword(from, to int) bool {
	ol := c.offsetLimit
	c.offsetLimit = to
	defer func() {
		c.offsetLimit = ol
	}()
	for _, kw := range c.keywords {
		c.offset = from
		kw.parse(c)
		if c.matchLast && c.offset == to {
			return true
		}
	}
	return false
}
func (c *context) success(to int) {
	c.offset = to
	c.matchLast = true
	if to > c.consumed {
		c.consumed = to
	}
}
func (c *context) fail(offset int) {
	c.offset = offset
	c.matchLast = false
}
func findLine(tokens []rune, offset int) (line, column int) {
	tokens = tokens[:offset]
	for i := range tokens {
		column++
		if tokens[i] == '\n' {
			column = 0
			line++
		}
	}
	return
}
func (c *context) parseError(p parser) error {
	definition := p.nodeName()
	flagIndex := strings.Index(definition, ":")
	if flagIndex > 0 {
		definition = definition[:flagIndex]
	}
	if c.failingParser == nil {
		c.failOffset = c.consumed
	}
	line, col := findLine(c.tokens, c.failOffset)
	return &ParseError{Offset: c.failOffset, Line: line, Column: col, Definition: definition}
}
func (c *context) finalizeParse(root parser) error {
	fp := c.failingParser
	if fp == nil {
		fp = root
	}
	to, match, found := c.results.longestResult(0, root.nodeID())
	if !found || !match || found && match && to < c.readOffset {
		return c.parseError(fp)
	}
	c.read()
	if c.eof {
		return nil
	}
	if c.readErr != nil {
		return c.readErr
	}
	return c.parseError(root)
}

type Node struct {
	Name     string
	Nodes    []*Node
	From, To int
	tokens   []rune
}

func (n *Node) Tokens() []rune {
	return n.tokens
}
func (n *Node) String() string {
	return fmt.Sprintf("%s:%d:%d:%s", n.Name, n.From, n.To, n.Text())
}
func (n *Node) Text() string {
	return string(n.Tokens()[n.From:n.To])
}

type CommitType int

const (
	None  CommitType = 0
	Alias CommitType = 1 << iota
	Whitespace
	NoWhitespace
	Keyword
	NoKeyword
	FailPass
	Root
	userDefined
)

type formatFlags int

const (
	formatNone   formatFlags = 0
	formatPretty formatFlags = 1 << iota
	formatIncludeComments
)

type ParseError struct {
	Input      string
	Offset     int
	Line       int
	Column     int
	Definition string
}
type parser interface {
	nodeName() string
	nodeID() int
	commitType() CommitType
	parse(*context)
}
type builder interface {
	nodeName() string
	nodeID() int
	build(*context) ([]*Node, bool)
}

var ErrInvalidUnicodeCharacter = errors.New("invalid unicode character")

func (pe *ParseError) Error() string {
	return fmt.Sprintf("%s:%d:%d:parse failed, parsing: %s", pe.Input, pe.Line+1, pe.Column+1, pe.Definition)
}
func parseInput(r io.Reader, p parser, b builder, kw []parser) (*Node, error) {
	c := newContext(bufio.NewReader(r), kw)
	p.parse(c)
	if c.readErr != nil {
		return nil, c.readErr
	}
	if err := c.finalizeParse(p); err != nil {
		if perr, ok := err.(*ParseError); ok {
			perr.Input = "<input>"
		}
		return nil, err
	}
	c.offset = 0
	c.results.resetPending()
	n, _ := b.build(c)
	return n[0], nil
}

func parse(r io.Reader) (*Node, error) {

	var p94 = sequenceParser{id: 94, commit: 128, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p92 = choiceParser{id: 92, commit: 2}
	var p91 = sequenceParser{id: 91, commit: 262, name: "whitespace", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{92}}
	var p1 = charParser{id: 1, chars: []rune{32, 8, 12, 13, 9, 11}}
	p91.items = []parser{&p1}
	p92.options = []parser{&p91}
	var p93 = sequenceParser{id: 93, commit: 258, name: "config:wsroot", ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var p88 = sequenceParser{id: 88, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p3 = sequenceParser{id: 3, commit: 266, name: "nl", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p2 = charParser{id: 2, chars: []rune{10}}
	p3.items = []parser{&p2}
	var p87 = sequenceParser{id: 87, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p87.items = []parser{&p92, &p3}
	p88.items = []parser{&p3, &p87}
	var p86 = sequenceParser{id: 86, commit: 2, ranges: [][]int{{1, 1}, {0, 1}}}
	var p81 = choiceParser{id: 81, commit: 258, name: "entry"}
	var p10 = sequenceParser{id: 10, commit: 258, name: "comment", ranges: [][]int{{1, 1}, {0, 1}}, generalizations: []int{81, 76}}
	var p5 = sequenceParser{id: 5, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p4 = charParser{id: 4, chars: []rune{35}}
	p5.items = []parser{&p4}
	var p9 = sequenceParser{id: 9, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p7 = sequenceParser{id: 7, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p6 = charParser{id: 6, not: true, chars: []rune{10}}
	p7.items = []parser{&p6}
	var p8 = sequenceParser{id: 8, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p8.items = []parser{&p92, &p7}
	p9.items = []parser{&p92, &p7, &p8}
	p10.items = []parser{&p5, &p9}
	var p80 = sequenceParser{id: 80, commit: 256, name: "group", ranges: [][]int{{1, 1}, {0, 1}}, generalizations: []int{81}}
	var p75 = choiceParser{id: 75, commit: 258, name: "group-key-form"}
	var p73 = sequenceParser{id: 73, commit: 256, name: "group-key", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{75}}
	var p70 = sequenceParser{id: 70, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p69 = charParser{id: 69, chars: []rune{91}}
	p70.items = []parser{&p69}
	var p65 = sequenceParser{id: 65, commit: 264, name: "key", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var p57 = sequenceParser{id: 57, commit: 264, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var p56 = sequenceParser{id: 56, commit: 258, name: "symbol-char", allChars: true, ranges: [][]int{{1, 1}}}
	var p55 = charParser{id: 55, chars: []rune{95, 45}, ranges: [][]rune{{97, 122}, {65, 90}, {48, 57}}}
	p56.items = []parser{&p55}
	p57.items = []parser{&p56}
	var p64 = sequenceParser{id: 64, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p63 = choiceParser{id: 63, commit: 258, name: "key-sep"}
	var p59 = sequenceParser{id: 59, commit: 2, allChars: true, ranges: [][]int{{1, 1}}, generalizations: []int{63}}
	var p58 = charParser{id: 58, chars: []rune{46}}
	p59.items = []parser{&p58}
	var p62 = sequenceParser{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{63}}
	var p60 = charParser{id: 60, chars: []rune{58}}
	var p61 = charParser{id: 61, chars: []rune{58}}
	p62.items = []parser{&p60, &p61}
	p63.options = []parser{&p59, &p62}
	p64.items = []parser{&p63, &p57}
	p65.items = []parser{&p57, &p64}
	var p72 = sequenceParser{id: 72, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p71 = charParser{id: 71, chars: []rune{93}}
	p72.items = []parser{&p71}
	p73.items = []parser{&p70, &p92, &p65, &p92, &p72}
	var p74 = sequenceParser{id: 74, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}, generalizations: []int{75}}
	p74.items = []parser{&p73, &p92, &p10}
	p75.options = []parser{&p73, &p74}
	var p79 = sequenceParser{id: 79, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p77 = sequenceParser{id: 77, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var p76 = choiceParser{id: 76, commit: 2}
	var p68 = sequenceParser{id: 68, commit: 256, name: "keyed-value", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{76, 81}}
	var p67 = sequenceParser{id: 67, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p66 = charParser{id: 66, chars: []rune{61}}
	p67.items = []parser{&p66}
	var p54 = choiceParser{id: 54, commit: 258, name: "value-form", generalizations: []int{76}}
	var p52 = choiceParser{id: 52, commit: 256, name: "value", generalizations: []int{54, 76}}
	var p51 = sequenceParser{id: 51, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{52, 54, 76}}
	var p49 = choiceParser{id: 49, commit: 258, name: "value-char"}
	var p43 = sequenceParser{id: 43, commit: 2, allChars: true, ranges: [][]int{{1, 1}}, generalizations: []int{49}}
	var p42 = charParser{id: 42, not: true, chars: []rune{10, 39, 34, 92, 91, 93, 61, 35}}
	p43.items = []parser{&p42}
	var p48 = sequenceParser{id: 48, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}, generalizations: []int{49}}
	var p45 = sequenceParser{id: 45, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p44 = charParser{id: 44, chars: []rune{92}}
	p45.items = []parser{&p44}
	var p47 = sequenceParser{id: 47, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p46 = charParser{id: 46, not: true}
	p47.items = []parser{&p46}
	p48.items = []parser{&p45, &p92, &p47}
	p49.options = []parser{&p43, &p48}
	var p50 = sequenceParser{id: 50, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p50.items = []parser{&p92, &p49}
	p51.items = []parser{&p49, &p50}
	var p41 = choiceParser{id: 41, commit: 256, name: "quote", generalizations: []int{52, 54, 76}}
	var p25 = sequenceParser{id: 25, commit: 258, name: "single-quote", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{41, 52, 54, 76}}
	var p12 = sequenceParser{id: 12, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p11 = charParser{id: 11, chars: []rune{39}}
	p12.items = []parser{&p11}
	var p24 = sequenceParser{id: 24, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p20 = choiceParser{id: 20, commit: 2}
	var p14 = sequenceParser{id: 14, commit: 2, allChars: true, ranges: [][]int{{1, 1}}, generalizations: []int{20}}
	var p13 = charParser{id: 13, not: true, chars: []rune{39, 92}}
	p14.items = []parser{&p13}
	var p19 = sequenceParser{id: 19, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}, generalizations: []int{20}}
	var p16 = sequenceParser{id: 16, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p15 = charParser{id: 15, chars: []rune{92}}
	p16.items = []parser{&p15}
	var p18 = sequenceParser{id: 18, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p17 = charParser{id: 17, not: true}
	p18.items = []parser{&p17}
	p19.items = []parser{&p16, &p92, &p18}
	p20.options = []parser{&p14, &p19}
	var p23 = sequenceParser{id: 23, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p23.items = []parser{&p92, &p20}
	p24.items = []parser{&p92, &p20, &p23}
	var p22 = sequenceParser{id: 22, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p21 = charParser{id: 21, chars: []rune{39}}
	p22.items = []parser{&p21}
	p25.items = []parser{&p12, &p24, &p92, &p22}
	var p40 = sequenceParser{id: 40, commit: 258, name: "double-quote", ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{41, 52, 54, 76}}
	var p27 = sequenceParser{id: 27, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p26 = charParser{id: 26, chars: []rune{34}}
	p27.items = []parser{&p26}
	var p39 = sequenceParser{id: 39, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p35 = choiceParser{id: 35, commit: 2}
	var p29 = sequenceParser{id: 29, commit: 2, allChars: true, ranges: [][]int{{1, 1}}, generalizations: []int{35}}
	var p28 = charParser{id: 28, not: true, chars: []rune{34, 92}}
	p29.items = []parser{&p28}
	var p34 = sequenceParser{id: 34, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}, generalizations: []int{35}}
	var p31 = sequenceParser{id: 31, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p30 = charParser{id: 30, chars: []rune{92}}
	p31.items = []parser{&p30}
	var p33 = sequenceParser{id: 33, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p32 = charParser{id: 32, not: true}
	p33.items = []parser{&p32}
	p34.items = []parser{&p31, &p92, &p33}
	p35.options = []parser{&p29, &p34}
	var p38 = sequenceParser{id: 38, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p38.items = []parser{&p92, &p35}
	p39.items = []parser{&p92, &p35, &p38}
	var p37 = sequenceParser{id: 37, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var p36 = charParser{id: 36, chars: []rune{34}}
	p37.items = []parser{&p36}
	p40.items = []parser{&p27, &p39, &p92, &p37}
	p41.options = []parser{&p25, &p40}
	p52.options = []parser{&p51, &p41}
	var p53 = sequenceParser{id: 53, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}, generalizations: []int{54, 76}}
	p53.items = []parser{&p52, &p92, &p10}
	p54.options = []parser{&p52, &p53}
	p68.items = []parser{&p65, &p92, &p67, &p92, &p54}
	p76.options = []parser{&p68, &p54, &p10}
	p77.items = []parser{&p3, &p92, &p76}
	var p78 = sequenceParser{id: 78, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p78.items = []parser{&p92, &p77}
	p79.items = []parser{&p92, &p77, &p78}
	p80.items = []parser{&p75, &p79}
	p81.options = []parser{&p10, &p80, &p68}
	var p85 = sequenceParser{id: 85, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p83 = sequenceParser{id: 83, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p82 = sequenceParser{id: 82, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p82.items = []parser{&p92, &p3}
	p83.items = []parser{&p3, &p82, &p92, &p81}
	var p84 = sequenceParser{id: 84, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p84.items = []parser{&p92, &p83}
	p85.items = []parser{&p92, &p83, &p84}
	p86.items = []parser{&p81, &p85}
	var p90 = sequenceParser{id: 90, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p89 = sequenceParser{id: 89, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p89.items = []parser{&p92, &p3}
	p90.items = []parser{&p92, &p3, &p89}
	p93.items = []parser{&p88, &p92, &p86, &p90}
	p94.items = []parser{&p92, &p93, &p92}
	var b94 = sequenceBuilder{id: 94, commit: 128, name: "config", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b92 = choiceBuilder{id: 92, commit: 2}
	var b91 = sequenceBuilder{id: 91, commit: 262, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{92}}
	var b1 = charBuilder{}
	b91.items = []builder{&b1}
	b92.options = []builder{&b91}
	var b93 = sequenceBuilder{id: 93, commit: 258, ranges: [][]int{{0, 1}, {0, -1}, {0, 1}, {0, 1}}}
	var b88 = sequenceBuilder{id: 88, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b3 = sequenceBuilder{id: 3, commit: 266, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b2 = charBuilder{}
	b3.items = []builder{&b2}
	var b87 = sequenceBuilder{id: 87, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b87.items = []builder{&b92, &b3}
	b88.items = []builder{&b3, &b87}
	var b86 = sequenceBuilder{id: 86, commit: 2, ranges: [][]int{{1, 1}, {0, 1}}}
	var b81 = choiceBuilder{id: 81, commit: 258}
	var b10 = sequenceBuilder{id: 10, commit: 258, ranges: [][]int{{1, 1}, {0, 1}}, generalizations: []int{81, 76}}
	var b5 = sequenceBuilder{id: 5, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b4 = charBuilder{}
	b5.items = []builder{&b4}
	var b9 = sequenceBuilder{id: 9, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b7 = sequenceBuilder{id: 7, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b6 = charBuilder{}
	b7.items = []builder{&b6}
	var b8 = sequenceBuilder{id: 8, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b8.items = []builder{&b92, &b7}
	b9.items = []builder{&b92, &b7, &b8}
	b10.items = []builder{&b5, &b9}
	var b80 = sequenceBuilder{id: 80, commit: 256, name: "group", ranges: [][]int{{1, 1}, {0, 1}}, generalizations: []int{81}}
	var b75 = choiceBuilder{id: 75, commit: 258}
	var b73 = sequenceBuilder{id: 73, commit: 256, name: "group-key", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{75}}
	var b70 = sequenceBuilder{id: 70, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b69 = charBuilder{}
	b70.items = []builder{&b69}
	var b65 = sequenceBuilder{id: 65, commit: 264, name: "key", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b57 = sequenceBuilder{id: 57, commit: 264, name: "symbol", ranges: [][]int{{1, -1}, {1, -1}}}
	var b56 = sequenceBuilder{id: 56, commit: 258, allChars: true, ranges: [][]int{{1, 1}}}
	var b55 = charBuilder{}
	b56.items = []builder{&b55}
	b57.items = []builder{&b56}
	var b64 = sequenceBuilder{id: 64, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b63 = choiceBuilder{id: 63, commit: 258}
	var b59 = sequenceBuilder{id: 59, commit: 2, allChars: true, ranges: [][]int{{1, 1}}, generalizations: []int{63}}
	var b58 = charBuilder{}
	b59.items = []builder{&b58}
	var b62 = sequenceBuilder{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{63}}
	var b60 = charBuilder{}
	var b61 = charBuilder{}
	b62.items = []builder{&b60, &b61}
	b63.options = []builder{&b59, &b62}
	b64.items = []builder{&b63, &b57}
	b65.items = []builder{&b57, &b64}
	var b72 = sequenceBuilder{id: 72, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b71 = charBuilder{}
	b72.items = []builder{&b71}
	b73.items = []builder{&b70, &b92, &b65, &b92, &b72}
	var b74 = sequenceBuilder{id: 74, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}, generalizations: []int{75}}
	b74.items = []builder{&b73, &b92, &b10}
	b75.options = []builder{&b73, &b74}
	var b79 = sequenceBuilder{id: 79, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b77 = sequenceBuilder{id: 77, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}}
	var b76 = choiceBuilder{id: 76, commit: 2}
	var b68 = sequenceBuilder{id: 68, commit: 256, name: "keyed-value", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{76, 81}}
	var b67 = sequenceBuilder{id: 67, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b66 = charBuilder{}
	b67.items = []builder{&b66}
	var b54 = choiceBuilder{id: 54, commit: 258, generalizations: []int{76}}
	var b52 = choiceBuilder{id: 52, commit: 256, name: "value", generalizations: []int{54, 76}}
	var b51 = sequenceBuilder{id: 51, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}, generalizations: []int{52, 54, 76}}
	var b49 = choiceBuilder{id: 49, commit: 258}
	var b43 = sequenceBuilder{id: 43, commit: 2, allChars: true, ranges: [][]int{{1, 1}}, generalizations: []int{49}}
	var b42 = charBuilder{}
	b43.items = []builder{&b42}
	var b48 = sequenceBuilder{id: 48, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}, generalizations: []int{49}}
	var b45 = sequenceBuilder{id: 45, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b44 = charBuilder{}
	b45.items = []builder{&b44}
	var b47 = sequenceBuilder{id: 47, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b46 = charBuilder{}
	b47.items = []builder{&b46}
	b48.items = []builder{&b45, &b92, &b47}
	b49.options = []builder{&b43, &b48}
	var b50 = sequenceBuilder{id: 50, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b50.items = []builder{&b92, &b49}
	b51.items = []builder{&b49, &b50}
	var b41 = choiceBuilder{id: 41, commit: 256, name: "quote", generalizations: []int{52, 54, 76}}
	var b25 = sequenceBuilder{id: 25, commit: 258, ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{41, 52, 54, 76}}
	var b12 = sequenceBuilder{id: 12, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b11 = charBuilder{}
	b12.items = []builder{&b11}
	var b24 = sequenceBuilder{id: 24, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b20 = choiceBuilder{id: 20, commit: 2}
	var b14 = sequenceBuilder{id: 14, commit: 2, allChars: true, ranges: [][]int{{1, 1}}, generalizations: []int{20}}
	var b13 = charBuilder{}
	b14.items = []builder{&b13}
	var b19 = sequenceBuilder{id: 19, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}, generalizations: []int{20}}
	var b16 = sequenceBuilder{id: 16, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b15 = charBuilder{}
	b16.items = []builder{&b15}
	var b18 = sequenceBuilder{id: 18, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b17 = charBuilder{}
	b18.items = []builder{&b17}
	b19.items = []builder{&b16, &b92, &b18}
	b20.options = []builder{&b14, &b19}
	var b23 = sequenceBuilder{id: 23, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b23.items = []builder{&b92, &b20}
	b24.items = []builder{&b92, &b20, &b23}
	var b22 = sequenceBuilder{id: 22, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b21 = charBuilder{}
	b22.items = []builder{&b21}
	b25.items = []builder{&b12, &b24, &b92, &b22}
	var b40 = sequenceBuilder{id: 40, commit: 258, ranges: [][]int{{1, 1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{41, 52, 54, 76}}
	var b27 = sequenceBuilder{id: 27, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b26 = charBuilder{}
	b27.items = []builder{&b26}
	var b39 = sequenceBuilder{id: 39, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b35 = choiceBuilder{id: 35, commit: 2}
	var b29 = sequenceBuilder{id: 29, commit: 2, allChars: true, ranges: [][]int{{1, 1}}, generalizations: []int{35}}
	var b28 = charBuilder{}
	b29.items = []builder{&b28}
	var b34 = sequenceBuilder{id: 34, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}, generalizations: []int{35}}
	var b31 = sequenceBuilder{id: 31, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b30 = charBuilder{}
	b31.items = []builder{&b30}
	var b33 = sequenceBuilder{id: 33, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b32 = charBuilder{}
	b33.items = []builder{&b32}
	b34.items = []builder{&b31, &b92, &b33}
	b35.options = []builder{&b29, &b34}
	var b38 = sequenceBuilder{id: 38, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b38.items = []builder{&b92, &b35}
	b39.items = []builder{&b92, &b35, &b38}
	var b37 = sequenceBuilder{id: 37, commit: 2, allChars: true, ranges: [][]int{{1, 1}}}
	var b36 = charBuilder{}
	b37.items = []builder{&b36}
	b40.items = []builder{&b27, &b39, &b92, &b37}
	b41.options = []builder{&b25, &b40}
	b52.options = []builder{&b51, &b41}
	var b53 = sequenceBuilder{id: 53, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}}, generalizations: []int{54, 76}}
	b53.items = []builder{&b52, &b92, &b10}
	b54.options = []builder{&b52, &b53}
	b68.items = []builder{&b65, &b92, &b67, &b92, &b54}
	b76.options = []builder{&b68, &b54, &b10}
	b77.items = []builder{&b3, &b92, &b76}
	var b78 = sequenceBuilder{id: 78, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b78.items = []builder{&b92, &b77}
	b79.items = []builder{&b92, &b77, &b78}
	b80.items = []builder{&b75, &b79}
	b81.options = []builder{&b10, &b80, &b68}
	var b85 = sequenceBuilder{id: 85, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b83 = sequenceBuilder{id: 83, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b82 = sequenceBuilder{id: 82, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b82.items = []builder{&b92, &b3}
	b83.items = []builder{&b3, &b82, &b92, &b81}
	var b84 = sequenceBuilder{id: 84, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b84.items = []builder{&b92, &b83}
	b85.items = []builder{&b92, &b83, &b84}
	b86.items = []builder{&b81, &b85}
	var b90 = sequenceBuilder{id: 90, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b89 = sequenceBuilder{id: 89, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b89.items = []builder{&b92, &b3}
	b90.items = []builder{&b92, &b3, &b89}
	b93.items = []builder{&b88, &b92, &b86, &b90}
	b94.items = []builder{&b92, &b93, &b92}

	var keywords = []parser{}

	return parseInput(r, &p94, &b94, keywords)
}
