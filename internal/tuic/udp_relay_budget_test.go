package tuic

import (
	"context"
	"errors"
	"testing"
)

func newRelayBudgetRegistry(t *testing.T) *udpAssociationRegistry {
	t.Helper()
	return newUdpAssociationRegistry(maxUdpRelayPacketSize)
}

// A client already holding the per-account relay budget must be refused before
// a SOCKS tunnel is opened, so one credential cannot exhaust the host's
// ephemeral ports across its connections.
func TestEnsureRelayRejectsClientPastRelayBudget(t *testing.T) {
	registry := newRelayBudgetRegistry(t)
	user := newUser([16]byte{2}, TuicClientSettings{Email: "relay-hog@example.com"})
	user.udpRelays.Store(maxUserUdpRelays)

	association := &udpAssociation{responseTransport: packetTransportDatagram}
	registry.associations[7] = association

	relay := &SocksRelay{Addr: "127.0.0.1:1", Password: "unused"}
	_, _, err := registry.ensureRelay(context.Background(), 7, association, user, relay)
	if !errors.Is(err, errUserUdpRelayBudget) {
		t.Fatalf("ensureRelay error = %v, want errUserUdpRelayBudget", err)
	}
	if got := user.udpRelays.Load(); got != maxUserUdpRelays {
		t.Fatalf("relay counter = %d, want it unchanged at %d", got, maxUserUdpRelays)
	}
	if association.relay != nil {
		t.Fatal("a refused association must not keep a relay")
	}
}

// A failed dial must hand the reservation back, otherwise repeated failures
// would drain the budget without any relay actually existing.
func TestEnsureRelayReleasesBudgetWhenDialFails(t *testing.T) {
	registry := newRelayBudgetRegistry(t)
	user := newUser([16]byte{3}, TuicClientSettings{Email: "dial-fail@example.com"})

	association := &udpAssociation{responseTransport: packetTransportDatagram}
	registry.associations[9] = association

	relay := &SocksRelay{Addr: "127.0.0.1:1", Password: "unused"}
	_, _, err := registry.ensureRelay(context.Background(), 9, association, user, relay)
	if err == nil {
		t.Fatal("expected the dial to fail against a closed port")
	}
	if errors.Is(err, errUserUdpRelayBudget) {
		t.Fatalf("dial failure must not surface as a budget error: %v", err)
	}
	if got := user.udpRelays.Load(); got != 0 {
		t.Fatalf("relay counter = %d after a failed dial, want 0", got)
	}
}

// Closing a relay through any registry path must return its slot exactly once.
func TestRelayBudgetReturnsOnCloseAndIsIdempotent(t *testing.T) {
	registry := newRelayBudgetRegistry(t)
	user := newUser([16]byte{4}, TuicClientSettings{Email: "close@example.com"})

	association := &udpAssociation{responseTransport: packetTransportDatagram}
	registry.associations[11] = association

	released := 0
	session := &SocksUDPSession{onClose: func() {
		released++
		user.udpRelays.Add(-1)
	}}
	association.relay = &udpRelaySession{relay: session, responseTransport: packetTransportDatagram}
	user.udpRelays.Store(1)

	_ = session.Close()
	_ = session.Close()

	if released != 1 {
		t.Fatalf("onClose ran %d times, want exactly 1", released)
	}
	if got := user.udpRelays.Load(); got != 0 {
		t.Fatalf("relay counter = %d after close, want 0", got)
	}

	registry.release(11, association)
	if got := user.udpRelays.Load(); got != 0 {
		t.Fatalf("release after close double-counted: counter = %d", got)
	}
}
