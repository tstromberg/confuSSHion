package holodeck

import (
	"strings"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
)

// PublicKeyHandler validates SSH public keys
func (h Holodeck) PublicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	klog.Info("Public key handler invoked")
	txt := strings.TrimSpace(string(gossh.MarshalAuthorizedKey(key)))
	klog.Infof("Session %s key provided: %s", ctx.SessionID(), txt)

	user := h.auth.ValidKey(txt)
	if user == nil {
		klog.Errorf("Unable to validate key: %s", key)
		return false
	}

	klog.Infof("Authenticated user: %s", user)
	h.authUser[ctx.SessionID()] = user
	return true
}
