package discovery

import (
	"strconv"
	"strings"
)

// ProtocolVersion is the on-wire protocol identifier advertised in TXT records
// and checked during the handshake.
const ProtocolVersion = "v1"

// ServiceType is the DNS-SD service type git-lan registers and browses.
const ServiceType = "_gitlan._tcp"

// Domain is the mDNS domain.
const Domain = "local."

// Advertisement is the metadata a host broadcasts about the repo it is sharing.
// It is serialized into mDNS TXT key=value records, which are PLAINTEXT by
// design. Never put keys, tokens, or anything sensitive here.
type Advertisement struct {
	Repo     string
	Branch   string
	Head     string
	Modified int
	Session  string
	Locked   bool
	Presence Presence
}

const (
	maxBranchLen = 20
	maxHeadLen   = 7
)

// TXT renders the advertisement as DNS-SD TXT records ("key=value" strings).
func (a Advertisement) TXT() []string {
	branch := a.Branch
	if len(branch) > maxBranchLen {
		branch = branch[:maxBranchLen]
	}
	head := a.Head
	if len(head) > maxHeadLen {
		head = head[:maxHeadLen]
	}
	locked := "0"
	if a.Locked {
		locked = "1"
	}
	presence := a.Presence
	if presence == "" {
		presence = PresenceOnline
	}
	return []string{
		"v=" + ProtocolVersion,
		"repo=" + a.Repo,
		"branch=" + branch,
		"head=" + head,
		"mod=" + strconv.Itoa(a.Modified),
		"session=" + a.Session,
		"lock=" + locked,
		"presence=" + string(presence),
	}
}

// ParseTXT decodes DNS-SD TXT records back into the metadata fields of a peer.
func ParseTXT(records []string) Advertisement {
	kv := make(map[string]string, len(records))
	for _, rec := range records {
		k, v, ok := strings.Cut(rec, "=")
		if !ok {
			continue
		}
		kv[k] = v
	}
	mod, _ := strconv.Atoi(kv["mod"])
	ad := Advertisement{
		Repo:     kv["repo"],
		Branch:   kv["branch"],
		Head:     kv["head"],
		Modified: mod,
		Session:  kv["session"],
		Locked:   kv["lock"] == "1",
		Presence: Presence(kv["presence"]),
	}
	if ad.Presence == "" {
		ad.Presence = PresenceOnline
	}
	return ad
}

// Protocol returns the protocol version from a set of TXT records, or "" if
// none is present.
func protocolOf(records []string) string {
	for _, rec := range records {
		if k, v, ok := strings.Cut(rec, "="); ok && k == "v" {
			return v
		}
	}
	return ""
}
