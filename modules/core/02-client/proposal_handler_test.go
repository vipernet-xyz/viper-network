package client_test

/*
	distributiontypes "github.com/vipernet-xyz/viper-network/x/distribution/types"
	govtypes "github.com/vipernet-xyz/viper-network/x/gov/types/v1beta1"
	sdk "github.com/vipernet-xyz/viper-network/types"

	ibctesting "github.com/vipernet-xyz/viper-network/testing"
	client "github.com/vipernet-xyz/viper-network/modules/core/02-client"
	clienttypes "github.com/vipernet-xyz/viper-network/modules/core/02-client/types"
	ibctm "github.com/vipernet-xyz/viper-network/modules/light-clients/07-tendermint"*/

/*func (suite *ClientTestSuite) TestNewClientUpdateProposalHandler() {
	var (
		content govtypes.Content
		err     error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"valid update client proposal", func() {
				subjectPath := ibctesting.NewPath(suite.chainA, suite.chainB)
				suite.coordinator.SetupClients(subjectPath)
				subjectClientState := suite.chainA.GetClientState(subjectPath.EndpointA.ClientID)

				substitutePath := ibctesting.NewPath(suite.chainA, suite.chainB)
				suite.coordinator.SetupClients(substitutePath)

				// update substitute twice
				err = substitutePath.EndpointA.UpdateClient()
				suite.Require().NoError(err)
				err = substitutePath.EndpointA.UpdateClient()
				suite.Require().NoError(err)
				substituteClientState := suite.chainA.GetClientState(substitutePath.EndpointA.ClientID)

				tmClientState, ok := subjectClientState.(*ibctm.ClientState)
				suite.Require().True(ok)
				tmClientState.AllowUpdateAfterMisbehaviour = true
				tmClientState.FrozenHeight = tmClientState.LatestHeight
				suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientState(suite.chainA.GetContext(), subjectPath.EndpointA.ClientID, tmClientState)

				// replicate changes to substitute (they must match)
				tmClientState, ok = substituteClientState.(*ibctm.ClientState)
				suite.Require().True(ok)
				tmClientState.AllowUpdateAfterMisbehaviour = true
				suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientState(suite.chainA.GetContext(), substitutePath.EndpointA.ClientID, tmClientState)

				content = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, subjectPath.EndpointA.ClientID, substitutePath.EndpointA.ClientID)
			}, true,
		},
		{
			"nil proposal", func() {
				content = nil
			}, false,
		},
		{
			"unsupported proposal type", func() {
				content = &distributiontypes.CommunityPoolSpendProposal{ //nolint:staticcheck
					Title:       ibctesting.Title,
					Description: ibctesting.Description,
					Recipient:   suite.chainA.SenderAccount.GetAddress().String(),
					Amount:      sdk.NewCoins(sdk.NewCoin("communityfunds", sdk.NewInt(10))),
				}
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			tc.malleate()

			proposalHandler := client.NewClientProposalHandler(suite.chainA.App.GetIBCKeeper().ClientKeeper)

			err = proposalHandler(suite.chainA.GetContext(), content)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
*/
