package schema

import (
	"context"

	"github.com/dagger/dagger/core"
	"github.com/dagger/dagger/dagql"
)

type llmSchema struct {
	srv *dagql.Server
}

var _ SchemaResolvers = &llmSchema{}

func (s llmSchema) Install() {
	dagql.Fields[*core.Query]{
		dagql.Func("llm", s.llm).
			Doc(`Initialize a Large Language Model (LLM)`).
			ArgDoc("model", "Model to use"),
	}.Install(s.srv)
	llmType := dagql.Fields[*core.Llm]{
		dagql.Func("model", s.model).
			Doc("return the model used by the llm"),
		dagql.Func("history", s.history).
			Doc("return the llm message history"),
		dagql.Func("lastReply", s.lastReply).
			Doc("return the last llm reply from the history"),
		dagql.Func("withPrompt", s.withPrompt).
			Doc("append a prompt to the llm context").
			ArgDoc("prompt", "The prompt to send"),
		dagql.Func("withPromptFile", s.withPromptFile).
			Doc("append the contents of a file to the llm context").
			ArgDoc("file", "The file to read the prompt from"),
		dagql.Func("withPromptVar", s.withPromptVar).
			Doc("set a variable for expansion in the prompt").
			ArgDoc("name", "The name of the variable").
			ArgDoc("value", "The value of the variable"),
		dagql.Func("loop", s.loop).
			Doc("send the context to the LLM endpoint, process replies and tool calls; continue in a loop").
			ArgDoc("maxLoops", "The maximum number of loops to allow."),
		dagql.Func("tools", s.tools).
			Doc("print documentation for available tools"),
	}
	llmType.Install(s.srv)
	s.srv.SetMiddleware(core.LlmMiddleware{Server: s.srv})
}

func (s *llmSchema) model(ctx context.Context, llm *core.Llm, args struct{}) (dagql.String, error) {
	return dagql.NewString(llm.Config.Model), nil
}

func (s *llmSchema) lastReply(ctx context.Context, llm *core.Llm, args struct{}) (dagql.String, error) {
	reply, err := llm.LastReply()
	if err != nil {
		return "", err
	}
	return dagql.NewString(reply), nil
}
func (s *llmSchema) withPrompt(ctx context.Context, llm *core.Llm, args struct {
	Prompt string
}) (*core.Llm, error) {
	return llm.WithPrompt(ctx, args.Prompt, s.srv)
}

func (s *llmSchema) withPromptVar(ctx context.Context, llm *core.Llm, args struct {
	Name  dagql.String
	Value dagql.String
}) (*core.Llm, error) {
	return llm.WithPromptVar(args.Name.String(), args.Value.String()), nil
}

func (s *llmSchema) withPromptFile(ctx context.Context, llm *core.Llm, args struct {
	File core.FileID
}) (*core.Llm, error) {
	file, err := args.File.Load(ctx, s.srv)
	if err != nil {
		return nil, err
	}
	return llm.WithPromptFile(ctx, file.Self, s.srv)
}

func (s *llmSchema) loop(ctx context.Context, llm *core.Llm, args struct {
	MaxLoops dagql.Optional[dagql.Int]
}) (*core.Llm, error) {
	maxLoops := args.MaxLoops.GetOr(0).Int()
	return llm.Loop(ctx, maxLoops, s.srv)
}

func (s *llmSchema) llm(ctx context.Context, parent *core.Query, args struct {
	Model dagql.Optional[dagql.String]
}) (*core.Llm, error) {
	var model string
	if args.Model.Valid {
		model = args.Model.Value.String()
	}
	return core.NewLlm(ctx, parent, s.srv, model)
}

func (s *llmSchema) history(ctx context.Context, llm *core.Llm, _ struct{}) (dagql.Array[dagql.String], error) {
	history, err := llm.History()
	if err != nil {
		return nil, err
	}
	return dagql.NewStringArray(history...), nil
}

func (s *llmSchema) tools(ctx context.Context, llm *core.Llm, _ struct{}) (dagql.String, error) {
	doc, err := llm.ToolsDoc(ctx, s.srv)
	return dagql.NewString(doc), err
}
