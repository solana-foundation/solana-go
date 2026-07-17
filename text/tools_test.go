package text

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestColorHelpersRespectDisableColors pins that every exported color helper
// honors the DisableColors escape hatch documented on Transaction.String.
func TestColorHelpersRespectDisableColors(t *testing.T) {
	old := DisableColors
	DisableColors = true
	defer func() { DisableColors = old }()

	helpers := map[string]func(string) string{
		"Black":         Black,
		"BlackBG":       BlackBG,
		"White":         White,
		"WhiteBG":       WhiteBG,
		"Lime":          Lime,
		"LimeBG":        LimeBG,
		"Yellow":        Yellow,
		"YellowBG":      YellowBG,
		"Orange":        Orange,
		"OrangeBG":      OrangeBG,
		"Red":           Red,
		"RedBG":         RedBG,
		"Shakespeare":   Shakespeare,
		"ShakespeareBG": ShakespeareBG,
		"Purple":        Purple,
		"PurpleBG":      PurpleBG,
		"Indigo":        Indigo,
		"IndigoBG":      IndigoBG,
		"Bold":          Bold,
	}
	for name, fn := range helpers {
		require.Equal(t, "sample", fn("sample"), "%s must return the input unchanged when DisableColors is set", name)
	}
}
