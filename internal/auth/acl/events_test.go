package acl

import "testing"

func TestEventsChannelAuditPermissions(t *testing.T) {
	perms := Permissions{ActionSubscribe}

	if !Events.AllowAll(ChannelAudit, RoleAdmin, perms) {
		t.Fatalf("expected admin to subscribe to audit events")
	}

	if !Events.AllowAll(ChannelAudit, RolePortal, perms) {
		t.Fatalf("expected portal to subscribe to audit events")
	}

	if Events.AllowAll(ChannelAudit, RoleUser, perms) {
		t.Fatalf("expected regular users to be denied audit event subscriptions")
	}
}
