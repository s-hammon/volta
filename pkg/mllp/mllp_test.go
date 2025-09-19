package mllp

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	w := bytes.NewBuffer([]byte{})
	err := Write(w, []byte{})
	require.NoError(t, err)
	require.Equal(t, []byte{sb, eb, cr}, w.Bytes())

	w.Reset()
	err = Write(w, []byte("test"))
	require.NoError(t, err)
	require.Equal(t, []byte{sb, 't', 'e', 's', 't', eb, cr}, w.Bytes())
}

func TestRead(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		want  []byte
		error bool
	}{
		{
			name: "valid empty",
			data: encodeStr(""),
			want: []byte{},
		},
		{
			name: "valid message",
			data: encodeStr("test"),
			want: []byte("test"),
		},
		{
			name:  "invalid empty",
			data:  []byte{},
			error: true,
		},
		{
			name:  "missing sb",
			data:  []byte{'t', 'e', 's', 't', eb, cr},
			error: true,
		},
		{
			name:  "missing eb",
			data:  []byte{sb, 't', 'e', 's', 't', cr},
			error: true,
		},
		{
			name:  "missing cr",
			data:  []byte{sb, 't', 'e', 's', 't', eb},
			error: true,
		},
		{
			name:  "end byte not cr",
			data:  []byte{sb, 't', 'e', 's', 't', eb, '\x0A'},
			error: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewBuffer(tt.data)
			mr := NewMsgReader(r)
			got, err := mr.Read()
			if tt.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}

	data := append(encodeStr("one"), encodeStr("two")...)

	r := bytes.NewBuffer(data)
	mr := NewMsgReader(r)

	got, err := mr.Read()
	require.NoError(t, err)
	require.Equal(t, []byte("one"), got)

	got, err = mr.Read()
	require.NoError(t, err)
	require.Equal(t, []byte("two"), got)

	_, err = mr.Read()
	require.Error(t, err)
}

func encodeStr(s string) []byte {
	b := []byte{sb}
	b = append(b, []byte(s)...)
	b = append(b, []byte{eb, cr}...)
	return b
}
