package event

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// receiveAudit publishes an audit event and returns the fields it was published with.
func receiveAudit(t *testing.T, level logrus.Level, ev []string, args ...any) Data {
	t.Helper()

	s := Subscribe(string(acl.ChannelAudit) + ".log.*")
	defer Unsubscribe(s)

	Audit(level, ev, args...)

	select {
	case msg := <-s.Receiver:
		return msg.Fields
	case <-time.After(time.Second):
		t.Fatal("no audit event was published")
		return nil
	}
}

func TestAudit(t *testing.T) {
	t.Run("ReportsThePeer", func(t *testing.T) {
		fields := receiveAudit(t, logrus.WarnLevel, []string{"203.0.113.55", "session %s", "login", status.Denied}, "sessxx78mbl7")

		assert.Equal(t, "203.0.113.55", fields["ip"])
		assert.Equal(t, "warning", fields["level"])
		assert.Contains(t, fields["message"], "sessxx78mbl7")
	})
	t.Run("ReportsAnIPv6Peer", func(t *testing.T) {
		fields := receiveAudit(t, logrus.WarnLevel, []string{"2001:db8::42", "session %s", "login", status.Denied}, "sessxx78mbl7")

		assert.Equal(t, "2001:db8::42", fields["ip"])
	})
	t.Run("CanonicalizesThePeer", func(t *testing.T) {
		fields := receiveAudit(t, logrus.WarnLevel, []string{"2001:0db8:0000:0000:0000:0000:0000:0042", "session", status.Denied})

		assert.Equal(t, "2001:db8::42", fields["ip"])
	})
	t.Run("WithoutAPeer", func(t *testing.T) {
		fields := receiveAudit(t, logrus.WarnLevel, []string{"cli", "users", "reset", status.Succeeded})

		assert.Equal(t, "", fields["ip"])
		assert.Contains(t, fields["message"], "reset")
	})
	t.Run("OnlyTheFirstSegment", func(t *testing.T) {
		fields := receiveAudit(t, logrus.WarnLevel, []string{"cli", "login as %s", status.Denied}, "backup-198.51.100.66")

		assert.Equal(t, "", fields["ip"])
		assert.Contains(t, fields["message"], "198.51.100.66")
	})
	t.Run("NoEvent", func(t *testing.T) {
		s := Subscribe(string(acl.ChannelAudit) + ".log.*")
		defer Unsubscribe(s)

		Audit(logrus.WarnLevel, nil)

		select {
		case <-s.Receiver:
			t.Fatal("an empty event must not be published")
		case <-time.After(100 * time.Millisecond):
		}
	})
}

func TestAuditIP(t *testing.T) {
	t.Run("IPv4", func(t *testing.T) {
		assert.Equal(t, "203.0.113.55", AuditIP([]string{"203.0.113.55", "session"}))
	})
	t.Run("IPv6", func(t *testing.T) {
		assert.Equal(t, "2001:db8::42", AuditIP([]string{"2001:0db8::0042", "session"}))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", AuditIP(nil))
		assert.Equal(t, "", AuditIP([]string{}))
		assert.Equal(t, "", AuditIP([]string{""}))
	})
	t.Run("NotAnAddress", func(t *testing.T) {
		assert.Equal(t, "", AuditIP([]string{"cli", "203.0.113.55"}))
		assert.Equal(t, "", AuditIP([]string{"backup-198.51.100.66"}))
		assert.Equal(t, "", AuditIP([]string{"198.51.100.66:1234"}))
		assert.Equal(t, "", AuditIP([]string{"user@example.com"}))
	})
}

func TestAuditLevels(t *testing.T) {
	t.Run("TraceIsNotPublished", func(t *testing.T) {
		s := Subscribe(string(acl.ChannelAudit) + ".log.*")
		defer Unsubscribe(s)

		AuditTrace([]string{"203.0.113.55", "session", status.Denied})
		AuditDebug([]string{"203.0.113.55", "session", status.Denied})

		select {
		case msg := <-s.Receiver:
			t.Fatalf("an event below info was published: %v", msg.Fields)
		case <-time.After(100 * time.Millisecond):
		}
	})
	t.Run("InfoAndAboveArePublished", func(t *testing.T) {
		for _, record := range []func(ev []string, args ...any){AuditInfo, AuditWarn, AuditErr} {
			s := Subscribe(string(acl.ChannelAudit) + ".log.*")

			record([]string{"203.0.113.55", "session", status.Denied})

			select {
			case msg := <-s.Receiver:
				require.Equal(t, "203.0.113.55", msg.Fields["ip"])
			case <-time.After(time.Second):
				t.Fatal("no audit event was published")
			}

			Unsubscribe(s)
		}
	})
}
