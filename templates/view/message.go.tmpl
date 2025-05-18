package view

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
)

const (
	retarget = "#messages"
	reswap   = "beforeend show:body:top"
)

var (
	messageTemplate         = template.Must(template.New("message.html").ParseFS(templatesFS, "templates/components/message.html"))
	messagesContextTemplate = template.Must(template.New("messages-context.html").ParseFS(templatesFS,
		"templates/components/messages-context.html", "templates/components/message.html"))
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
		if r.Method == http.MethodDelete {
			err := messagesContextTemplate.Execute(w, m)
			if err != nil {
				slog.Error("render messages-context", "err", err)
			}
			return err
		}
		w.Header().Set("HX-Retarget", retarget)
		w.Header().Set("HX-Reswap", reswap)
	}

	err := messageTemplate.Execute(w, m)
	if err != nil {
		slog.Error("render message", "err", err)
	}
	return err
}
