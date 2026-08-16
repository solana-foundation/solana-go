//go:build amd64 && !amd64.v3 && !purego

package base58

import "golang.org/x/sys/cpu"

var useAVX2 = cpu.X86.HasAVX2
