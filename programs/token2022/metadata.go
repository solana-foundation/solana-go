package token2022

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go"
)

const DEFAULT_METADATA_MINT_LEN = 234

/*
Before the Token Extensions Program and the Token Metadata Interface, the process of adding extra data to a Mint Account required creating a Metadata Account through the Metaplex Metadata Program.

The MetadataPointer extension now enables a Mint Account to specify the address of its corresponding Metadata Account. This flexibility allows the Mint Account to point to any account owned by a program that implements the Token Metadata Interface.

The Token Extensions Program directly implements the Token Metadata Interface, made accessible through the TokenMetadata extension. With the TokenMetadata extension, the Mint Account itself can now store the metadata.

Example:

	tokenMetadata := TokenMetadata{
		UpdateAuthority:    updateAuthority,
		Mint:               mint,
		Name:               "OPOS",
		Symbol:             "OPOS",
		Uri:                "https://raw.githubusercontent.com/solana-developers/opos-asset/main/assets/DeveloperPortal/metadata.json",
		AdditionalMetadata: map[string]string{"description": "Only Possible On Solana"},
	}

	lamports, err := rpcClient.GetMinimumBalanceForRentExemption(ctx, tokenMetadata.LenForLamports(), rpc.CommitmentFinalized)
	if err != nil {
		return err
	}

	createAccountInstruction := system.NewCreateAccountInstruction(
		lamports,
		234,
		token2022.ProgramID,
		payer,
		mint.PublicKey(),
	).Build()

);
*/
type TokenMetadata struct {
	// The authority that can sign to update the metadata
	UpdateAuthority *solana.PublicKey
	// The associated mint, used to counter spoofing to be sure that metadata belongs to a particular mint
	Mint solana.PublicKey
	// The longer name of the token
	Name string
	// The shortened symbol for the token
	Symbol string
	// The URI pointing to richer metadata
	Uri string
	// Any additional metadata about the token as key-value pairs
	AdditionalMetadata map[string]string
}

// Convert the metadata to a byte array
func (meta *TokenMetadata) Pack() []byte {
	// If no updateAuthority given, set it to the None/Zero PublicKey for encoding
	updateAuthority, _ := solana.PublicKeyFromBase58("11111111111111111111111111111111")
	if meta.UpdateAuthority != nil {
		updateAuthority = *meta.UpdateAuthority
	}
	var buf bytes.Buffer
	buf.Write(fixCodecSize(getBytesCodec(), 32)(updateAuthority.Bytes()))
	buf.Write(fixCodecSize(getBytesCodec(), 32)(meta.Mint.Bytes()))
	buf.Write(getStringCodec()(meta.Name))
	buf.Write(getStringCodec()(meta.Symbol))
	buf.Write(getStringCodec()(meta.Uri))
	buf.Write(getMapCodec(getStringCodec())(meta.AdditionalMetadata))
	return buf.Bytes()
}

// Use this in conjuntion with GetMinimumBalanceForRentExemption to calculate the lamports needed to create the account
func (meta *TokenMetadata) LenForLamports() uint64 {
	return DEFAULT_METADATA_MINT_LEN + uint64(len(meta.Pack()))
}

// Construct an Initialize MetadataPointer instruction
func NewInitializeMetadataPointerInstruction(
	mint solana.PublicKey,
	authority solana.PublicKey,
	metadataAddress solana.PublicKey,
) solana.Instruction {
	programID := ProgramID

	pointerData := initializeMetadataPointerData{
		Instruction:                MetadataPointerExtension,
		MetadataPointerInstruction: initialize,
		Authority:                  authority,
		MetadataAddress:            metadataAddress,
	}

	ix := &createInitializeMetadataPointerInstruction{
		programID: programID,
		accounts: []*solana.AccountMeta{
			solana.Meta(mint).WRITE(),
		},
		data: pointerData.encode(),
	}

	return ix
}

type createInitializeMetadataPointerInstruction struct {
	programID solana.PublicKey
	accounts  []*solana.AccountMeta
	data      []byte
}

func (inst *createInitializeMetadataPointerInstruction) ProgramID() solana.PublicKey {
	return inst.programID
}

func (inst *createInitializeMetadataPointerInstruction) Accounts() (out []*solana.AccountMeta) {
	return inst.accounts
}

func (inst *createInitializeMetadataPointerInstruction) Data() ([]byte, error) {
	return inst.data, nil
}

type initializeMetadataPointerData struct {
	Instruction                TokenInstruction
	MetadataPointerInstruction programInstruction
	Authority                  solana.PublicKey
	MetadataAddress            solana.PublicKey
}

func (data *initializeMetadataPointerData) encode() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, data.Instruction)
	binary.Write(&buf, binary.LittleEndian, data.MetadataPointerInstruction)
	buf.Write(data.Authority.Bytes())
	buf.Write(data.MetadataAddress.Bytes())
	return buf.Bytes()
}

// TODO: Must test this!
func NewUpdateMetadataPointerInstruction(
	mint solana.PublicKey,
	authority solana.PublicKey,
	metadataAddress *solana.PublicKey,
	multiSigners []any,
) solana.Instruction {
	programID := ProgramID

	keys := addSigners([]*solana.AccountMeta{
		solana.Meta(mint).WRITE(),
	}, authority, multiSigners)

	pointerData := updateMetadataPointerData{
		Instruction:                MetadataPointerExtension,
		MetadataPointerInstruction: update,
		MetadataAddress:            *metadataAddress,
	}

	ix := &createUpdateMetadataPointerInstruction{
		programID: programID,
		accounts:  keys,
		data:      pointerData.encode(),
	}
	return ix
}

func ParseTokenMetadata(data []byte) (TokenMetadata, error) {
	if len(data) < DEFAULT_METADATA_MINT_LEN {
		return TokenMetadata{}, fmt.Errorf("data length is less than the minimum required length of 234")
	}

	var updateAuth *solana.PublicKey = nil
	if data[238] != 0 {
		a := solana.PublicKeyFromBytes(data[238 : 238+32])
		updateAuth = &a
	}

	mint := solana.PublicKeyFromBytes(data[202 : 202+32])

	if len(data) < 306 {
		return TokenMetadata{
			UpdateAuthority:    updateAuth,
			Mint:               mint,
			Name:               "",
			Symbol:             "",
			Uri:                "",
			AdditionalMetadata: map[string]string{},
		}, fmt.Errorf("data length is less than what is necessary to parse metadata")
	}

	nameLen := int(binary.LittleEndian.Uint32(data[302 : 302+4]))
	if len(data) < 302+4+int(nameLen) {
		return TokenMetadata{
			UpdateAuthority:    updateAuth,
			Mint:               mint,
			Name:               "",
			Symbol:             "",
			Uri:                "",
			AdditionalMetadata: map[string]string{},
		}, fmt.Errorf("data length is less than what is necessary to parse metadata")
	}
	name := string(data[302+4 : 302+4+nameLen])

	if len(data) < 302+4+nameLen+4 {
		return TokenMetadata{
			UpdateAuthority:    updateAuth,
			Mint:               mint,
			Name:               name,
			Symbol:             "",
			Uri:                "",
			AdditionalMetadata: map[string]string{},
		}, fmt.Errorf("data length is less than what is necessary to parse metadata")
	}

	symbolLen := int(binary.LittleEndian.Uint32(data[302+4+nameLen : 302+4+nameLen+4]))
	if len(data) < 302+4+nameLen+4+symbolLen {
		return TokenMetadata{
			UpdateAuthority:    updateAuth,
			Mint:               mint,
			Name:               name,
			Symbol:             "",
			Uri:                "",
			AdditionalMetadata: map[string]string{},
		}, fmt.Errorf("data length is less than what is necessary to parse metadata")
	}
	symbol := string(data[302+4+nameLen+4 : 302+4+nameLen+4+symbolLen])

	if len(data) < 302+4+nameLen+4+symbolLen+4 {
		return TokenMetadata{
			UpdateAuthority:    updateAuth,
			Mint:               mint,
			Name:               name,
			Symbol:             symbol,
			Uri:                "",
			AdditionalMetadata: map[string]string{},
		}, fmt.Errorf("data length is less than what is necessary to parse metadata")
	}

	uriLen := int(binary.LittleEndian.Uint32(data[302+4+nameLen+4+symbolLen : 302+4+nameLen+4+symbolLen+4]))
	if len(data) < 302+4+nameLen+4+symbolLen+4+uriLen {
		return TokenMetadata{
			UpdateAuthority:    updateAuth,
			Mint:               mint,
			Name:               name,
			Symbol:             symbol,
			Uri:                "",
			AdditionalMetadata: map[string]string{},
		}, fmt.Errorf("data length is less than what is necessary to parse metadata")
	}
	uri := string(data[302+4+nameLen+4+symbolLen+4 : 302+4+nameLen+4+symbolLen+4+uriLen])

	meta := TokenMetadata{
		UpdateAuthority:    updateAuth,
		Mint:               mint,
		Name:               name,
		Symbol:             symbol,
		Uri:                uri,
		AdditionalMetadata: map[string]string{},
	}

	return meta, nil
}

// u32('mintAuthorityOption'),
// publicKey('mintAuthority'),
// u64('supply'),
// u8('decimals'),
// bool('isInitialized'),
// u32('freezeAuthorityOption'),
// publicKey('freezeAuthority'),

func addSigners(keys []*solana.AccountMeta, authority solana.PublicKey, multiSigners []interface{}) []*solana.AccountMeta {
	keys = append(keys, solana.Meta(authority).SIGNER())
	for _, signer := range multiSigners {
		switch v := signer.(type) {
		case solana.PublicKey:
			keys = append(keys, solana.Meta(v).SIGNER())
		case solana.PrivateKey:
			keys = append(keys, solana.Meta(v.PublicKey()).SIGNER())
		}
	}
	return keys
}

type updateMetadataPointerData struct {
	Instruction                TokenInstruction
	MetadataPointerInstruction programInstruction
	MetadataAddress            solana.PublicKey
}

func (data *updateMetadataPointerData) encode() []byte {
	var buf bytes.Buffer
	buf.WriteByte(byte(data.Instruction))
	buf.WriteByte(byte(data.MetadataPointerInstruction))
	buf.Write(data.MetadataAddress.Bytes())
	return buf.Bytes()
}

type createUpdateMetadataPointerInstruction struct {
	programID solana.PublicKey
	accounts  []*solana.AccountMeta
	data      []byte
}

func (inst *createUpdateMetadataPointerInstruction) ProgramID() solana.PublicKey {
	return inst.programID
}

func (inst *createUpdateMetadataPointerInstruction) Accounts() (out []*solana.AccountMeta) {
	return inst.accounts
}

func (inst *createUpdateMetadataPointerInstruction) Data() ([]byte, error) {
	return inst.data, nil
}
