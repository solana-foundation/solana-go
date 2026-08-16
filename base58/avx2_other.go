//go:build !amd64 || purego

package base58

const useAVX2 = false

func encode32AVX2(*[32]byte, *[48]byte) int  { panic("unreachable") }
func encode64AVX2(*[64]byte, *[96]byte) int  { panic("unreachable") }
func decode32AVX2(*byte, int, *[32]byte) int { panic("unreachable") }
func decode64AVX2(*byte, int, *[64]byte) int { panic("unreachable") }
