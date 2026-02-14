package cmd

import (
	"reflect"
	"slices"
	"testing"

	"github.com/alecthomas/kong"
)

func TestCompleteWordsRootAndSubcommands(t *testing.T) {
	suggestions, err := completeWords(1, []string{"attio", ""})
	if err != nil {
		t.Fatalf("completeWords root: %v", err)
	}
	if !slices.Contains(suggestions, "objects") {
		t.Fatalf("expected objects suggestion, got %#v", suggestions)
	}
	if !slices.Contains(suggestions, "--json") {
		t.Fatalf("expected --json flag suggestion, got %#v", suggestions)
	}

	suggestions, err = completeWords(2, []string{"attio", "records", ""})
	if err != nil {
		t.Fatalf("completeWords records: %v", err)
	}
	if !slices.Contains(suggestions, "query") || !slices.Contains(suggestions, "search") {
		t.Fatalf("expected records subcommands, got %#v", suggestions)
	}
}

func TestCompleteWordsStopsForValueAndTerminator(t *testing.T) {
	suggestions, err := completeWords(2, []string{"attio", "--profile", ""})
	if err != nil {
		t.Fatalf("completeWords flag value: %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected no suggestions for pending flag value, got %#v", suggestions)
	}

	suggestions, err = completeWords(2, []string{"attio", "--", "rest"})
	if err != nil {
		t.Fatalf("completeWords terminator: %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected no suggestions after terminator, got %#v", suggestions)
	}
}

func TestCompleteWordsFlagPrefixes(t *testing.T) {
	suggestions, err := completeWords(1, []string{"attio", "--j"})
	if err != nil {
		t.Fatalf("completeWords flag prefix: %v", err)
	}
	if !slices.Contains(suggestions, "--json") {
		t.Fatalf("expected --json suggestion, got %#v", suggestions)
	}
}

func TestCompletionRootNodeAndBuilders(t *testing.T) {
	root, err := completionRootNode()
	if err != nil {
		t.Fatalf("completionRootNode: %v", err)
	}
	if root == nil {
		t.Fatalf("expected non-nil completion root")
	}
	if _, ok := root.children["records"]; !ok {
		t.Fatalf("expected records child in completion root")
	}
	if _, ok := root.flags["--json"]; !ok {
		t.Fatalf("expected --json flag in completion root")
	}
}

func TestCompletionHelpers(t *testing.T) {
	if got := normalizeCword(-1, 2); got != 1 {
		t.Fatalf("unexpected normalized cword: %d", got)
	}
	if got := normalizeCword(5, 2); got != 2 {
		t.Fatalf("unexpected normalized cword upper clamp: %d", got)
	}
	if got := normalizeCword(-1, 0); got != -1 {
		t.Fatalf("unexpected normalized cword for empty words: %d", got)
	}

	if got := completionStartIndex([]string{"attio", "records"}); got != 1 {
		t.Fatalf("expected start index 1 for program name, got %d", got)
	}
	if got := completionStartIndex([]string{"records"}); got != 0 {
		t.Fatalf("expected start index 0 for non-program, got %d", got)
	}

	if !isProgramName("/usr/local/bin/attio") {
		t.Fatalf("expected attio binary to be recognized")
	}
	if !isProgramName("ATTIO.EXE") {
		t.Fatalf("expected attio.exe to be recognized case-insensitively")
	}
	if isProgramName("something-else") {
		t.Fatalf("did not expect other binary to be recognized")
	}

	flag, hasValue := splitFlagToken("--profile=dev")
	if flag != "--profile" || !hasValue {
		t.Fatalf("unexpected splitFlagToken result: %q %v", flag, hasValue)
	}
	flag, hasValue = splitFlagToken("--profile")
	if flag != "--profile" || hasValue {
		t.Fatalf("unexpected splitFlagToken without value: %q %v", flag, hasValue)
	}
}

func TestAdvanceCompletionNodeAndMatchingHelpers(t *testing.T) {
	node := &completionNode{
		children: map[string]*completionNode{
			"records": {
				children: map[string]*completionNode{
					"query": {children: map[string]*completionNode{}, flags: map[string]completionFlag{}},
				},
				flags: map[string]completionFlag{"--limit": {takesValue: true}},
			},
		},
		flags: map[string]completionFlag{"--json": {takesValue: false}, "--profile": {takesValue: true}},
	}

	gotNode, termIdx, needsValue := advanceCompletionNode(node, []string{"attio", "records", "query"}, 1, 3)
	if needsValue {
		t.Fatalf("did not expect pending value")
	}
	if termIdx != -1 {
		t.Fatalf("unexpected terminator index: %d", termIdx)
	}
	if _, ok := gotNode.children["query"]; ok {
		t.Fatalf("expected to advance to records child node")
	}

	_, _, needsValue = advanceCompletionNode(node, []string{"attio", "--profile", ""}, 1, 2)
	if !needsValue {
		t.Fatalf("expected pending value for --profile")
	}

	if !shouldStopAfterTerminator(1, 2, []string{"attio", "--", "x"}) {
		t.Fatalf("expected stop after explicit terminator")
	}
	if !shouldStopAfterTerminator(-1, 1, []string{"attio", "--"}) {
		t.Fatalf("expected stop when current token is terminator")
	}

	if !expectsFlagValue(node, 2, []string{"attio", "--profile", ""}, 1) {
		t.Fatalf("expected flag value detection")
	}
	if expectsFlagValue(node, 2, []string{"attio", "records", ""}, 1) {
		t.Fatalf("did not expect flag value detection for command token")
	}

	cmds := matchingCommands(node, "rec")
	if !reflect.DeepEqual(cmds, []string{"records"}) {
		t.Fatalf("unexpected matching commands: %#v", cmds)
	}
	flags := matchingFlags(node, "--j")
	if !reflect.DeepEqual(flags, []string{"--json"}) {
		t.Fatalf("unexpected matching flags: %#v", flags)
	}
}

func TestNegatedFlagNameBranches(t *testing.T) {
	plain := &kong.Flag{Value: &kong.Value{Name: "json", Tag: &kong.Tag{Negatable: ""}}}
	if got := negatedFlagName(plain); got != "" {
		t.Fatalf("expected empty negated flag for non-negatable input, got %q", got)
	}

	auto := &kong.Flag{Value: &kong.Value{Name: "json", Tag: &kong.Tag{Negatable: "_"}}}
	if got := negatedFlagName(auto); got != "--no-json" {
		t.Fatalf("expected auto negated flag '--no-json', got %q", got)
	}

	custom := &kong.Flag{Value: &kong.Value{Name: "json", Tag: &kong.Tag{Negatable: "without-json"}}}
	if got := negatedFlagName(custom); got != "--without-json" {
		t.Fatalf("expected custom negated flag '--without-json', got %q", got)
	}
}
