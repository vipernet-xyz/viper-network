package types

import (
	"fmt"
	"reflect"
	"strings"

	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/vipernet-xyz/viper-network/codec/types"
	sdkerrors "github.com/vipernet-xyz/viper-network/types/errors"
	govtypes "github.com/vipernet-xyz/viper-network/x/governance/types"

	"github.com/vipernet-xyz/viper-network/modules/core/exported"
)

const (
	// ProposalTypeClientUpdate defines the type for a ClientUpdateProposal
	ProposalTypeClientUpdate = "ClientUpdate"
	ProposalTypeUpgrade      = "IBCUpgrade"
)

var (
	_ govtypes.Content                   = &ClientUpdateProposal{}
	_ govtypes.Content                   = &UpgradeProposal{}
	_ codectypes.UnpackInterfacesMessage = &UpgradeProposal{}
)

func init() {
	RegisterProposalType(ProposalTypeClientUpdate)
	RegisterProposalType(ProposalTypeUpgrade)
}

// NewClientUpdateProposal creates a new client update proposal.
func NewClientUpdateProposal(title, description, subjectClientID, substituteClientID string) govtypes.Content {
	return &ClientUpdateProposal{
		Title:              title,
		Description:        description,
		SubjectClientId:    subjectClientID,
		SubstituteClientId: substituteClientID,
	}
}

// GetTitle returns the title of a client update proposal.
func (cup *ClientUpdateProposal) GetTitle() string { return cup.Title }

// GetDescription returns the description of a client update proposal.
func (cup *ClientUpdateProposal) GetDescription() string { return cup.Description }

// ProposalRoute returns the routing key of a client update proposal.
func (cup *ClientUpdateProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a client update proposal.
func (cup *ClientUpdateProposal) ProposalType() string { return ProposalTypeClientUpdate }

// ValidateBasic runs basic stateless validity checks
func (cup *ClientUpdateProposal) ValidateBasic() error {
	err := ValidateAbstract(cup)
	if err != nil {
		return err
	}

	if cup.SubjectClientId == cup.SubstituteClientId {
		return errorsmod.Wrap(ErrInvalidSubstitute, "subject and substitute client identifiers are equal")
	}
	if _, _, err := ParseClientIdentifier(cup.SubjectClientId); err != nil {
		return err
	}
	if _, _, err := ParseClientIdentifier(cup.SubstituteClientId); err != nil {
		return err
	}

	return nil
}

// NewUpgradeProposal creates a new IBC breaking upgrade proposal.
func NewUpgradeProposal(title, description string, plan Plan, upgradedClientState exported.ClientState) (govtypes.Content, error) {
	protoAny, err := PackClientState(upgradedClientState)
	if err != nil {
		return nil, err
	}

	return &UpgradeProposal{
		Title:               title,
		Description:         description,
		Plan:                plan,
		UpgradedClientState: protoAny,
	}, nil
}

// GetTitle returns the title of a upgrade proposal.
func (up *UpgradeProposal) GetTitle() string { return up.Title }

// GetDescription returns the description of a upgrade proposal.
func (up *UpgradeProposal) GetDescription() string { return up.Description }

// ProposalRoute returns the routing key of a upgrade proposal.
func (up *UpgradeProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the upgrade proposal type.
func (up *UpgradeProposal) ProposalType() string { return ProposalTypeUpgrade }

// ValidateBasic runs basic stateless validity checks
func (up *UpgradeProposal) ValidateBasic() error {
	if err := ValidateAbstract(up); err != nil {
		return err
	}

	if err := up.Plan.ValidateBasic(); err != nil {
		return err
	}

	if up.UpgradedClientState == nil {
		return errorsmod.Wrap(ErrInvalidUpgradeProposal, "upgraded client state cannot be nil")
	}

	clientState, err := UnpackClientState(up.UpgradedClientState)
	if err != nil {
		return errorsmod.Wrap(err, "failed to unpack upgraded client state")
	}

	if !reflect.DeepEqual(clientState, clientState.ZeroCustomFields()) {
		return errorsmod.Wrap(ErrInvalidUpgradeProposal, "upgraded client state is not zeroed out")
	}

	return nil
}

// String returns the string representation of the UpgradeProposal.
func (up UpgradeProposal) String() string {
	var upgradedClientStr string
	upgradedClient, err := UnpackClientState(up.UpgradedClientState)
	if err != nil {
		upgradedClientStr = "invalid IBC Client State"
	} else {
		upgradedClientStr = upgradedClient.String()
	}

	return fmt.Sprintf(`IBC Upgrade Proposal
  Title: %s
  Description: %s
  %s
  Upgraded IBC Client: %s`, up.Title, up.Description, up.Plan, upgradedClientStr)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (up UpgradeProposal) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(up.UpgradedClientState, new(exported.ClientState))
}

// Proposal types
const (
	ProposalTypeText string = "Text"

	// Constants pertaining to a Content object
	MaxDescriptionLength int = 10000
	MaxTitleLength       int = 140
)

var validProposalTypes = map[string]struct{}{
	ProposalTypeText: {},
}

// RegisterProposalType registers a proposal type. It will panic if the type is
// already registered.
func RegisterProposalType(ty string) {
	if _, ok := validProposalTypes[ty]; ok {
		panic(fmt.Sprintf("already registered proposal type: %s", ty))
	}

	validProposalTypes[ty] = struct{}{}
}

// ValidateAbstract validates a proposal's abstract contents returning an error
// if invalid.
func ValidateAbstract(c govtypes.Content) error {
	title := c.GetTitle()
	if len(strings.TrimSpace(title)) == 0 {
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalContent, "proposal title cannot be blank")
	}
	if len(title) > MaxTitleLength {
		return sdkerrors.Wrapf(govtypes.ErrInvalidProposalContent, "proposal title is longer than max length of %d", MaxTitleLength)
	}

	description := c.GetDescription()
	if len(description) == 0 {
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalContent, "proposal description cannot be blank")
	}
	if len(description) > MaxDescriptionLength {
		return sdkerrors.Wrapf(govtypes.ErrInvalidProposalContent, "proposal description is longer than max length of %d", MaxDescriptionLength)
	}

	return nil
}

// ValidateBasic does basic validation of a Plan
func (p Plan) ValidateBasic() error {
	if !p.Time.IsZero() {
		return sdkerrors.ErrInvalidRequest.Wrap("time-based upgrades have been deprecated in the SDK")
	}
	if p.UpgradedClientState != nil {
		return sdkerrors.ErrInvalidRequest.Wrap("upgrade logic for IBC has been moved to the IBC module")
	}
	if len(p.Name) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "name cannot be empty")
	}
	if p.Height <= 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "height must be greater than 0")
	}

	return nil
}
