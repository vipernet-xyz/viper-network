package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
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
	cdc.RegisterStructure(MsgClaim{}, "vipernet/claim")
	cdc.RegisterStructure(MsgProtoProof{}, "vipernet/protoProof")
	cdc.RegisterStructure(MsgProof{}, "vipernet/proof")
	cdc.RegisterStructure(Relay{}, "vipernet/relay")
	cdc.RegisterStructure(Session{}, "vipernet/session")
	cdc.RegisterStructure(RelayResponse{}, "vipernet/relay_response")
	cdc.RegisterStructure(RelayProof{}, "vipernet/relay_proof")
	cdc.RegisterStructure(ChallengeProofInvalidData{}, "vipernet/challenge_proof_invalid_data")
	cdc.RegisterStructure(ProofI_RelayProof{}, "vipernet/proto_relay_proofI")
	cdc.RegisterStructure(ProofI_ChallengeProof{}, "vipernet/proto_challenge_proofI")
	cdc.RegisterStructure(ProtoEvidence{}, "vipernet/evidence_persisted")
	cdc.RegisterStructure(servicersTypes.Validator{}, "pos/8.0Validator")    // todo does this really need to depend on servicers/types
	cdc.RegisterStructure(servicersTypes.LegacyValidator{}, "pos/Validator") // todo does this really need to depend on servicers/types
	cdc.RegisterInterface("x.vipernet.Proof", (*Proof)(nil), &RelayProof{}, &ChallengeProofInvalidData{})
	cdc.RegisterInterface("types.isProofI_Proof", (*isProofI_Proof)(nil))
	cdc.RegisterImplementation((*sdk.ProtoMsg)(nil), &MsgClaim{}, &MsgProof{})
	cdc.RegisterImplementation((*sdk.Msg)(nil), &MsgClaim{}, &MsgProof{})
	ModuleCdc = cdc
}
