package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	nodesTypes "github.com/vipernet-xyz/viper-network/x/nodes/types"
)

// module wide codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.NewCodec(types.NewInterfaceRegistry())
	RegisterCodec(ModuleCdc)
	crypto.RegisterAmino(ModuleCdc.AminoCodec().Amino)
}

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterStructure(MsgClaim{}, "vipercore/claim")
	cdc.RegisterStructure(MsgProtoProof{}, "vipercore/protoProof")
	cdc.RegisterStructure(MsgProof{}, "vipercore/proof")
	cdc.RegisterStructure(Relay{}, "vipercore/relay")
	cdc.RegisterStructure(Session{}, "vipercore/session")
	cdc.RegisterStructure(RelayResponse{}, "vipercore/relay_response")
	cdc.RegisterStructure(RelayProof{}, "vipercore/relay_proof")
	cdc.RegisterStructure(ChallengeProofInvalidData{}, "vipercore/challenge_proof_invalid_data")
	cdc.RegisterStructure(ProofI_RelayProof{}, "vipercore/proto_relay_proofI")
	cdc.RegisterStructure(ProofI_ChallengeProof{}, "vipercore/proto_challenge_proofI")
	cdc.RegisterStructure(ProtoEvidence{}, "vipercore/evidence_persisted")
	cdc.RegisterStructure(nodesTypes.Validator{}, "pos/8.0Validator")    // todo does this really need to depend on nodes/types
	cdc.RegisterStructure(nodesTypes.LegacyValidator{}, "pos/Validator") // todo does this really need to depend on nodes/types
	cdc.RegisterInterface("x.vipercore.Proof", (*Proof)(nil), &RelayProof{}, &ChallengeProofInvalidData{})
	cdc.RegisterInterface("types.isProofI_Proof", (*isProofI_Proof)(nil))
	cdc.RegisterImplementation((*sdk.ProtoMsg)(nil), &MsgClaim{}, &MsgProof{})
	cdc.RegisterImplementation((*sdk.Msg)(nil), &MsgClaim{}, &MsgProof{})
	ModuleCdc = cdc
}
