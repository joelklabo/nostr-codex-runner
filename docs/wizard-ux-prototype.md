# Wizard UX Prototype Findings – Issue 3oa.19

Objective: pick a CLI input library that supports masked secrets, validation, and simple branching without heavy TUI boilerplate.

Options reviewed

- **AlecAivazis/survey** (recommended)
  - Pros: mature, supports password/confirm prompts, validation funcs, select/multiselect, nicely handles Ctrl+C. Minimal code to mask secrets.
  - Cons: adds dependency; limited styling (ok for MVP).

- **promptui**
  - Pros: lightweight, familiar API.
  - Cons: weaker password handling (needs manual masking), less ergonomic validation, poorer Windows support.

- **charmbracelet/bubbletea**
  - Pros: rich TUI, great for fancy flows.
  - Cons: overkill for MVP; more code; heavier deps.

- **bufio + fmt (hand-rolled)**
  - Pros: zero deps.
  - Cons: no masked input; must re-implement validation and select UI; higher bug risk.

Prototype sketch (survey)

```go
import "github.com/AlecAivazis/survey/v2"

func askPrivateKey() (string, error) {
    prompt := &survey.Password{Message: "Enter nostr private key (hex):"}
    var key string
    err := survey.AskOne(prompt, &key, survey.WithValidator(func(ans interface{}) error {
        s := ans.(string)
        if len(s) != 64 { return fmt.Errorf("expected 64 hex chars") }
        if _, err := hex.DecodeString(s); err != nil { return fmt.Errorf("not hex") }
        return nil
    }))
    return key, err
}
```

- Masked input verified; validation triggers re-prompt with error message.

- Ctrl+C returns `ErrInterrupt` which we can wrap.

Branching pattern

- Use `survey.Select` for transport/agent choices; store selections and skip irrelevant prompts.

- For yes/no confirmations, use `survey.Confirm` to gate shell action enabling and file overwrite.

Decision

- Adopt **survey** for wizard MVP.

- Keep UI uncolored; rely on short prompt text and defaults.

Next steps

- Add dependency in go.mod when implementing wizard (issue 3oa.20).

- Build minimal POC command under `internal/wizard` using survey; ensure tests simulate stdin (issue 3oa.21).
