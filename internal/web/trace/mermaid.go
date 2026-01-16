package trace

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
)

// MermaidBuilder builds Mermaid diagram syntax
type MermaidBuilder struct {
	sb    strings.Builder
	nodes map[string]string // label -> nodeID
}

// NewMermaidBuilder creates a new builder for flowchart diagrams
func NewMermaidBuilder() *MermaidBuilder {
	b := &MermaidBuilder{
		nodes: make(map[string]string, 64),
	}
	b.sb.WriteString("flowchart TD\n")
	return b
}

// Node adds or retrieves a node by label
func (b *MermaidBuilder) Node(label string) string {
	if id, ok := b.nodes[label]; ok {
		return id
	}
	id := "n" + hashID(label)
	b.nodes[label] = id

	b.sb.WriteString("  ")
	b.sb.WriteString(id)
	b.sb.WriteString("[\"")
	b.sb.WriteString(escapeMermaidLabel(label))
	b.sb.WriteString("\"]\n")

	return id
}

// Edge adds an edge between two nodes
func (b *MermaidBuilder) Edge(fromID, toID, label string) {
	b.sb.WriteString("  ")
	b.sb.WriteString(fromID)
	b.sb.WriteString(" -->")

	if label != "" {
		b.sb.WriteString("|\"")
		b.sb.WriteString(escapeMermaidLabel(label))
		b.sb.WriteString("\"|")
	}

	b.sb.WriteString(" ")
	b.sb.WriteString(toID)
	b.sb.WriteString("\n")
}

// String returns the built diagram
func (b *MermaidBuilder) String() string { return b.sb.String() }

// BuildFlowchart creates a flowchart from trace events
func BuildFlowchart(events []Event) string {
	sort.Slice(events, func(i, j int) bool {
		if events[i].TS.Equal(events[j].TS) {
			return i < j
		}
		return events[i].TS.Before(events[j].TS)
	})

	mb := NewMermaidBuilder()

	for _, ev := range events {
		localID := mb.Node("local: " + ev.Local)
		remoteID := mb.Node("remote: " + ev.Remote)

		edgeLabel := fmt.Sprintf(
			"%s len=%d cid=%d",
			ev.TS.Format("15:04:05.000"),
			ev.Len,
			ev.ClientID,
		)

		switch ev.Dir {
		case In:
			mb.Edge(remoteID, localID, "IN "+edgeLabel)
		case Out:
			mb.Edge(localID, remoteID, "OUT "+edgeLabel)
		default:
			mb.Edge(remoteID, localID, string(ev.Dir)+" "+edgeLabel)
		}
	}

	return mb.String()
}

// BuildSequenceDiagram creates a sequence diagram from trace events
func BuildSequenceDiagram(events []Event) string {
	sort.Slice(events, func(i, j int) bool {
		if events[i].TS.Equal(events[j].TS) {
			return i < j
		}
		return events[i].TS.Before(events[j].TS)
	})

	var sb strings.Builder
	sb.WriteString("sequenceDiagram\n")

	serverLabel := "Server"
	if len(events) > 0 && events[0].Local != "" {
		serverLabel = "Server " + events[0].Local
	}
	sb.WriteString("  participant S as \"")
	sb.WriteString(escapeMermaidLabel(serverLabel))
	sb.WriteString("\"\n")

	clientPID := map[string]string{}

	getClientPID := func(remote string, cid int) string {
		if remote == "" {
			remote = "unknown"
		}
		if pid, ok := clientPID[remote]; ok {
			return pid
		}
		pid := "C" + hashID(remote)
		clientPID[remote] = pid

		lbl := fmt.Sprintf("Client cid=%d\\n%s", cid, remote)
		sb.WriteString("  participant ")
		sb.WriteString(pid)
		sb.WriteString(" as \"")
		sb.WriteString(escapeMermaidLabel(lbl))
		sb.WriteString("\"\n")

		return pid
	}

	for _, ev := range events {
		c := getClientPID(ev.Remote, ev.ClientID)

		label := fmt.Sprintf("%s len=%d cid=%d",
			ev.TS.Format("15:04:05.000"),
			ev.Len,
			ev.ClientID,
		)

		switch ev.Dir {
		case In:
			sb.WriteString("  ")
			sb.WriteString(c)
			sb.WriteString("->>S: ")
			sb.WriteString(escapeMermaidLabel("SEND " + label))
			sb.WriteString("\n")

		case Out:
			sb.WriteString("  S->>")
			sb.WriteString(c)
			sb.WriteString(": ")
			sb.WriteString(escapeMermaidLabel("SEND " + label))
			sb.WriteString("\n")

		default:
			sb.WriteString("  Note over S,")
			sb.WriteString(c)
			sb.WriteString(": ")
			sb.WriteString(escapeMermaidLabel("dir=" + string(ev.Dir) + " " + label))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func hashID(s string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum32())
}

func escapeMermaidLabel(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return s
}
