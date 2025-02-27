package common

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type DialPriority uint64

const (
	PriorityRequestedDial DialPriority = 1
	PriorityRandomDial    DialPriority = 10
)

const (
	DiscProto     = "/disc/0.1"
	IdentityProto = "/id/0.1"
)

var (
	ErrInvalidMultiaddr        = errors.New("invalid multiaddr")
	ErrCouldNotCreateMultiaddr = errors.New("could not create a multi address")

	ErrMultiaddrNoFoundIP   = errors.New("no IP address found")
	ErrMultiaddrContainsDNS = errors.New("multiaddr contains dns")
)

// DNSRegex is a regex string to match against a valid dns/dns4/dns6 addr
const DNSRegex = `^/?(dns)(4|6)?/[^-|^/][A-Za-z0-9-]([^-|^/]?)+([\\-\\.]{1}[a-z0-9]+)*\\.[A-Za-z]{2,}(/?)$`

func StringToAddrInfo(addr string) (*peer.AddrInfo, error) {
	addr0, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return nil, err
	}

	addr1, err := peer.AddrInfoFromP2pAddr(addr0)

	if err != nil {
		return nil, err
	}

	return addr1, nil
}

var (
	// Regex used for matching loopback addresses (IPv4, IPv6, DNS)
	// This regex will match:
	// /ip4/localhost/tcp/<port>
	// /ip4/127.0.0.1/tcp/<port>
	// /ip4/<any other loopback>/tcp/<port>
	// /ip6/<any loopback>/tcp/<port>
	// /dns/foobar.com/tcp/<port>
	loopbackRegex = regexp.MustCompile(
		//nolint:lll
		fmt.Sprintf(`^\/ip4\/127(?:\.[0-9]+){0,2}\.[0-9]+\/tcp\/\d+$|^\/ip4\/localhost\/tcp\/\d+$|^\/ip6\/(?:0*\:)*?:?0*1\/tcp\/\d+$|%s`, DNSRegex),
	)

	dnsRegex = "^/?(dns)(4|6)?/[^-|^/][A-Za-z0-9-]([^-|^/]?)+([\\-\\.]{1}[a-z0-9]+)*\\.[A-Za-z]{2,}(/?)$"
)

// AddrInfoToString converts an AddrInfo into a string representation that can be dialed from another node
func AddrInfoToString(addr *peer.AddrInfo) string {
	// Safety check
	if len(addr.Addrs) == 0 {
		panic("No dial addresses found")
	}

	dialAddress := addr.Addrs[0].String()

	// Try to see if a non loopback address is present in the list
	if len(addr.Addrs) > 1 && loopbackRegex.MatchString(dialAddress) {
		// Find an address that's not a loopback address
		for _, address := range addr.Addrs {
			if !loopbackRegex.MatchString(address.String()) {
				// Not a loopback address, dial address found
				dialAddress = address.String()

				break
			}
		}
	}

	// Format output and return
	return dialAddress + "/p2p/" + addr.ID.String()
}

// MultiAddrFromDNS constructs a multiAddr from the passed in DNS address and port combination
func MultiAddrFromDNS(addr string, port int) (multiaddr.Multiaddr, error) {
	var (
		version string
		domain  string
	)

	match, err := regexp.MatchString(
		dnsRegex,
		addr,
	)
	if err != nil || !match {
		return nil, ErrInvalidMultiaddr
	}

	s := strings.Trim(addr, "/")
	split := strings.Split(s, "/")

	if len(split) != 2 {
		return nil, ErrInvalidMultiaddr
	}

	switch split[0] {
	case "dns":
		version = "dns"
	case "dns4":
		version = "dns4"
	case "dns6":
		version = "dns6"
	default:
		return nil, ErrInvalidMultiaddr
	}

	domain = split[1]

	multiAddr, err := multiaddr.NewMultiaddr(
		fmt.Sprintf(
			"/%s/%s/tcp/%d",
			version,
			domain,
			port,
		),
	)

	if err != nil {
		return nil, ErrCouldNotCreateMultiaddr
	}

	return multiAddr, nil
}

// ParseMultiaddrIP parse a multiaddr into a net.IP
func ParseMultiaddrIP(ma multiaddr.Multiaddr) (net.IP, error) {
	_, err := ma.ValueForProtocol(multiaddr.P_DNS)
	if err == nil {
		return nil, ErrMultiaddrContainsDNS
	}

	ip, err := ma.ValueForProtocol(multiaddr.P_IP4)
	if err == nil {
		return net.ParseIP(ip), nil
	}

	ip, err = ma.ValueForProtocol(multiaddr.P_IP6)
	if err == nil {
		return net.ParseIP(ip), nil
	}

	return nil, ErrMultiaddrNoFoundIP
}
