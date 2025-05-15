package htmx

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
)

const messagesSelector = "#messages"

var (
	//go:embed message.html
	messageHTML string

	messageTemplate = template.Must(template.New("message").Parse(messageHTML))
)

type MessageType string

const (
	TypeInfo    = MessageType("info")
	TypeSuccess = MessageType("success")
	TypeError   = MessageType("error")
	TypeWarning = MessageType("warning")
)

func (t MessageType) Icon() string {
	switch t {
	case TypeInfo:
		return "info-fill"
	case TypeSuccess:
		return "check-circle-fill"
	default:
		return "exclamation-triangle-fill"
	}
}

func (t MessageType) Class() string {
	switch t {
	case TypeInfo:
		return "primary"
	case TypeSuccess:
		return "success"
	case TypeError:
		return "danger"
	default:
		return "warning"
	}
}

type Message struct {
	Code int         `json:"code"`
	Text string      `json:"text"`
	Type MessageType `json:"type"`
}

func NewMessage(code int, text string, typ MessageType) Message {
	return Message{Code: code,
		Text: text,
		Type: typ}
}

func ErrorMessage(code int, text string) Message {
	return NewMessage(code, text, TypeError)
}

func InfoMessage(code int, text string) Message {
	return NewMessage(code, text, TypeInfo)
}

func SuccessMessage(code int, text string) Message {
	return NewMessage(code, text, TypeSuccess)
}

func WarningMessage(code int, text string) Message {
	return NewMessage(code, text, TypeWarning)
}

func (m Message) Render(w http.ResponseWriter, r *http.Request) error {
	if strings.Contains(r.Header.Get("accept"), "application/json") {
		w.WriteHeader(m.Code)
		return json.NewEncoder(w).Encode(m)
	}

	if HXRequest(r) {
		w.Header().Set("HX-Retarget", messagesSelector)
		w.Header().Set("HX-Reswap", "beforeend")
	}

	err := messageTemplate.Execute(w, m)
	if err != nil {
		slog.Error("render message", "err", err)
	}
	return err
}

func Info(w http.ResponseWriter, r *http.Request, code int, text string) error {
	return NewMessage(code, text, TypeInfo).Render(w, r)
}

func Success(w http.ResponseWriter, r *http.Request, code int, text string) error {
	return NewMessage(code, text, TypeSuccess).Render(w, r)
}

func Error(w http.ResponseWriter, r *http.Request, code int, text string) error {
	return NewMessage(code, text, TypeError).Render(w, r)
}

func Warning(w http.ResponseWriter, r *http.Request, code int, text string) error {
	return NewMessage(code, text, TypeWarning).Render(w, r)
}
