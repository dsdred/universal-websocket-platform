package runtimediagnosticstransport

import "encoding/json"

const outputLimit = 16 * 1024

type event struct {
	SchemaVersion uint8  `json:"schemaVersion"`
	Code          uint8  `json:"code"`
	Sequence      string `json:"sequence"`
}

type receipt struct {
	ReadAttempts  uint8  `json:"readAttempts"`
	Observed      uint8  `json:"observed"`
	SinkSucceeded uint8  `json:"sinkSucceeded"`
	SinkFailed    uint8  `json:"sinkFailed"`
	SinkPanicked  uint8  `json:"sinkPanicked"`
	Discarded     uint8  `json:"discarded"`
	Stop          string `json:"stop"`
}

type success struct {
	APIVersion uint8   `json:"apiVersion"`
	Count      uint8   `json:"count"`
	Events     []event `json:"events"`
	Receipt    receipt `json:"receipt"`
}

func project(data []byte, budget uint8) []byte {
	if len(data) > outputLimit || budget < 1 || budget > 64 {
		return nil
	}
	f, err := object(data, "apiVersion", "count", "events", "receipt")
	if err != nil {
		return nil
	}
	defer clearFields(f)
	rf, err := object(f["receipt"], "readAttempts", "observed", "sinkSucceeded", "sinkFailed", "sinkPanicked", "discarded", "stop")
	if err != nil {
		return nil
	}
	defer clearFields(rf)
	var events []json.RawMessage
	defer func() {
		for _, raw := range events {
			clear(raw)
		}
		clear(events)
	}()
	if json.Unmarshal(f["events"], &events) != nil || len(events) > 64 {
		return nil
	}
	for _, raw := range events {
		ef, err := object(raw, "schemaVersion", "code", "sequence")
		if err != nil {
			return nil
		}
		clearFields(ef)
	}
	var s success
	defer func() { clear(s.Events); s = success{} }()
	if json.Unmarshal(data, &s) != nil || s.APIVersion != 1 || int(s.Count) != len(s.Events) || s.Count > budget {
		return nil
	}
	r := s.Receipt
	if r.Observed != s.Count || r.SinkSucceeded != s.Count || r.SinkFailed != 0 || r.SinkPanicked != 0 || r.Discarded != 0 || r.ReadAttempts > budget {
		return nil
	}
	if !(r.Stop == "no_item" && int(r.ReadAttempts) == int(s.Count)+1 || r.Stop == "budget" && r.ReadAttempts == budget && s.Count == budget) {
		return nil
	}
	for _, e := range s.Events {
		if _, valid := decimal(e.Sequence); !valid || e.SchemaVersion != 1 || e.Code != 1 && e.Code != 2 {
			return nil
		}
	}
	body, err := json.Marshal(s)
	if err != nil || len(body) > outputLimit {
		clear(body)
		return nil
	}
	return body
}
