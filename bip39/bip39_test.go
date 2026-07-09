package bip39

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/gagliardetto/solana-go/bip39/wordlists"
)

type vector struct {
	entropy  string
	mnemonic string
	seed     string
}

func TestGetWordList(t *testing.T) {
	assertEqualStringSlices(t, wordlists.English, GetWordList())
}

func TestGetWordIndex(t *testing.T) {
	for expectedIdx, word := range wordList {
		actualIdx, ok := GetWordIndex(word)
		assertTrue(t, ok)
		assertEqual(t, actualIdx, expectedIdx)
	}

	for _, word := range []string{"a", "set", "of", "invalid", "words"} {
		actualIdx, ok := GetWordIndex(word)
		assertFalse(t, ok)
		assertEqual(t, actualIdx, 0)
	}
}

func TestNewMnemonic(t *testing.T) {
	for _, vector := range testVectors() {
		entropy, err := hex.DecodeString(vector.entropy)
		assertNil(t, err)

		mnemonic, err := NewMnemonic(entropy)
		assertNil(t, err)
		assertEqualString(t, vector.mnemonic, mnemonic)

		_, err = NewSeedWithErrorChecking(mnemonic, "TREZOR")
		assertNil(t, err)

		seed := NewSeed(mnemonic, "TREZOR")
		assertEqualString(t, vector.seed, hex.EncodeToString(seed))
	}
}

func TestNewMnemonicInvalidEntropy(t *testing.T) {
	_, err := NewMnemonic([]byte{})
	assertNotNil(t, err)
}

func TestNewSeedWithErrorCheckingInvalidMnemonics(t *testing.T) {
	for _, vector := range badMnemonicSentences() {
		_, err := NewSeedWithErrorChecking(vector.mnemonic, "TREZOR")
		assertNotNil(t, err)
	}
}

func TestIsMnemonicValid(t *testing.T) {
	for _, vector := range badMnemonicSentences() {
		assertFalse(t, IsMnemonicValid(vector.mnemonic))
	}

	for _, vector := range testVectors() {
		assertTrue(t, IsMnemonicValid(vector.mnemonic))
	}
}

func TestMnemonicToByteArrayInvalidMnemonic(t *testing.T) {
	for _, vector := range badMnemonicSentences() {
		_, err := MnemonicToByteArray(vector.mnemonic)
		assertNotNil(t, err)
	}

	_, err := MnemonicToByteArray("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon yellow")
	assertNotNil(t, err)
	assertEqual(t, err, ErrInvalidMnemonic)
}

func TestNewEntropy(t *testing.T) {
	// Good tests.
	for i := 128; i <= 256; i += 32 {
		_, err := NewEntropy(i)
		assertNil(t, err)
	}
	// Bad Values
	for i := 0; i <= 256; i++ {
		if i%8 != 0 {
			_, err := NewEntropy(i)
			assertNotNil(t, err)
		}
	}
}

func TestMnemonicToByteArrayForDifferentArrayLengths(t *testing.T) {
	max := 1000
	for i := 0; i < max; i++ {
		//16, 20, 24, 28, 32
		length := 16 + (i%5)*4
		seed := make([]byte, length)
		if n, err := rand.Read(seed); err != nil {
			t.Errorf("%v", err)
		} else if n != length {
			t.Errorf("Wrong number of bytes read: %d", n)
		}

		mnemonic, err := NewMnemonic(seed)
		if err != nil {
			t.Errorf("%v", err)
		}

		_, err = MnemonicToByteArray(mnemonic)
		if err != nil {
			t.Errorf("Failed for %x - %v", seed, mnemonic)
		}
	}
}
func TestPadByteSlice(t *testing.T) {
	assertEqualByteSlices(t, []byte{0}, padByteSlice([]byte{}, 1))
	assertEqualByteSlices(t, []byte{0, 1}, padByteSlice([]byte{1}, 2))
	assertEqualByteSlices(t, []byte{1, 1}, padByteSlice([]byte{1, 1}, 2))
	assertEqualByteSlices(t, []byte{1, 1, 1}, padByteSlice([]byte{1, 1, 1}, 2))
}

func TestCompareByteSlices(t *testing.T) {
	assertTrue(t, compareByteSlices([]byte{}, []byte{}))
	assertTrue(t, compareByteSlices([]byte{1}, []byte{1}))
	assertFalse(t, compareByteSlices([]byte{1}, []byte{0}))
	assertFalse(t, compareByteSlices([]byte{1}, []byte{}))
	assertFalse(t, compareByteSlices([]byte{1}, nil))
}

func TestMnemonicToByteArrayForZeroLeadingSeeds(t *testing.T) {
	ms := []string{
		"00000000000000000000000000000000",
		"00a84c51041d49acca66e6160c1fa999",
		"00ca45df1673c76537a2020bfed1dafd",
		"0019d5871c7b81fd83d474ef1c1e1dae",
		"00dcb021afb35ffcdd1d032d2056fc86",
		"0062be7bd09a27288b6cf0eb565ec739",
		"00dc705b5efa0adf25b9734226ba60d4",
		"0017747418d54c6003fa64fade83374b",
		"000d44d3ee7c3dfa45e608c65384431b",
		"008241c1ef976b0323061affe5bf24b9",
		"00a6aec77e4d16bea80b50a34991aaba",
		"0011527b8c6ddecb9d0c20beccdeb58d",
		"001c938c503c8f5a2bba2248ff621546",
		"0002f90aaf7a8327698f0031b6317c36",
		"00bff43071ed7e07f77b14f615993bac",
		"00da143e00ef17fc63b6fb22dcc2c326",
		"00ffc6764fb32a354cab1a3ddefb015d",
		"0062ef47e0985e8953f24760b7598cdd",
		"003bf9765064f71d304908d906c065f5",
		"00993851503471439d154b3613947474",
		"007ad0ffe9eae753a483a76af06dfa67",
		"00091824db9ec19e663bee51d64c83cc",
		"00f48ac621f7e3cb39b2012ac3121543",
		"0072917415cdca24dfa66c4a92c885b4",
		"0027ced2b279ea8a91d29364487cdbf4",
		"00b9c0d37fb10ba272e55842ad812583",
		"004b3d0d2b9285946c687a5350479c8c",
		"00c7c12a37d3a7f8c1532b17c89b724c",
		"00f400c5545f06ae17ad00f3041e4e26",
		"001e290be10df4d209f247ac5878662b",
		"00bf0f74568e582a7dd1ee64f792ec8b",
		"00d2e43ecde6b72b847db1539ed89e23",
		"00cecba6678505bb7bfec8ed307251f6",
		"000aeed1a9edcbb4bc88f610d3ce84eb",
		"00d06206aadfc25c2b21805d283f15ae",
		"00a31789a2ab2d54f8fadd5331010287",
		"003493c5f520e8d5c0483e895a121dc9",
		"004706112800b76001ece2e268bc830e",
		"00ab31e28bb5305be56e38337dbfa486",
		"006872fe85df6b0fa945248e6f9379d1",
		"00717e5e375da6934e3cfdf57edaf3bd",
		"007f1b46e7b9c4c76e77c434b9bccd6b",
		"00dc93735aa35def3b9a2ff676560205",
		"002cd5dcd881a49c7b87714c6a570a76",
		"0013b5af9e13fac87e0c505686cfb6bf",
		"007ab1ec9526b0bc04b64ae65fd42631",
		"00abb4e11d8385c1cca905a6a65e9144",
		"00574fc62a0501ad8afada2e246708c3",
		"005207e0a815bb2da6b4c35ec1f2bf52",
		"00f3460f136fb9700080099cbd62bc18",
		"007a591f204c03ca7b93981237112526",
		"00cfe0befd428f8e5f83a5bfc801472e",
		"00987551ac7a879bf0c09b8bc474d9af",
		"00cadd3ce3d78e49fbc933a85682df3f",
		"00bfbf2e346c855ccc360d03281455a1",
		"004cdf55d429d028f715544ce22d4f31",
		"0075c84a7d15e0ac85e1e41025eed23b",
		"00807dddd61f71725d336cab844d2cb5",
		"00422f21b77fe20e367467ed98c18410",
		"00b44d0ac622907119c626c850a462fd",
		"00363f5e7f22fc49f3cd662a28956563",
		"000fe5837e68397bbf58db9f221bdc4e",
		"0056af33835c888ef0c22599686445d3",
		"00790a8647fd3dfb38b7e2b6f578f2c6",
		"00da8d9009675cb7beec930e263014fb",
		"00d4b384540a5bb54aa760edaa4fb2fe",
		"00be9b1479ed680fdd5d91a41eb926d0",
		"009182347502af97077c40a6e74b4b5c",
		"00f5c90ee1c67fa77fd821f8e9fab4f1",
		"005568f9a2dd6b0c0cc2f5ba3d9cac38",
		"008b481f8678577d9cf6aa3f6cd6056b",
		"00c4323ece5e4fe3b6cd4c5c932931af",
		"009791f7550c3798c5a214cb2d0ea773",
		"008a7baab22481f0ad8167dd9f90d55c",
		"00f0e601519aafdc8ff94975e64c946d",
		"0083b61e0daa9219df59d697c270cd31",
	}

	for _, m := range ms {
		seed, _ := hex.DecodeString(m)

		mnemonic, err := NewMnemonic(seed)
		if err != nil {
			t.Errorf("%v", err)
		}

		_, err = MnemonicToByteArray(mnemonic)
		if err != nil {
			t.Errorf("Failed for %x - %v", seed, mnemonic)
		}
	}
}
func TestEntropyFromMnemonic128(t *testing.T) {
	testEntropyFromMnemonic(t, 128)
}

func TestEntropyFromMnemonic160(t *testing.T) {
	testEntropyFromMnemonic(t, 160)
}

func TestEntropyFromMnemonic192(t *testing.T) {
	testEntropyFromMnemonic(t, 192)
}

func TestEntropyFromMnemonic224(t *testing.T) {
	testEntropyFromMnemonic(t, 224)
}

func TestEntropyFromMnemonic256(t *testing.T) {
	testEntropyFromMnemonic(t, 256)
}

func TestEntropyFromMnemonicInvalidChecksum(t *testing.T) {
	_, err := EntropyFromMnemonic("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon yellow")
	assertEqual(t, ErrChecksumIncorrect, err)
}

func TestEntropyFromMnemonicInvalidMnemonicSize(t *testing.T) {
	for _, mnemonic := range []string{
		"a a a a a a a a a a a a a a a a a a a a a a a a a", // Too many words
		"a",                           // Too few
		"a a a a a a a a a a a a a a", // Not multiple of 3
	} {
		_, err := EntropyFromMnemonic(mnemonic)
		assertEqual(t, ErrInvalidMnemonic, err)
	}
}

func testEntropyFromMnemonic(t *testing.T, bitSize int) {
	for i := 0; i < 512; i++ {
		expectedEntropy, err := NewEntropy(bitSize)
		assertNil(t, err)
		assertTrue(t, len(expectedEntropy) != 0)

		mnemonic, err := NewMnemonic(expectedEntropy)
		assertNil(t, err)
		assertTrue(t, len(mnemonic) != 0)

		actualEntropy, err := EntropyFromMnemonic(mnemonic)
		assertNil(t, err)
		assertEqualByteSlices(t, expectedEntropy, actualEntropy)
	}
}

func testVectors() []vector {
	return []vector{
		{
			entropy:  "00000000000000000000000000000000",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			seed:     "c55257c360c07c72029aebc1b53c05ed0362ada38ead3e3e9efa3708e53495531f09a6987599d18264c1e1c92f2cf141630c7a3c4ab7c81b2f001698e7463b04",
		},
		{
			entropy:  "7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f",
			mnemonic: "legal winner thank year wave sausage worth useful legal winner thank yellow",
			seed:     "2e8905819b8723fe2c1d161860e5ee1830318dbf49a83bd451cfb8440c28bd6fa457fe1296106559a3c80937a1c1069be3a3a5bd381ee6260e8d9739fce1f607",
		},
		{
			entropy:  "80808080808080808080808080808080",
			mnemonic: "letter advice cage absurd amount doctor acoustic avoid letter advice cage above",
			seed:     "d71de856f81a8acc65e6fc851a38d4d7ec216fd0796d0a6827a3ad6ed5511a30fa280f12eb2e47ed2ac03b5c462a0358d18d69fe4f985ec81778c1b370b652a8",
		},
		{
			entropy:  "ffffffffffffffffffffffffffffffff",
			mnemonic: "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong",
			seed:     "ac27495480225222079d7be181583751e86f571027b0497b5b5d11218e0a8a13332572917f0f8e5a589620c6f15b11c61dee327651a14c34e18231052e48c069",
		},
		{
			entropy:  "000000000000000000000000000000000000000000000000",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon agent",
			seed:     "035895f2f481b1b0f01fcf8c289c794660b289981a78f8106447707fdd9666ca06da5a9a565181599b79f53b844d8a71dd9f439c52a3d7b3e8a79c906ac845fa",
		},
		{
			entropy:  "7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f",
			mnemonic: "legal winner thank year wave sausage worth useful legal winner thank year wave sausage worth useful legal will",
			seed:     "f2b94508732bcbacbcc020faefecfc89feafa6649a5491b8c952cede496c214a0c7b3c392d168748f2d4a612bada0753b52a1c7ac53c1e93abd5c6320b9e95dd",
		},
		{
			entropy:  "808080808080808080808080808080808080808080808080",
			mnemonic: "letter advice cage absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter always",
			seed:     "107d7c02a5aa6f38c58083ff74f04c607c2d2c0ecc55501dadd72d025b751bc27fe913ffb796f841c49b1d33b610cf0e91d3aa239027f5e99fe4ce9e5088cd65",
		},
		{
			entropy:  "ffffffffffffffffffffffffffffffffffffffffffffffff",
			mnemonic: "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo when",
			seed:     "0cd6e5d827bb62eb8fc1e262254223817fd068a74b5b449cc2f667c3f1f985a76379b43348d952e2265b4cd129090758b3e3c2c49103b5051aac2eaeb890a528",
		},
		{
			entropy:  "0000000000000000000000000000000000000000000000000000000000000000",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
			seed:     "bda85446c68413707090a52022edd26a1c9462295029f2e60cd7c4f2bbd3097170af7a4d73245cafa9c3cca8d561a7c3de6f5d4a10be8ed2a5e608d68f92fcc8",
		},
		{
			entropy:  "7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f7f",
			mnemonic: "legal winner thank year wave sausage worth useful legal winner thank year wave sausage worth useful legal winner thank year wave sausage worth title",
			seed:     "bc09fca1804f7e69da93c2f2028eb238c227f2e9dda30cd63699232578480a4021b146ad717fbb7e451ce9eb835f43620bf5c514db0f8add49f5d121449d3e87",
		},
		{
			entropy:  "8080808080808080808080808080808080808080808080808080808080808080",
			mnemonic: "letter advice cage absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic bless",
			seed:     "c0c519bd0e91a2ed54357d9d1ebef6f5af218a153624cf4f2da911a0ed8f7a09e2ef61af0aca007096df430022f7a2b6fb91661a9589097069720d015e4e982f",
		},
		{
			entropy:  "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			mnemonic: "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo vote",
			seed:     "dd48c104698c30cfe2b6142103248622fb7bb0ff692eebb00089b32d22484e1613912f0a5b694407be899ffd31ed3992c456cdf60f5d4564b8ba3f05a69890ad",
		},
		{
			entropy:  "77c2b00716cec7213839159e404db50d",
			mnemonic: "jelly better achieve collect unaware mountain thought cargo oxygen act hood bridge",
			seed:     "b5b6d0127db1a9d2226af0c3346031d77af31e918dba64287a1b44b8ebf63cdd52676f672a290aae502472cf2d602c051f3e6f18055e84e4c43897fc4e51a6ff",
		},
		{
			entropy:  "b63a9c59a6e641f288ebc103017f1da9f8290b3da6bdef7b",
			mnemonic: "renew stay biology evidence goat welcome casual join adapt armor shuffle fault little machine walk stumble urge swap",
			seed:     "9248d83e06f4cd98debf5b6f010542760df925ce46cf38a1bdb4e4de7d21f5c39366941c69e1bdbf2966e0f6e6dbece898a0e2f0a4c2b3e640953dfe8b7bbdc5",
		},
		{
			entropy:  "3e141609b97933b66a060dcddc71fad1d91677db872031e85f4c015c5e7e8982",
			mnemonic: "dignity pass list indicate nasty swamp pool script soccer toe leaf photo multiply desk host tomato cradle drill spread actor shine dismiss champion exotic",
			seed:     "ff7f3184df8696d8bef94b6c03114dbee0ef89ff938712301d27ed8336ca89ef9635da20af07d4175f2bf5f3de130f39c9d9e8dd0472489c19b1a020a940da67",
		},
		{
			entropy:  "0460ef47585604c5660618db2e6a7e7f",
			mnemonic: "afford alter spike radar gate glance object seek swamp infant panel yellow",
			seed:     "65f93a9f36b6c85cbe634ffc1f99f2b82cbb10b31edc7f087b4f6cb9e976e9faf76ff41f8f27c99afdf38f7a303ba1136ee48a4c1e7fcd3dba7aa876113a36e4",
		},
		{
			entropy:  "72f60ebac5dd8add8d2a25a797102c3ce21bc029c200076f",
			mnemonic: "indicate race push merry suffer human cruise dwarf pole review arch keep canvas theme poem divorce alter left",
			seed:     "3bbf9daa0dfad8229786ace5ddb4e00fa98a044ae4c4975ffd5e094dba9e0bb289349dbe2091761f30f382d4e35c4a670ee8ab50758d2c55881be69e327117ba",
		},
		{
			entropy:  "2c85efc7f24ee4573d2b81a6ec66cee209b2dcbd09d8eddc51e0215b0b68e416",
			mnemonic: "clutch control vehicle tonight unusual clog visa ice plunge glimpse recipe series open hour vintage deposit universe tip job dress radar refuse motion taste",
			seed:     "fe908f96f46668b2d5b37d82f558c77ed0d69dd0e7e043a5b0511c48c2f1064694a956f86360c93dd04052a8899497ce9e985ebe0c8c52b955e6ae86d4ff4449",
		},
		{
			entropy:  "eaebabb2383351fd31d703840b32e9e2",
			mnemonic: "turtle front uncle idea crush write shrug there lottery flower risk shell",
			seed:     "bdfb76a0759f301b0b899a1e3985227e53b3f51e67e3f2a65363caedf3e32fde42a66c404f18d7b05818c95ef3ca1e5146646856c461c073169467511680876c",
		},
		{
			entropy:  "7ac45cfe7722ee6c7ba84fbc2d5bd61b45cb2fe5eb65aa78",
			mnemonic: "kiss carry display unusual confirm curtain upgrade antique rotate hello void custom frequent obey nut hole price segment",
			seed:     "ed56ff6c833c07982eb7119a8f48fd363c4a9b1601cd2de736b01045c5eb8ab4f57b079403485d1c4924f0790dc10a971763337cb9f9c62226f64fff26397c79",
		},
		{
			entropy:  "4fa1a8bc3e6d80ee1316050e862c1812031493212b7ec3f3bb1b08f168cabeef",
			mnemonic: "exile ask congress lamp submit jacket era scheme attend cousin alcohol catch course end lucky hurt sentence oven short ball bird grab wing top",
			seed:     "095ee6f817b4c2cb30a5a797360a81a40ab0f9a4e25ecd672a3f58a0b5ba0687c096a6b14d2c0deb3bdefce4f61d01ae07417d502429352e27695163f7447a8c",
		},
		{
			entropy:  "18ab19a9f54a9274f03e5209a2ac8a91",
			mnemonic: "board flee heavy tunnel powder denial science ski answer betray cargo cat",
			seed:     "6eff1bb21562918509c73cb990260db07c0ce34ff0e3cc4a8cb3276129fbcb300bddfe005831350efd633909f476c45c88253276d9fd0df6ef48609e8bb7dca8",
		},
		{
			entropy:  "18a2e1d81b8ecfb2a333adcb0c17a5b9eb76cc5d05db91a4",
			mnemonic: "board blade invite damage undo sun mimic interest slam gaze truly inherit resist great inject rocket museum chief",
			seed:     "f84521c777a13b61564234bf8f8b62b3afce27fc4062b51bb5e62bdfecb23864ee6ecf07c1d5a97c0834307c5c852d8ceb88e7c97923c0a3b496bedd4e5f88a9",
		},
		{
			entropy:  "15da872c95a13dd738fbf50e427583ad61f18fd99f628c417a61cf8343c90419",
			mnemonic: "beyond stage sleep clip because twist token leaf atom beauty genius food business side grid unable middle armed observe pair crouch tonight away coconut",
			seed:     "b15509eaa2d09d3efd3e006ef42151b30367dc6e3aa5e44caba3fe4d3e352e65101fbdb86a96776b91946ff06f8eac594dc6ee1d3e82a42dfe1b40fef6bcc3fd",
		},
	}
}

func badMnemonicSentences() []vector {
	return []vector{
		{mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"},
		{mnemonic: "legal winner thank year wave sausage worth useful legal winner thank yellow yellow"},
		{mnemonic: "letter advice cage absurd amount doctor acoustic avoid letter advice caged above"},
		{mnemonic: "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo, wrong"},
		{mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"},
		{mnemonic: "legal winner thank year wave sausage worth useful legal winner thank year wave sausage worth useful legal will will will"},
		{mnemonic: "letter advice cage absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter always."},
		{mnemonic: "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo why"},
		{mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art art"},
		{mnemonic: "legal winner thank year wave sausage worth useful legal winner thanks year wave worth useful legal winner thank year wave sausage worth title"},
		{mnemonic: "letter advice cage absurd amount doctor acoustic avoid letters advice cage absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic bless"},
		{mnemonic: "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo voted"},
		{mnemonic: "jello better achieve collect unaware mountain thought cargo oxygen act hood bridge"},
		{mnemonic: "renew, stay, biology, evidence, goat, welcome, casual, join, adapt, armor, shuffle, fault, little, machine, walk, stumble, urge, swap"},
		{mnemonic: "dignity pass list indicate nasty"},

		// From issue 32
		{mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon letter"},
	}
}

func assertNil(t *testing.T, object interface{}) {
	if object != nil {
		t.Errorf("Expected nil, got %v", object)
	}
}

func assertNotNil(t *testing.T, object interface{}) {
	if object == nil {
		t.Error("Expected not nil")
	}
}

func assertTrue(t *testing.T, a bool) {
	if !a {
		t.Error("Expected true, got false")
	}
}

func assertFalse(t *testing.T, a bool) {
	if a {
		t.Error("Expected false, got true")
	}
}

func assertEqual(t *testing.T, a, b interface{}) {
	if a != b {
		t.Errorf("Objects not equal, expected `%s` and got `%s`", a, b)
	}
}

func assertEqualString(t *testing.T, a, b string) {
	if a != b {
		t.Errorf("Strings not equal, expected `%s` and got `%s`", a, b)
	}
}

func assertEqualStringSlices(t *testing.T, a, b []string) {
	if len(a) != len(b) {
		t.Errorf("String slices not equal, expected %v and got %v", a, b)
		return
	}
	for i := range a {
		if a[i] != b[i] {
			t.Errorf("String slices not equal, expected %v and got %v", a, b)
			return
		}
	}
}

func assertEqualByteSlices(t *testing.T, a, b []byte) {
	if len(a) != len(b) {
		t.Errorf("Byte slices not equal, expected %v and got %v", a, b)
		return
	}
	for i := range a {
		if a[i] != b[i] {
			t.Errorf("Byte slices not equal, expected %v and got %v", a, b)
			return
		}
	}
}
