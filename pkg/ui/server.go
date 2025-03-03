package ui

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/tstromberg/confuSSHion/pkg/history"
	"k8s.io/klog/v2"
)

// Server provides a web interface for viewing SSH session histories.
type Server struct {
	store     *history.Store
	port      int
	templates *template.Template
}

// NewServer creates a new UI server.
func NewServer(store *history.Store, port int) (*Server, error) {
	// Initialize templates
	tmpl := template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	})

	// Parse each template
	for name, content := range map[string]string{
		"index.html":   indexTemplate,
		"session.html": sessionTemplate,
	} {
		if _, err := tmpl.New(name).Parse(content); err != nil {
			return nil, fmt.Errorf("parse template %s: %w", name, err)
		}
	}

	return &Server{
		store:     store,
		port:      port,
		templates: tmpl,
	}, nil
}

// Start begins serving the web UI.
func (s *Server) Start() error {
	if s.store == nil {
		return fmt.Errorf("history store is nil")
	}

	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/session/", s.handleSessionDetail)

	addr := fmt.Sprintf(":%d", s.port)
	klog.Infof("Starting web UI on http://localhost%s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleIndex shows a list of all sessions.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	sessionIDs, err := s.store.ListSessions()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing sessions: %v", err), http.StatusInternalServerError)
		return
	}

	sessions := []*history.Session{}
	for _, sid := range sessionIDs {
		session, err := s.store.GetSession(sid)
		if err != nil {
			klog.Errorf("Failed to get session %s: %v", sid, err)
			continue
		}
		sessions = append(sessions, session)
	}

	// Sort sessions by descending time order (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "index.html", struct {
		Title    string
		Time     time.Time
		Sessions []*history.Session
	}{
		Title:    "confuSSHion Sessions",
		Time:     time.Now(),
		Sessions: sessions,
	}); err != nil {
		klog.Errorf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// handleSessionDetail shows details of a specific session.
func (s *Server) handleSessionDetail(w http.ResponseWriter, r *http.Request) {
	sid := strings.TrimPrefix(r.URL.Path, "/session/")
	if sid == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	session, err := s.store.GetSession(sid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Session not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "session.html", struct {
		Title   string
		Time    time.Time
		Session *history.Session
	}{
		Title:   fmt.Sprintf("Session %s", sid),
		Time:    time.Now(),
		Session: session,
	}); err != nil {
		klog.Errorf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

const indexTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>SSH Sessions</h1>

    {{if .Sessions}}
        <ul>
        {{range .Sessions}}
            <li>
                <a href="/session/{{.SID}}">{{.SID}}</a> -
                User: {{.UserInfo.RemoteUser}} -
                IP: {{.UserInfo.RemoteAddr}} -
                Started: {{.StartTime.Format "2006-01-02 15:04:05"}} -
                Duration: {{.EndTime.Sub .StartTime}} -
                Commands: {{len .Log}}
            </li>
        {{end}}
        </ul>
    {{else}}
        <p>No sessions recorded yet.</p>
    {{end}}

    <p><small>Generated at {{.Time.Format "2006-01-02 15:04:05"}}</small></p>
</body>
</html>
`

// sessionTemplate is the HTML template for the session detail page.
const sessionTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>SSH Session: {{.Session.SID}}</h1>

    <p><a href="/">Back to sessions list</a></p>

    <h2>Session Info</h2>
    <pre>
User: {{.Session.UserInfo.RemoteUser}}
{{if .Session.UserInfo.AuthUser}}Auth User: {{.Session.UserInfo.AuthUser}}{{end}}
Remote Address: {{.Session.UserInfo.RemoteAddr}}
Started: {{.Session.StartTime.Format "2006-01-02 15:04:05"}}
Ended: {{.Session.EndTime.Format "2006-01-02 15:04:05"}}
Duration: {{.Session.EndTime.Sub .Session.StartTime}}
OS: {{.Session.NodeConfig.OS}}
{{if .Session.NodeConfig.Hostname}}Hostname: {{.Session.NodeConfig.Hostname}}{{end}}
    </pre>

    <h2>Session Log</h2>
    <pre>
{{range $i, $entry := .Session.Log}}======== Command {{add $i 1}} - {{$entry.Timestamp.Format "2006-01-02 15:04:05"}} ========
{{if eq $entry.Input "LOGIN"}}--- Login Banner ---{{else}}$ {{$entry.Input}}{{end}}
{{$entry.Output}}

{{end}}
    </pre>

    <p><small>Generated at {{.Time.Format "2006-01-02 15:04:05"}}</small></p>
</body>
</html>`
