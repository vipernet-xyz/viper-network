package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"

	host "github.com/vipernet-xyz/viper-network/modules/core/24-host"
)

const (

	// PlanByte specifies the Byte under which a pending upgrade plan is stored in the store
	PlanByte = 0x0

	// SubModuleName defines the IBC client name
	SubModuleName string = "client"

	// RouterKey is the message route for IBC client
	RouterKey string = SubModuleName

	// QuerierRoute is the querier route for IBC client
	QuerierRoute string = SubModuleName

	// KeyNextClientSequence is the key used to store the next client sequence in
	// the keeper.
	KeyNextClientSequence = "nextClientSequence"

	// DoneByte is a prefix to look up completed upgrade plan by name
	DoneByte = 0x1

	// VersionMapByte is a prefix to look up module names (key) and versions (value)
	VersionMapByte = 0x2

	// ProtocolVersionByte is a prefix to look up Protocol Version
	ProtocolVersionByte = 0x3

	// KeyUpgradedIBCState is the key under which upgraded ibc state is stored in the upgrade store
	KeyUpgradedIBCState = "upgradedIBCState"

	// KeyUpgradedClient is the sub-key under which upgraded client state will be stored
	KeyUpgradedClient = "upgradedClient"

	// KeyUpgradedConsState is the sub-key under which upgraded consensus state will be stored
	KeyUpgradedConsState = "upgradedConsState"
)

// FormatClientIdentifier returns the client identifier with the sequence appended.
// This is a SDK specific format not enforced by IBC protocol.
func FormatClientIdentifier(clientType string, sequence uint64) string {
	return fmt.Sprintf("%s-%d", clientType, sequence)
}

// IsClientIDFormat checks if a clientID is in the format required on the SDK for
// parsing client identifiers. The client identifier must be in the form: `{client-type}-{N}
// which per the specification only permits ASCII for the {client-type} segment and
// 1 to 20 digits for the {N} segment.
// `([\w-]+\w)?` allows for a letter or hyphen, with the {client-type} starting with a letter
// and ending with a letter, i.e. `letter+(letter|hypen+letter)?`.
var IsClientIDFormat = regexp.MustCompile(`^\w+([\w-]+\w)?-[0-9]{1,20}$`).MatchString

// IsValidClientID checks if the clientID is valid and can be parsed into the client
// identifier format.
func IsValidClientID(clientID string) bool {
	_, _, err := ParseClientIdentifier(clientID)
	return err == nil
}

// ParseClientIdentifier parses the client type and sequence from the client identifier.
func ParseClientIdentifier(clientID string) (string, uint64, error) {
	if !IsClientIDFormat(clientID) {
		return "", 0, errorsmod.Wrapf(host.ErrInvalidID, "invalid client identifier %s is not in format: `{client-type}-{N}`", clientID)
	}

	splitStr := strings.Split(clientID, "-")
	lastIndex := len(splitStr) - 1

	clientType := strings.Join(splitStr[:lastIndex], "-")
	if strings.TrimSpace(clientType) == "" {
		return "", 0, errorsmod.Wrap(host.ErrInvalidID, "client identifier must be in format: `{client-type}-{N}` and client type cannot be blank")
	}

	sequence, err := strconv.ParseUint(splitStr[lastIndex], 10, 64)
	if err != nil {
		return "", 0, errorsmod.Wrap(err, "failed to parse client identifier sequence")
	}

	return clientType, sequence, nil
}

// PlanKey is the key under which the current plan is saved
// We store PlanByte as a const to keep it immutable (unlike a []byte)
func PlanKey() []byte {
	return []byte{PlanByte}
}

// UpgradedConsStateKey is the key under which the upgraded consensus state is saved
// Connecting IBC chains can verify against the upgraded consensus state in this path before
// upgrading their clients.
func UpgradedConsStateKey(height int64) []byte {
	return []byte(fmt.Sprintf("%s/%d/%s", KeyUpgradedIBCState, height, KeyUpgradedConsState))
}

// UpgradedClientKey is the key under which the upgraded client state is saved
// Connecting IBC chains can verify against the upgraded client in this path before
// upgrading their clients
func UpgradedClientKey(height int64) []byte {
	return []byte(fmt.Sprintf("%s/%d/%s", KeyUpgradedIBCState, height, KeyUpgradedClient))
}

var UpgradeKey = []byte("upgrade")
