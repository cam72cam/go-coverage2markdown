package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Data struct {
	File       string
	LineNumber string
	Function   string
	Percent    string
}

var re = regexp.MustCompile("\t+")

func ParseLine(line string) Data {
	sp := strings.Split(line, ":")

	if sp[0] == "total" {
		sp = []string{sp[0], "", sp[1]}
	}
	ss := strings.Split(strings.Trim(re.ReplaceAllString(sp[2], " "), " "), " ")

	return Data{
		File:       sp[0],
		LineNumber: sp[1],
		Function:   ss[0],
		Percent:    ss[1],
	}
}

type Node struct {
	Children map[string]*Node
	Data     []Data
}

func (n *Node) AddData(data Data) {
	current := n
	parts := strings.Split(data.File, "/")
	if len(parts) > 1 {
		parts = parts[:len(parts)-1]
	}
	for _, part := range parts {
		child, ok := current.Children[part]
		if !ok {
			child = &Node{
				Children: make(map[string]*Node),
				Data:     make([]Data, 0),
			}
			current.Children[part] = child
		}
		current = child
	}
	current.Data = append(current.Data, data)
}

func (n *Node) Print(prefix string) {
	if len(n.Children) == 1 && len(n.Data) == 0 {
		// Pass through prefix
		for key, child := range n.Children {
			child.Print(fmt.Sprintf("%s/%s", prefix, key))
		}
	} else {
		if prefix != "" && prefix != "total" {
			fmt.Printf("<details><summary>%s</summary>\n", prefix)
		}
		if len(n.Children) > 0 {
			if prefix != "" {
				fmt.Printf("\n<blockquote>\n")
			}

			keys := make([]string, 0)

			for key, _ := range n.Children {
				keys = append(keys, key)
			}

			sort.Strings(keys)

			for _, key := range keys {
				child := n.Children[key]
				child.Print(key)
			}
			if prefix != "" {
				fmt.Printf("</blockquote>\n")
			}
		}
		if len(n.Data) > 0 {
			fmt.Printf("\n| Location | Function | Coverage |\n")
			fmt.Printf("| -------- | -------- | -------- |\n")
			for _, data := range n.Data {
				parts := strings.Split(data.File, "/")
				marker := ""
				pct, _ := strconv.ParseFloat(data.Percent[:len(data.Percent)-1], 32)
				if pct > 75 {
					marker = "🟢"
				} else if pct >= 50 {
					marker = "🟡"
				} else {
					marker = "🔴"
				}
				fmt.Printf("| %s:%s | %s %s | %s |\n", parts[len(parts)-1], data.LineNumber, marker, data.Function, data.Percent)
			}
		}
		if prefix != "" && prefix != "total" {
			fmt.Printf("</details>\n")
		}
	}
}

func main() {
	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(content), "\n")

	root := &Node{
		Children: make(map[string]*Node),
		Data:     make([]Data, 0),
	}

	fmt.Printf("## Coverage Report:\n")

	for _, line := range lines {
		if line != "" {
			root.AddData(ParseLine(line))
		}
	}

	root.Print("")
}
