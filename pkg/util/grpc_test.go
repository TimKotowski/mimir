// SPDX-License-Identifier: AGPL-3.0-only

package util

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCloseAndExhaust(t *testing.T) {
	t.Run("CloseSend returns an error", func(t *testing.T) {
		expectedErr := errors.New("something went wrong")
		stream := &mockStream{closeSendError: expectedErr}

		actualErr := CloseAndExhaust[string](stream)
		require.Equal(t, expectedErr, actualErr)
	})

	t.Run("Recv returns error immediately", func(t *testing.T) {
		stream := &mockStream{recvErrors: []error{io.EOF}}
		err := CloseAndExhaust[string](stream)
		require.NoError(t, err, "CloseAndExhaust should ignore errors from Recv()")
	})

	t.Run("Recv returns error after multiple calls", func(t *testing.T) {
		stream := &mockStream{recvErrors: []error{nil, nil, io.EOF}}
		err := CloseAndExhaust[string](stream)
		require.NoError(t, err, "CloseAndExhaust should ignore errors from Recv()")
	})

	t.Run("Recv blocks forever", func(t *testing.T) {
		stream := &mockStream{}
		returned := make(chan error)

		go func() {
			returned <- CloseAndExhaust[string](stream)
		}()

		select {
		case err := <-returned:
			require.Equal(t, ErrCloseAndExhaustTimedOut, err)
		case <-time.After(500 * time.Millisecond):
			require.FailNow(t, "expected CloseAndExhaust to time out waiting for Recv() to return, but it did not")
		}
	})
}

type mockStream struct {
	closeSendError error
	recvErrors     []error
}

func (m *mockStream) CloseSend() error {
	return m.closeSendError
}

func (m *mockStream) Recv() (string, error) {
	if len(m.recvErrors) == 0 {
		// Block forever.
		<-make(chan struct{})
	}

	err := m.recvErrors[0]
	m.recvErrors = m.recvErrors[1:]

	return "", err
}
