package cmd

import (
	"context"
	"os"
	"sort"

	"github.com/alecthomas/kong"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type SchemaCmd struct {
}

type commandSchema struct {
	Name        string          `json:"name"`
	Path        string          `json:"path"`
	Help        string          `json:"help,omitempty"`
	Hidden      bool            `json:"hidden,omitempty"`
	Aliases     []string        `json:"aliases,omitempty"`
	Flags       []flagSchema    `json:"flags,omitempty"`
	Positionals []valueSchema   `json:"positionals,omitempty"`
	Commands    []commandSchema `json:"commands,omitempty"`
}

type flagSchema struct {
	Name      string   `json:"name"`
	Short     string   `json:"short,omitempty"`
	Help      string   `json:"help,omitempty"`
	Required  bool     `json:"required,omitempty"`
	Default   string   `json:"default,omitempty"`
	Type      string   `json:"type,omitempty"`
	Enum      []string `json:"enum,omitempty"`
	Env       []string `json:"env,omitempty"`
	Aliases   []string `json:"aliases,omitempty"`
	Negatable bool     `json:"negatable,omitempty"`
}

type valueSchema struct {
	Name     string   `json:"name"`
	Help     string   `json:"help,omitempty"`
	Required bool     `json:"required,omitempty"`
	Default  string   `json:"default,omitempty"`
	Type     string   `json:"type,omitempty"`
	Enum     []string `json:"enum,omitempty"`
}

func (c *SchemaCmd) Run(ctx context.Context, parser *kong.Kong) error {
	root := buildCommandSchema(parser.Model.Node)
	return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
		"version": buildVersionString(),
		"root":    root,
	})
}

func buildCommandSchema(node *kong.Node) commandSchema {
	schema := commandSchema{
		Name:    node.Name,
		Path:    node.FullPath(),
		Help:    node.Help,
		Hidden:  node.Hidden,
		Aliases: append([]string(nil), node.Aliases...),
	}

	schema.Flags = make([]flagSchema, 0, len(node.Flags))
	for _, flag := range node.Flags {
		if flag == nil || flag.Hidden {
			continue
		}
		f := flagSchema{
			Name:      flag.Name,
			Help:      flag.Help,
			Required:  flag.Required,
			Default:   flag.Default,
			Aliases:   append([]string(nil), flag.Aliases...),
			Env:       append([]string(nil), flag.Envs...),
			Negatable: flag.Negated,
		}
		if flag.Short != 0 {
			f.Short = string(flag.Short)
		}
		if flag.Target.IsValid() {
			f.Type = flag.Target.Type().String()
		}
		if enum := flag.EnumSlice(); len(enum) > 0 {
			f.Enum = enum
		}
		schema.Flags = append(schema.Flags, f)
	}
	sort.Slice(schema.Flags, func(i, j int) bool {
		return schema.Flags[i].Name < schema.Flags[j].Name
	})

	schema.Positionals = make([]valueSchema, 0, len(node.Positional))
	for _, positional := range node.Positional {
		if positional == nil {
			continue
		}
		v := valueSchema{
			Name:     positional.Name,
			Help:     positional.Help,
			Required: positional.Required,
			Default:  positional.Default,
		}
		if positional.Target.IsValid() {
			v.Type = positional.Target.Type().String()
		}
		if enum := positional.EnumSlice(); len(enum) > 0 {
			v.Enum = enum
		}
		schema.Positionals = append(schema.Positionals, v)
	}

	children := make([]*kong.Node, 0, len(node.Children))
	for _, child := range node.Children {
		if child == nil || child.Hidden {
			continue
		}
		children = append(children, child)
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i].Name < children[j].Name
	})
	schema.Commands = make([]commandSchema, 0, len(children))
	for _, child := range children {
		schema.Commands = append(schema.Commands, buildCommandSchema(child))
	}

	return schema
}
