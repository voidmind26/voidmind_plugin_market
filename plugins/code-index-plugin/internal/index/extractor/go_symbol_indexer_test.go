package extractor

import (
	"testing"

	"code-index-plugin/internal/index/model"
)

func TestExtractGoSymbolsReturnsFunctionsAndTypes(t *testing.T) {
	content := []byte(`package svc

// PaymentService coordinates payment callbacks.
type PaymentService struct{}

type CallbackHandler interface {
	HandleCallback(id string) error
}

type PaymentAlias = PaymentService

type ValidatorA struct{}
type ValidatorB struct{}

// HandleCallback processes provider callbacks.
func (s *PaymentService) HandleCallback(id string) error {
	return nil
}

func (ValidatorA) Validate() error { return nil }
func (*ValidatorB) Validate() error { return nil }

func helper() {}
`)

	symbols, err := ExtractGoSymbols("service/payment.go", content)
	if err != nil {
		t.Fatalf("ExtractGoSymbols returned error: %v", err)
	}
	if len(symbols) != 9 {
		t.Fatalf("expected 9 symbols, got %d (%v)", len(symbols), symbols)
	}

	assertGoSymbol(t, symbols[0], "PaymentService", "type", "service/payment.go", 4, 4)
	assertContains(t, symbols[0].Keywords, "paymentservice")
	assertContains(t, symbols[0].Keywords, "struct")
	if symbols[0].Summary == "" {
		t.Fatalf("expected type summary to be non-empty")
	}

	assertGoSymbol(t, symbols[1], "CallbackHandler", "type", "service/payment.go", 6, 8)
	assertContains(t, symbols[1].Keywords, "interface")

	assertGoSymbol(t, symbols[2], "PaymentAlias", "type", "service/payment.go", 10, 10)
	assertContains(t, symbols[2].Keywords, "alias")

	assertGoSymbol(t, symbols[3], "ValidatorA", "type", "service/payment.go", 12, 12)
	assertGoSymbol(t, symbols[4], "ValidatorB", "type", "service/payment.go", 13, 13)

	assertGoSymbol(t, symbols[5], "HandleCallback", "method", "service/payment.go", 16, 18)
	assertContains(t, symbols[5].Keywords, "paymentservice")
	if symbols[5].Summary == "" {
		t.Fatalf("expected method summary to be non-empty")
	}
	if symbols[5].Receiver != "PaymentService" {
		t.Fatalf("expected receiver PaymentService, got %q", symbols[5].Receiver)
	}

	assertGoSymbol(t, symbols[6], "Validate", "method", "service/payment.go", 20, 20)
	if symbols[6].Receiver != "ValidatorA" {
		t.Fatalf("expected receiver ValidatorA, got %q", symbols[6].Receiver)
	}

	assertGoSymbol(t, symbols[7], "Validate", "method", "service/payment.go", 21, 21)
	if symbols[7].Receiver != "ValidatorB" {
		t.Fatalf("expected receiver ValidatorB, got %q", symbols[7].Receiver)
	}

	if symbols[6].Receiver == symbols[7].Receiver {
		t.Fatalf("expected same-name methods to be distinguishable by receiver, got %q and %q", symbols[6].Receiver, symbols[7].Receiver)
	}

	assertGoSymbol(t, symbols[8], "helper", "func", "service/payment.go", 23, 23)
}

func TestSplitTermsSplitsCamelCaseIdentifiers(t *testing.T) {
	terms := splitTerms("GetLoginUserInfo user_id HTTPServer")
	assertContains(t, terms, "getloginuserinfo")
	assertContains(t, terms, "get")
	assertContains(t, terms, "login")
	assertContains(t, terms, "user")
	assertContains(t, terms, "info")
	assertContains(t, terms, "httpserver")
	assertContains(t, terms, "http")
	assertContains(t, terms, "server")
}

func assertGoSymbol(t *testing.T, got model.SymbolRecord, wantName, wantType, wantPath string, wantStartLine, wantEndLine int) {
	t.Helper()
	if got.SymbolName != wantName {
		t.Fatalf("expected symbol name %q, got %q", wantName, got.SymbolName)
	}
	if got.SymbolType != wantType {
		t.Fatalf("expected symbol type %q, got %q", wantType, got.SymbolType)
	}
	if got.Path != wantPath {
		t.Fatalf("expected path %q, got %q", wantPath, got.Path)
	}
	if got.StartLine != wantStartLine || got.EndLine != wantEndLine {
		t.Fatalf("expected lines %d-%d, got %d-%d", wantStartLine, wantEndLine, got.StartLine, got.EndLine)
	}
}
