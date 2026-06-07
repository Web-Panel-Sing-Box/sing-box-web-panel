package singbox

import (
	"fmt"
	"strconv"
	"strings"
)

// ClientOutboundTagPrefix is the prefix the generator stamps on every
// per-client direct outbound. Sing-box reports it back in
// /connections.chains[0], which is how the panel attributes traffic to
// individual clients without any log parsing.
const ClientOutboundTagPrefix = "user-"

// ClientOutboundTag builds the outbound tag for a client. The numeric ID keeps
// the tag in sing-box's safe charset regardless of how the client is named
// (Cyrillic, emoji, etc. only ever land in the route rule's `user` matcher).
func ClientOutboundTag(id int64) string {
	return fmt.Sprintf("%s%d", ClientOutboundTagPrefix, id)
}

// ParseClientOutboundTag is the inverse of ClientOutboundTag. Returns ok=false
// for any tag the panel did not generate (e.g. "direct", panel UI traffic).
func ParseClientOutboundTag(tag string) (int64, bool) {
	rest, ok := strings.CutPrefix(tag, ClientOutboundTagPrefix)
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseInt(rest, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}
