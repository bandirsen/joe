package joetest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-joe/joe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBot(t *testing.T) {
	b := NewBot(t)

	var seenEvents []TestEvent
	b.Brain.RegisterHandler(func(evt TestEvent) {
		seenEvents = append(seenEvents, evt)
	})

	b.Start()
	assert.Equal(t, "test > ", b.ReadOutput())

	b.EmitSync(TestEvent{N: 123})
	b.Stop()

	assert.Equal(t, []TestEvent{{N: 123}}, seenEvents)
}

func TestBotEmitSyncTimeout(t *testing.T) {
	mock := new(mockT)
	b := NewBot(mock)
	b.Timeout = time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b.Brain.RegisterHandler(func(evt TestEvent) {
		<-ctx.Done()
	})

	b.Start()
	b.EmitSync(TestEvent{})
	b.Stop()

	require.Len(t, mock.Errors, 2)
	assert.True(t, mock.failed)
	assert.True(t, mock.fatal)
	assert.Equal(t, "EmitSync timed out", mock.Errors[0])
	assert.Equal(t, "Stop timed out", mock.Errors[1])
}

func TestBot_RegistrationErrors(t *testing.T) {
	b := NewBot(t)

	b.Brain.RegisterHandler(func(evt *TestEvent) {
		// handlers cannot use pointers for the event type so registering this
		// handler function should create a registration error.
	})

	b.Start()

	select {
	case err := <-b.runErr:
		require.Error(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), "invalid event handlers: "))
		t.Log(err.Error())
	case <-time.After(b.Timeout):
		b.T.Errorf("Timeout")
	}
}

func TestBot_RegistrationErrors2(t *testing.T) {
	b := NewBot(t)

	b.RespondRegex("invalid regex: (", func(joe.Message) error {
		return joe.ErrNotImplemented
	})

	b.Start()

	select {
	case err := <-b.runErr:
		require.Error(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), "invalid event handlers: "))
		t.Log(err.Error())
	case <-time.After(b.Timeout):
		b.T.Errorf("Timeout")
	}
}
