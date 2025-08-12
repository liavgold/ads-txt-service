package parser

import (
	"bufio"
	"io"
	"strings"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseAdsTxt(r io.Reader) map[string]int {
	out := make(map[string]int)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) == 0 {
			continue
		}
		ad := strings.TrimSpace(parts[0])
		if ad == "" {
			continue
		}
		out[ad]++
	}
	return out
}
