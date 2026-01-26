package search

import (
	"fmt"
	"time"

	"github.com/google/cel-go/cel"
)

// NewCELEnv creates the CEL environment with search-specific variables.
func NewCELEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("user", cel.StringType),
		cel.Variable("now", cel.TimestampType),
		cel.Variable("item_type", cel.StringType),
		cel.Variable("state", cel.StringType),
		cel.Variable("title", cel.StringType),
		cel.Variable("body", cel.StringType),
		cel.Variable("author", cel.StringType),
		cel.Variable("assigned", cel.ListType(cel.StringType)),
		cel.Variable("reviewers", cel.ListType(cel.StringType)),
		cel.Variable("labels", cel.ListType(cel.StringType)),
		cel.Variable("created", cel.TimestampType),
		cel.Variable("updated", cel.TimestampType),
		cel.Variable("owner", cel.StringType),
		cel.Variable("repo", cel.StringType),
		cel.Variable("number", cel.IntType),
	)
}

// CompileCELFilter compiles a CEL filter expression into a program.
func CompileCELFilter(expr string) (cel.Program, error) {
	env, err := NewCELEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("failed to compile CEL expression: %w", issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL program: %w", err)
	}

	return prg, nil
}

// EvaluateFilter evaluates a compiled CEL program against a variable map.
func EvaluateFilter(prg cel.Program, vars map[string]any) (bool, error) {
	// Add current time
	vars["now"] = time.Now()

	out, _, err := prg.Eval(vars)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate CEL expression: %w", err)
	}

	result, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("CEL expression must return a boolean, got %T", out.Value())
	}

	return result, nil
}
