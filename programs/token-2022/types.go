package token2022

import (
	ag_binary "github.com/gagliardetto/solana-go/binary"
)

type AuthorityType ag_binary.BorshEnum

const (
	// Authority to mint new tokens.
	AuthorityMintTokens AuthorityType = iota

	// Authority to freeze any account associated with the Mint.
	AuthorityFreezeAccount

	// Owner of a given token account.
	AuthorityAccountOwner

	// Authority to close a token account.
	AuthorityCloseAccount

	// Authority to set the transfer fee.
	AuthorityTransferFeeConfig

	// Authority to withdraw withheld tokens from a mint.
	AuthorityWithheldWithdraw

	// Authority to close a mint account.
	AuthorityMintCloseAccount

	// Authority to set the interest rate.
	AuthorityInterestRate

	// Authority to transfer or burn any tokens from any account.
	AuthorityPermanentDelegate

	// Authority to update confidential transfer mint parameters.
	AuthorityConfidentialTransferMint

	// Authority to set the transfer hook program id.
	AuthorityTransferHookProgramId

	// Authority to update confidential transfer fee parameters.
	AuthorityConfidentialTransferFeeConfig

	// Authority to set the metadata pointer.
	AuthorityMetadataPointer

	// Authority to set the group pointer.
	AuthorityGroupPointer

	// Authority to set the group member pointer.
	AuthorityGroupMemberPointer

	// Authority to set the scaled UI amount.
	AuthorityScaledUiAmount

	// Authority to pause or resume token operations.
	AuthorityPausable

	// Authority to control permissioned burn.
	AuthorityPermissionedBurn
)

type AccountState ag_binary.BorshEnum

const (
	// Account is not yet initialized.
	AccountStateUninitialized AccountState = iota

	// Account is initialized; the account owner and/or delegate may perform
	// permitted operations on this account.
	AccountStateInitialized

	// Account has been frozen by the mint freeze authority. Neither the account
	// owner nor the delegate are able to perform operations on this account.
	AccountStateFrozen
)

// ExtensionType identifies token-2022 extensions that can be applied to mints or accounts.
type ExtensionType uint16

const (
	ExtensionUninitialized                 ExtensionType = 0
	ExtensionTransferFeeConfig             ExtensionType = 1
	ExtensionTransferFeeAmount             ExtensionType = 2
	ExtensionMintCloseAuthority            ExtensionType = 3
	ExtensionConfidentialTransferMint      ExtensionType = 4
	ExtensionConfidentialTransferAccount   ExtensionType = 5
	ExtensionDefaultAccountState           ExtensionType = 6
	ExtensionImmutableOwner                ExtensionType = 7
	ExtensionMemoTransfer                  ExtensionType = 8
	ExtensionNonTransferable               ExtensionType = 9
	ExtensionInterestBearingConfig         ExtensionType = 10
	ExtensionCpiGuard                      ExtensionType = 11
	ExtensionPermanentDelegate             ExtensionType = 12
	ExtensionNonTransferableAccount        ExtensionType = 13
	ExtensionTransferHook                  ExtensionType = 14
	ExtensionTransferHookAccount           ExtensionType = 15
	ExtensionConfidentialTransferFeeConfig ExtensionType = 16
	ExtensionConfidentialTransferFeeAmount ExtensionType = 17
	ExtensionMetadataPointer               ExtensionType = 18
	ExtensionTokenMetadata                 ExtensionType = 19
	ExtensionGroupPointer                  ExtensionType = 20
	ExtensionTokenGroup                    ExtensionType = 21
	ExtensionGroupMemberPointer            ExtensionType = 22
	ExtensionTokenGroupMember              ExtensionType = 23
	ExtensionConfidentialMintBurn          ExtensionType = 24
	ExtensionScaledUiAmount                ExtensionType = 25
	ExtensionPausable                      ExtensionType = 26
	ExtensionPausableAccount               ExtensionType = 27
	ExtensionPermissionedBurn              ExtensionType = 28
)
