package token2022

import "bytes"

// TODO: Everything here should either already exist or needs to be moved to a more appropriate package.


func fixCodecSize(_ func([]byte) []byte, size int) func([]byte) []byte {
    return func(b []byte) []byte {
        if len(b) > size {
            return b[:size]
        }
        padded := make([]byte, size)
        copy(padded, b)
        return padded
    }
}

func getBytesCodec() func([]byte) []byte {
    return func(b []byte) []byte {
        return b
    }
}

func getStringCodec() func(string) []byte {
    return func(s string) []byte {
        return []byte(s)
    }
}


func getMapCodec(codec func(string) []byte) func(map[string]string) []byte {
    return func(m map[string]string) []byte {
        var buf bytes.Buffer
        for k, v := range m {
            buf.Write(codec(k))
            buf.Write(codec(v))
        }
        return buf.Bytes()
    }
}